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

package packagerevision

import (
	"context"
	"strings"
	"time"

	"github.com/hasheddan/crank/apis/v1alpha1"
	"github.com/hasheddan/crank/pkg/unpack"
	"github.com/pkg/errors"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

const (
	finalizer        = "finalizer.crank.crossplane.io"
	reconcileTimeout = 1 * time.Minute

	aShortWait = 30 * time.Second
)

// ReconcilerOption is used to configure the Reconciler.
type ReconcilerOption func(*Reconciler)

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

// Reconciler reconciles package revisions.
type Reconciler struct {
	client resource.ClientApplicator
	log    logging.Logger
	record event.Recorder

	newPackageRevision func() v1alpha1.PackageRevision
}

// SetupProviderRevision adds a controller that reconciles ProviderRevisions.
func SetupProviderRevision(mgr ctrl.Manager, l logging.Logger) error {
	name := "packages/" + strings.ToLower(v1alpha1.ProviderRevisionGroupKind)

	nr := func() v1alpha1.PackageRevision { return &v1alpha1.ProviderRevision{} }

	r := NewReconciler(mgr,
		WithNewPackageRevisionFn(nr),
		WithLogger(l.WithValues("controller", name)),
		WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor("providerrevision"))),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ProviderRevision{}).
		Complete(r)
}

// SetupConfigurationRevision adds a controller that reconciles ConfigurationRevisions.
func SetupConfigurationRevision(mgr ctrl.Manager, l logging.Logger) error {
	name := "packages/" + strings.ToLower(v1alpha1.ConfigurationRevisionGroupKind)

	nr := func() v1alpha1.PackageRevision { return &v1alpha1.ConfigurationRevision{} }

	r := NewReconciler(mgr,
		WithNewPackageRevisionFn(nr),
		WithLogger(l.WithValues("controller", name)),
		WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor("configurationrevision"))),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ConfigurationRevision{}).
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

// Reconcile package revision.
func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) { // nolint:gocyclo
	log := r.log.WithValues("request", req)
	log.Debug("Reconciling")

	ctx, cancel := context.WithTimeout(context.Background(), reconcileTimeout)
	defer cancel()

	pr := r.newPackageRevision()
	if err := r.client.Get(ctx, req.NamespacedName, pr); err != nil {
		// There's no need to requeue if we no longer exist. Otherwise we'll be
		// requeued implicitly because we return an error.
		log.Debug("Cannot get package revision", "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), "cannot get PackageRevision")
	}

	crds, _, _, err := unpack.Resources(pr.GetSource())
	if err != nil {
		log.Debug("Cannot unpack resources", "error", err)
		return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(err, "cannot unpack PackageRevision resources")
	}

	for _, c := range crds {
		if pr.GetDesiredState() == v1alpha1.PackageRevisionInactive {
			meta.AddOwnerReference(&c, meta.AsOwner(meta.ReferenceTo(pr, pr.GetObjectKind().GroupVersionKind())))
			if err := r.client.Applicator.Apply(ctx, &c, ownerReferenceApplicator(meta.AsOwner(meta.ReferenceTo(pr, pr.GetObjectKind().GroupVersionKind())))); err != nil {
				log.Debug("Cannot apply crds", "error", "crd", c.Name, err)
				pr.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
				return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, pr), "cannot update package status")
			}
		} else {
			meta.AddOwnerReference(&c, meta.AsController(meta.ReferenceTo(pr, pr.GetObjectKind().GroupVersionKind())))
			if err := r.client.Applicator.Apply(ctx, &c, resource.MustBeControllableBy(pr.GetUID()), ownerReferenceApplicator(meta.AsController(meta.ReferenceTo(pr, pr.GetObjectKind().GroupVersionKind())))); err != nil {
				log.Debug("Cannot apply crds", "error", "crd", c.Name, err)
				pr.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
				return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, pr), "cannot update package status")
			}
		}
	}

	pr.SetConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())
	return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, pr), "cannot update package status")
}

// TODO(hasheddan): pretty sure there is a cleaner way to do this
func ownerReferenceApplicator(r v1.OwnerReference) resource.ApplyOption {
	return func(_ context.Context, current, desired runtime.Object) error {
		cr, ok := current.(*v1beta1.CustomResourceDefinition)
		if !ok {
			return errors.New("not crd")
		}
		dr, ok := desired.(*v1beta1.CustomResourceDefinition)
		if !ok {
			return errors.New("not crd")
		}
		dr.OwnerReferences = cr.OwnerReferences
		meta.AddOwnerReference(cr, r)
		return nil
	}
}
