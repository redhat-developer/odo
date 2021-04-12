package kclient

import (
	"encoding/json"
	"fmt"
	"github.com/kylelemons/godebug/pretty"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"strings"
	"testing"

	scv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	ktesting "k8s.io/client-go/testing"
)

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

func fakePlanInstanceCreateParameterSchemasRaw() ([][]byte, error) {
	planInstanceCreateParameterSchema1 := make(map[string][]string)
	planInstanceCreateParameterSchema1["required"] = []string{"PLAN_DATABASE_URI", "PLAN_DATABASE_USERNAME", "PLAN_DATABASE_PASSWORD"}

	planInstanceCreateParameterSchema2 := make(map[string][]string)
	planInstanceCreateParameterSchema2["required"] = []string{"PLAN_DATABASE_USERNAME_2", "PLAN_DATABASE_PASSWORD"}

	planInstanceCreateParameterSchemaRaw1, err := json.Marshal(planInstanceCreateParameterSchema1)
	if err != nil {
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
	}

	planInstanceCreateParameterSchemaRaw2, err := json.Marshal(planInstanceCreateParameterSchema2)
	if err != nil {
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
	}

	var data [][]byte
	data = append(data, planInstanceCreateParameterSchemaRaw1)
	data = append(data, planInstanceCreateParameterSchemaRaw2)

	return data, nil
}

