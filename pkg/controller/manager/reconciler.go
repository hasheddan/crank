package manager

import (
	"context"
	"strings"
	"time"

	"github.com/hasheddan/crank/apis/v1alpha1"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
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
		record: event.NewNopRecorder(),
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
	p.Status.SetConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())
	return reconcile.Result{Requeue: false}, errors.Wrap(r.client.Status().Update(ctx, p), "cannot update package status")
}
