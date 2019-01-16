package testingutil

import (
	"encoding/json"
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// M is an alias for map[string]interface{}
type M map[string]interface{}

// FakeClusterServiceClass creates a fake service class with the specified name for testing purposes
func FakeClusterServiceClass(name string, tags ...string) v1beta1.ClusterServiceClass {
	class := v1beta1.ClusterServiceClass{
		Spec: v1beta1.ClusterServiceClassSpec{
			CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
				ExternalName: name,
			},
		},
	}

	if len(tags) > 0 {
		class.Spec.Tags = tags
	}

	return class
}

// FakeClusterServicePlan creates a fake plan with the specified external name and using planNumber to customize description,
// metadata and parameter values
func FakeClusterServicePlan(name string, planNumber int) v1beta1.ClusterServicePlan {
	return v1beta1.ClusterServicePlan{
		Spec: v1beta1.ClusterServicePlanSpec{
			ClusterServiceClassRef: v1beta1.ClusterObjectReference{
				Name: "1dda1477cace09730bd8ed7a6505607e",
			},
			CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
				ExternalName:                         name,
				Description:                          fmt.Sprintf("this is a example description %d", planNumber),
				ExternalMetadata:                     &runtime.RawExtension{Raw: fakePlanExternalMetaData(planNumber)},
				ServiceInstanceCreateParameterSchema: &runtime.RawExtension{Raw: FakePlanServiceInstanceCreateParameterSchemasRaw()[(planNumber-1)%2]},
			},
		},
	}
}

// fakePlanExternalMetaData creates fake plan metadata using "plan-name-<number>" as "displayName"
func fakePlanExternalMetaData(number int) []byte {
	planExternalMetaData1 := make(map[string]string)
	planExternalMetaData1["displayName"] = fmt.Sprintf("plan-name-%d", number)
	planExternalMetaDataRaw1, err := json.Marshal(planExternalMetaData1)
	if err != nil {
		panic(err)
	}
	return planExternalMetaDataRaw1
}

// FakePlanServiceInstanceCreateParameterSchemasRaw creates 2 create parameter schemas for testing purposes
func FakePlanServiceInstanceCreateParameterSchemasRaw() [][]byte {
	planServiceInstanceCreateParameterSchema1 := make(M)
	planServiceInstanceCreateParameterSchema1["required"] = []string{"PLAN_DATABASE_URI", "PLAN_DATABASE_USERNAME", "PLAN_DATABASE_PASSWORD"}
	planServiceInstanceCreateParameterSchema1["properties"] = map[string]M{
		"PLAN_DATABASE_URI": {
			"default": "someuri",
			"type":    "string",
		},
		"PLAN_DATABASE_USERNAME": {
			"default": "name",
			"type":    "string",
		},
		"PLAN_DATABASE_PASSWORD": {
			"type": "string",
		},
		"SOME_OTHER": {
			"default": "other",
			"type":    "string",
		},
	}

	planServiceInstanceCreateParameterSchema2 := make(M)
	planServiceInstanceCreateParameterSchema2["required"] = []string{"PLAN_DATABASE_USERNAME_2", "PLAN_DATABASE_PASSWORD"}
	planServiceInstanceCreateParameterSchema2["properties"] = map[string]M{
		"PLAN_DATABASE_USERNAME_2": {
			"default": "user2",
			"type":    "string",
		},
		"PLAN_DATABASE_PASSWORD": {
			"type": "string",
		},
	}

	planServiceInstanceCreateParameterSchemaRaw1, err := json.Marshal(planServiceInstanceCreateParameterSchema1)
	if err != nil {
		if err != nil {
			panic(err)
		}
	}

	planServiceInstanceCreateParameterSchemaRaw2, err := json.Marshal(planServiceInstanceCreateParameterSchema2)
	if err != nil {
		if err != nil {
			panic(err)
		}
	}

	var data [][]byte
	data = append(data, planServiceInstanceCreateParameterSchemaRaw1)
	data = append(data, planServiceInstanceCreateParameterSchemaRaw2)

	return data
}