func TestDeleteServiceInstance(t *testing.T) {

	tests := []struct {
		name        string
		serviceName string
		labels      map[string]string
		serviceList scv1beta1.ServiceInstanceList
		wantErr     bool
	}{
		{
			name:        "Delete service instance",
			serviceName: "mongodb",
			labels: map[string]string{
				applabels.ApplicationLabel:     "app",
				componentlabels.ComponentLabel: "mongodb",
			},
			serviceList: scv1beta1.ServiceInstanceList{
				Items: []scv1beta1.ServiceInstance{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "mongodb",
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "mongodb",
								componentlabels.ComponentTypeLabel: "mongodb-persistent",
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()

			//fake the services listing
			fkclientset.ServiceCatalogClientSet.PrependReactor("list", "serviceinstances", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &tt.serviceList, nil
			})

			// Fake the servicebinding delete
			fkclientset.ServiceCatalogClientSet.PrependReactor("delete", "servicebindings", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			// Fake the serviceinstance delete
			fkclientset.ServiceCatalogClientSet.PrependReactor("delete", "serviceinstances", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			err := fkclient.DeleteServiceInstance(tt.labels)
			// Checks for error in positive cases
			if !tt.wantErr && (err != nil) {
				t.Errorf(" client.DeleteServiceInstance(labels) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			// Check for validating actions performed
			// deleting based on the labels means listing the services and then delete the instance and binding for each
			// thus we have 1 list action that always takes place, plus another 2 (delete instance, delete binding)
			// for each service
			expectedNumberOfServiceCatalogActions := 1 + (2 * len(tt.serviceList.Items))
			if len(fkclientset.ServiceCatalogClientSet.Actions()) != expectedNumberOfServiceCatalogActions && !tt.wantErr {
				t.Errorf("expected %d action in CreateServiceInstace got: %v",
					expectedNumberOfServiceCatalogActions, fkclientset.ServiceCatalogClientSet.Actions())
			}

			// Check that the correct service binding was deleted
			DeletedServiceBinding := fkclientset.ServiceCatalogClientSet.Actions()[1].(ktesting.DeleteAction).GetName()
			if DeletedServiceBinding != tt.serviceName {
				t.Errorf("Delete action is performed with wrong ServiceBinding, expected: %s, got %s", tt.serviceName, DeletedServiceBinding)
			}

			// Check that the correct service instance was deleted
			DeletedServiceInstance := fkclientset.ServiceCatalogClientSet.Actions()[2].(ktesting.DeleteAction).GetName()
			if DeletedServiceInstance != tt.serviceName {
				t.Errorf("Delete action is performed with wrong ServiceInstance, expected: %s, got %s", tt.serviceName, DeletedServiceInstance)
			}
		})
	}
}

func TestListServiceInstances(t *testing.T) {

	type args struct {
		Project  string
		Selector string
	}

	tests := []struct {
		name        string
		args        args
		serviceList scv1beta1.ServiceInstanceList
		output      []scv1beta1.ServiceInstance
		wantErr     bool
	}{
		{
			name: "test case 1",
			args: args{
				Project:  "myproject",
				Selector: "app.kubernetes.io/instance=mysql-persistent,app.kubernetes.io/part-of=app",
			},
			serviceList: scv1beta1.ServiceInstanceList{
				Items: []scv1beta1.ServiceInstance{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:       "mysql-persistent",
							Finalizers: []string{"kubernetes-incubator/service-catalog"},
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "mysql-persistent",
								componentlabels.ComponentTypeLabel: "mysql-persistent",
							},
							Namespace: "myproject",
						},
						Spec: scv1beta1.ServiceInstanceSpec{
							PlanReference: scv1beta1.PlanReference{
								ClusterServiceClassExternalName: "mysql-persistent",
								ClusterServicePlanExternalName:  "default",
							},
						},
						Status: scv1beta1.ServiceInstanceStatus{
							Conditions: []scv1beta1.ServiceInstanceCondition{
								{
									Reason: "ProvisionedSuccessfully",
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:       "jenkins-persistent",
							Finalizers: []string{"kubernetes-incubator/service-catalog"},
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "jenkins-persistent",
								componentlabels.ComponentTypeLabel: "jenkins-persistent",
							},
							Namespace: "myproject",
						},
						Spec: scv1beta1.ServiceInstanceSpec{
							PlanReference: scv1beta1.PlanReference{
								ClusterServiceClassExternalName: "jenkins-persistent",
								ClusterServicePlanExternalName:  "default",
							},
						},
						Status: scv1beta1.ServiceInstanceStatus{
							Conditions: []scv1beta1.ServiceInstanceCondition{
								{
									Reason: "ProvisionedSuccessfully",
								},
							},
						},
					},
				},
			},
			output: []scv1beta1.ServiceInstance{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "mysql-persistent",
						Finalizers: []string{"kubernetes-incubator/service-catalog"},
						Labels: map[string]string{
							applabels.ApplicationLabel:         "app",
							componentlabels.ComponentLabel:     "mysql-persistent",
							componentlabels.ComponentTypeLabel: "mysql-persistent",
						},
						Namespace: "myproject",
					},
					Spec: scv1beta1.ServiceInstanceSpec{
						PlanReference: scv1beta1.PlanReference{
							ClusterServiceClassExternalName: "mysql-persistent",
							ClusterServicePlanExternalName:  "default",
						},
					},
					Status: scv1beta1.ServiceInstanceStatus{
						Conditions: []scv1beta1.ServiceInstanceCondition{
							{
								Reason: "ProvisionedSuccessfully",
							},
						},
					},
				},
			},

			wantErr: false,
		},
	}

	for _, tt := range tests {
		client, fakeClientSet := FakeNew()

		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "serviceinstances", func(action ktesting.Action) (bool, runtime.Object, error) {
			if !reflect.DeepEqual(action.(ktesting.ListAction).GetListRestrictions().Labels.String(), tt.args.Selector) {
				return true, nil, fmt.Errorf("labels not matching with expected values, expected:%s, got:%s", tt.args.Selector, action.(ktesting.ListAction).GetListRestrictions())
			}
			return true, &tt.serviceList, nil
		})

		svcInstanceList, err := client.ListServiceInstances(tt.args.Selector)

		if !reflect.DeepEqual(tt.output, svcInstanceList) {
			t.Errorf("expected output: %#v,got: %#v", tt.serviceList, svcInstanceList)
		}

		if err == nil && !tt.wantErr {
			if (len(fakeClientSet.ServiceCatalogClientSet.Actions()) != 1) && (tt.wantErr != true) {
				t.Errorf("expected 1 action in ListServicecatalog got: %v", fakeClientSet.ServiceCatalogClientSet.Actions())
			}
		} else if err == nil && tt.wantErr {
			t.Error("test failed, expected: false, got true")
		} else if err != nil && !tt.wantErr {
			t.Errorf("test failed, expected: no error, got error: %s", err.Error())
		}
	}
}

