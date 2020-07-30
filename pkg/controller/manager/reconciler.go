/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package manager

import (
	"context"
	"strings"
	"time"

	"github.com/hasheddan/crank/apis/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/hasheddan/crank/pkg/dag"
	"github.com/hasheddan/crank/pkg/unpack"
)

const (
	finalizer        = "finalizer.crank.crossplane.io"
	reconcileTimeout = 1 * time.Minute

	aShortWait = 30 * time.Second
)

// Reconciler reconciles packages.
type Reconciler struct {
	client resource.ClientApplicator
	log    logging.Logger
	record event.Recorder
}

// Setup adds a controller that reconciles Packages.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	name := "packages/" + strings.ToLower(v1alpha1.PackageGroupKind)

	r := &Reconciler{
		client: resource.ClientApplicator{
			Client:     mgr.GetClient(),
			Applicator: resource.NewAPIPatchingApplicator(mgr.GetClient()),
		},
		log:    l.WithValues("controller", name),
		record: event.NewAPIRecorder(mgr.GetEventRecorderFor("packages")),
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Package{}).
		Owns(&v1alpha1.PackageRevision{}).
		Complete(r)
}

// Reconcile package.
func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) { // nolint:gocyclo
	log := r.log.WithValues("request", req)
	log.Debug("Reconciling")

	ctx, cancel := context.WithTimeout(context.Background(), reconcileTimeout)
	defer cancel()

	p := &v1alpha1.Package{}
	if err := r.client.Get(ctx, req.NamespacedName, p); err != nil {
		// There's no need to requeue if we no longer exist. Otherwise we'll be
		// requeued implicitly because we return an error.
		log.Debug("Cannot get package", "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), "cannot get Package")
	}

	m := &v1alpha1.PackageLock{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: "packages"}, m); err != nil {
		log.Debug("Cannot get package package lock", "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), "cannot get PackageLock")
	}

	// Construct DAG for PackageLock.
	d, err := dag.New(m.Spec.Packages)
	if err != nil {
		p.Status.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
		r.record.Event(p, event.Warning(event.Reason("failed building state"), err))
		return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
	}

	// Unpack this image.
	// TODO(hasheddan): no need to unpack each time.
	digest, deps, err := unpack.Unpack(p.Spec.Package)
	if err != nil {
		p.Status.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
		r.record.Event(p, event.Warning(event.Reason("failed to unpack package"), err))
		return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
	}
	// If node does not exist in DAG then we want to add it and make sure it is
	// valid.
	if !d.NodeExists(strings.Split(p.Spec.Package, ":")[0]) {
		// If node does not exist already then adding it should always be successful.
		if err := d.AddNode(strings.Split(p.Spec.Package, ":")[0]); err != nil {
			p.Status.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
			r.record.Event(p, event.Warning(event.Reason("failed adding node"), err))
			return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
		}
		// Check to see if all dependencies are satisfied.
		if err := d.AddEdges(map[string][]string{strings.Split(p.Spec.Package, ":")[0]: deps}); err != nil {
			// If dependencies are not satisfied, we need to install them.
			p.Status.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
			r.record.Event(p, event.Warning(event.Reason("failed adding package dependencies"), err))
			return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
		}
		// Check that adding dependencies does not result in cyclical dependency.
		if _, err := d.Sort(); err != nil {
			p.Status.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
			r.record.Event(p, event.Warning(event.Reason("cyclical dependency"), err))
			return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
		}
		// Append the package to the PackageLock and attempt to update.
		if m.Spec.Packages == nil {
			m.Spec.Packages = map[string]v1alpha1.PackageDependencies{}
		}
		m.Spec.Packages[strings.Split(p.Spec.Package, ":")[0]] = v1alpha1.PackageDependencies{
			Name:         p.Name,
			Image:        p.Spec.Package,
			Dependencies: deps,
		}
		// If another package has updated the PackageLock since we read it then this will
		// fail. Try again after short wait.
		if err := r.client.Update(ctx, m); err != nil {
			log.Debug("Cannot update package package lock", "error", err)
			return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(err, "cannot update PackageLock")
		}
	}
	// If updating the PackageLock was successful then we can create PackageRevision safely.
	pr := &v1alpha1.PackageRevision{
		ObjectMeta: metav1.ObjectMeta{
			Name:   digest,
			Labels: map[string]string{"crank.crossplane.io/package": p.Name},
		},
		Spec: v1alpha1.PackageRevisionSpec{
			DesiredState: v1alpha1.PackageRevisionInactive,
			Revision:     1,
			Image:        p.Spec.Package,
		},
	}
	meta.AddOwnerReference(pr, meta.AsController(meta.ReferenceTo(p, v1alpha1.PackageGroupVersionKind)))
	if err := r.client.Applicator.Apply(ctx, pr, resource.MustBeControllableBy(p.UID), desiredStateApplicator()); err != nil {
		log.Debug("Cannot create PackageRevision", "error", err)
		p.Status.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
		return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot create PackageRevision")
	}
	p.Status.CurrentRevision = digest
	p.Status.SetConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())
	return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
}

// TODO(hasheddan): pretty sure there is a cleaner way to do this
func desiredStateApplicator() resource.ApplyOption {
	return func(_ context.Context, current, desired runtime.Object) error {
		cr, ok := current.(*v1alpha1.PackageRevision)
		if !ok {
			return errors.New("not package revision")
		}
		dr, ok := desired.(*v1alpha1.PackageRevision)
		if !ok {
			return errors.New("not package revision")
		}
		dr.Spec.DesiredState = cr.Spec.DesiredState
		return nil
	}
}
