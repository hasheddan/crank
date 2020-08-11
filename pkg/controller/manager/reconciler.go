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

// ReconcilerOption is used to configure the Reconciler.
type ReconcilerOption func(*Reconciler)

// WithNewPackageFn determines the type of package being reconciled.
func WithNewPackageFn(f func() v1alpha1.Package) ReconcilerOption {
	return func(r *Reconciler) {
		r.newPackage = f
	}
}

// WithNewPackageRevisionFn determines the type of package being reconciled.
func WithNewPackageRevisionFn(f func() v1alpha1.PackageRevision) ReconcilerOption {
	return func(r *Reconciler) {
		r.newPackageRevision = f
	}
}

// WithLogger specifies how the Reconciler should log messages.
func WithLogger(log logging.Logger) ReconcilerOption {
	return func(r *Reconciler) {
		r.log = log
	}
}

// WithRecorder specifies how the Reconciler should record Kubernetes events.
func WithRecorder(er event.Recorder) ReconcilerOption {
	return func(r *Reconciler) {
		r.record = er
	}
}

// Reconciler reconciles packages.
type Reconciler struct {
	client resource.ClientApplicator
	log    logging.Logger
	record event.Recorder

	newPackage         func() v1alpha1.Package
	newPackageRevision func() v1alpha1.PackageRevision
}

// SetupProvider adds a controller that reconciles Providers.
func SetupProvider(mgr ctrl.Manager, l logging.Logger) error {
	name := "packages/" + strings.ToLower(v1alpha1.ProviderGroupKind)
	np := func() v1alpha1.Package { return &v1alpha1.Provider{} }
	nr := func() v1alpha1.PackageRevision { return &v1alpha1.ProviderRevision{} }

	r := NewReconciler(mgr,
		WithNewPackageFn(np),
		WithNewPackageRevisionFn(nr),
		WithLogger(l.WithValues("controller", name)),
		WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor("provider"))),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Provider{}).
		Owns(&v1alpha1.ProviderRevision{}).
		Complete(r)
}

// SetupConfiguration adds a controller that reconciles Configurations.
func SetupConfiguration(mgr ctrl.Manager, l logging.Logger) error {
	name := "packages/" + strings.ToLower(v1alpha1.ConfigurationGroupKind)
	np := func() v1alpha1.Package { return &v1alpha1.Configuration{} }
	nr := func() v1alpha1.PackageRevision { return &v1alpha1.ConfigurationRevision{} }

	r := NewReconciler(mgr,
		WithNewPackageFn(np),
		WithNewPackageRevisionFn(nr),
		WithLogger(l.WithValues("controller", name)),
		WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor("configuration"))),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Configuration{}).
		Owns(&v1alpha1.ConfigurationRevision{}).
		Complete(r)
}

// NewReconciler creates a new package reconciler.
func NewReconciler(mgr ctrl.Manager, opts ...ReconcilerOption) *Reconciler {
	r := &Reconciler{
		client: resource.ClientApplicator{
			Client:     mgr.GetClient(),
			Applicator: resource.NewAPIPatchingApplicator(mgr.GetClient()),
		},
		log:    logging.NewNopLogger(),
		record: event.NewNopRecorder(),
	}

	for _, f := range opts {
		f(r)
	}

	return r
}

// Reconcile package.
func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) { // nolint:gocyclo
	log := r.log.WithValues("request", req)
	log.Debug("Reconciling")

	ctx, cancel := context.WithTimeout(context.Background(), reconcileTimeout)
	defer cancel()

	p := r.newPackage()
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
		p.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
		r.record.Event(p, event.Warning(event.Reason("failed building state"), err))
		return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
	}

	// Unpack this image.
	// TODO(hasheddan): no need to unpack each time.
	digest, deps, err := unpack.Unpack(p.GetSource())
	if err != nil {
		p.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
		r.record.Event(p, event.Warning(event.Reason("failed to unpack package"), err))
		return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
	}
	// If node does not exist in DAG then we want to add it and make sure it is
	// valid.
	if !d.NodeExists(strings.Split(p.GetSource(), ":")[0]) {
		// If node does not exist already then adding it should always be successful.
		if err := d.AddNode(strings.Split(p.GetSource(), ":")[0]); err != nil {
			p.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
			r.record.Event(p, event.Warning(event.Reason("failed adding node"), err))
			return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
		}
		// Check to see if all dependencies are satisfied.
		if err := d.AddEdges(map[string][]string{strings.Split(p.GetSource(), ":")[0]: deps}); err != nil {
			// If dependencies are not satisfied, we need to install them.
			p.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
			r.record.Event(p, event.Warning(event.Reason("failed adding package dependencies"), err))
			return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
		}
		// Check that adding dependencies does not result in cyclical dependency.
		if _, err := d.Sort(); err != nil {
			p.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
			r.record.Event(p, event.Warning(event.Reason("cyclical dependency"), err))
			return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
		}
		// Append the package to the PackageLock and attempt to update.
		if m.Spec.Packages == nil {
			m.Spec.Packages = map[string]v1alpha1.PackageDependencies{}
		}
		m.Spec.Packages[strings.Split(p.GetSource(), ":")[0]] = v1alpha1.PackageDependencies{
			Name:         p.GetName(),
			Image:        p.GetSource(),
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
	pr := r.newPackageRevision()
	pr.SetName(digest)
	pr.SetLabels(map[string]string{"crank.crossplane.io/package": p.GetName()})
	pr.SetDesiredState(v1alpha1.PackageRevisionInactive)
	pr.SetSource(p.GetSource())
	pr.SetRevision(1)

	meta.AddOwnerReference(pr, meta.AsController(meta.ReferenceTo(p, p.GetObjectKind().GroupVersionKind())))
	if err := r.client.Applicator.Apply(ctx, pr, resource.MustBeControllableBy(p.GetUID()), desiredStateApplicator()); err != nil {
		log.Debug("Cannot create PackageRevision", "error", err)
		p.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
		return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot create PackageRevision")
	}
	p.SetCurrentRevision(digest)
	p.SetConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())
	return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
}

// TODO(hasheddan): pretty sure there is a cleaner way to do this
func desiredStateApplicator() resource.ApplyOption {
	return func(_ context.Context, current, desired runtime.Object) error {
		cr, ok := current.(v1alpha1.PackageRevision)
		if !ok {
			return errors.New("not package revision")
		}
		dr, ok := desired.(v1alpha1.PackageRevision)
		if !ok {
			return errors.New("not package revision")
		}
		dr.SetDesiredState(cr.GetDesiredState())
		return nil
	}
}
