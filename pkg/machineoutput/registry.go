package machineoutput

import (
	"github.com/redhat-developer/odo/pkg/preference"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RegistryListOutput struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	RegistryList      *[]preference.Registry `json:"registries,omitempty"`
}

func NewRegistryListOutput(registryList *[]preference.Registry) RegistryListOutput {
	return RegistryListOutput{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: APIVersion,
		},
		RegistryList: registryList,
	}
}
