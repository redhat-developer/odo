package service

import (
	"testing"

	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

// MockCRDescriptionOne a mock description
func MockCRDescriptionOne() *olm.CRDDescription {
	return &olm.CRDDescription{
		Name:        "etcdclusters.etcd.database.coreos.com",
		Version:     "v1beta2",
		Kind:        "EtcdCluster",
		DisplayName: "etcd Cluster",
		Resources: []olm.APIResourceReference{
			{Kind: "Service", Version: "v1"},
			{Kind: "Pod", Version: "v1"},
		},
		SpecDescriptors: []olm.SpecDescriptor{
			{
				Path:        "size",
				DisplayName: "Size",
				Description: "The desired number of member Pods for the etcd cluster.",
				XDescriptors: []string{
					"urn:alm:descriptor:com.tectonic.ui:podCount",
				},
			},
			{
				Path:        "pod.resources",
				DisplayName: "Resource Requirements",
				Description: "Limits describes the minimum/maximum amount of compute resources required/allowed",
				XDescriptors: []string{
					"urn:alm:descriptor:com.tectonic.ui:resourceRequirements",
				},
			},
		},
	}
}

func MockCRDescriptionTwo() *olm.CRDDescription {
	return &olm.CRDDescription{}
}
func TestCRBuilderMap(t *testing.T) {

}