func TestGetClusterServiceClass(t *testing.T) {
	classExternalMetaData := make(map[string]interface{})
	classExternalMetaData["longDescription"] = "example long description"
	classExternalMetaData["dependencies"] = []string{"docker.io/centos/7", "docker.io/centos/8"}

	classExternalMetaDataRaw, err := json.Marshal(classExternalMetaData)
	if err != nil {
		fmt.Printf("error occured %v during marshalling", err)
		return
	}

	type args struct {
		serviceName string
	}
	tests := []struct {
		name                    string
		args                    args
		returnedServicesClasses *scv1beta1.ClusterServiceClassList
		wantedServiceClass      *scv1beta1.ClusterServiceClass
		wantErr                 bool
	}{
		{
			name: "test case 1: with one valid service class returned",
			args: args{
				serviceName: "class name",
			},
			returnedServicesClasses: &scv1beta1.ClusterServiceClassList{
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
			wantedServiceClass: &scv1beta1.ClusterServiceClass{
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
			wantErr: false,
		},
		{
			name: "test case 2: with two service classes returned",
			args: args{
				serviceName: "class name",
			},
			returnedServicesClasses: &scv1beta1.ClusterServiceClassList{
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
					{
						ObjectMeta: metav1.ObjectMeta{Name: "1dda1477cace09730bd8ed7a6505607e"},
						Spec: scv1beta1.ClusterServiceClassSpec{
							CommonServiceClassSpec: scv1beta1.CommonServiceClassSpec{
								ExternalName:     "class name",
								Bindable:         false,
								Description:      "example description",
								Tags:             []string{"java"},
								ExternalMetadata: &runtime.RawExtension{Raw: classExternalMetaDataRaw},
							},
							ClusterServiceBrokerName: "broker name 1",
						},
					},
				},
			},
			wantedServiceClass: &scv1beta1.ClusterServiceClass{},
			wantErr:            true,
		},
		{
			name: "test case 3: with no service classes returned",
			args: args{
				serviceName: "class name",
			},
			returnedServicesClasses: &scv1beta1.ClusterServiceClassList{
				Items: []scv1beta1.ClusterServiceClass{},
			},
			wantedServiceClass: &scv1beta1.ClusterServiceClass{},
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		client, fakeClientSet := FakeNew()

		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "clusterserviceclasses", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			if action.(ktesting.ListAction).GetListRestrictions().Fields.String() != fmt.Sprintf("spec.externalName=%v", tt.args.serviceName) {
				t.Errorf("got a different service name got: %v , expected: %v", action.(ktesting.ListAction).GetListRestrictions().Fields.String(), fmt.Sprintf("spec.externalName=%v", tt.args.serviceName))
			}
			return true, tt.returnedServicesClasses, nil
		})

		gotServiceClass, err := client.GetClusterServiceClass(tt.args.serviceName)
		if err == nil && !tt.wantErr {
			if len(fakeClientSet.ServiceCatalogClientSet.Actions()) != 1 {
				t.Errorf("expected 1 action in GetServiceClassAndPlans got: %v", fakeClientSet.ServiceCatalogClientSet.Actions())
			}

			if !reflect.DeepEqual(gotServiceClass.Spec, tt.wantedServiceClass.Spec) {
				t.Errorf("different service class spec value expected: %v", pretty.Compare(gotServiceClass.Spec, tt.wantedServiceClass.Spec))
			}
			if !reflect.DeepEqual(gotServiceClass.Name, tt.wantedServiceClass.Name) {
				t.Errorf("different service class name value expected got: %v , expected: %v", gotServiceClass.Name, tt.wantedServiceClass.Name)
			}
		} else if err == nil && tt.wantErr {
			t.Error("test failed, expected: false, got true")
		} else if err != nil && !tt.wantErr {
			t.Errorf("test failed, expected: no error, got error: %s", err.Error())
		}
	}

}

