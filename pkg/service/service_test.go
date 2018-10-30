package service

import (
	"encoding/json"
	"fmt"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/occlient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
	"reflect"
	"sort"
	"testing"
)

// M is an alias for map[string]interface{}
type M map[string]interface{}

func fakePlanExternalMetaDataRaw() ([][]byte, error) {
	planExternalMetaData1 := make(map[string]string)
	planExternalMetaData1["displayName"] = "plan-name-1"

	planExternalMetaData2 := make(map[string]string)
	planExternalMetaData2["displayName"] = "plan-name-2"

	planExternalMetaDataRaw1, err := json.Marshal(planExternalMetaData1)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	planExternalMetaDataRaw2, err := json.Marshal(planExternalMetaData2)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	var data [][]byte
	data = append(data, planExternalMetaDataRaw1)
	data = append(data, planExternalMetaDataRaw2)

	return data, nil
}

func fakePlanServiceInstanceCreateParameterSchemasRaw() ([][]byte, error) {
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
			return nil, errors.Wrap(err, "")
		}
	}

	planServiceInstanceCreateParameterSchemaRaw2, err := json.Marshal(planServiceInstanceCreateParameterSchema2)
	if err != nil {
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
	}

	var data [][]byte
	data = append(data, planServiceInstanceCreateParameterSchemaRaw1)
	data = append(data, planServiceInstanceCreateParameterSchemaRaw2)

	return data, nil
}

