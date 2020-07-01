/*
Copyright 2019 The Tekton Authors

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

package v1alpha1

import (
	"fmt"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

// Check that EventListener may be validated and defaulted.
var _ apis.Validatable = (*EventListener)(nil)
var _ apis.Defaultable = (*EventListener)(nil)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EventListener exposes a service to accept HTTP event payloads.
//
// +k8s:openapi-gen=true
type EventListener struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec holds the desired state of the EventListener from the client
	// +optional
	Spec EventListenerSpec `json:"spec"`
	// +optional
	Status EventListenerStatus `json:"status,omitempty"`
}

// EventListenerSpec defines the desired state of the EventListener, represented
// by a list of Triggers.
type EventListenerSpec struct {
	ServiceAccountName string                 `json:"serviceAccountName"`
	Triggers           []EventListenerTrigger `json:"triggers"`
	ServiceType        corev1.ServiceType     `json:"serviceType,omitempty"`
}

// EventListenerTrigger represents a connection between TriggerBinding, Params,
// and TriggerTemplate; TriggerBinding provides extracted values for
// TriggerTemplate to then create resources from.
type EventListenerTrigger struct {
	Bindings []*EventListenerBinding `json:"bindings"`
	Template EventListenerTemplate   `json:"template"`
	// +optional
	Name         string              `json:"name,omitempty"`
	Interceptors []*EventInterceptor `json:"interceptors,omitempty"`
	// ServiceAccount optionally associates credentials with each trigger;
	// more granular authorization for
	// who is allowed to utilize the associated pipeline
	// vs. defaulting to whatever permissions are associated
	// with the entire EventListener and associated sink facilitates
	// multi-tenant model based scenarios
	// TODO do we want to restrict this to the event listener namespace and just ask for the service account name here?
	// +optional
	ServiceAccount *corev1.ObjectReference `json:"serviceAccount,omitempty"`
}

// EventInterceptor provides a hook to intercept and pre-process events
type EventInterceptor struct {
	Webhook *WebhookInterceptor `json:"webhook,omitempty"`
	GitHub  *GitHubInterceptor  `json:"github,omitempty"`
	GitLab  *GitLabInterceptor  `json:"gitlab,omitempty"`
	CEL     *CELInterceptor     `json:"cel,omitempty"`
}

// WebhookInterceptor provides a webhook to intercept and pre-process events
type WebhookInterceptor struct {
	// ObjectRef is a reference to an object that will resolve to a cluster DNS
	// name to use as the EventInterceptor. Either objectRef or url can be specified
	// +optional
	ObjectRef *corev1.ObjectReference `json:"objectRef,omitempty"`
	// Header is a group of key-value pairs that can be appended to the
	// interceptor request headers. This allows the interceptor to make
	// decisions specific to an EventListenerTrigger.
	Header []v1beta1.Param `json:"header,omitempty"`
}

// GitHubInterceptor provides a webhook to intercept and pre-process events
type GitHubInterceptor struct {
	SecretRef  *SecretRef `json:"secretRef,omitempty"`
	EventTypes []string   `json:"eventTypes,omitempty"`
}

// GitLabInterceptor provides a webhook to intercept and pre-process events
type GitLabInterceptor struct {
	SecretRef  *SecretRef `json:"secretRef,omitempty"`
	EventTypes []string   `json:"eventTypes,omitempty"`
}

// CELInterceptor provides a webhook to intercept and pre-process events
type CELInterceptor struct {
	Filter   string       `json:"filter,omitempty"`
	Overlays []CELOverlay `json:"overlays,omitempty"`
}

// CELOverlay provides a way to modify the request body using CEL expressions
type CELOverlay struct {
	Key        string `json:"key,omitempty"`
	Expression string `json:"expression,omitempty"`
}

// SecretRef contains the information required to reference a single secret string
// This is needed because the other secretRef types are not cross-namespace and do not
// actually contain the "SecretName" field, which allows us to access a single secret value.
type SecretRef struct {
	SecretKey  string `json:"secretKey,omitempty"`
	SecretName string `json:"secretName,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
}

// EventListenerBinding refers to a particular TriggerBinding or ClusterTriggerBindingresource.
type EventListenerBinding struct {
	Name       string              `json:"name,omitempty"`
	Kind       TriggerBindingKind  `json:"kind,omitempty"`
	Ref        string              `json:"ref,omitempty"`
	Spec       *TriggerBindingSpec `json:"spec,omitempty"`
	APIVersion string              `json:"apiversion,omitempty"`
}

// EventListenerTemplate refers to a particular TriggerTemplate resource.
type EventListenerTemplate struct {
	Name       string `json:"name"`
	APIVersion string `json:"apiversion,omitempty"`
}

// EventListenerList contains a list of TriggerBinding
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type EventListenerList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EventListener `json:"items"`
}

// EventListenerStatus holds the status of the EventListener
// +k8s:deepcopy-gen=true
type EventListenerStatus struct {
	duckv1beta1.Status `json:",inline"`

	// EventListener is Addressable. It currently exposes the service DNS
	// address of the the EventListener sink
	duckv1alpha1.AddressStatus `json:",inline"`

	// Configuration stores configuration for the EventListener service
	Configuration EventListenerConfig `json:"configuration"`
}

// EventListenerConfig stores configuration for resources generated by the
// EventListener
type EventListenerConfig struct {
	// GeneratedResourceName is the name given to all resources reconciled by
	// the EventListener
	GeneratedResourceName string `json:"generatedName"`
}

// The conditions that are internally resolved by the EventListener reconciler
const (
	// ServiceExists is the ConditionType set on the EventListener, which
	// specifies Service existence.
	ServiceExists apis.ConditionType = "Service"
	// DeploymentExists is the ConditionType set on the EventListener, which
	// specifies Deployment existence.
	DeploymentExists apis.ConditionType = "Deployment"
)

// Check that EventListener may be validated and defaulted.
// TriggerBindingKind defines the type of TriggerBinding used by the EventListener.
type TriggerBindingKind string

const (
	// NamespacedTriggerBindingKind indicates that triggerbinding type has a namespace scope.
	NamespacedTriggerBindingKind TriggerBindingKind = "TriggerBinding"
	// ClusterTriggerBindingKind indicates that triggerbinding type has a cluster scope.
	ClusterTriggerBindingKind TriggerBindingKind = "ClusterTriggerBinding"
)

var eventListenerCondSet = apis.NewLivingConditionSet(ServiceExists, DeploymentExists)

// GetCondition returns the Condition matching the given type.
func (els *EventListenerStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return eventListenerCondSet.Manage(els).GetCondition(t)
}

// SetCondition sets the condition, unsetting previous conditions with the same
// type as necessary. This is a local change and needs to be persisted to the
// K8s API elsewhere.
func (els *EventListenerStatus) SetCondition(newCond *apis.Condition) {
	if newCond != nil {
		eventListenerCondSet.Manage(els).SetCondition(*newCond)
	}
}

// SetDeploymentConditions sets the Deployment conditions on the EventListener,
// which is a reflection of the actual Deployment status.
func (els *EventListenerStatus) SetDeploymentConditions(deploymentConditions []appsv1.DeploymentCondition) {
	// Manually remove the DeploymentReplicaFailure condition since it does
	// not always exist and would stay around otherwise
	replicaFailureIndex := -1
	for i := range els.Conditions {
		if els.Conditions[i].Type == apis.ConditionType(appsv1.DeploymentReplicaFailure) {
			replicaFailureIndex = i
			break
		}
	}
	if replicaFailureIndex != -1 {
		els.Conditions = append(els.Conditions[:replicaFailureIndex], els.Conditions[replicaFailureIndex+1:]...)
	}
	for _, cond := range deploymentConditions {
		els.SetCondition(&apis.Condition{
			Type:    apis.ConditionType(cond.Type),
			Status:  cond.Status,
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}
}

// SetExistsCondition simplifies setting the exists conditions on the
// EventListenerStatus.
func (els *EventListenerStatus) SetExistsCondition(cond apis.ConditionType, err error) {
	if err != nil {
		els.SetCondition(&apis.Condition{
			Type:    cond,
			Status:  corev1.ConditionFalse,
			Message: err.Error(),
		})
	} else {
		els.SetCondition(&apis.Condition{
			Type:    cond,
			Status:  corev1.ConditionTrue,
			Message: fmt.Sprintf("%s exists", cond),
		})
	}
}

// InitializeConditions will set all conditions in eventListenerCondSet to false
// for the EventListener. This does not use the InitializeCondition() provided
// by the conditionsImpl to avoid setting the happy condition. This is a local
// change and needs to be persisted to the K8s API elsewhere.
func (els *EventListenerStatus) InitializeConditions() {
	for _, condition := range []apis.ConditionType{
		ServiceExists,
		DeploymentExists,
	} {
		els.SetCondition(&apis.Condition{
			Type:   condition,
			Status: corev1.ConditionFalse,
		})
	}
}

// GetOwnerReference gets the EventListener as owner reference for any related
// objects.
func (el *EventListener) GetOwnerReference() *metav1.OwnerReference {
	return metav1.NewControllerRef(el, schema.GroupVersionKind{
		Group:   SchemeGroupVersion.Group,
		Version: SchemeGroupVersion.Version,
		Kind:    "EventListener",
	})
}

// SetAddress sets the address (as part of Addressable contract) and marks the correct condition.
func (els *EventListenerStatus) SetAddress(hostname string) {
	if els.Address == nil {
		els.Address = &duckv1alpha1.Addressable{}
	}
	if hostname != "" {
		els.Address.URL = &apis.URL{
			Scheme: "http",
			Host:   hostname,
		}
	} else {
		els.Address.URL = nil
	}
}
