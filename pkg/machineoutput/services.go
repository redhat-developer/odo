package machineoutput

import (
	"github.com/openshift/odo/pkg/catalog"
	olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CatalogListOutput struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Services          *catalog.ServiceTypeList `json:"services,omitempty"`
	// list of clusterserviceversions (installed by Operators)
	Operators *olm.ClusterServiceVersionList `json:"operators,omitempty"`
}

func NewCatalogListOutput(services *catalog.ServiceTypeList, operators *olm.ClusterServiceVersionList) CatalogListOutput {
	return CatalogListOutput{
		TypeMeta: metav1.TypeMeta{
			Kind: "CatalogListOutput",
		},
		Services:  services,
		Operators: operators,
	}
}
