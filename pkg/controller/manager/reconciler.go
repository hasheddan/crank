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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
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
	client client.Client
	log    logging.Logger
	record event.Recorder
}

// Setup adds a controller that reconciles Packages.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	name := "packages/" + strings.ToLower(v1alpha1.PackageGroupKind)

	r := &Reconciler{
		client: mgr.GetClient(),
		log:    l.WithValues("controller", name),
		record: event.NewAPIRecorder(mgr.GetEventRecorderFor("packages")),
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Package{}).
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

	m := &v1alpha1.Module{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: "packages"}, m); err != nil {
		log.Debug("Cannot get package module", "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), "cannot get Module")
	}

	d := dag.New()
	for _, pkg := range m.Spec.Packages {
		if pkg.Image == p.Spec.Package {
			p.Status.SetConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())
			return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
		}
		if err := d.AddNode(pkg.Image); err != nil {
			p.Status.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
			r.record.Event(p, event.Warning(event.Reason("failed building state"), err))
			return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
		}
	}
	for _, pkg := range m.Spec.Packages {
		if err := d.AddEdges(map[string][]string{pkg.Image: pkg.Dependencies}); err != nil {
			p.Status.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
			r.record.Event(p, event.Warning(event.Reason("failed building state"), err))
			return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
		}
	}
	if err := d.AddNode(p.Spec.Package); err != nil {
		p.Status.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
		r.record.Event(p, event.Warning(event.Reason("failed adding node"), err))
		return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
	}
	deps, err := unpack.Unpack(p.Spec.Package)
	if err != nil {
		p.Status.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
		r.record.Event(p, event.Warning(event.Reason("failed to unpack package"), err))
		return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
	}
	if err := d.AddEdges(map[string][]string{p.Spec.Package: deps}); err != nil {
		p.Status.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
		r.record.Event(p, event.Warning(event.Reason("failed adding package dependencies"), err))
		return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
	}
	if _, err := d.Sort(); err != nil {
		p.Status.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
		r.record.Event(p, event.Warning(event.Reason("cyclical dependency"), err))
		return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
	}
	m.Spec.Packages = append(m.Spec.Packages, v1alpha1.PackageDependencies{
		Name:         p.Name,
		Image:        p.Spec.Package,
		Dependencies: deps,
	})
	if err := r.client.Update(ctx, m); err != nil {
		log.Debug("Cannot update package module", "error", err)
		return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(resource.IgnoreNotFound(err), "cannot update Module")
	}
	p.Status.SetConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())
	return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
}