func TestListClusterServicePlansByServiceName(t *testing.T) {
	planExternalMetaDataRaw, err := fakePlanExternalMetaDataRaw()
	if err != nil {
		fmt.Printf("error occured %v during marshalling", err)
		return
	}

	planInstanceCreateParameterSchemasRaw, err := fakePlanInstanceCreateParameterSchemasRaw()
	if err != nil {
		fmt.Printf("error occured %v during marshalling", err)
		return
	}

	type args struct {
		serviceClassName string
	}
	tests := []struct {
		name    string
		args    args
		want    []scv1beta1.ClusterServicePlan
		wantErr bool
	}{
		{
			name:    "test case 1 : plans found for the service class",
			args:    args{serviceClassName: "1dda1477cace09730bd8ed7a6505607e"},
			wantErr: false,
			want: []scv1beta1.ClusterServicePlan{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "67042296c7c95e84142f21f58da2ebfe",
					},
					Spec: scv1beta1.ClusterServicePlanSpec{
						ClusterServiceClassRef: scv1beta1.ClusterObjectReference{
							Name: "1dda1477cace09730bd8ed7a6505607e",
						},
						CommonServicePlanSpec: scv1beta1.CommonServicePlanSpec{
							ExternalName:                  "dev",
							Description:                   "this is a example description 1",
							ExternalMetadata:              &runtime.RawExtension{Raw: planExternalMetaDataRaw[0]},
							InstanceCreateParameterSchema: &runtime.RawExtension{Raw: planInstanceCreateParameterSchemasRaw[0]},
						},
					},
				},

				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "7f88be6129622f72554c20af879a8ce0",
					},
					Spec: scv1beta1.ClusterServicePlanSpec{
						ClusterServiceClassRef: scv1beta1.ClusterObjectReference{
							Name: "1dda1477cace09730bd8ed7a6505607e",
						},
						CommonServicePlanSpec: scv1beta1.CommonServicePlanSpec{
							ExternalName:                  "prod",
							Description:                   "this is a example description 2",
							ExternalMetadata:              &runtime.RawExtension{Raw: planExternalMetaDataRaw[1]},
							InstanceCreateParameterSchema: &runtime.RawExtension{Raw: planInstanceCreateParameterSchemasRaw[1]},
						},
					},
				},
			},
		},
		{
			name:    "test case 2 : no plans found for the service class",
			args:    args{serviceClassName: "1dda1477cace09730bd8"},
			wantErr: false,
			want:    []scv1beta1.ClusterServicePlan{},
		},
	}

	planList := scv1beta1.ClusterServicePlanList{
		Items: []scv1beta1.ClusterServicePlan{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "67042296c7c95e84142f21f58da2ebfe",
				},
				Spec: scv1beta1.ClusterServicePlanSpec{
					ClusterServiceClassRef: scv1beta1.ClusterObjectReference{
						Name: "1dda1477cace09730bd8ed7a6505607e",
					},
					CommonServicePlanSpec: scv1beta1.CommonServicePlanSpec{
						ExternalName:                  "dev",
						Description:                   "this is a example description 1",
						ExternalMetadata:              &runtime.RawExtension{Raw: planExternalMetaDataRaw[0]},
						InstanceCreateParameterSchema: &runtime.RawExtension{Raw: planInstanceCreateParameterSchemasRaw[0]},
					},
				},
			},

			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "7f88be6129622f72554c20af879a8ce0",
				},
				Spec: scv1beta1.ClusterServicePlanSpec{
					ClusterServiceClassRef: scv1beta1.ClusterObjectReference{
						Name: "1dda1477cace09730bd8ed7a6505607e",
					},
					CommonServicePlanSpec: scv1beta1.CommonServicePlanSpec{
						ExternalName:                  "prod",
						Description:                   "this is a example description 2",
						ExternalMetadata:              &runtime.RawExtension{Raw: planExternalMetaDataRaw[1]},
						InstanceCreateParameterSchema: &runtime.RawExtension{Raw: planInstanceCreateParameterSchemasRaw[1]},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, fakeClientSet := FakeNew()

			fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "clusterserviceplans", func(action ktesting.Action) (bool, runtime.Object, error) {
				var pList []scv1beta1.ClusterServicePlan
				for _, plan := range planList.Items {
					if plan.Spec.ClusterServiceClassRef.Name == strings.Split(action.(ktesting.ListAction).GetListRestrictions().Fields.String(), "=")[1] {
						pList = append(pList, plan)
					}
				}

				return true, &scv1beta1.ClusterServicePlanList{Items: pList}, nil
			})

			gotPlans, err := client.ListClusterServicePlansByServiceName(tt.args.serviceClassName)
			if err == nil && !tt.wantErr {
				if len(fakeClientSet.ServiceCatalogClientSet.Actions()) != 1 {
					t.Errorf("expected 2 actions in GetServiceClassAndPlans got: %v", fakeClientSet.ServiceCatalogClientSet.Actions())
				}

				for _, wantedServicePlan := range tt.want {
					found := false
					for _, gotServicePlan := range gotPlans {
						if reflect.DeepEqual(wantedServicePlan.Spec.ExternalName, gotServicePlan.Spec.ExternalName) {
							found = true
						} else {
							continue
						}

						if !reflect.DeepEqual(wantedServicePlan.Name, gotServicePlan.Name) {
							t.Errorf("different plan name expected got: %v , expected: %v", wantedServicePlan.Name, gotServicePlan.Name)
						}

						if !reflect.DeepEqual(wantedServicePlan.Spec, gotServicePlan.Spec) {
							t.Errorf("different plan spec value expected: %v", pretty.Compare(wantedServicePlan.Spec, gotServicePlan.Spec))
						}
					}

					if !found {
						t.Errorf("service plan %v not found", wantedServicePlan.Spec.ExternalName)
					}
				}
			} else if err == nil && tt.wantErr {
				t.Error("test failed, expected: false, got true")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, expected: no error, got error: %s", err.Error())
			}
		})
	}
}

