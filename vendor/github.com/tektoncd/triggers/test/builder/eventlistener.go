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

package builder

import (
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
)

// EventListenerOp is an operation which modifies the EventListener.
type EventListenerOp func(*v1alpha1.EventListener)

// EventListenerSpecOp is an operation which modifies the EventListenerSpec.
type EventListenerSpecOp func(*v1alpha1.EventListenerSpec)

// EventListenerStatusOp is an operation which modifies the EventListenerStatus.
type EventListenerStatusOp func(*v1alpha1.EventListenerStatus)

// EventListenerTriggerOp is an operation which modifies the Trigger.
type EventListenerTriggerOp func(*v1alpha1.EventListenerTrigger)

// EventInterceptorOp is an operation which modifies the EventInterceptor.
type EventInterceptorOp func(*v1alpha1.EventInterceptor)

// EventListener creates an EventListener with default values.
// Any number of EventListenerOp modifiers can be passed to transform it.
func EventListener(name, namespace string, ops ...EventListenerOp) *v1alpha1.EventListener {
	e := &v1alpha1.EventListener{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	for _, op := range ops {
		op(e)
	}

	return e
}

// EventListenerMeta sets the Meta structs of the EventListener.
// Any number of MetaOp modifiers can be passed.
func EventListenerMeta(ops ...MetaOp) EventListenerOp {
	return func(e *v1alpha1.EventListener) {
		for _, op := range ops {
			switch o := op.(type) {
			case ObjectMetaOp:
				o(&e.ObjectMeta)
			case TypeMetaOp:
				o(&e.TypeMeta)
			}
		}
	}
}

// EventListenerSpec sets the specified spec of the EventListener.
// Any number of EventListenerSpecOp modifiers can be passed to create/modify it.
func EventListenerSpec(ops ...EventListenerSpecOp) EventListenerOp {
	return func(e *v1alpha1.EventListener) {
		for _, op := range ops {
			op(&e.Spec)
		}
	}
}

// EventListenerServiceAccount sets the specified ServiceAccount of the EventListener.
func EventListenerServiceAccount(saName string) EventListenerSpecOp {
	return func(spec *v1alpha1.EventListenerSpec) {
		spec.ServiceAccountName = saName
	}
}

// EventListenerTrigger adds an EventListenerTrigger to the EventListenerSpec Triggers.
// Any number of EventListenerTriggerOp modifiers can be passed to create/modify it.
func EventListenerTrigger(ttName, apiVersion string, ops ...EventListenerTriggerOp) EventListenerSpecOp {
	return func(spec *v1alpha1.EventListenerSpec) {
		spec.Triggers = append(spec.Triggers, Trigger(ttName, apiVersion, ops...))
	}
}

// EventListenerStatus sets the specified status of the EventListener.
// Any number of EventListenerStatusOp modifiers can be passed to create/modify it.
func EventListenerStatus(ops ...EventListenerStatusOp) EventListenerOp {
	return func(e *v1alpha1.EventListener) {
		for _, op := range ops {
			op(&e.Status)
		}
	}
}

// EventListenerCondition sets the specified condition on the EventListenerStatus.
func EventListenerCondition(t apis.ConditionType, status corev1.ConditionStatus, message, reason string) EventListenerStatusOp {
	return func(e *v1alpha1.EventListenerStatus) {
		e.SetCondition(&apis.Condition{
			Type:    t,
			Status:  status,
			Message: message,
			Reason:  reason,
		})
	}
}

// EventListenerConfig sets the EventListenerConfiguration on the EventListenerStatus.
func EventListenerConfig(generatedResourceName string) EventListenerStatusOp {
	return func(e *v1alpha1.EventListenerStatus) {
		e.Configuration.GeneratedResourceName = generatedResourceName
	}
}

// EventListenerAddress sets the EventListenerAddress on the EventListenerStatus
func EventListenerAddress(hostname string) EventListenerStatusOp {
	return func(e *v1alpha1.EventListenerStatus) {
		e.Address = NewAddressable(hostname)
	}
}

func NewAddressable(hostname string) *duckv1alpha1.Addressable {
	addressable := &duckv1alpha1.Addressable{}
	addressable.URL = &apis.URL{
		Scheme: "http",
		Host:   hostname,
	}
	return addressable
}

// Trigger creates an EventListenerTrigger. Any number of EventListenerTriggerOp
// modifiers can be passed to create/modify it. For creating an EventListenerBinding
// you have to pass a EventListenerTriggerOp
func Trigger(ttName, apiVersion string, ops ...EventListenerTriggerOp) v1alpha1.EventListenerTrigger {
	t := v1alpha1.EventListenerTrigger{
		Template: v1alpha1.EventListenerTemplate{
			Name:       ttName,
			APIVersion: apiVersion,
		},
	}

	for _, op := range ops {
		op(&t)
	}

	return t
}

// EventListenerTriggerName adds a Name to the Trigger in EventListenerSpec Triggers.
func EventListenerTriggerName(name string) EventListenerTriggerOp {
	return func(trigger *v1alpha1.EventListenerTrigger) {
		trigger.Name = name
	}
}

// EventListenerTriggerServiceAccount set the specified ServiceAccount of the EventListenerTrigger.
func EventListenerTriggerServiceAccount(saName, namespace string) EventListenerTriggerOp {
	return func(trigger *v1alpha1.EventListenerTrigger) {
		trigger.ServiceAccount = &corev1.ObjectReference{
			Namespace: saName,
			Name:      namespace,
		}
	}
}

// EventListenerTriggerBinding adds a Binding to the Trigger in EventListenerSpec Triggers.
func EventListenerTriggerBinding(ref, kind, name, apiVersion string, ops ...TriggerBindingSpecOp) EventListenerTriggerOp {
	return func(trigger *v1alpha1.EventListenerTrigger) {
		binding := &v1alpha1.EventListenerBinding{
			Name:       name,
			APIVersion: apiVersion,
		}

		if len(ref) != 0 {
			binding.Ref = ref
			if kind == "ClusterTriggerBinding" {
				binding.Kind = v1alpha1.ClusterTriggerBindingKind
			}

			if kind == "TriggerBinding" || kind == "" {
				binding.Kind = v1alpha1.NamespacedTriggerBindingKind
			}
		}
		if len(ops) != 0 {
			binding.Spec = &v1alpha1.TriggerBindingSpec{}
			for _, op := range ops {
				op(binding.Spec)
			}
		}
		trigger.Bindings = append(trigger.Bindings, binding)
	}
}

// EventListenerTriggerInterceptor adds an objectRef to an interceptor Service to the EventListenerTrigger.
func EventListenerTriggerInterceptor(name, version, kind, namespace string, ops ...EventInterceptorOp) EventListenerTriggerOp {
	return func(t *v1alpha1.EventListenerTrigger) {
		i := &v1alpha1.EventInterceptor{
			Webhook: &v1alpha1.WebhookInterceptor{
				ObjectRef: &corev1.ObjectReference{
					Kind:       kind,
					Name:       name,
					APIVersion: version,
					Namespace:  namespace,
				},
			},
		}
		for _, op := range ops {
			op(i)
		}
		t.Interceptors = append(t.Interceptors, i)
	}
}

// EventInterceptorParam adds a parameter to the EventInterceptor.
func EventInterceptorParam(name, value string) EventInterceptorOp {
	return func(i *v1alpha1.EventInterceptor) {
		if i.Webhook != nil {
			for _, param := range i.Webhook.Header {
				if param.Name == name {
					param.Value.ArrayVal = append(param.Value.ArrayVal, value)
					return
				}
			}
			i.Webhook.Header = append(i.Webhook.Header,
				pipelinev1.Param{
					Name: name,
					Value: pipelinev1.ArrayOrString{
						ArrayVal: []string{value},
						Type:     pipelinev1.ParamTypeArray,
					},
				})
		}
	}
}

// EventListenerCELInterceptor adds a CEL filter to the EventListenerTrigger.
func EventListenerCELInterceptor(filter string, ops ...EventInterceptorOp) EventListenerTriggerOp {
	return func(t *v1alpha1.EventListenerTrigger) {
		i := &v1alpha1.EventInterceptor{
			CEL: &v1alpha1.CELInterceptor{
				Filter: filter,
			},
		}
		for _, op := range ops {
			op(i)
		}
		t.Interceptors = append(t.Interceptors, i)
	}
}

func EventListenerCELOverlay(key, expression string) EventInterceptorOp {
	return func(i *v1alpha1.EventInterceptor) {
		if i.CEL != nil {
			i.CEL.Overlays = append(i.CEL.Overlays, v1alpha1.CELOverlay{
				Key:        key,
				Expression: expression,
			})
		}
	}
}
