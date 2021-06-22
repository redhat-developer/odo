package service

import (
	"testing"

	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/stretchr/testify/require"
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

// MockCRDescriptionTwo a mock description
func MockCRDescriptionTwo() *olm.CRDDescription {
	return &olm.CRDDescription{
		Name:        "pgclusters.crunchydata.com",
		Version:     "v1",
		Kind:        "Pgcluster",
		DisplayName: "Postgres Primary Cluster Member",
		Description: "Represents a Postgres primary cluster member",
		Resources: []olm.APIResourceReference{
			{Kind: "Pgcluster", Version: "v1"},
			{Kind: "ConfigMap", Version: "v1"},
			{Kind: "Deployment", Version: "v1"},
			{Kind: "Job", Version: "v1"},
			{Kind: "Pod", Version: "v1"},
			{Kind: "ReplicaSet", Version: "v1"},
			{Kind: "Secret", Version: "v1"},
			{Kind: "Service", Version: "v1"},
			{Kind: "PersistentVolumeClaim", Version: "v1"},
		},
		SpecDescriptors: []olm.SpecDescriptor{
			{
				Path:        "cppimage",
				DisplayName: "PostgreSQL Image",
				Description: "The Crunchy PostgreSQL image to use. Possible values are \"crunchy-postgres-ha\" and \"crunchy-postgres-gis-ha\"",
			},
			{
				Path:        "cppimagetag",
				DisplayName: "PostgresSQL Image Tag",
				Description: "The tag of the PostgreSQL image to use. Example is \"ubi7-12.4-4.5.0\"",
			},
			{
				Path:        "host",
				DisplayName: "PostgreSQL Host",
			},
		},
	}
}
func TestCRBuilderMap(t *testing.T) {
	builder := NewCRBuilder(MockCRDescriptionTwo())
	require.Nil(t, builder.SetAndValidate("host", "10.1.10.2"), "set shouldn't fail")
	require.Nil(t, builder.SetAndValidate("cppimage", "crunchy-postgres-ha"), "set shouldn't fail")
	require.Nil(t, builder.SetAndValidate("cppimagetag", "2.5"), "set shouldn't fail")
	outMap, err := builder.Map()
	require.Nil(t, err, "error shouldn't be nil")
	expected := map[string]interface{}{
		"apiVersion": "crunchydata.com/v1",
		"kind":       "Pgcluster",
		"metadata":   map[string]interface{}{},
		"spec": map[string]interface{}{
			"cppimage":    "crunchy-postgres-ha",
			"cppimagetag": 2.5,
			"host":        "10.1.10.2",
		},
	}
	require.Equal(t, outMap, expected, "The map output doesn't match the expected out")
}