func TestCreateServiceInstance(t *testing.T) {
	type args struct {
		serviceName string
		serviceType string
		labels      map[string]string
		plan        string
		parameters  map[string]string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create service instance",
			args: args{
				serviceName: "jenkins",
				serviceType: "jenkins",
				labels: map[string]string{
					"name":      "mongodb",
					"namespace": "blog",
				},
				plan:       "dev",
				parameters: map[string]string{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := FakeNew()

			_, err := fkclient.CreateServiceInstance(tt.args.serviceName, tt.args.serviceType, tt.args.plan, tt.args.parameters, tt.args.labels)
			// Checks for error in positive cases
			if tt.wantErr == false && (err != nil) {
				t.Errorf(" client.CreateServiceInstance(serviceName,serviceType, labels) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			// Check for validating actions performed
			// creating a service instance also means creating a serviceBinding
			// which is why we expect 2 actions
			if len(fkclientset.ServiceCatalogClientSet.Actions()) != 2 && !tt.wantErr {
				t.Errorf("expected 1 action in CreateServiceInstace got: %v", fkclientset.ServiceCatalogClientSet.Actions())
			}

			createdServiceInstance := fkclientset.ServiceCatalogClientSet.Actions()[0].(ktesting.CreateAction).GetObject().(*scv1beta1.ServiceInstance)
			if !reflect.DeepEqual(createdServiceInstance.Labels, tt.args.labels) {
				t.Errorf("labels in created serviceInstance is not matching expected labels, expected: %v, got: %v", tt.args.labels, createdServiceInstance.Labels)
			}
			if createdServiceInstance.Name != tt.args.serviceName {
				t.Errorf("labels in created serviceInstance is not matching expected labels, expected: %v, got: %v", tt.args.serviceName, createdServiceInstance.Name)
			}
			if !reflect.DeepEqual(createdServiceInstance.Spec.ClusterServiceClassExternalName, tt.args.serviceType) {
				t.Errorf("labels in created serviceInstance is not matching expected labels, expected: %v, got: %v", tt.args.serviceType, createdServiceInstance.Spec.ClusterServiceClassExternalName)
			}
		})
	}
}

func TestGetServiceBinding(t *testing.T) {
	tests := []struct {
		name        string
		namespace   string
		serviceName string
		wantErr     bool
		want        *scv1beta1.ServiceBinding
	}{
		{
			name:        "Case: Valid request for retrieving a service binding",
			namespace:   "",
			serviceName: "foo",
			want: &scv1beta1.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Invalid request for retrieving a service binding",
			namespace:   "",
			serviceName: "foo2",
			want: &scv1beta1.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			// Fake getting Secret
			fakeClientSet.ServiceCatalogClientSet.PrependReactor("get", "servicebindings", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.want.Name != tt.serviceName {
					return true, nil, fmt.Errorf("'get' called with a different serviebinding name")
				}
				return true, tt.want, nil
			})

			returnValue, err := fakeClient.GetServiceBinding(tt.serviceName, tt.namespace)

			// Check for validating return value
			if err == nil && returnValue != tt.want {
				t.Errorf("error in return value got: %v, expected %v", returnValue, tt.want)
			}

			if !tt.wantErr == (err != nil) {
				t.Errorf("\nclient.GetServiceBinding(serviceName, namespace) unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateServiceBinding(t *testing.T) {
	tests := []struct {
		name        string
		bindingNS   string
		bindingName string
		labels      map[string]string
		wantErr     bool
	}{
		{
			name:        "Case: Valid request for creating a secret",
			bindingNS:   "",
			bindingName: "foo",
			labels:      map[string]string{"app": "app"},
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			err := fakeClient.CreateServiceBinding(tt.bindingName, tt.bindingNS, tt.labels)

			if err == nil && !tt.wantErr {
				if len(fakeClientSet.ServiceCatalogClientSet.Actions()) != 1 {
					t.Errorf("expected 1 ServiceCatalogClientSet.Actions() in CreateServiceBinding, got: %v", fakeClientSet.ServiceCatalogClientSet.Actions())
				}
				createdBinding := fakeClientSet.ServiceCatalogClientSet.Actions()[0].(ktesting.CreateAction).GetObject().(*scv1beta1.ServiceBinding)
				if createdBinding.Name != tt.bindingName {
					t.Errorf("the name of servicebinding was not correct, expected: %s, got: %s", tt.bindingName, createdBinding.Name)
				}
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}

		})
	}
}

func TestListServiceClassesByCategory(t *testing.T) {
	t.Run("ListServiceClassesByCategory should work", func(t *testing.T) {
		client, fakeClientSet := FakeNew()
		foo := testingutil.FakeClusterServiceClass("foo", "footag", "footag2")
		bar := testingutil.FakeClusterServiceClass("bar", "")
		boo := testingutil.FakeClusterServiceClass("boo")
		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "clusterserviceclasses", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, &scv1beta1.ClusterServiceClassList{
				Items: []scv1beta1.ClusterServiceClass{
					foo,
					bar,
					boo,
				},
			}, nil
		})

		expected := map[string][]scv1beta1.ClusterServiceClass{"footag": {foo}, "other": {bar, boo}}
		categories, err := client.ListServiceClassesByCategory()

		if err != nil {
			t.Errorf("test failed due to %s", err.Error())
		}

		if !reflect.DeepEqual(expected, categories) {
			t.Errorf("test failed, expected %v, got %v", expected, categories)
		}
	})
}

func TestListMatchingPlans(t *testing.T) {
	t.Run("ListMatchingPlans should work", func(t *testing.T) {
		client, fakeClientSet := FakeNew()
		foo := testingutil.FakeClusterServiceClass("foo", "footag", "footag2")
		dev := testingutil.FakeClusterServicePlan("dev", 1)
		classId := foo.Spec.ExternalID
		dev.Spec.ClusterServiceClassRef.Name = classId
		prod := testingutil.FakeClusterServicePlan("prod", 2)
		prod.Spec.ClusterServiceClassRef.Name = classId

		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "clusterserviceplans", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			value, _ := action.(ktesting.ListAction).GetListRestrictions().Fields.RequiresExactMatch("spec.clusterServiceClassRef.name")
			if value != classId {
				t.Errorf("cluster service plans list should have been filtered on 'spec.clusterServiceClassRef.name==%s'", classId)
			}

			return true, &scv1beta1.ClusterServicePlanList{
				Items: []scv1beta1.ClusterServicePlan{
					dev,
					prod,
				},
			}, nil
		})

		expected := map[string]scv1beta1.ClusterServicePlan{"dev": dev, "prod": prod}
		plans, err := client.ListMatchingPlans(foo)

		if err != nil {
			t.Errorf("test failed due to %s", err.Error())
		}

		if !reflect.DeepEqual(expected, plans) {
			t.Errorf("test failed, expected %v, got %v", expected, plans)
		}
	})
}

