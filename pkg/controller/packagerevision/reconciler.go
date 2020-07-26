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

// Reconciler reconciles package revisions.
type Reconciler struct {
	client resource.ClientApplicator
	log    logging.Logger
	record event.Recorder
}

// Setup adds a controller that reconciles PackageRevisions.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	name := "packages/" + strings.ToLower(v1alpha1.PackageRevisionGroupKind)

	r := &Reconciler{
		client: resource.ClientApplicator{
			Client:     mgr.GetClient(),
			Applicator: resource.NewAPIPatchingApplicator(mgr.GetClient()),
		},
		log:    l.WithValues("controller", name),
		record: event.NewAPIRecorder(mgr.GetEventRecorderFor("packagerevision")),
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.PackageRevision{}).
		Complete(r)
}

// Reconcile package revision.
func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) { // nolint:gocyclo
	log := r.log.WithValues("request", req)
	log.Debug("Reconciling")

	ctx, cancel := context.WithTimeout(context.Background(), reconcileTimeout)
	defer cancel()

	p := &v1alpha1.PackageRevision{}
	if err := r.client.Get(ctx, req.NamespacedName, p); err != nil {
		// There's no need to requeue if we no longer exist. Otherwise we'll be
		// requeued implicitly because we return an error.
		log.Debug("Cannot get package revision", "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), "cannot get PackageRevision")
	}

	crds, _, _, err := unpack.Resources(p.Spec.Image)
	if err != nil {
		log.Debug("Cannot unpack resources", "error", err)
		return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(err, "cannot unpack PackageRevision resources")
	}

	for _, c := range crds {
		if p.Spec.DesiredState == v1alpha1.PackageRevisionInactive {
			meta.AddOwnerReference(&c, meta.AsOwner(meta.ReferenceTo(p, v1alpha1.PackageRevisionGroupVersionKind)))
			if err := r.client.Applicator.Apply(ctx, &c, ownerReferenceApplicator(meta.AsOwner(meta.ReferenceTo(p, v1alpha1.PackageRevisionGroupVersionKind)))); err != nil {
				log.Debug("Cannot apply crds", "error", "crd", c.Name, err)
				p.Status.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
				return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
			}
		} else {
			meta.AddOwnerReference(&c, meta.AsController(meta.ReferenceTo(p, v1alpha1.PackageRevisionGroupVersionKind)))
			if err := r.client.Applicator.Apply(ctx, &c, resource.MustBeControllableBy(p.UID), ownerReferenceApplicator(meta.AsController(meta.ReferenceTo(p, v1alpha1.PackageRevisionGroupVersionKind)))); err != nil {
				log.Debug("Cannot apply crds", "error", "crd", c.Name, err)
				p.Status.SetConditions(runtimev1alpha1.Unavailable(), runtimev1alpha1.ReconcileSuccess())
				return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
			}
		}
	}

	p.Status.SetConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())
	return reconcile.Result{RequeueAfter: aShortWait}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
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
