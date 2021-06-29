package testingutil

import (
	"encoding/json"
	"fmt"

	scv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// M is an alias for map[string]interface{}
type M map[string]interface{}

// FakeClusterServiceClass creates a fake service class with the specified name for testing purposes
func FakeClusterServiceClass(name string, tags ...string) scv1beta1.ClusterServiceClass {
	classExternalMetaData := make(map[string]interface{})
	classExternalMetaDataRaw, err := json.Marshal(classExternalMetaData)
	if err != nil {
		panic(err)
	}

	class := scv1beta1.ClusterServiceClass{
		Spec: scv1beta1.ClusterServiceClassSpec{
			CommonServiceClassSpec: scv1beta1.CommonServiceClassSpec{
				ExternalName:     name,
				ExternalMetadata: &runtime.RawExtension{Raw: classExternalMetaDataRaw},
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
func FakeClusterServicePlan(name string, planNumber int) scv1beta1.ClusterServicePlan {
	return scv1beta1.ClusterServicePlan{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: scv1beta1.ClusterServicePlanSpec{
			ClusterServiceClassRef: scv1beta1.ClusterObjectReference{
				Name: "1dda1477cace09730bd8ed7a6505607e",
			},
			CommonServicePlanSpec: scv1beta1.CommonServicePlanSpec{
				ExternalName:                  name,
				Description:                   fmt.Sprintf("this is a example description %d", planNumber),
				ExternalMetadata:              SingleValuedRawExtension("displayName", fmt.Sprintf("plan-name-%d", planNumber)),
				InstanceCreateParameterSchema: &runtime.RawExtension{Raw: FakePlanInstanceCreateParameterSchemasRaw()[(planNumber-1)%2]},
			},
		},
	}
}

// SingleValuedRawExtension creates a simple, single valued (name=value), raw extension
func SingleValuedRawExtension(name, value string) *runtime.RawExtension {
	metadata := make(map[string]string)
	metadata[name] = value
	serialized, err := json.Marshal(metadata)
	if err != nil {
		panic(err)
	}
	return &runtime.RawExtension{Raw: serialized}
}

// FakePlanInstanceCreateParameterSchemasRaw creates 2 create parameter schemas for testing purposes
func FakePlanInstanceCreateParameterSchemasRaw() [][]byte {
	planInstanceCreateParameterSchema1 := make(M)
	planInstanceCreateParameterSchema1["required"] = []string{"PLAN_DATABASE_URI", "PLAN_DATABASE_USERNAME", "PLAN_DATABASE_PASSWORD"}
	planInstanceCreateParameterSchema1["properties"] = map[string]M{
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

	planInstanceCreateParameterSchema2 := make(M)
	planInstanceCreateParameterSchema2["required"] = []string{"PLAN_DATABASE_USERNAME_2", "PLAN_DATABASE_PASSWORD"}
	planInstanceCreateParameterSchema2["properties"] = map[string]M{
		"PLAN_DATABASE_USERNAME_2": {
			"default": "user2",
			"type":    "string",
		},
		"PLAN_DATABASE_PASSWORD": {
			"type": "string",
		},
	}

	planInstanceCreateParameterSchemaRaw1, err := json.Marshal(planInstanceCreateParameterSchema1)
	if err != nil {
		if err != nil {
			panic(err)
		}
	}

	planInstanceCreateParameterSchemaRaw2, err := json.Marshal(planInstanceCreateParameterSchema2)
	if err != nil {
		if err != nil {
			panic(err)
		}
	}

	var data [][]byte
	data = append(data, planInstanceCreateParameterSchemaRaw1)
	data = append(data, planInstanceCreateParameterSchemaRaw2)

	return data
}

// FakeServiceClassInstance creates and returns a simple service class instance for testing purpose
// serviceInstanceName is the name of the service class instance
// serviceClassName is the name of the service class
// planName is the name of the plan
// status is the status of the service instance
func FakeServiceClassInstance(serviceInstanceName string, serviceClassName string, planName string, status string) scv1beta1.ServiceInstance {
	var service = scv1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceInstanceName,
			Labels: map[string]string{
				applabels.ApplicationLabel:         "app",
				componentlabels.ComponentLabel:     serviceInstanceName,
				componentlabels.ComponentTypeLabel: "dh-mariadb-apb",
			},
			Namespace: "myproject",
		},
		Spec: scv1beta1.ServiceInstanceSpec{
			PlanReference: scv1beta1.PlanReference{
				ClusterServiceClassExternalName: serviceClassName,
				ClusterServicePlanExternalName:  planName,
			},
		},
		Status: scv1beta1.ServiceInstanceStatus{
			Conditions: []scv1beta1.ServiceInstanceCondition{
				{
					Reason: status,
				},
			},
		},
	}
	return service
}

func FakeKubeService(componentName, serviceName string) corev1.Service {
	return corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   serviceName,
			Labels: componentlabels.GetLabels(componentName, "app", false),
		},
	}
}

func FakeKubeServices(componentName string) []corev1.Service {
	return []corev1.Service{
		FakeKubeService(componentName, "service-1"),
		FakeKubeService(componentName, "service-2"),
		FakeKubeService(componentName, "service-3"),
	}
}
