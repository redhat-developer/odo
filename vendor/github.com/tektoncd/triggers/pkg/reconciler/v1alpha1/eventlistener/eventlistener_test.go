/*
Copyright 2019 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package eventlistener

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/tektoncd/pipeline/pkg/system"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	"github.com/tektoncd/triggers/test"
	bldr "github.com/tektoncd/triggers/test/builder"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/rand"
	k8stest "k8s.io/client-go/testing"
	"knative.dev/pkg/apis"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
	fakekubeclient "knative.dev/pkg/client/injection/kube/client/fake"
	"knative.dev/pkg/configmap"
	rtesting "knative.dev/pkg/reconciler/testing"
)

func init() {
	rand.Seed(0)
	generatedResourceName = fmt.Sprintf("%s-cbhtc", eventListenerName)
	eventListener0 = bldr.EventListener(eventListenerName, namespace,
		bldr.EventListenerSpec(
			bldr.EventListenerServiceAccount("sa"),
		),
		bldr.EventListenerStatus(
			bldr.EventListenerConfig(generatedResourceName),
		),
	)
}

var (
	generatedResourceName    string
	ignoreLastTransitionTime = cmpopts.IgnoreTypes(apis.Condition{}.LastTransitionTime.Inner.Time)

	// 0 indicates pre-reconciliation
	eventListener0    *v1alpha1.EventListener
	eventListenerName = "my-eventlistener"
	namespace         = "tekton-pipelines"
	namespaceResource = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	reconcileKey                 = fmt.Sprintf("%s/%s", namespace, eventListenerName)
	updateLabel                  = map[string]string{"update": "true"}
	updatedSa                    = "updatedSa"
	deploymentAvailableCondition = appsv1.DeploymentCondition{
		Type:    appsv1.DeploymentAvailable,
		Status:  corev1.ConditionTrue,
		Message: "Deployment has minimum availability",
		Reason:  "MinimumReplicasAvailable",
	}
	deploymentProgressingCondition = appsv1.DeploymentCondition{
		Type:    appsv1.DeploymentProgressing,
		Status:  corev1.ConditionTrue,
		Message: fmt.Sprintf("ReplicaSet \"%s\" has successfully progressed.", eventListenerName),
		Reason:  "NewReplicaSetAvailable",
	}
	generatedLabels = GenerateResourceLabels(eventListenerName)
)

// getEventListenerTestAssets returns TestAssets that have been seeded with the
// given test.Resources r where r represents the state of the system
func getEventListenerTestAssets(t *testing.T, r test.Resources) (test.Assets, context.CancelFunc) {
	t.Helper()
	ctx, _ := rtesting.SetupFakeContext(t)
	ctx, cancel := context.WithCancel(ctx)
	kubeClient := fakekubeclient.Get(ctx)
	// Fake client reactor chain ignores non handled reactors until v1.40.0
	// Test deployment/service resource should set their defaults
	// See: https://github.com/kubernetes/kubernetes/pull/73601
	kubeClient.PrependReactor("create", "deployments",
		func(action k8stest.Action) (bool, runtime.Object, error) {
			deployment := action.(k8stest.CreateActionImpl).GetObject().(*appsv1.Deployment)
			// Only add conditions when they don't exist
			// Test seeding expected resources "creates", which would create duplicates
			if len(deployment.Status.Conditions) == 0 {
				deployment.Status.Conditions = append(deployment.Status.Conditions, deploymentAvailableCondition)
				deployment.Status.Conditions = append(deployment.Status.Conditions, deploymentProgressingCondition)
			}
			// Pass modified resource and react using the default catch all reactor
			return kubeClient.ReactionChain[len(kubeClient.ReactionChain)-1].React(action)
		})
	clients := test.SeedResources(t, ctx, r)
	cmw := configmap.NewInformedWatcher(clients.Kube, system.GetNamespace())
	return test.Assets{
		Controller: NewController(ctx, cmw),
		Clients:    clients,
	}, cancel
}

func Test_reconcileService(t *testing.T) {
	eventListener1 := eventListener0.DeepCopy()
	eventListener1.Status.SetExistsCondition(v1alpha1.ServiceExists, nil)
	eventListener1.Status.Address = &duckv1alpha1.Addressable{}
	eventListener1.Status.Address.URL = &apis.URL{
		Scheme: "http",
		Host:   listenerHostname(generatedResourceName, namespace, *ElPort),
	}

	eventListener2 := eventListener1.DeepCopy()
	eventListener2.Labels = updateLabel

	service1 := &corev1.Service{
		ObjectMeta: generateObjectMeta(eventListener0),
		Spec: corev1.ServiceSpec{
			Selector: generatedLabels,
			Type:     eventListener1.Spec.ServiceType,
			Ports: []corev1.ServicePort{
				{
					Name:     eventListenerServicePortName,
					Protocol: corev1.ProtocolTCP,
					Port:     int32(*ElPort),
					TargetPort: intstr.IntOrString{
						IntVal: int32(*ElPort),
					},
				},
			},
		},
	}
	service2 := service1.DeepCopy()
	service2.Labels = mergeLabels(generatedLabels, updateLabel)
	service2.Spec.Selector = mergeLabels(generatedLabels, updateLabel)

	service3 := service1.DeepCopy()
	service3.Spec.Ports[0].NodePort = 30000

	tests := []struct {
		name           string
		startResources test.Resources
		endResources   test.Resources
	}{
		{
			name: "create-service",
			startResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener0},
			},
			endResources: test.Resources{
				EventListeners: []*v1alpha1.EventListener{eventListener1},
				Services:       []*corev1.Service{service1},
			},
		},
		{
			name: "eventlistener-label-update",
			startResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener2},
				Services:       []*corev1.Service{service1},
			},
			endResources: test.Resources{
				EventListeners: []*v1alpha1.EventListener{eventListener2},
				Services:       []*corev1.Service{service2},
			},
		},
		{
			name: "service-label-update",
			startResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener1},
				Services:       []*corev1.Service{service2},
			},
			endResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener1},
				Services:       []*corev1.Service{service1},
			},
		},
		{
			name: "service-nodeport-update",
			startResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener1},
				Services:       []*corev1.Service{service3},
			},
			endResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener1},
				Services:       []*corev1.Service{service3},
			},
		},
	}
	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			// Setup
			testAssets, cancel := getEventListenerTestAssets(t, tests[i].startResources)
			defer cancel()

			// Run Reconcile
			err := testAssets.Controller.Reconciler.(*Reconciler).reconcileService(tests[i].startResources.EventListeners[0])
			if err != nil {
				t.Errorf("eventlistener.Reconcile() returned error: %s", err)
				return
			}
			// Grab test resource results
			actualEndResources, err := test.GetResourcesFromClients(testAssets.Clients)
			if err != nil {
				t.Fatal(err)
			}
			// Compare services
			// Semantic equality since VolatileTime will not match using cmp.Diff
			if diff := cmp.Diff(tests[i].endResources.Services, actualEndResources.Services, ignoreLastTransitionTime); diff != "" {
				t.Errorf("eventlistener.Reconcile() equality mismatch. Diff request body: -want +got: %s", diff)
			}

			// Compare EventListener
			// The updates to EventListener are not persisted within reconcileService
			if diff := cmp.Diff(tests[i].endResources.EventListeners[0], tests[i].startResources.EventListeners[0], ignoreLastTransitionTime); diff != "" {
				t.Errorf("eventlistener.Reconcile() equality mismatch. Diff request body: -want +got: %s", diff)
			}
		})
	}
}

func Test_reconcileDeployment(t *testing.T) {
	eventListener1 := eventListener0.DeepCopy()
	eventListener1.Status.SetExistsCondition(v1alpha1.DeploymentExists, nil)
	eventListener1.Status.SetDeploymentConditions([]appsv1.DeploymentCondition{
		deploymentAvailableCondition,
		deploymentProgressingCondition,
	})

	eventListener2 := eventListener1.DeepCopy()
	eventListener2.Labels = updateLabel

	eventListener3 := eventListener1.DeepCopy()
	eventListener3.Status.SetCondition(&apis.Condition{
		Type: apis.ConditionType(appsv1.DeploymentReplicaFailure),
	})

	eventListener4 := eventListener1.DeepCopy()
	eventListener4.Spec.ServiceAccountName = updatedSa

	var replicas int32 = 1
	// deployment1 == initial deployment
	deployment1 := &appsv1.Deployment{
		ObjectMeta: generateObjectMeta(eventListener0),
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: generatedLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: generatedLabels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: eventListener0.Spec.ServiceAccountName,
					Containers: []corev1.Container{
						{
							Name:  "event-listener",
							Image: *elImage,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(*ElPort),
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Args: []string{
								"-el-name", eventListenerName,
								"-el-namespace", namespace,
								"-port", strconv.Itoa(*ElPort),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config-logging",
									MountPath: "/etc/config-logging",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "SYSTEM_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "config-logging",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: eventListenerConfigMapName,
									},
								},
							},
						},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{
				deploymentAvailableCondition,
				deploymentProgressingCondition,
			},
		},
	}

	// deployment 2 == initial deployment + labels from eventListener
	deployment2 := deployment1.DeepCopy()
	deployment2.Labels = mergeLabels(generatedLabels, updateLabel)
	deployment2.Spec.Selector.MatchLabels = mergeLabels(generatedLabels, updateLabel)
	deployment2.Spec.Template.Labels = mergeLabels(generatedLabels, updateLabel)

	// deployment 3 == initial deployment + updated replicas
	deployment3 := deployment1.DeepCopy()
	var updateReplicas int32 = 5
	deployment3.Spec.Replicas = &updateReplicas

	deployment4 := deployment1.DeepCopy()
	deployment4.Spec.Template.Spec.ServiceAccountName = updatedSa

	deploymentMissingVolumes := deployment1.DeepCopy()
	deploymentMissingVolumes.Spec.Template.Spec.Volumes = nil
	deploymentMissingVolumes.Spec.Template.Spec.Containers[0].VolumeMounts = nil

	tests := []struct {
		name           string
		startResources test.Resources
		endResources   test.Resources
	}{
		{
			name: "create-deployment",
			startResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener0},
			},
			endResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener1},
				Deployments:    []*appsv1.Deployment{deployment1},
			},
		},
		{
			name: "eventlistener-label-update",
			startResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener2},
				Deployments:    []*appsv1.Deployment{deployment1},
			},
			endResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener2},
				Deployments:    []*appsv1.Deployment{deployment2},
			},
		},
		{
			name: "deployment-label-update",
			startResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener1},
				Deployments:    []*appsv1.Deployment{deployment2},
			},
			endResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener1},
				Deployments:    []*appsv1.Deployment{deployment1},
			},
		},
		{
			name: "deployment-replica-update",
			startResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener1},
				Deployments:    []*appsv1.Deployment{deployment3},
			},
			endResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener1},
				Deployments:    []*appsv1.Deployment{deployment3},
			},
		},
		{
			name: "eventlistener-replica-failure-status-update",
			startResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener3},
				Deployments:    []*appsv1.Deployment{deployment1},
			},
			endResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener1},
				Deployments:    []*appsv1.Deployment{deployment1},
			},
		},
		{
			name: "eventlistener-serviceaccount-update",
			startResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener4},
				Deployments:    []*appsv1.Deployment{deployment1},
			},
			endResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener4},
				Deployments:    []*appsv1.Deployment{deployment4},
			},
		}, {
			name: "eventlistener-config-volume-mount-update",
			startResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener2},
				Deployments:    []*appsv1.Deployment{deploymentMissingVolumes},
			},
			endResources: test.Resources{
				Namespaces:     []*corev1.Namespace{namespaceResource},
				EventListeners: []*v1alpha1.EventListener{eventListener2},
				Deployments:    []*appsv1.Deployment{deployment2},
			},
		},
	}
	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			// Setup
			testAssets, cancel := getEventListenerTestAssets(t, tests[i].startResources)
			defer cancel()

			// Run Reconcile
			err := testAssets.Controller.Reconciler.(*Reconciler).reconcileDeployment(tests[i].startResources.EventListeners[0])
			if err != nil {
				t.Errorf("eventlistener.Reconcile() returned error: %s", err)
				return
			}
			// Grab test resource results
			actualEndResources, err := test.GetResourcesFromClients(testAssets.Clients)
			if err != nil {
				t.Fatal(err)
			}
			// Compare Deployments
			// Semantic equality since VolatileTime will not match using cmp.Diff
			if !equality.Semantic.DeepEqual(tests[i].endResources.Deployments, actualEndResources.Deployments) {
				t.Error("eventlistener.Reconcile() equality mismatch. Ignore semantic time mismatch")
				diff := cmp.Diff(tests[i].endResources.Deployments, actualEndResources.Deployments)
				t.Errorf("Diff request body: -want +got: %s", diff)
			}
			// Compare EventListener
			// The updates to EventListener are not persisted within reconcileService
			if !equality.Semantic.DeepEqual(tests[i].endResources.EventListeners[0], tests[i].startResources.EventListeners[0]) {
				t.Error("eventlistener.Reconcile() equality mismatch. Ignore semantic time mismatch")
				diff := cmp.Diff(tests[i].endResources.EventListeners[0], tests[i].startResources.EventListeners[0])
				t.Errorf("Diff request body: -want +got: %s", diff)
			}
		})
	}
}

func TestReconcile(t *testing.T) {
	eventListener1 := bldr.EventListener(eventListenerName, namespace,
		bldr.EventListenerSpec(
			bldr.EventListenerServiceAccount("sa"),
		),
		bldr.EventListenerStatus(
			bldr.EventListenerConfig(generatedResourceName),
			bldr.EventListenerAddress(listenerHostname(generatedResourceName, namespace, *ElPort)),
			bldr.EventListenerCondition(
				v1alpha1.ServiceExists,
				corev1.ConditionTrue,
				"Service exists", "",
			),
			bldr.EventListenerCondition(
				v1alpha1.DeploymentExists,
				corev1.ConditionTrue,
				"Deployment exists", "",
			),
			bldr.EventListenerCondition(
				apis.ConditionType(appsv1.DeploymentAvailable),
				corev1.ConditionTrue,
				"Deployment has minimum availability",
				"MinimumReplicasAvailable",
			),
			bldr.EventListenerCondition(
				apis.ConditionType(appsv1.DeploymentProgressing),
				corev1.ConditionTrue,
				fmt.Sprintf("ReplicaSet \"%s\" has successfully progressed.", eventListenerName),
				"NewReplicaSetAvailable",
			),
		),
	)

	eventListener2 := eventListener1.DeepCopy()
	eventListener2.Labels = updateLabel

	eventListener3 := eventListener2.DeepCopy()
	eventListener3.Spec.ServiceAccountName = updatedSa

	eventListener4 := eventListener3.DeepCopy()
	eventListener4.Spec.ServiceType = corev1.ServiceTypeNodePort

	var replicas int32 = 1
	deployment1 := &appsv1.Deployment{
		ObjectMeta: generateObjectMeta(eventListener0),
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: generatedLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: generatedLabels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: eventListener0.Spec.ServiceAccountName,
					Containers: []corev1.Container{{
						Name:  "event-listener",
						Image: *elImage,
						Ports: []corev1.ContainerPort{{
							ContainerPort: int32(*ElPort),
							Protocol:      corev1.ProtocolTCP,
						}},
						Args: []string{
							"-el-name", eventListenerName,
							"-el-namespace", namespace,
							"-port", strconv.Itoa(*ElPort),
						},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "config-logging",
							MountPath: "/etc/config-logging",
						}},
						Env: []corev1.EnvVar{{
							Name: "SYSTEM_NAMESPACE",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.namespace",
								},
							},
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "config-logging",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: eventListenerConfigMapName,
								},
							},
						},
					}},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{
				deploymentAvailableCondition,
				deploymentProgressingCondition,
			},
		},
	}

	deployment2 := deployment1.DeepCopy()
	deployment2.Labels = mergeLabels(updateLabel, generatedLabels)
	deployment2.Spec.Selector.MatchLabels = mergeLabels(updateLabel, generatedLabels)
	deployment2.Spec.Template.Labels = mergeLabels(updateLabel, generatedLabels)

	deployment3 := deployment2.DeepCopy()
	deployment3.Spec.Template.Spec.ServiceAccountName = updatedSa

	service1 := &corev1.Service{
		ObjectMeta: generateObjectMeta(eventListener0),
		Spec: corev1.ServiceSpec{
			Selector: generatedLabels,
			Type:     eventListener1.Spec.ServiceType,
			Ports: []corev1.ServicePort{
				{
					Name:     eventListenerServicePortName,
					Protocol: corev1.ProtocolTCP,
					Port:     int32(*ElPort),
					TargetPort: intstr.IntOrString{
						IntVal: int32(*ElPort),
					},
				},
			},
		},
	}

	service2 := service1.DeepCopy()
	service2.Labels = mergeLabels(updateLabel, generatedLabels)
	service2.Spec.Selector = mergeLabels(updateLabel, generatedLabels)

	service3 := service2.DeepCopy()
	service3.Spec.Type = corev1.ServiceTypeNodePort

	loggingConfigMap := defaultLoggingConfigMap()
	loggingConfigMap.ObjectMeta.Namespace = namespace

	tests := []struct {
		name           string
		key            string
		startResources test.Resources
		endResources   test.Resources
	}{{
		name: "create-eventlistener",
		key:  reconcileKey,
		startResources: test.Resources{
			Namespaces:     []*corev1.Namespace{namespaceResource},
			EventListeners: []*v1alpha1.EventListener{eventListener0},
		},
		endResources: test.Resources{
			Namespaces:     []*corev1.Namespace{namespaceResource},
			EventListeners: []*v1alpha1.EventListener{eventListener1},
			Deployments:    []*appsv1.Deployment{deployment1},
			Services:       []*corev1.Service{service1},
			ConfigMaps:     []*corev1.ConfigMap{loggingConfigMap},
		},
	}, {
		name: "update-eventlistener-labels",
		key:  reconcileKey,
		startResources: test.Resources{
			Namespaces:     []*corev1.Namespace{namespaceResource},
			EventListeners: []*v1alpha1.EventListener{eventListener2},
			Deployments:    []*appsv1.Deployment{deployment1},
			Services:       []*corev1.Service{service1},
		},
		endResources: test.Resources{
			Namespaces:     []*corev1.Namespace{namespaceResource},
			EventListeners: []*v1alpha1.EventListener{eventListener2},
			Deployments:    []*appsv1.Deployment{deployment2},
			Services:       []*corev1.Service{service2},
			ConfigMaps:     []*corev1.ConfigMap{loggingConfigMap},
		},
	}, {
		name: "update-eventlistener-serviceaccount",
		key:  reconcileKey,
		startResources: test.Resources{
			Namespaces:     []*corev1.Namespace{namespaceResource},
			EventListeners: []*v1alpha1.EventListener{eventListener3},
			Deployments:    []*appsv1.Deployment{deployment2},
			Services:       []*corev1.Service{service2},
		},
		endResources: test.Resources{
			Namespaces:     []*corev1.Namespace{namespaceResource},
			EventListeners: []*v1alpha1.EventListener{eventListener3},
			Deployments:    []*appsv1.Deployment{deployment3},
			Services:       []*corev1.Service{service2},
			ConfigMaps:     []*corev1.ConfigMap{loggingConfigMap},
		},
	}, {
		name: "update-eventlistener-servicetype",
		key:  reconcileKey,
		startResources: test.Resources{
			Namespaces:     []*corev1.Namespace{namespaceResource},
			EventListeners: []*v1alpha1.EventListener{eventListener4},
			Deployments:    []*appsv1.Deployment{deployment3},
			Services:       []*corev1.Service{service2},
		},
		endResources: test.Resources{
			Namespaces:     []*corev1.Namespace{namespaceResource},
			EventListeners: []*v1alpha1.EventListener{eventListener4},
			Deployments:    []*appsv1.Deployment{deployment3},
			Services:       []*corev1.Service{service3},
			ConfigMaps:     []*corev1.ConfigMap{loggingConfigMap},
		},
	}, {
		name:           "delete-eventlistener",
		key:            reconcileKey,
		startResources: test.Resources{},
		endResources:   test.Resources{},
	}, {
		name: "delete-last-eventlistener",
		key:  reconcileKey,
		startResources: test.Resources{
			Namespaces: []*corev1.Namespace{namespaceResource},
			ConfigMaps: []*corev1.ConfigMap{loggingConfigMap},
		},
		endResources: test.Resources{
			Namespaces: []*corev1.Namespace{namespaceResource},
		},
	}, {
		name: "delete-eventlistener-with-remaining-eventlistener",
		key:  reconcileKey,
		startResources: test.Resources{
			Namespaces:     []*corev1.Namespace{namespaceResource},
			EventListeners: []*v1alpha1.EventListener{eventListener1},
			ConfigMaps:     []*corev1.ConfigMap{loggingConfigMap},
		},
		endResources: test.Resources{
			Namespaces:     []*corev1.Namespace{namespaceResource},
			EventListeners: []*v1alpha1.EventListener{eventListener1},
			ConfigMaps:     []*corev1.ConfigMap{loggingConfigMap},
			Deployments:    []*appsv1.Deployment{deployment1},
			Services:       []*corev1.Service{service1},
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup with startResources
			testAssets, cancel := getEventListenerTestAssets(t, tt.startResources)
			defer cancel()

			// Run Reconcile
			err := testAssets.Controller.Reconciler.Reconcile(context.Background(), tt.key)
			if err != nil {
				t.Errorf("eventlistener.Reconcile() returned error: %s", err)
				return
			}
			// Grab test resource results
			actualEndResources, err := test.GetResourcesFromClients(testAssets.Clients)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.endResources, *actualEndResources, ignoreLastTransitionTime); diff != "" {
				t.Errorf("eventlistener.Reconcile() equality mismatch. Diff request body: -want +got: %s", diff)
			}
		})
	}
}

func Test_wrapError(t *testing.T) {
	tests := []struct {
		name           string
		error1, error2 error
		expectedError  error
	}{{
		name:          "Both error empty",
		error1:        nil,
		error2:        nil,
		expectedError: nil,
	}, {
		name:          "Error one empty",
		error1:        nil,
		error2:        fmt.Errorf("error"),
		expectedError: fmt.Errorf("error"),
	}, {
		name:          "Error two empty",
		error1:        fmt.Errorf("error"),
		error2:        nil,
		expectedError: fmt.Errorf("error"),
	}, {
		name:          "Both errors",
		error1:        fmt.Errorf("error1"),
		error2:        fmt.Errorf("error2"),
		expectedError: fmt.Errorf("error1 : error2"),
	}}
	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			actualError := wrapError(tests[i].error1, tests[i].error2)
			// Compare strings since errors have unexported fields that panic
			var expectedErrorString string
			var actualErrorString string
			if tests[i].expectedError != nil {
				expectedErrorString = tests[i].expectedError.Error()
			}
			if actualError != nil {
				actualErrorString = actualError.Error()
			}
			if diff := cmp.Diff(expectedErrorString, actualErrorString); diff != "" {
				t.Errorf("wrapError() did not return expected. -want, +got: %s", diff)
			}
		})
	}
}

func Test_mergeLabels(t *testing.T) {
	tests := []struct {
		name           string
		l1, l2         map[string]string
		expectedLabels map[string]string
	}{{
		name:           "Both maps empty",
		l1:             nil,
		l2:             nil,
		expectedLabels: map[string]string{},
	}, {
		name:           "Map one empty",
		l1:             nil,
		l2:             map[string]string{"k": "v"},
		expectedLabels: map[string]string{"k": "v"},
	}, {
		name:           "Map two empty",
		l1:             map[string]string{"k": "v"},
		l2:             nil,
		expectedLabels: map[string]string{"k": "v"},
	}, {
		name:           "Both maps",
		l1:             map[string]string{"k1": "v1"},
		l2:             map[string]string{"k2": "v2"},
		expectedLabels: map[string]string{"k1": "v1", "k2": "v2"},
	}, {
		name:           "Both maps with clobber",
		l1:             map[string]string{"k1": "v1"},
		l2:             map[string]string{"k1": "v2"},
		expectedLabels: map[string]string{"k1": "v2"},
	}}
	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			actualLabels := mergeLabels(tests[i].l1, tests[i].l2)
			if diff := cmp.Diff(tests[i].expectedLabels, actualLabels); diff != "" {
				t.Errorf("mergeLabels() did not return expected. -want, +got: %s", diff)
			}
		})
	}
}

func TestGenerateResourceLabels(t *testing.T) {
	expectedLabels := mergeLabels(StaticResourceLabels, map[string]string{"eventlistener": eventListenerName})
	actualLabels := GenerateResourceLabels(eventListenerName)
	if diff := cmp.Diff(expectedLabels, actualLabels); diff != "" {
		t.Errorf("mergeLabels() did not return expected. -want, +got: %s", diff)
	}
}

func Test_generateObjectMeta(t *testing.T) {
	blockOwnerDeletion := true
	isController := true
	tests := []struct {
		name               string
		el                 *v1alpha1.EventListener
		expectedObjectMeta metav1.ObjectMeta
	}{{
		name: "Empty EventListener",
		el:   bldr.EventListener(eventListenerName, ""),
		expectedObjectMeta: metav1.ObjectMeta{
			Namespace: "",
			Name:      "",
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         "tekton.dev/v1alpha1",
				Kind:               "EventListener",
				Name:               eventListenerName,
				UID:                "",
				Controller:         &isController,
				BlockOwnerDeletion: &blockOwnerDeletion,
			}},
			Labels: generatedLabels,
		},
	}, {
		name: "EventListener with Configuration",
		el: bldr.EventListener(eventListenerName, "",
			bldr.EventListenerStatus(
				bldr.EventListenerConfig("generatedName"),
			),
		),
		expectedObjectMeta: metav1.ObjectMeta{
			Namespace: "",
			Name:      "generatedName",
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         "tekton.dev/v1alpha1",
				Kind:               "EventListener",
				Name:               eventListenerName,
				UID:                "",
				Controller:         &isController,
				BlockOwnerDeletion: &blockOwnerDeletion,
			}},
			Labels: generatedLabels,
		},
	}, {
		name: "EventListener with Labels",
		el: bldr.EventListener(eventListenerName, "",
			bldr.EventListenerMeta(
				bldr.Label("k", "v"),
			),
		),
		expectedObjectMeta: metav1.ObjectMeta{
			Namespace: "",
			Name:      "",
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         "tekton.dev/v1alpha1",
				Kind:               "EventListener",
				Name:               eventListenerName,
				UID:                "",
				Controller:         &isController,
				BlockOwnerDeletion: &blockOwnerDeletion,
			}},
			Labels: mergeLabels(map[string]string{"k": "v"}, generatedLabels),
		},
	}}
	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			actualObjectMeta := generateObjectMeta(tests[i].el)
			if diff := cmp.Diff(tests[i].expectedObjectMeta, actualObjectMeta); diff != "" {
				t.Errorf("generateObjectMeta() did not return expected. -want, +got: %s", diff)
			}
		})
	}
}
