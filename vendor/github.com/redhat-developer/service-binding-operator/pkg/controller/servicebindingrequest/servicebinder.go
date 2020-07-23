package servicebindingrequest

import (
	"context"
	"fmt"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	"gotest.tools/assert/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

const (
	// BindingSuccess binding has succeeded
	BindingSuccess = "BindingSuccess"
	// BindingFail binding has failed
	BindingFail = "BindingFail"
	//Finalizer annotation used in finalizer steps
	Finalizer = "finalizer.servicebindingrequest.openshift.io"
	// time in seconds to wait before requeuing requests
	requeueAfter int64 = 45
)

// GroupVersion represents the service binding request resource's group version.
var GroupVersion = v1alpha1.SchemeGroupVersion.WithResource(ServiceBindingRequestResource)

// message converts the error to string for the Message field in the Status condition
func (b *ServiceBinder) message(err error) string {
	return err.Error()
}

// ServiceBinderOptions is BuildServiceBinder arguments.
type ServiceBinderOptions struct {
	Logger                 *log.Log
	DynClient              dynamic.Interface
	DetectBindingResources bool
	SBR                    *v1alpha1.ServiceBindingRequest
	Objects                []*unstructured.Unstructured
	EnvVars                map[string][]byte
	EnvVarPrefix           string
	Binding                *Binding
	RESTMapper             meta.RESTMapper
}

// ErrInvalidServiceBinderOptions is returned when ServiceBinderOptions contains an invalid value.
type ErrInvalidServiceBinderOptions string

func (e ErrInvalidServiceBinderOptions) Error() string {
	return fmt.Sprintf("option %q is missing", string(e))
}

// Valid returns an error if the receiver is invalid, nil otherwise.
func (o *ServiceBinderOptions) Valid() error {
	if o.SBR == nil {
		return ErrInvalidServiceBinderOptions("SBR")
	}

	if o.DynClient == nil {
		return ErrInvalidServiceBinderOptions("DynClient")
	}

	if o.Binding == nil {
		return ErrInvalidServiceBinderOptions("Binding")
	}

	if o.RESTMapper == nil {
		return ErrInvalidServiceBinderOptions("RESTMapper")
	}

	return nil
}

// ServiceBinder manages binding for a Service Binding Request and associated objects.
type ServiceBinder struct {
	// Binder is responsible for interacting with the cluster and apply binding related changes.
	Binder *Binder
	// EnvVars contains the environment variables to bind.
	EnvVars map[string][]byte
	// DynClient is the Kubernetes dynamic client used to interact with the cluster.
	DynClient dynamic.Interface
	// Logger provides logging facilities for internal components.
	Logger *log.Log
	// Objects is a list of additional unstructured objects related to the Service Binding Request.
	Objects []*unstructured.Unstructured
	// SBR is the ServiceBindingRequest associated with binding.
	SBR *v1alpha1.ServiceBindingRequest
	// Secret is the Secret associated with the Service Binding Request.
	Secret *Secret
}

// updateServiceBindingRequest execute update API call on a SBR request. It can return errors from
// this action.
func updateServiceBindingRequest(
	dynClient dynamic.Interface,
	sbr *v1alpha1.ServiceBindingRequest,
) (*v1alpha1.ServiceBindingRequest, error) {
	u, err := converter.ToUnstructured(sbr)
	if err != nil {
		return nil, err
	}

	nsClient := dynClient.
		Resource(GroupVersion).
		Namespace(sbr.GetNamespace())

	u, err = nsClient.Update(u, v1.UpdateOptions{})

	if err != nil {
		return nil, err
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, sbr)
	if err != nil {
		return nil, err
	}

	return sbr, nil
}

// updateServiceBindingRequest execute update API call on a SBR request. It can return errors from
// this action.
func (b *ServiceBinder) updateServiceBindingRequest(
	sbr *v1alpha1.ServiceBindingRequest,
) (*v1alpha1.ServiceBindingRequest, error) {
	return updateServiceBindingRequest(b.DynClient, sbr)
}

// Unbind removes the relationship between a Service Binding Request and its related objects.
func (b *ServiceBinder) Unbind() (reconcile.Result, error) {
	logger := b.Logger.WithName("Unbind")

	// when finalizer is not found anymore, it can be safely removed
	if !containsStringSlice(b.SBR.GetFinalizers(), Finalizer) {
		logger.Info("Resource can be safely deleted!")
		return Done()
	}

	logger.Info("Cleaning related objects from operator's annotations...")
	if err := RemoveAndUpdateSBRAnnotations(b.DynClient, b.Objects); err != nil {
		logger.Error(err, "On removing annotations from related objects.")
		return RequeueError(err)
	}

	if err := b.Binder.Unbind(); err != nil {
		logger.Error(err, "On unbinding related objects")
		return RequeueError(err)
	}

	logger.Info("Deleting intermediary secret")
	if err := b.Secret.Delete(); err != nil {
		logger.Error(err, "On deleting intermediary secret.")
		return RequeueError(err)
	}

	logger.Debug("Removing resource finalizers...")
	b.SBR.SetFinalizers(removeStringSlice(b.SBR.GetFinalizers(), Finalizer))
	if _, err := b.updateServiceBindingRequest(b.SBR); err != nil {
		return NoRequeue(err)
	}

	return Done()
}

// UpdateServiceBindingRequestStatus execute update API call on a SBR Status. It can return errors from
// this action.
func updateServiceBindingRequestStatus(
	dynClient dynamic.Interface,
	sbr *v1alpha1.ServiceBindingRequest,
) (*v1alpha1.ServiceBindingRequest, error) {
	u, err := converter.ToUnstructured(sbr)
	if err != nil {
		return nil, err
	}

	nsClient := dynClient.
		Resource(GroupVersion).
		Namespace(sbr.GetNamespace())

	u, err = nsClient.UpdateStatus(u, v1.UpdateOptions{})

	if err != nil {
		return nil, err
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, sbr)
	if err != nil {
		return nil, err
	}

	return sbr, nil
}

// updateStatusServiceBindingRequest updates the Service Binding Request's status field.
func (b *ServiceBinder) updateStatusServiceBindingRequest(
	sbr *v1alpha1.ServiceBindingRequest,
	sbrStatus *v1alpha1.ServiceBindingRequestStatus,
) (
	*v1alpha1.ServiceBindingRequest,
	error,
) {
	// do not update if both statuses are equal
	if result := cmp.DeepEqual(sbr.Status, sbrStatus)(); result.Success() {
		return sbr, nil
	}

	// coping status over informed object
	sbr.Status = *sbrStatus

	return updateServiceBindingRequestStatus(b.DynClient, sbr)
}

// onError comprise the update of ServiceBindingRequest status to set error flag, and inspect
// informed error to apply a different behavior for not-founds.
func (b *ServiceBinder) onError(
	err error,
	sbr *v1alpha1.ServiceBindingRequest,
	sbrStatus *v1alpha1.ServiceBindingRequestStatus,
	objs []*unstructured.Unstructured,
) (reconcile.Result, error) {

	if objs != nil {
		b.setApplicationObjects(sbrStatus, objs)
	}
	conditionsv1.SetStatusCondition(&sbrStatus.Conditions, conditionsv1.Condition{
		Type:    BindingReady,
		Status:  corev1.ConditionFalse,
		Reason:  BindingFail,
		Message: b.message(err),
	})
	newSbr, errStatus := b.updateStatusServiceBindingRequest(sbr, sbrStatus)
	if errStatus != nil {
		return RequeueError(errStatus)
	}
	b.SBR = newSbr

	return RequeueOnNotFound(err, requeueAfter)
}

// Bind configures binding between the Service Binding Request and its related objects.
func (b *ServiceBinder) Bind() (reconcile.Result, error) {
	sbrStatus := b.SBR.Status.DeepCopy()

	b.Logger.Info("Saving data on intermediary secret...")
	secretObj, err := b.Secret.Commit(b.EnvVars)
	if err != nil {
		b.Logger.Error(err, "On saving secret data..")
		return b.onError(err, b.SBR, sbrStatus, nil)
	}
	sbrStatus.Secret = secretObj.GetName()

	updatedObjects, err := b.Binder.Bind()
	if err != nil {
		b.Logger.Error(err, "On binding application.")
		return b.onError(err, b.SBR, sbrStatus, updatedObjects)
	}
	b.setApplicationObjects(sbrStatus, updatedObjects)

	// annotating objects related to binding
	namespacedName := types.NamespacedName{Namespace: b.SBR.GetNamespace(), Name: b.SBR.GetName()}
	if err = SetAndUpdateSBRAnnotations(b.DynClient, namespacedName, append(b.Objects, secretObj)); err != nil {
		b.Logger.Error(err, "On setting annotations in related objects.")
		return b.onError(err, b.SBR, sbrStatus, updatedObjects)
	}

	conditionsv1.SetStatusCondition(&sbrStatus.Conditions, conditionsv1.Condition{
		Type:   BindingReady,
		Status: corev1.ConditionTrue,
	})

	// updating status of request instance
	sbr, err := b.updateStatusServiceBindingRequest(b.SBR, sbrStatus)
	if err != nil {
		return RequeueOnConflict(err)
	}

	// appending finalizer, should be later removed upon resource deletion
	sbr.SetFinalizers(append(removeStringSlice(b.SBR.GetFinalizers(), Finalizer), Finalizer))
	if _, err = b.updateServiceBindingRequest(sbr); err != nil {
		return NoRequeue(err)
	}

	b.Logger.Info("All done!")
	return Done()
}

// setApplicationObjects replaces the Status's equivalent field.
func (b *ServiceBinder) setApplicationObjects(
	sbrStatus *v1alpha1.ServiceBindingRequestStatus,
	objs []*unstructured.Unstructured,
) {
	boundApps := []v1alpha1.BoundApplication{}
	for _, obj := range objs {
		boundApp := v1alpha1.BoundApplication{
			GroupVersionKind: v1.GroupVersionKind{
				Group:   obj.GroupVersionKind().Group,
				Version: obj.GroupVersionKind().Version,
				Kind:    obj.GetKind(),
			},
			LocalObjectReference: corev1.LocalObjectReference{
				Name: obj.GetName(),
			},
		}
		boundApps = append(boundApps, boundApp)
	}
	sbrStatus.Applications = boundApps
}

// BuildServiceBinder creates a new binding manager according to options.
func BuildServiceBinder(
	ctx context.Context,
	options *ServiceBinderOptions,
) (
	*ServiceBinder,
	error,
) {
	if err := options.Valid(); err != nil {
		return nil, err
	}

	// FIXME(isuttonl): review whether it is possible to move Secret.Commit() and Secret.Delete() to
	// ServiceBinder.
	secret := NewSecret(
		options.DynClient,
		options.SBR.GetNamespace(),
		options.SBR.GetName(),
	)

	// FIXME(isuttonl): review whether binder can be lazily created in Bind() and Unbind(); also
	// consider renaming to ResourceBinder
	binder := NewBinder(
		ctx,
		options.DynClient,
		options.SBR,
		options.Binding.VolumeKeys,
		options.RESTMapper,
	)

	return &ServiceBinder{
		Logger:    options.Logger,
		Binder:    binder,
		DynClient: options.DynClient,
		SBR:       options.SBR,
		Objects:   options.Objects,
		EnvVars:   options.Binding.EnvVars,
		Secret:    secret,
	}, nil
}

type Binding struct {
	EnvVars    map[string][]byte
	VolumeKeys []string
}

func buildBinding(
	client dynamic.Interface,
	customEnvVar []corev1.EnvVar,
	svcCtxs ServiceContextList,
	globalEnvVarPrefix string,
) (*Binding, error) {
	envVars, volumeKeys, err := NewRetriever(client).
		ProcessServiceContexts(globalEnvVarPrefix, svcCtxs, customEnvVar)
	if err != nil {
		return nil, err
	}

	return &Binding{
		EnvVars:    envVars,
		VolumeKeys: volumeKeys,
	}, nil
}
