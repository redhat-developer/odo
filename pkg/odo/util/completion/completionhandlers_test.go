package completion

import (
	"github.com/posener/complete"
	"github.com/redhat-developer/odo/pkg/testingutil"
	"reflect"
	"sort"
	"testing"

	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"

	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	appsv1 "github.com/openshift/api/apps/v1"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestServicePlanCompletionHandler(t *testing.T) {
	serviceClassList := &scv1beta1.ClusterServiceClassList{
		Items: []scv1beta1.ClusterServiceClass{testingutil.FakeClusterServiceClass("class name", "dummy")},
	}
	tests := []struct {
		name                 string
		returnedServiceClass *scv1beta1.ClusterServiceClassList
		returnedServicePlan  []scv1beta1.ClusterServicePlan
		output               []string
		parsedArgs           parsedArgs
	}{
		{
			name: "Case 0: no service name supplied",
			parsedArgs: parsedArgs{
				original: complete.Args{
					Completed: []string{"create"},
				},
			},
			output: []string{},
		},
		{
			name:                 "Case 1: single plan exists",
			returnedServiceClass: serviceClassList,
			returnedServicePlan:  []scv1beta1.ClusterServicePlan{testingutil.FakeClusterServicePlan("default", 1)},
			parsedArgs: parsedArgs{
				original: complete.Args{
					Completed: []string{"create", "class name"},
				},
			},
			output: []string{"default"},
		},
		{
			name:                 "Case 2: multiple plans exist",
			returnedServiceClass: serviceClassList,
			returnedServicePlan: []scv1beta1.ClusterServicePlan{
				testingutil.FakeClusterServicePlan("plan1", 1),
				testingutil.FakeClusterServicePlan("plan2", 2),
			},
			parsedArgs: parsedArgs{
				original: complete.Args{
					Completed: []string{"create", "class name"},
				},
			},
			output: []string{"plan1", "plan2"},
		},
	}

	for _, tt := range tests {
		client, fakeClientSet := occlient.FakeNew()
		context := genericclioptions.NewFakeContext("project", "app", "component", client)

		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "clusterserviceclasses", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, tt.returnedServiceClass, nil
		})

		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "clusterserviceplans", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, &scv1beta1.ClusterServicePlanList{Items: tt.returnedServicePlan}, nil
		})

		completions := ServicePlanCompletionHandler(nil, tt.parsedArgs, context)
		sort.Strings(completions)

		// Sort the output and expected o/p in-order to avoid issues due to order as its not important
		sort.Strings(completions)
		sort.Strings(tt.output)

		if !reflect.DeepEqual(tt.output, completions) {
			t.Errorf("expected output: %#v,got: %#v", tt.output, completions)
		}
	}
}

func TestServiceParameterCompletionHandler(t *testing.T) {
	serviceClassList := &scv1beta1.ClusterServiceClassList{
		Items: []scv1beta1.ClusterServiceClass{testingutil.FakeClusterServiceClass("class name", "dummy")},
	}
	tests := []struct {
		name                 string
		returnedServiceClass *scv1beta1.ClusterServiceClassList
		returnedServicePlan  []scv1beta1.ClusterServicePlan
		output               []string
		parsedArgs           parsedArgs
	}{
		{
			name: "Case 0: no service name supplied",
			parsedArgs: parsedArgs{
				original: complete.Args{
					Completed: []string{"create"},
				},
			},
			output: []string{},
		},
		{
			name:                 "Case 1: no plan supplied and single plan exists",
			returnedServiceClass: serviceClassList,
			returnedServicePlan:  []scv1beta1.ClusterServicePlan{testingutil.FakeClusterServicePlan("default", 1)},
			parsedArgs: parsedArgs{
				original: complete.Args{
					Completed: []string{"create", "class name"},
				},
			},
			output: []string{"PLAN_DATABASE_URI", "PLAN_DATABASE_USERNAME", "PLAN_DATABASE_PASSWORD", "SOME_OTHER"},
		},
		{
			name:                 "Case 2: no plan supplied and multiple plans exists",
			returnedServiceClass: serviceClassList,
			returnedServicePlan: []scv1beta1.ClusterServicePlan{
				testingutil.FakeClusterServicePlan("plan1", 1),
				testingutil.FakeClusterServicePlan("plan2", 2),
			},
			parsedArgs: parsedArgs{
				original: complete.Args{
					Completed: []string{"create", "class name"},
				},
			},
			output: []string{},
		},
		{
			name:                 "Case 3: plan supplied but doesn't match",
			returnedServiceClass: serviceClassList,
			returnedServicePlan:  []scv1beta1.ClusterServicePlan{testingutil.FakeClusterServicePlan("default", 1)},
			parsedArgs: parsedArgs{
				original: complete.Args{
					Completed: []string{"create", "class name"},
				},
				flagValues: map[string]string{"plan": "other"},
			},
			output: []string{},
		},
		{
			name:                 "Case 4: matching plan supplied and no other parameters supplied",
			returnedServiceClass: serviceClassList,
			returnedServicePlan: []scv1beta1.ClusterServicePlan{
				testingutil.FakeClusterServicePlan("plan2", 2),
				testingutil.FakeClusterServicePlan("plan1", 1),
			},
			parsedArgs: parsedArgs{
				original: complete.Args{
					Completed: []string{"create", "class name"},
				},
				flagValues: map[string]string{"plan": "plan1"},
			},
			output: []string{"PLAN_DATABASE_URI", "PLAN_DATABASE_USERNAME", "PLAN_DATABASE_PASSWORD", "SOME_OTHER"},
		},
		{
			name:                 "Case 5: no plan supplied but some other parameters supplied",
			returnedServiceClass: serviceClassList,
			returnedServicePlan:  []scv1beta1.ClusterServicePlan{testingutil.FakeClusterServicePlan("default", 1)},
			parsedArgs: parsedArgs{
				original: complete.Args{
					Completed: []string{"create", "class name"},
				},
				flagValues: map[string]string{"parameters": "[PLAN_DATABASE_USERNAME, SOME_OTHER]"},
			},
			output: []string{"PLAN_DATABASE_URI", "PLAN_DATABASE_PASSWORD"},
		},
		{
			name:                 "Case 6: matching plan supplied but some other parameters supplied",
			returnedServiceClass: serviceClassList,
			returnedServicePlan:  []scv1beta1.ClusterServicePlan{testingutil.FakeClusterServicePlan("default", 1)},
			parsedArgs: parsedArgs{
				original: complete.Args{
					Completed: []string{"create", "class name"},
				},
				flagValues: map[string]string{"plan": "default", "parameters": "[PLAN_DATABASE_USERNAME]"},
			},
			output: []string{"PLAN_DATABASE_URI", "PLAN_DATABASE_PASSWORD", "SOME_OTHER"},
		},
	}

	for _, tt := range tests {
		client, fakeClientSet := occlient.FakeNew()
		context := genericclioptions.NewFakeContext("project", "app", "component", client)

		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "clusterserviceclasses", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, tt.returnedServiceClass, nil
		})

		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "clusterserviceplans", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, &scv1beta1.ClusterServicePlanList{Items: tt.returnedServicePlan}, nil
		})

		completions := ServiceParameterCompletionHandler(nil, tt.parsedArgs, context)
		sort.Strings(completions)

		// Sort the output and expected o/p in-order to avoid issues due to order as its not important
		sort.Strings(completions)
		sort.Strings(tt.output)

		if !reflect.DeepEqual(tt.output, completions) {
			t.Errorf("expected output: %#v,got: %#v", tt.output, completions)
		}
	}
}

