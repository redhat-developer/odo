package occlient

import (
	"fmt"
	"sync"
	"testing"
	"time"

	appsv1 "github.com/openshift/api/apps/v1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
)

func fakeDeploymentConfigWithReplicas(name string, image string, requestedReplicas int, availableReplicas int) *appsv1.DeploymentConfig {

	//deploymentConfig := fakeDeploymentConfig(name, image, nil, nil)

	// save component type as label
	labels := componentlabels.GetLabels(name, name, true)
	labels[componentlabels.ComponentTypeLabel] = image
	labels[componentlabels.ComponentTypeVersion] = "latest"
	labels[applabels.ApplicationLabel] = name

	// save source path as annotation
	annotations := map[string]string{"app.openshift.io/vcs-uri": "./",
		"app.kubernetes.io/component-source-type": "local",
	}

	// Create CommonObjectMeta to be passed in
	commonObjectMeta := metav1.ObjectMeta{
		Name:        name,
		Labels:      labels,
		Annotations: annotations,
	}

	commonImageMeta := CommonImageMeta{
		Name:      name,
		Tag:       "latest",
		Namespace: "openshift",
		Ports:     []corev1.ContainerPort{{Name: "foo", HostPort: 80, ContainerPort: 80}},
	}

	// Generate the DeploymentConfig that will be used.
	dc := generateSupervisordDeploymentConfig(
		commonObjectMeta,
		commonImageMeta,
		nil,
		nil,
		fakeResourceRequirements(),
	)

	dc.Status.Replicas = int32(requestedReplicas)
	dc.Status.AvailableReplicas = int32(availableReplicas)

	return &dc
}

func TestWaitForEverything(t *testing.T) {
	mu := sync.Mutex{}
	type args struct {
		requestedReplicas      int
		availableReplicas      int
		podStatus              corev1.PodPhase
		deploymentConfigStatus appsv1.DeploymentConfig
		timeout                time.Duration
	}
	tests := []struct {
		name    string
		podName string
		args    args
		wantErr bool
	}{
		{
			name:    "Case 1 - Successful pod deployment test",
			podName: "foobar",
			args: args{
				requestedReplicas: 1,
				availableReplicas: 1,
				podStatus:         corev1.PodRunning,
				timeout:           3 * time.Second,
			},
			wantErr: false,
		},
		{
			// Case 1.5 because why not, we haven't yet implemented multiple pod / availability, only 1 pod is only launched
			name:    "Case 1.5 - Successful pod deployment test with multiple replicas (even though this feature isn't supported yet)",
			podName: "foobar",
			args: args{
				requestedReplicas: 2,
				availableReplicas: 2,
				podStatus:         corev1.PodRunning,
				timeout:           3 * time.Second,
			},
			wantErr: false,
		},
		{
			name:    "Case 2 - Fail / timeout if the pod is still running and DC is 1/0",
			podName: "foobar",
			args: args{
				requestedReplicas: 1,
				availableReplicas: 0,
				podStatus:         corev1.PodRunning,
				timeout:           1 * time.Second,
			},
			wantErr: true,
		},
		{
			name:    "Case 3 - Fail when a pod is unsuccessful (failed)",
			podName: "foobar",
			args: args{
				requestedReplicas: 1,
				availableReplicas: 0,
				podStatus:         corev1.PodFailed,
				timeout:           3 * time.Second,
			},
			wantErr: true,
		},
		{
			name:    "Case 4 - Fail with a failed pod, even though deploymentconfig is correct",
			podName: "foobar",
			args: args{
				requestedReplicas: 1,
				availableReplicas: 1,
				podStatus:         corev1.PodFailed,
				timeout:           3 * time.Second,
			},
			wantErr: true,
		},
		{
			name:    "Case 5 - Fail when the pod is unknown",
			podName: "foobar",
			args: args{
				requestedReplicas: 1,
				availableReplicas: 0,
				podStatus:         corev1.PodUnknown,
				timeout:           3 * time.Second,
			},
			wantErr: true,
		},
		{
			name:    "Case 6 - Fail / timeout if the pod is running, BUT deployment config is still 1/0 of available replicas",
			podName: "foobar",
			args: args{
				requestedReplicas: 1,
				availableReplicas: 0,
				podStatus:         corev1.PodRunning,
				timeout:           1 * time.Second,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fakeClient, fakeClientSet := FakeNew()
			fakePodWatch := watch.NewFake()
			fakeDeploymentConfigWatch := watch.NewFake()

			// Fake the pod status
			fakePod := fakePodStatus(tt.args.podStatus, tt.podName)
			go func(pod *corev1.Pod) {
				mu.Lock()
				fakePodWatch.Modify(fakePodStatus(tt.args.podStatus, tt.podName))
				mu.Unlock()
			}(fakePod)

			// Fake deployment config
			dc := fakeDeploymentConfigWithReplicas(tt.name, tt.name, tt.args.requestedReplicas, tt.args.availableReplicas)
			go func(dc *appsv1.DeploymentConfig) {
				mu.Lock()
				fakeDeploymentConfigWatch.Modify(dc)
				mu.Unlock()
			}(dc)

			// Add the reactors
			fakeClientSet.Kubernetes.PrependWatchReactor("pods", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fakePodWatch, nil
			})
			fakeClientSet.AppsClientset.PrependWatchReactor("deploymentconfigs", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fakeDeploymentConfigWatch, nil
			})

			// Run function WaitForEverything
			podSelector := fmt.Sprintf("deploymentconfig=%s", tt.podName)
			err := fakeClient.WaitForEverything(podSelector, tt.podName, tt.args.timeout)

			if err == nil && tt.wantErr {
				t.Error("test failed, expected: false, got true")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, expected: no error, got error: %s", err.Error())
			}

		})
	}
}
