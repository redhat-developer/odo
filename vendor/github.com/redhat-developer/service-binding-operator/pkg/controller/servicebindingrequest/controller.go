package servicebindingrequest

import (
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Add creates a new ServiceBindingRequest Controller and adds it to the Manager. The Manager will
// set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	client, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return err
	}
	r, err := newReconciler(mgr, client)
	if err != nil {
		return err
	}
	return add(mgr, r, client)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, client dynamic.Interface) (reconcile.Reconciler, error) {
	return &Reconciler{
		dynClient:  client,
		scheme:     mgr.GetScheme(),
		RestMapper: mgr.GetRESTMapper(),
	}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler.
func add(mgr manager.Manager, r reconcile.Reconciler, client dynamic.Interface) error {
	opts := controller.Options{Reconciler: r}
	c, err := NewSBRController(mgr, opts, client)
	if err != nil {
		return err
	}
	return c.Watch()
}

// blank assignment to verify that ReconcileServiceBindingRequest implements reconcile.Reconciler
var _ reconcile.Reconciler = &Reconciler{}
