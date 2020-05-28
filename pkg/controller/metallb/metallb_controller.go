package metallb

import (
	"context"
	"path/filepath"

	loadbalancerv1alpha1 "github.com/openshift/metallb-operator/pkg/apis/loadbalancer/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/openshift/cluster-network-operator/pkg/apply"
	"github.com/openshift/cluster-network-operator/pkg/render"
	"github.com/pkg/errors"
)

var log = logf.Log.WithName("controller_metallb")

// Add creates a new MetalLB Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMetalLB{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("metallb-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MetalLB
	err = c.Watch(&source.Kind{Type: &loadbalancerv1alpha1.MetalLB{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner MetalLB
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &loadbalancerv1alpha1.MetalLB{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMetalLB implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMetalLB{}

// ReconcileMetalLB reconciles a MetalLB object
type ReconcileMetalLB struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MetalLB object and makes changes based on the state read
// and what is in the MetalLB.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMetalLB) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MetalLB")

	// Fetch the MetalLB instance
	instance := &loadbalancerv1alpha1.MetalLB{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Create namespace
	err = r.applyNamespace(instance)
	if err != nil {
		errors.Wrap(err, "Failed to apply namespace")
		return reconcile.Result{}, err
	}

	// Iterate decomposed metallb manifest

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Already deployed")
	return reconcile.Result{}, nil
}

// applyNamespace creates the metallb-system namespace
func (r *ReconcileMetalLB) applyNamespace(instance *loadbalancerv1alpha1.MetalLB) error {
	data := render.MakeRenderData()
	return r.renderAndApply(instance, data, "namespace", true)
}

func (r *ReconcileMetalLB) renderAndApply(instance *loadbalancerv1alpha1.MetalLB, data render.RenderData, sourceDirectory string, setControllerReference bool) error {
	var err error
	objs := []*uns.Unstructured{}

	objs, err = render.RenderDir(filepath.Join("/manifests", sourceDirectory), &data)
	if err != nil {
		return errors.Wrapf(err, "Failed to render metallb %s", sourceDirectory)
	}

	for _, obj := range objs {
		// RenderDir seems to add an extra null entry to the list. It appears to be because of the
		// nested templates. This just makes sure we don't try to apply an empty obj.
		if obj.GetName() == "" {
			continue
		}
		if setControllerReference {
			// Set the controller refernce. When the CR is removed, it will remove the CRDs as well
			err = controllerutil.SetControllerReference(instance, obj, r.scheme)
			if err != nil {
				return errors.Wrap(err, "Failed to set owner reference")
			}
		}

		// Now apply the object
		err = apply.ApplyObject(context.TODO(), r.client, obj)
		if err != nil {
			return errors.Wrapf(err, "Failed to apply object %v", obj)
		}
	}
	return nil
}
