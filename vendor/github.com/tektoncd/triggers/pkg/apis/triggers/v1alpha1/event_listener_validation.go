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
	"context"
	"fmt"
	"net/http"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"k8s.io/apimachinery/pkg/util/validation"
	"knative.dev/pkg/apis"
)

// Validate EventListener.
func (e *EventListener) Validate(ctx context.Context) *apis.FieldError {
	return e.Spec.validate(ctx, e)
}

func (s *EventListenerSpec) validate(ctx context.Context, el *EventListener) *apis.FieldError {
	if len(s.Triggers) == 0 {
		return apis.ErrMissingField("spec.triggers")
	}
	for i, trigger := range s.Triggers {
		if err := trigger.validate(ctx).ViaField(fmt.Sprintf("spec.triggers[%d]", i)); err != nil {
			return err
		}
	}
	return nil
}

func (t *EventListenerTrigger) validate(ctx context.Context) *apis.FieldError {
	// Validate that only one of binding or bindings is set
	if t.DeprecatedBinding != nil && len(t.Bindings) > 0 {
		return apis.ErrMultipleOneOf("binding", "bindings")
	}
	// Validate that only one of inteceptor or interceptors is set
	if t.DeprecatedInterceptor != nil && len(t.Interceptors) > 0 {
		return apis.ErrMultipleOneOf("interceptor", "interceptors")
	}

	// Validate optional TriggerBinding
	for i, b := range t.Bindings {
		if b.Name == "" {
			return apis.ErrMissingField(fmt.Sprintf("bindings[%d].name", i))
		}
	}
	// Validate required TriggerTemplate
	// Optional explicit match
	if t.Template.APIVersion != "" {
		if t.Template.APIVersion != "v1alpha1" {
			return apis.ErrInvalidValue(fmt.Errorf("invalid apiVersion"), "template.apiVersion")
		}
	}
	if t.Template.Name == "" {
		return apis.ErrMissingField(fmt.Sprintf("template.name"))
	}
	for i, interceptor := range t.Interceptors {
		if err := interceptor.validate(ctx).ViaField(fmt.Sprintf("interceptors[%d]", i)); err != nil {
			return err
		}
	}

	// The trigger name is added as a label value for 'tekton.dev/trigger' so it must follow the k8s label guidelines:
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set
	if errs := validation.IsValidLabelValue(t.Name); len(errs) > 0 {
		return apis.ErrInvalidValue(fmt.Sprintf("trigger name '%s' must be a valid label value", t.Name), "name")
	}

	return nil
}

func (i *EventInterceptor) validate(ctx context.Context) *apis.FieldError {
	if i.Webhook == nil && i.GitHub == nil && i.GitLab == nil && i.CEL == nil {
		return apis.ErrMissingField("interceptor")
	}

	// Enforce oneof
	numSet := 0
	if i.Webhook != nil {
		numSet++
	}
	if i.GitHub != nil {
		numSet++
	}
	if i.GitLab != nil {
		numSet++
	}

	if numSet > 1 {
		return apis.ErrMultipleOneOf("interceptor.webhook", "interceptor.github", "interceptor.gitlab")
	}

	if i.Webhook != nil {
		if i.Webhook.ObjectRef == nil || i.Webhook.ObjectRef.Name == "" {
			return apis.ErrMissingField("interceptor.webhook.objectRef")
		}
		w := i.Webhook
		if w.ObjectRef.Kind != "Service" {
			return apis.ErrInvalidValue(fmt.Errorf("invalid kind"), "interceptor.webhook.objectRef.kind")
		}

		// Optional explicit match
		if w.ObjectRef.APIVersion != "v1" {
			return apis.ErrInvalidValue(fmt.Errorf("invalid apiVersion"), "interceptor.webhook.objectRef.apiVersion")
		}

		for i, header := range w.Header {
			// Enforce non-empty canonical header keys
			if len(header.Name) == 0 || http.CanonicalHeaderKey(header.Name) != header.Name {
				return apis.ErrInvalidValue(fmt.Errorf("invalid header name"), fmt.Sprintf("interceptor.webhook.header[%d].name", i))
			}
			// Enforce non-empty header values
			if header.Value.Type == pipelinev1.ParamTypeString {
				if len(header.Value.StringVal) == 0 {
					return apis.ErrInvalidValue(fmt.Errorf("invalid header value"), fmt.Sprintf("interceptor.webhook.header[%d].value", i))
				}
			} else if len(header.Value.ArrayVal) == 0 {
				return apis.ErrInvalidValue(fmt.Errorf("invalid header value"), fmt.Sprintf("interceptor.webhook.header[%d].value", i))
			}
		}
	}

	// No github validation required yet.
	// if i.GitHub != nil {
	//
	// }

	// No gitlab validation required yet.
	// if i.GitLab != nil {
	//
	// }

	if i.CEL != nil {
		if i.CEL.Filter == "" {
			return apis.ErrMissingField("interceptor.cel.filter")
		}
	}
	return nil
}
