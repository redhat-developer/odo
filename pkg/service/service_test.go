package service

import (
	"encoding/json"
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"

	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data"
	devfileFileSystem "github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/kylelemons/godebug/pretty"
	"github.com/onsi/gomega/matchers"
	"github.com/openshift/odo/pkg/testingutil"

	"reflect"
	"sort"
	"testing"

	scv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	appsv1 "github.com/openshift/api/apps/v1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/occlient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestGetServiceClassAndPlans(t *testing.T) {

	classExternalMetaData := make(map[string]interface{})
	classExternalMetaData["longDescription"] = "example long description"
	classExternalMetaData["dependencies"] = []string{"docker.io/centos/7", "docker.io/centos/8"}

	classExternalMetaDataRaw, err := json.Marshal(classExternalMetaData)
	if err != nil {
		fmt.Printf("error occured %v during marshalling", err)
		return
	}

	type args struct {
		ServiceName string
	}
	plan1 := testingutil.FakeClusterServicePlan("dev", 1)
	plan2 := testingutil.FakeClusterServicePlan("prod", 2)
	tests := []struct {
		name                 string
		args                 args
		returnedClassID      string
		returnedServiceClass *scv1beta1.ClusterServiceClassList
		returnedServicePlan  []scv1beta1.ClusterServicePlan
		wantedServiceClass   ServiceClass
		wantedServicePlans   []ServicePlan
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
			returnedServicePlan: []scv1beta1.ClusterServicePlan{plan1, plan2},
			wantedServiceClass: ServiceClass{
				Name:              "class name",
				ShortDescription:  "example description",
				LongDescription:   "example long description",
				Tags:              []string{"php", "java"},
				Bindable:          false,
				ServiceBrokerName: "broker name",
				VersionsAvailable: []string{"docker.io/centos/7", "docker.io/centos/8"},
			},
			wantedServicePlans: []ServicePlan{
				{
					Name:        "dev",
					Description: "this is a example description 1",
					DisplayName: "plan-name-1",
					Parameters: []ServicePlanParameter{
						NewServicePlanParameter("PLAN_DATABASE_URI", "string", "someuri", true),
						NewServicePlanParameter("PLAN_DATABASE_USERNAME", "string", "name", true),
						NewServicePlanParameter("PLAN_DATABASE_PASSWORD", "string", "", true),
						NewServicePlanParameter("SOME_OTHER", "string", "other", false),
					},
				},
				{
					Name:        "prod",
					Description: "this is a example description 2",
					DisplayName: "plan-name-2",
					Parameters: []ServicePlanParameter{
						NewServicePlanParameter("PLAN_DATABASE_USERNAME_2", "string", "user2", true),
						NewServicePlanParameter("PLAN_DATABASE_PASSWORD", "string", "", true),
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
						t.Errorf("Different plan parameters value. Expected: %v , got: %v", wantedServicePlan.Parameters, gotServicePlan.Parameters)
					}

					if !reflect.DeepEqual(wantedServicePlan.DisplayName, gotServicePlan.DisplayName) {
						t.Errorf("Different plan display name value. Expected: %v , got: %v", wantedServicePlan.DisplayName, gotServicePlan.DisplayName)
					}

					if !reflect.DeepEqual(wantedServicePlan.Description, gotServicePlan.Description) {
						t.Errorf("Different plan description value. Expected: %v , got: %v", wantedServicePlan.Description, gotServicePlan.Description)
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

func TestListWithDetailedStatus(t *testing.T) {

	type args struct {
		Project  string
		Selector string
	}

	tests := []struct {
		name        string
		args        args
		serviceList scv1beta1.ServiceInstanceList
		secretList  corev1.SecretList
		dcList      appsv1.DeploymentConfigList
		output      []Service
	}{
		{
			name: "Case 1: services with various statuses, some bound and some linked",
			args: args{
				Project:  "myproject",
				Selector: "app.kubernetes.io/instance=mysql-persistent,app.kubernetes.io/part-of=app",
			},
			serviceList: scv1beta1.ServiceInstanceList{
				Items: []scv1beta1.ServiceInstance{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "mysql-persistent",
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
							Name: "postgresql-ephemeral",
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "postgresql-ephemeral",
								componentlabels.ComponentTypeLabel: "postgresql-ephemeral",
							},
							Namespace: "myproject",
						},
						Spec: scv1beta1.ServiceInstanceSpec{
							PlanReference: scv1beta1.PlanReference{
								ClusterServiceClassExternalName: "postgresql-ephemeral",
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
							Name: "mongodb",
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "mongodb",
								componentlabels.ComponentTypeLabel: "mongodb",
							},
							Namespace: "myproject",
						},
						Spec: scv1beta1.ServiceInstanceSpec{
							PlanReference: scv1beta1.PlanReference{
								ClusterServiceClassExternalName: "mongodb",
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
							Name: "jenkins-persistent",
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
									Reason: "Provisioning",
								},
							},
						},
					},
				},
			},
			secretList: corev1.SecretList{
				Items: []corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "dummySecret",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "postgresql-ephemeral",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "mysql-persistent",
						},
					},
				},
			},
			dcList: appsv1.DeploymentConfigList{
				Items: []appsv1.DeploymentConfig{
					{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								applabels.ApplicationLabel: "app",
							},
						},
						Spec: appsv1.DeploymentConfigSpec{
							Template: &corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name: "dummyContainer",
										},
									},
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								applabels.ApplicationLabel: "app",
							},
						},
						Spec: appsv1.DeploymentConfigSpec{
							Template: &corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											EnvFrom: []corev1.EnvFromSource{
												{
													SecretRef: &corev1.SecretEnvSource{
														LocalObjectReference: corev1.LocalObjectReference{
															Name: "mysql-persistent",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			output: []Service{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "odo.dev/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "mysql-persistent",
					},
					Spec: ServiceSpec{
						Type: "mysql-persistent",
						Plan: "default",
					},
					Status: ServiceStatus{
						Status: "ProvisionedAndLinked",
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "odo.dev/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "postgresql-ephemeral",
					},
					Spec: ServiceSpec{
						Type: "postgresql-ephemeral",
						Plan: "default",
					},
					Status: ServiceStatus{
						Status: "ProvisionedAndBound",
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "odo.dev/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "mongodb",
					},
					Spec: ServiceSpec{
						Type: "mongodb",
						Plan: "default",
					},
					Status: ServiceStatus{
						Status: "ProvisionedSuccessfully",
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "odo.dev/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "jenkins-persistent",
					},
					Spec: ServiceSpec{
						Type: "jenkins-persistent",
						Plan: "default",
					},
					Status: ServiceStatus{
						Status: "Provisioning",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		client, fakeClientSet := occlient.FakeNew()

		//fake the services
		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "serviceinstances", func(action ktesting.Action) (bool, runtime.Object, error) {
			return true, &tt.serviceList, nil
		})

		//fake the secrets
		fakeClientSet.Kubernetes.PrependReactor("list", "secrets", func(action ktesting.Action) (bool, runtime.Object, error) {
			return true, &tt.secretList, nil
		})

		//fake the dcs
		fakeClientSet.AppsClientset.PrependReactor("list", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
			return true, &tt.dcList, nil
		})

		svcInstanceList, _ := ListWithDetailedStatus(client, "app")

		if !reflect.DeepEqual(tt.output, svcInstanceList.Items) {
			t.Error(fmt.Sprintf("Expected output: %v", pretty.Compare(tt.serviceList, svcInstanceList.Items)))
		}
	}
}

func TestDeleteServiceAndUnlinkComponents(t *testing.T) {
	const appName = "app"
	type args struct {
		ServiceName string
	}
	tests := []struct {
		name                       string
		args                       args
		serviceList                scv1beta1.ServiceInstanceList
		dcList                     appsv1.DeploymentConfigList
		expectedDCNamesToBeUpdated []string
		wantErr                    bool
	}{
		{
			name: "Case 1: Delete service that has linked component",
			args: args{
				ServiceName: "mysql",
			},
			wantErr:                    false,
			expectedDCNamesToBeUpdated: []string{"component-with-matching-link"},
			serviceList: scv1beta1.ServiceInstanceList{
				Items: []scv1beta1.ServiceInstance{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "mysql",
							Labels: map[string]string{
								applabels.ApplicationLabel:         appName,
								componentlabels.ComponentLabel:     "mysql",
								componentlabels.ComponentTypeLabel: "mysql-persistent",
							},
						},
					},
				},
			},
			dcList: appsv1.DeploymentConfigList{
				Items: []appsv1.DeploymentConfig{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "component-with-no-links" + "-" + appName,
							Labels: map[string]string{
								applabels.ApplicationLabel:     appName,
								componentlabels.ComponentLabel: "component-with-no-links",
							},
						},
						Spec: appsv1.DeploymentConfigSpec{
							Template: &corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name: "dummyContainer",
										},
									},
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "component-with-matching-link" + "-" + appName,
							Labels: map[string]string{
								applabels.ApplicationLabel:     appName,
								componentlabels.ComponentLabel: "component-with-matching-link",
							},
						},
						Spec: appsv1.DeploymentConfigSpec{
							Template: &corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											EnvFrom: []corev1.EnvFromSource{
												{
													SecretRef: &corev1.SecretEnvSource{
														LocalObjectReference: corev1.LocalObjectReference{
															Name: "mysql",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "component-with-non-matching-link" + "-" + appName,
							Labels: map[string]string{
								applabels.ApplicationLabel:     appName,
								componentlabels.ComponentLabel: "component-with-non-matching-link",
							},
						},
						Spec: appsv1.DeploymentConfigSpec{
							Template: &corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											EnvFrom: []corev1.EnvFromSource{
												{
													SecretRef: &corev1.SecretEnvSource{
														LocalObjectReference: corev1.LocalObjectReference{
															Name: "other",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		client, fakeClientSet := occlient.FakeNew()

		//fake the services listing
		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "serviceinstances", func(action ktesting.Action) (bool, runtime.Object, error) {
			return true, &tt.serviceList, nil
		})

		// Fake the servicebinding delete
		fakeClientSet.ServiceCatalogClientSet.PrependReactor("delete", "servicebindings", func(action ktesting.Action) (bool, runtime.Object, error) {
			return true, nil, nil
		})

		// Fake the serviceinstance delete
		fakeClientSet.ServiceCatalogClientSet.PrependReactor("delete", "serviceinstances", func(action ktesting.Action) (bool, runtime.Object, error) {
			return true, nil, nil
		})

		//fake the dc listing
		fakeClientSet.AppsClientset.PrependReactor("list", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
			return true, &tt.dcList, nil
		})

		//fake the dc get
		fakeClientSet.AppsClientset.PrependReactor("get", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
			dcNameToFind := action.(ktesting.GetAction).GetName()
			var matchingDC appsv1.DeploymentConfig
			found := false
			for _, dc := range tt.dcList.Items {
				if dc.Name == dcNameToFind {
					matchingDC = dc
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected to find DeploymentConfig named %s in the dcList", dcNameToFind)
			}

			return true, &matchingDC, nil
		})

		err := DeleteServiceAndUnlinkComponents(client, tt.args.ServiceName, "app")

		if !tt.wantErr == (err != nil) {
			t.Errorf("service.DeleteServiceAndUnlinkComponents(...) unexpected error %v, wantErr %v", err, tt.wantErr)
		}

		// ensure we deleted the service
		if len(fakeClientSet.ServiceCatalogClientSet.Actions()) != 3 && !tt.wantErr {
			t.Errorf("service was deleted properly, got actions: %v", fakeClientSet.ServiceCatalogClientSet.Actions())
		}

		// ensure we updated the correct number of deployments
		// there should always be a list action
		// then each update to a dc is 2 actions, a get and an update
		expectedNumberOfDCActions := 1 + (2 * len(tt.expectedDCNamesToBeUpdated))
		if len(fakeClientSet.AppsClientset.Actions()) != 3 && !tt.wantErr {
			t.Errorf("expected to see %d actions, got: %v", expectedNumberOfDCActions, fakeClientSet.AppsClientset.Actions())
		}
	}
}

func TestServicePlanParameterUnmarshalling(t *testing.T) {
	parameter := NewServicePlanParameter("name", "string", "default", true)
	parameter.Title = "title"
	parameter.Description = "description"

	tests := []struct {
		name     string
		json     string
		expected ServicePlanParameter
	}{
		{
			name: "full",
			json: `{
"name": "name",
"title": "title",
"description": "description",
"default": "default",
"required": true,
"type": "string"
}`,
			expected: parameter,
		},
		{
			name: "not required",
			json: `{
"name": "name",
"default": "default",
"type": "integer"
}`,
			expected: NewServicePlanParameter("name", "integer", "default", false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spp := &ServicePlanParameter{}
			err := json.Unmarshal([]byte(tt.json), &spp)
			if err != nil {
				t.Errorf("unmarshalling failed: %v", err)
			}
			if !reflect.DeepEqual(tt.expected, *spp) {
				t.Errorf("param: %v, got: %v", tt.expected, *spp)
			}
		})
	}
}

func TestServicePlanParameterMarshalling(t *testing.T) {
	parameter := NewServicePlanParameter("name", "string", "default", true)
	parameter.Title = "title"
	parameter.Description = "description"

	tests := []struct {
		name  string
		json  string
		param ServicePlanParameter
	}{
		{
			name: "full",
			json: `{
"name": "name",
"title": "title",
"description": "description",
"default": "default",
"required": true,
"type": "string"
}`,
			param: parameter,
		},
		{
			name: "not required",
			json: `{
"name": "name",
"default": "default",
"type": "integer"
}`,
			param: NewServicePlanParameter("name", "integer", "default", false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := json.Marshal(tt.param)
			if err != nil {
				t.Errorf("marshalling failed: %v", err)
			}
			s := string(actual)
			matcher := matchers.MatchJSONMatcher{JSONToMatch: tt.json}
			success, err := matcher.Match(s)
			if err != nil {
				t.Errorf("couldn't match json: %v", err)
			}
			if !success {
				t.Errorf("param: %v, got: %v", tt.json, s)
			}
		})
	}
}

func TestAddKubernetesComponentToDevfile(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	type args struct {
		crd        string
		name       string
		devfileObj parser.DevfileObj
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []v1alpha2.Component
	}{
		{
			name: "Case 1: Add service CRD to devfile.yaml",
			args: args{
				crd:  "test CRD",
				name: "testName",
				devfileObj: parser.DevfileObj{
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
						if err != nil {
							t.Error(err)
						}
						return devfileData
					}(),
					Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				},
			},
			wantErr: false,
			want: []v1alpha2.Component{{
				Name: "testName",
				ComponentUnion: devfile.ComponentUnion{
					Kubernetes: &devfile.KubernetesComponent{
						K8sLikeComponent: devfile.K8sLikeComponent{
							BaseComponent: devfile.BaseComponent{},
							K8sLikeComponentLocation: devfile.K8sLikeComponentLocation{
								Inlined: "test CRD",
							},
						},
					},
				},
			},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AddKubernetesComponentToDevfile(tt.args.crd, tt.args.name, tt.args.devfileObj); (err != nil) != tt.wantErr {
				t.Errorf("AddKubernetesComponentToDevfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := tt.args.devfileObj.Data.GetComponents(common.DevfileOptions{})
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteKubernetesComponentFromDevfile(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	type args struct {
		name       string
		devfileObj parser.DevfileObj
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []v1alpha2.Component
	}{
		{
			name: "Case 1: Remove a CRD from devfile.yaml",
			args: args{
				name: "testName",
				devfileObj: parser.DevfileObj{
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
						if err != nil {
							t.Error(err)
						}
						err = devfileData.AddComponents([]v1alpha2.Component{{
							Name: "testName",
							ComponentUnion: devfile.ComponentUnion{
								Kubernetes: &devfile.KubernetesComponent{
									K8sLikeComponent: devfile.K8sLikeComponent{
										BaseComponent: devfile.BaseComponent{},
										K8sLikeComponentLocation: devfile.K8sLikeComponentLocation{
											Inlined: "test CRD",
										},
									},
								},
							},
						},
						})
						if err != nil {
							t.Error(err)
						}
						return devfileData
					}(),
					Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				},
			},
			wantErr: false,
			want:    []v1alpha2.Component{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteKubernetesComponentFromDevfile(tt.args.name, tt.args.devfileObj); (err != nil) != tt.wantErr {
				t.Errorf("DeleteKubernetesComponentFromDevfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := tt.args.devfileObj.Data.GetComponents(common.DevfileOptions{})
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}