func TestLinkCompletionHandler(t *testing.T) {

	tests := []struct {
		name        string
		component   string
		dcList      appsv1.DeploymentConfigList
		serviceList scv1beta1.ServiceInstanceList
		output      []string
	}{
		{
			name:      "Case 1: both components and services are present",
			component: "frontend",
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
									Reason: "Provisioning",
								},
							},
						},
					},
				},
			},
			dcList: appsv1.DeploymentConfigList{
				Items: []appsv1.DeploymentConfig{
					{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "backend",
								componentlabels.ComponentTypeLabel: "java",
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
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "frontend",
								componentlabels.ComponentTypeLabel: "nodejs",
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
				},
			},
			// make sure that the 'component' is not part of the suggestions
			output: []string{"backend", "mysql-persistent", "postgresql-ephemeral"},
		},
	}

	for _, tt := range tests {
		client, fakeClientSet := occlient.FakeNew()
		parsedArgs := parsedArgs{
			commands: make(map[string]bool),
		}
		context := genericclioptions.NewFakeContext("project", "app", tt.component, client)

		//fake the services
		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "serviceinstances", func(action ktesting.Action) (bool, runtime.Object, error) {
			return true, &tt.serviceList, nil
		})

		//fake the dcs
		fakeClientSet.AppsClientset.PrependReactor("list", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
			return true, &tt.dcList, nil
		})

		completions := LinkCompletionHandler(nil, parsedArgs, context)
		sort.Strings(completions)

		if !reflect.DeepEqual(tt.output, completions) {
			t.Errorf("expected output: %#v,got: %#v", tt.output, completions)
		}
	}
}

func TestUnlinkCompletionHandler(t *testing.T) {

	tests := []struct {
		name        string
		component   string
		dcList      appsv1.DeploymentConfigList
		serviceList scv1beta1.ServiceInstanceList
		output      []string
	}{
		{
			name:      "Case 1: both components and services are present",
			component: "frontend",
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
				},
			},
			dcList: appsv1.DeploymentConfigList{
				Items: []appsv1.DeploymentConfig{
					{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "backend",
								componentlabels.ComponentTypeLabel: "java",
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
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "backend2",
								componentlabels.ComponentTypeLabel: "java",
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
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "frontend",
								componentlabels.ComponentTypeLabel: "nodejs",
							},
						},
						Spec: appsv1.DeploymentConfigSpec{
							Template: &corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name: "dummyContainer",
											EnvFrom: []corev1.EnvFromSource{
												{
													SecretRef: &corev1.SecretEnvSource{
														LocalObjectReference: corev1.LocalObjectReference{Name: "postgresql-ephemeral"},
													},
												},
												{
													SecretRef: &corev1.SecretEnvSource{
														LocalObjectReference: corev1.LocalObjectReference{Name: "backend-8080"},
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
			// make sure that the 'component' is not part of the suggestions and that only actually linked components/services show up
			output: []string{"backend", "postgresql-ephemeral"},
		},
	}

	for _, tt := range tests {
		client, fakeClientSet := occlient.FakeNew()
		parsedArgs := parsedArgs{
			commands: make(map[string]bool),
		}
		context := genericclioptions.NewFakeContext("project", "app", tt.component, client)

		//fake the services
		fakeClientSet.ServiceCatalogClientSet.PrependReactor("list", "serviceinstances", func(action ktesting.Action) (bool, runtime.Object, error) {
			return true, &tt.serviceList, nil
		})

		//fake the dcs
		fakeClientSet.AppsClientset.PrependReactor("list", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
			return true, &tt.dcList, nil
		})

		completions := UnlinkCompletionHandler(nil, parsedArgs, context)
		sort.Strings(completions)

		if !reflect.DeepEqual(tt.output, completions) {
			t.Errorf("expected output: %#v,got: %#v", tt.output, completions)
		}
	}
}