func TestGetServiceClassAndPlans(t *testing.T) {

	classExternalMetaData := make(map[string]interface{})
	classExternalMetaData["longDescription"] = "example long description"
	classExternalMetaData["dependencies"] = []string{"docker.io/centos/7", "docker.io/centos/8"}

	classExternalMetaDataRaw, err := json.Marshal(classExternalMetaData)
	if err != nil {
		fmt.Printf("error occured %v during marshalling", err)
		return
	}

	planExternalMetaDataRaw, err := fakePlanExternalMetaDataRaw()
	if err != nil {
		fmt.Printf("error occured %v during marshalling", err)
		return
	}

	planServiceInstanceCreateParameterSchemasRaw, err := fakePlanServiceInstanceCreateParameterSchemasRaw()
	if err != nil {
		fmt.Printf("error occured %v during marshalling", err)
		return
	}

	type args struct {
		ServiceName string
	}
	tests := []struct {
		name                 string
		args                 args
		returnedClassID      string
		returnedServiceClass *scv1beta1.ClusterServiceClassList
		returnedServicePlan  []scv1beta1.ClusterServicePlan
		wantedServiceClass   ServiceClass
		wantedServicePlans   []ServicePlans
		wantErr              bool
	}{
		{
			name: "test 1 : with correct values",
			args: args{
				ServiceName: "class name",
			},
			returnedClassID: "1dda1477cace09730bd8ed7a6505607e",
			returnedServiceClass: &scv1beta1.ClusterServiceClassList{
				Items: []scv1beta1.ClusterServiceClass{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "1dda1477cace09730bd8ed7a6505607e"},
						Spec: scv1beta1.ClusterServiceClassSpec{
							CommonServiceClassSpec: scv1beta1.CommonServiceClassSpec{
								ExternalName:     "class name",
								Bindable:         false,
								Description:      "example description",
								Tags:             []string{"php", "java"},
								ExternalMetadata: &runtime.RawExtension{Raw: classExternalMetaDataRaw},
							},
							ClusterServiceBrokerName: "broker name",
						},
					},
				},
			},
			returnedServicePlan: []scv1beta1.ClusterServicePlan{
				{
					Spec: scv1beta1.ClusterServicePlanSpec{
						ClusterServiceClassRef: scv1beta1.ClusterObjectReference{
							Name: "1dda1477cace09730bd8ed7a6505607e",
						},
						CommonServicePlanSpec: scv1beta1.CommonServicePlanSpec{
							ExternalName:                         "dev",
							Description:                          "this is a example description 1",
							ExternalMetadata:                     &runtime.RawExtension{Raw: planExternalMetaDataRaw[0]},
							ServiceInstanceCreateParameterSchema: &runtime.RawExtension{Raw: planServiceInstanceCreateParameterSchemasRaw[0]},
						},
					},
				},
				{
					Spec: scv1beta1.ClusterServicePlanSpec{
						ClusterServiceClassRef: scv1beta1.ClusterObjectReference{
							Name: "1dda1477cace09730bd8ed7a6505607e",
						},
						CommonServicePlanSpec: scv1beta1.CommonServicePlanSpec{
							ExternalName:                         "prod",
							Description:                          "this is a example description 2",
							ExternalMetadata:                     &runtime.RawExtension{Raw: planExternalMetaDataRaw[1]},
							ServiceInstanceCreateParameterSchema: &runtime.RawExtension{Raw: planServiceInstanceCreateParameterSchemasRaw[1]},
						},
					},
				},
			},
			wantedServiceClass: ServiceClass{
				Name:              "class name",
				ShortDescription:  "example description",
				LongDescription:   "example long description",
				Tags:              []string{"php", "java"},
				Bindable:          false,
				ServiceBrokerName: "broker name",
				VersionsAvailable: []string{"docker.io/centos/7", "docker.io/centos/8"},
			},
			wantedServicePlans: []ServicePlans{
				{
					Name:        "dev",
					Description: "this is a example description 1",
					DisplayName: "plan-name-1",
					Parameters: []ServicePlanParameter{
						{
							Name:            "PLAN_DATABASE_URI",
							Required:        true,
							DefaultValue:    "someuri",
							HasDefaultValue: true,
							Type:            "string",
						},
						{
							Name:            "PLAN_DATABASE_USERNAME",
							Required:        true,
							DefaultValue:    "name",
							HasDefaultValue: true,
							Type:            "string",
						},
						{
							Name:            "PLAN_DATABASE_PASSWORD",
							Required:        true,
							HasDefaultValue: false,
							Type:            "string",
						},
						{
							Name:            "SOME_OTHER",
							Required:        false,
							DefaultValue:    "other",
							HasDefaultValue: true,
							Type:            "string",
						},
					},
				},
				{
					Name:        "prod",
					Description: "this is a example description 2",
					DisplayName: "plan-name-2",
					Parameters: []ServicePlanParameter{
						{
							Name:            "PLAN_DATABASE_USERNAME_2",
							Required:        true,
							DefaultValue:    "user2",
							HasDefaultValue: true,
							Type:            "string",
						},
						{
							Name:            "PLAN_DATABASE_PASSWORD",
							Required:        true,
							HasDefaultValue: false,
							Type:            "string",
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		client, fakeClientSet := occlient.FakeNew()

		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "clusterserviceclasses", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			if action.(ktesting.ListAction).GetListRestrictions().Fields.String() != fmt.Sprintf("spec.externalName=%v", tt.args.ServiceName) {
				t.Errorf("got a different service name got: %v , expected: %v", action.(ktesting.ListAction).GetListRestrictions().Fields.String(), fmt.Sprintf("spec.externalName=%v", tt.args.ServiceName))
			}
			return true, tt.returnedServiceClass, nil
		})

		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "clusterserviceplans", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			if action.(ktesting.ListAction).GetListRestrictions().Fields.String() != fmt.Sprintf("spec.clusterServiceClassRef.name=%v", tt.returnedClassID) {
				t.Errorf("got a different service name got: %v , expected: %v", action.(ktesting.ListAction).GetListRestrictions().Fields.String(), fmt.Sprintf("spec.clusterServiceClassRef.name=%v", tt.returnedClassID))
			}
			return true, &scv1beta1.ClusterServicePlanList{Items: tt.returnedServicePlan}, nil
		})

		serviceClass, servicePlans, err := GetServiceClassAndPlans(client, tt.args.ServiceName)

		if err == nil && !tt.wantErr {
			if len(fakeClientSet.ServiceCatalogClientSet.Actions()) != 2 {
				t.Errorf("expected 2 actions in GetServiceClassAndPlans got: %v", fakeClientSet.ServiceCatalogClientSet.Actions())
			}

			if !reflect.DeepEqual(tt.wantedServiceClass.Name, serviceClass.Name) {
				t.Errorf("different service class name expected got: %v , expected: %v", serviceClass.Name, tt.wantedServiceClass.Name)
			}

			if !reflect.DeepEqual(tt.wantedServiceClass.Bindable, serviceClass.Bindable) {
				t.Errorf("different service class bindable value expected got: %v , expected: %v", serviceClass.Bindable, tt.wantedServiceClass.Bindable)
			}

			if !reflect.DeepEqual(tt.wantedServiceClass.ShortDescription, serviceClass.ShortDescription) {
				t.Errorf("different short description value expected got: %v , expected: %v", serviceClass.ShortDescription, tt.wantedServiceClass.ShortDescription)
			}

			if !reflect.DeepEqual(tt.wantedServiceClass.LongDescription, serviceClass.LongDescription) {
				t.Errorf("different long description value expected got: %v , expected: %v", serviceClass.LongDescription, tt.wantedServiceClass.LongDescription)
			}

			if !reflect.DeepEqual(tt.wantedServiceClass.ServiceBrokerName, serviceClass.ServiceBrokerName) {
				t.Errorf("different service broker name value expected got: %v , expected: %v", serviceClass.ServiceBrokerName, tt.wantedServiceClass.ServiceBrokerName)
			}

			if !reflect.DeepEqual(tt.wantedServiceClass.Tags, serviceClass.Tags) {
				t.Errorf("different service class tags value expected got: %v , expected: %v", serviceClass.Tags, tt.wantedServiceClass.Tags)
			}

			for _, wantedServicePlan := range tt.wantedServicePlans {

				// make sure that the plans are sorted so we can compare them later
				sort.Slice(wantedServicePlan.Parameters, func(i, j int) bool {
					return wantedServicePlan.Parameters[i].Name < wantedServicePlan.Parameters[j].Name
				})

				found := false
				for _, gotServicePlan := range servicePlans {
					if reflect.DeepEqual(wantedServicePlan.Name, gotServicePlan.Name) {
						found = true
					} else {
						continue
					}

					// make sure that the plans are sorted so we can compare them
					sort.Slice(gotServicePlan.Parameters, func(i, j int) bool {
						return gotServicePlan.Parameters[i].Name < gotServicePlan.Parameters[j].Name
					})

					if !reflect.DeepEqual(wantedServicePlan.Parameters, gotServicePlan.Parameters) {
						t.Errorf("Different plan parameters value. Expected: %v , got: %v", gotServicePlan.Parameters, wantedServicePlan.Parameters)
					}

					if !reflect.DeepEqual(wantedServicePlan.DisplayName, gotServicePlan.DisplayName) {
						t.Errorf("Different plan display name value. Expected: %v , got: %v", gotServicePlan.DisplayName, wantedServicePlan.DisplayName)
					}

					if !reflect.DeepEqual(wantedServicePlan.Description, gotServicePlan.Description) {
						t.Errorf("Different plan description value. Expected: %v , got: %v", gotServicePlan.Description, wantedServicePlan.Description)
					}
				}

				if !found {
					t.Errorf("service plan %v not found", wantedServicePlan.Name)
				}
			}
		} else if err == nil && tt.wantErr {
			t.Error("test failed, expected: false, got true")
		} else if err != nil && !tt.wantErr {
			t.Errorf("test failed, expected: no error, got error: %s", err.Error())
		}
	}
}