func TestListServiceInstanceLabelValues(t *testing.T) {
	type args struct {
		serviceList    scv1beta1.ServiceInstanceList
		expectedOutput []string
		// dcBefore appsv1.DeploymentConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		actions int
	}{
		{
			name: "Case 1 - Retrieve list",
			args: args{
				expectedOutput: []string{"app", "app2"},
				serviceList: scv1beta1.ServiceInstanceList{
					Items: []scv1beta1.ServiceInstance{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:       "mysql-persistent",
								Finalizers: []string{"kubernetes-incubator/service-catalog"},
								Labels: map[string]string{
									applabels.ApplicationLabel:         "app",
									componentlabels.ComponentLabel:     "mysql-persistent",
									componentlabels.ComponentTypeLabel: "mysql-persistent",
								},
								Namespace: "myproject",
							},
							Spec: scv1beta1.ServiceInstanceSpec{
								PlanReference: scv1beta1.PlanReference{
									ClusterServiceClassExternalName: "mysql-persistent",
									ClusterServicePlanExternalName:  "default",
								},
							},
							Status: scv1beta1.ServiceInstanceStatus{
								Conditions: []scv1beta1.ServiceInstanceCondition{
									{
										Reason: "ProvisionedSuccessfully",
									},
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:       "jenkins-persistent",
								Finalizers: []string{"kubernetes-incubator/service-catalog"},
								Labels: map[string]string{
									applabels.ApplicationLabel:         "app2",
									componentlabels.ComponentLabel:     "jenkins-persistent",
									componentlabels.ComponentTypeLabel: "jenkins-persistent",
								},
								Namespace: "myproject",
							},
							Spec: scv1beta1.ServiceInstanceSpec{
								PlanReference: scv1beta1.PlanReference{
									ClusterServiceClassExternalName: "jenkins-persistent",
									ClusterServicePlanExternalName:  "default",
								},
							},
							Status: scv1beta1.ServiceInstanceStatus{
								Conditions: []scv1beta1.ServiceInstanceCondition{
									{
										Reason: "ProvisionedSuccessfully",
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			actions: 1,
		},
		{
			name: "Case 2 - Retrieve list, different order",
			args: args{
				expectedOutput: []string{"app", "app2"},
				serviceList: scv1beta1.ServiceInstanceList{
					Items: []scv1beta1.ServiceInstance{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:       "mysql-persistent",
								Finalizers: []string{"kubernetes-incubator/service-catalog"},
								Labels: map[string]string{
									applabels.ApplicationLabel:         "app2",
									componentlabels.ComponentLabel:     "mysql-persistent",
									componentlabels.ComponentTypeLabel: "mysql-persistent",
								},
								Namespace: "myproject",
							},
							Spec: scv1beta1.ServiceInstanceSpec{
								PlanReference: scv1beta1.PlanReference{
									ClusterServiceClassExternalName: "mysql-persistent",
									ClusterServicePlanExternalName:  "default",
								},
							},
							Status: scv1beta1.ServiceInstanceStatus{
								Conditions: []scv1beta1.ServiceInstanceCondition{
									{
										Reason: "ProvisionedSuccessfully",
									},
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:       "jenkins-persistent",
								Finalizers: []string{"kubernetes-incubator/service-catalog"},
								Labels: map[string]string{
									applabels.ApplicationLabel:         "app",
									componentlabels.ComponentLabel:     "jenkins-persistent",
									componentlabels.ComponentTypeLabel: "jenkins-persistent",
								},
								Namespace: "myproject",
							},
							Spec: scv1beta1.ServiceInstanceSpec{
								PlanReference: scv1beta1.PlanReference{
									ClusterServiceClassExternalName: "jenkins-persistent",
									ClusterServicePlanExternalName:  "default",
								},
							},
							Status: scv1beta1.ServiceInstanceStatus{
								Conditions: []scv1beta1.ServiceInstanceCondition{
									{
										Reason: "ProvisionedSuccessfully",
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			actions: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fakeClient, fakeClientSet := FakeNew()

			fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "serviceinstances", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &tt.args.serviceList, nil
			})

			// Run function ListServiceInstanceLabelValues
			list, err := fakeClient.ListServiceInstanceLabelValues(applabels.ApplicationLabel, applabels.ApplicationLabel)

			if err == nil && !tt.wantErr {

				// Compare arrays
				if !reflect.DeepEqual(list, tt.args.expectedOutput) {
					t.Errorf("expected %s output, got %s", tt.args.expectedOutput, list)
				}

				if (len(fakeClientSet.ServiceCatalogClientSet.Actions()) != tt.actions) && !tt.wantErr {
					t.Errorf("expected %v action(s) in ListServiceInstanceLabelValues got %v: %v", tt.actions, len(fakeClientSet.ServiceCatalogClientSet.Actions()), fakeClientSet.ServiceCatalogClientSet.Actions())
				}

			} else if err == nil && tt.wantErr {
				t.Error("test failed, expected: false, got true")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, expected: no error, got error: %s", err.Error())
			}
		})
	}
}
