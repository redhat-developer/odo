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

package v1alpha1

import (
	"errors"

	"github.com/redhat-developer/service-binding-operator/apis"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var log = logf.Log.WithName("WebHook ServiceBinding")

func (r *ServiceBinding) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/validate-binding-operators-coreos-com-v1alpha1-servicebinding,mutating=false,failurePolicy=fail,sideEffects=None,groups=binding.operators.coreos.com,resources=servicebindings,verbs=create;update,versions=v1alpha1,name=vservicebinding.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ServiceBinding{}

func checkNameAndSelector(r *ServiceBinding) error {
	if r.Spec.Application.Name != "" && r.Spec.Application.LabelSelector != nil && r.Spec.Application.LabelSelector.MatchLabels != nil {
		err := errors.New("name and selector MUST NOT be defined in the application reference")
		log.Error(err, "name and selector check failed")
		return err
	}
	return nil
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ServiceBinding) ValidateCreate() error {
	return checkNameAndSelector(r)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ServiceBinding) ValidateUpdate(old runtime.Object) error {
	oldSb, ok := old.(*ServiceBinding)
	if !ok {
		return errors.New("Old object is not service binding")
	}
	err := apis.CanUpdateBinding(r, oldSb)
	if err != nil {
		log.Error(err, "Update failed")
		return err
	}
	return checkNameAndSelector(r)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ServiceBinding) ValidateDelete() error {
	return nil
}
