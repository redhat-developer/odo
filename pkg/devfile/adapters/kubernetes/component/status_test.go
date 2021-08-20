package component

import (
	"github.com/openshift/odo/pkg/util"
	"testing"

	"github.com/devfile/library/pkg/devfile/parser/data"

	"github.com/openshift/odo/pkg/envinfo"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/testingutil"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/occlient"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ktesting "k8s.io/client-go/testing"
)

func TestGetDeploymentStatus(t *testing.T) {

	testComponentName := "component"
	testAppName := "app"

	deploymentName, err := util.NamespaceKubernetesObject(testComponentName, testAppName)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	tests := []struct {
		name                  string
		envInfo               envinfo.EnvSpecificInfo
		running               bool
		wantErr               bool
		deployment            v1.Deployment
		replicaSet            v1.ReplicaSetList
		podSet                corev1.PodList
		expectedDeploymentUID string
		expectedReplicaSetUID string
		expectedPodUID        string
	}{
		{
			name:    "Case 1: A single deployment, matching replica, and matching pod",
			envInfo: envinfo.EnvSpecificInfo{},
			running: false,
			wantErr: false,
			deployment: v1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       kclient.DeploymentKind,
					APIVersion: kclient.DeploymentAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: deploymentName,
					UID:  types.UID("deployment-uid"),
					Labels: map[string]string{
						componentlabels.ComponentLabel: testComponentName,
						applabels.ApplicationLabel:     testAppName,
					},
				},
			},
			replicaSet: v1.ReplicaSetList{
				Items: []v1.ReplicaSet{
					{
						ObjectMeta: metav1.ObjectMeta{
							UID: "replica-set-uid",
							OwnerReferences: []metav1.OwnerReference{
								{
									UID: types.UID("deployment-uid"),
								},
							},
						},
						Spec: v1.ReplicaSetSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{},
							},
						},
					},
				},
			},
			podSet: corev1.PodList{
				Items: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							UID: "pod-uid",
							OwnerReferences: []metav1.OwnerReference{
								{
									UID: types.UID("replica-set-uid"),
								},
							},
						},
					},
				},
			},
			expectedDeploymentUID: "deployment-uid",
			expectedReplicaSetUID: "replica-set-uid",
			expectedPodUID:        "pod-uid",
		},
		{
			name:    "Case 2: A single deployment, multiple replicas with different generations, and a single matching pod",
			envInfo: envinfo.EnvSpecificInfo{},
			running: false,
			wantErr: false,
			deployment: v1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       kclient.DeploymentKind,
					APIVersion: kclient.DeploymentAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: deploymentName,
					UID:  types.UID("deployment-uid"),
					Labels: map[string]string{
						componentlabels.ComponentLabel: testComponentName,
						applabels.ApplicationLabel:     testAppName,
					},
				},
			},
			replicaSet: v1.ReplicaSetList{
				Items: []v1.ReplicaSet{
					{
						ObjectMeta: metav1.ObjectMeta{
							UID:        "replica-set-uid1",
							Generation: 1,
							OwnerReferences: []metav1.OwnerReference{
								{
									UID: types.UID("deployment-uid"),
								},
							},
						},
						Spec: v1.ReplicaSetSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							UID:        "replica-set-uid2",
							Generation: 2,
							OwnerReferences: []metav1.OwnerReference{
								{
									UID: types.UID("deployment-uid"),
								},
							},
						},
						Spec: v1.ReplicaSetSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{},
							},
						},
					},
				},
			},
			podSet: corev1.PodList{
				Items: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							UID: "pod-uid",
							OwnerReferences: []metav1.OwnerReference{
								{
									UID: types.UID("replica-set-uid2"),
								},
							},
						},
					},
				},
			},
			expectedDeploymentUID: "deployment-uid",
			expectedReplicaSetUID: "replica-set-uid2",
			expectedPodUID:        "pod-uid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			comp := testingutil.GetFakeContainerComponent(testComponentName)
			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{comp})
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands([]devfilev1.Command{getExecCommand("run", devfilev1.RunCommandGroupKind)})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				AppName:       testAppName,
				Devfile:       devObj,
			}

			fkclient, fkclientset := occlient.FakeNew()

			// Return test case's deployment, when requested
			fkclientset.Kubernetes.PrependReactor("get", "*", func(action ktesting.Action) (bool, runtime.Object, error) {
				if getAction, is := action.(ktesting.GetAction); is && getAction.GetName() == deploymentName {
					return true, &tt.deployment, nil
				}
				return false, nil, nil
			})

			// Return test case's deployment, when requested
			fkclientset.Kubernetes.PrependReactor("patch", "*", func(action ktesting.Action) (bool, runtime.Object, error) {
				if patchAction, is := action.(ktesting.PatchAction); is && patchAction.GetName() == deploymentName {
					return true, &tt.deployment, nil
				}
				return false, nil, nil
			})

			// Return test case's deployment, when requested
			fkclientset.Kubernetes.PrependReactor("apply", "*", func(action ktesting.Action) (bool, runtime.Object, error) {
				if patchAction, is := action.(ktesting.PatchAction); is && patchAction.GetName() == deploymentName {
					return true, &tt.deployment, nil
				}
				return false, nil, nil
			})

			// Return test cases's replicasets, or pods, when requested
			fkclientset.Kubernetes.PrependReactor("list", "*", func(action ktesting.Action) (bool, runtime.Object, error) {
				switch action.GetResource().Resource {
				case "replicasets":
					return true, &tt.replicaSet, nil
				case "pods":
					return true, &tt.podSet, nil
				case "deployments":
					return true, &v1.DeploymentList{Items: []v1.Deployment{tt.deployment}}, nil
				}
				return false, nil, nil
			})

			tt.envInfo.EnvInfo = *envinfo.GetFakeEnvInfo(envinfo.ComponentSettings{
				Name:    testComponentName,
				AppName: testAppName,
			})

			componentAdapter := New(adapterCtx, *fkclient)
			fkclient.Namespace = componentAdapter.Client.Namespace
			err := componentAdapter.createOrUpdateComponent(tt.running, tt.envInfo)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Call the function to test
			result, err := componentAdapter.getDeploymentStatus()
			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Fatalf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			if string(result.DeploymentUID) != tt.expectedDeploymentUID {
				t.Fatalf("could not find expected deployment UID %s %s", string(result.DeploymentUID), tt.expectedDeploymentUID)
			}

			if string(result.ReplicaSetUID) != tt.expectedReplicaSetUID {
				t.Fatalf("could not find expected replica set UID %s %s", string(result.ReplicaSetUID), tt.expectedReplicaSetUID)
			}

			if result.Pods == nil || len(result.Pods) != 1 {
				t.Fatalf("results of this test should match 1 pod")
			}

			if string(result.Pods[0].UID) != tt.expectedPodUID {
				t.Fatalf("pod UID did not match expected pod UID: %s %s", string(result.Pods[0].UID), tt.expectedPodUID)
			}

		})
	}

}
