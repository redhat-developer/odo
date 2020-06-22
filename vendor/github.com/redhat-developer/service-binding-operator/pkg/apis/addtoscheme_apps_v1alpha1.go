package apis

import (
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
	AddToSchemes = append(AddToSchemes, olmv1alpha1.SchemeBuilder.AddToScheme)
}
