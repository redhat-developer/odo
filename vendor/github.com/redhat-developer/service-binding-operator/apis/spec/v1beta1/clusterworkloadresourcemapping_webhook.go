/*
Copyright 2021.

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

package v1beta1

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/util/jsonpath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (r *ClusterWorkloadResourceMapping) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-servicebinding-io-v1beta1-clusterworkloadresourcemapping,mutating=false,failurePolicy=fail,sideEffects=None,groups=servicebinding.io,resources=clusterworkloadresourcemappings,verbs=create;update,versions=v1beta1,name=vclusterworkloadresourcemapping.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ClusterWorkloadResourceMapping{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (m *ClusterWorkloadResourceMapping) ValidateCreate() error {
	return m.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (m *ClusterWorkloadResourceMapping) ValidateUpdate(old runtime.Object) error {
	return m.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (m *ClusterWorkloadResourceMapping) ValidateDelete() error {
	// don't currently need validation of resource removal
	return nil
}

func (m *ClusterWorkloadResourceMapping) validate() error {
	errs := field.ErrorList{}
	path := field.NewPath("spec")
	versions := map[string]int{}

	for i, template := range m.Spec.Versions {
		childPath := path.Child(fmt.Sprintf("versions[%d]", i))
		if template.Version == "" {
			errs = append(errs, field.Required(childPath.Child("version"), "field \"version\" required"))
		} else if strings.TrimSpace(template.Version) == "" {
			errs = append(errs, field.Invalid(childPath.Child("version"), template.Version, "Whitespace-only version field forbidden"))
		} else if _, ok := versions[template.Version]; ok {
			errs = append(errs, field.Duplicate(childPath, template.Version))
		}

		versions[template.Version] = i
		errs = append(errs, template.validate(childPath)...)
	}

	return errs.ToAggregate()
}

func (template *ClusterWorkloadResourceMappingTemplate) validate(path *field.Path) field.ErrorList {
	errs := field.ErrorList{}
	for i, container := range template.Containers {
		child := path.Child(fmt.Sprintf("containers[%d]", i))
		errs = append(errs, container.validate(child)...)
	}

	errs = append(errs, validateRestrictedPath(path.Child("volumes"), template.Volumes)...)
	errs = append(errs, validateRestrictedPath(path.Child("annotations"), template.Annotations)...)

	return errs
}

func (container *ClusterWorkloadResourceMappingContainer) validate(path *field.Path) field.ErrorList {
	errs := field.ErrorList{}

	errs = append(errs, validateRestrictedPath(path.Child("name"), container.Name)...)
	errs = append(errs, validateRestrictedPath(path.Child("env"), container.Env)...)
	errs = append(errs, validateRestrictedPath(path.Child("volumeMounts"), container.VolumeMounts)...)

	if container.Path == "" {
		errs = append(errs, field.Required(path.Child("path"), "field \"path\" required"))
	} else {
		jsonpath := jsonpath.New("")
		formatted := fmt.Sprintf("{%s}", container.Path)
		if err := jsonpath.Parse(formatted); err != nil {
			errs = append(errs, field.Invalid(path.Child("path"), container.Path, "Invalid JSONPath"))
		}
	}

	return errs
}

func validateRestrictedPath(fieldPath *field.Path, value string) field.ErrorList {
	if value != "" {
		return isValidRestrictedJsonPath(fieldPath, value)
	}
	return nil
}

func isValidRestrictedJsonPath(fieldPath *field.Path, path string) field.ErrorList {
	errs := field.ErrorList{}
	parser, err := jsonpath.Parse("", fmt.Sprintf("{%s}", path))
	if err != nil {
		return append(errs, field.Invalid(fieldPath, path, "Unable to parse fixed JSONPath"))
	}
	if len(verifyJsonPath(parser.Root, fieldPath, path)) != 0 {
		errs = append(errs, field.Invalid(fieldPath, path, "Invalid fixed JSONPath"))
	}
	return errs
}

func verifyJsonPath(node jsonpath.Node, path *field.Path, value string) field.ErrorList {
	errs := field.ErrorList{}
	switch node.Type() {
	case jsonpath.NodeField:
		break
	case jsonpath.NodeList:
		list := node.(*jsonpath.ListNode)
		for _, node := range list.Nodes {
			errs = append(errs, verifyJsonPath(node, path, value)...)
		}
	default:
		errs = append(errs, field.Invalid(path, node.Type().String(), "Invalid node type"))
	}
	return errs
}
