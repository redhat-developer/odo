package servicebindingrequest

import (
	"context"
	"errors"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

const (
	// BindingReady indicates that the binding succeeded
	BindingReady conditionsv1.ConditionType = "Ready"
	// EmptyServiceSelectorsReason is used when the ServiceBindingRequest has empty
	// backingServiceSelectors.
	EmptyServiceSelectorsReason = "EmptyServiceSelectors"
)

// Reconciler reconciles a ServiceBindingRequest object
type Reconciler struct {
	dynClient  dynamic.Interface // kubernetes dynamic api client
	scheme     *runtime.Scheme   // api scheme
	RestMapper meta.RESTMapper   // restMapper to convert GVK and GVR
}

// reconcilerLog local logger instance
var reconcilerLog = log.NewLog("reconciler")

//// validateServiceBindingRequest check for unsupported settings in SBR.
//func (r *Reconciler) validateServiceBindingRequest(sbr *v1alpha1.ServiceBindingRequest) error {
//	// check if application ResourceRef and MatchLabels, one of them is required.
//	if sbr.Spec.ApplicationSelector.ResourceRef == "" &&
//		sbr.Spec.ApplicationSelector.LabelSelector == nil {
//		return fmt.Errorf("both ResourceRef and LabelSelector are not set")
//	}
//	return nil
//}

// getServiceBindingRequest retrieve the SBR object based on namespaced-name.
func (r *Reconciler) getServiceBindingRequest(
	namespacedName types.NamespacedName,
) (*v1alpha1.ServiceBindingRequest, error) {
	gr := v1alpha1.SchemeGroupVersion.WithResource(ServiceBindingRequestResource)
	resourceClient := r.dynClient.Resource(gr).Namespace(namespacedName.Namespace)
	u, err := resourceClient.Get(namespacedName.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	sbr := &v1alpha1.ServiceBindingRequest{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, sbr)
	if err != nil {
		return nil, err
	}

	return sbr, nil
}

// extractServiceSelectors returns a list of all BackingServiceSelector items from a
// ServiceBindingRequest.
//
// NOTE(isuttonl): remove this method when spec.backingServiceSelector is deprecated
func extractServiceSelectors(
	sbr *v1alpha1.ServiceBindingRequest,
) []v1alpha1.BackingServiceSelector {
	selector := sbr.Spec.BackingServiceSelector
	inSelectors := sbr.Spec.BackingServiceSelectors
	var selectors []v1alpha1.BackingServiceSelector

	if selector != nil {
		selectors = append(selectors, *selector)
	}
	if inSelectors != nil {
		selectors = append(selectors, *inSelectors...)
	}
	return selectors
}

// Reconcile a ServiceBindingRequest by the following steps:
// 1. Inspecting SBR in order to identify backend service. The service is composed by a CRD name and
//    kind, and by inspecting "connects-to" label identify the name of service instance;
// 2. Using OperatorLifecycleManager standards, identifying which items are intersting for binding
//    by parsing CustomResourceDefinitionDescripton object. Alternatively, this informmation may
// 	  also come from special annotations in the CR/CRD;
// 3. Search and read contents identified in previous step, creating an intermediary secret to hold
//    data formatted as environment variables key/value;
// 4. Search applications that are interested to bind with given service, by inspecting labels. The
//    Deployment (and other kinds) will be updated in "spec" level.
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := reconcilerLog.WithValues(
		"Request.Namespace", request.Namespace,
		"Request.Name", request.Name,
	)

	logger.Info("Reconciling ServiceBindingRequest...")

	// fetch and validate namespaced ServiceBindingRequest instance
	sbr, err := r.getServiceBindingRequest(request.NamespacedName)
	if err != nil {
		if errors.Is(err, ApplicationNotFound) {
			logger.Info("SBR deleted after application deletion")
			return Done()
		}
		logger.Error(err, "On retrieving service-binding-request instance.")
		return DoneOnNotFound(err)
	}

	// validate namespaced ServiceBindingRequest instance (this check has been disabled until test data has been
	// adjusted to reflect the validation)
	//
	//if err = r.validateServiceBindingRequest(sbr); err != nil {
	//	logger.Error(err, "On validating service-binding-request instance.")
	//	return Done()
	//}

	logger = logger.WithValues("ServiceBindingRequest.Name", sbr.Name)
	logger.Debug("Found service binding request to inspect")

	ctx := context.Background()

	selectors := extractServiceSelectors(sbr)
	if len(selectors) == 0 {
		conditionsv1.SetStatusCondition(&sbr.Status.Conditions, conditionsv1.Condition{
			Type:    BindingReady,
			Status:  corev1.ConditionFalse,
			Reason:  EmptyServiceSelectorsReason,
			Message: "The spec.backingServiceSelectors field is empty.",
		})
		_, updateErr := updateServiceBindingRequestStatus(r.dynClient, sbr)
		if updateErr == nil {
			return Done()
		}
		// TODO: do not requeue here
		//
		// Since there are nothing to recover from in the case service selectors is empty, it is
		// still required to requeue due to some watches not being implemented. This is known issue
		// being worked in https://github.com/redhat-developer/service-binding-operator/pull/442.
		return RequeueError(ErrEmptyBackingServiceSelectors)
	}

	serviceCtxs, err := buildServiceContexts(
		r.dynClient,
		sbr.GetNamespace(),
		selectors,
		sbr.Spec.DetectBindingResources,
		r.RestMapper,
	)
	if err != nil {
		return RequeueError(err)
	}

	binding, err := buildBinding(
		r.dynClient,
		sbr.Spec.CustomEnvVar,
		serviceCtxs,
		sbr.Spec.EnvVarPrefix,
	)
	if err != nil {
		return RequeueError(err)
	}

	options := &ServiceBinderOptions{
		DynClient:              r.dynClient,
		DetectBindingResources: sbr.Spec.DetectBindingResources,
		SBR:                    sbr,
		Logger:                 logger,
		Objects:                serviceCtxs.GetServices(),
		Binding:                binding,
		RESTMapper:             r.RestMapper,
	}

	sb, err := BuildServiceBinder(ctx, options)
	if err != nil {
		// BuildServiceBinder can return only InvalidOptionsErr, and it is a programmer's error so
		// just bail out without re-queueing nor updating conditions.
		logger.Error(err, "Building ServiceBinder")
		return NoRequeue(err)
	}

	if sbr.GetDeletionTimestamp() != nil {
		logger := logger.WithName("unbind")
		logger.Info("Executing unbinding steps...")
		return sb.Unbind()
	}

	logger.Info("Binding applications with intermediary secret...")
	return sb.Bind()
}
