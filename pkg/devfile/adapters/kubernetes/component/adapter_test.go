package component

import (
	"testing"

	"github.com/openshift/odo/pkg/devfile"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	versionsCommon "github.com/openshift/odo/pkg/devfile/versions/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/testingutil"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
)

func TestStart(t *testing.T) {

	testComponentName := "test"

	tests := []struct {
		name          string
		componentType versionsCommon.DevfileComponentType
		wantErr       bool
	}{
		{
			name:          "Case: Invalid devfile",
			componentType: "",
			wantErr:       true,
		},
		{
			name:          "Case: Valid devfile",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfile.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			fkclient, fkclientset := kclient.FakeNew()
			fkWatch := watch.NewFake()

			// Change the status
			go func() {
				podStatus := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: testComponentName,
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				}
				fkWatch.Modify(podStatus)
			}()
			fkclientset.Kubernetes.PrependWatchReactor("pods", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch, nil
			})

			componentAdapter := New(adapterCtx, *fkclient)
			err := componentAdapter.Start()

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter start unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestDoesComponentExist(t *testing.T) {

	fakeComponentName := "fake-component"

	tests := []struct {
		name          string
		componentType versionsCommon.DevfileComponentType
		componentName string
	}{
		{
			name:          "Case: Valid devfile",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			componentName: "test-name",
		},
		{
			name:          "Case: Valid devfile, empty component name",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			componentName: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfile.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: tt.componentName,
				Devfile:       devObj,
			}

			fkclient, fkclientset := kclient.FakeNew()
			fkWatch := watch.NewFake()

			// Change the status
			go func() {
				podStatus := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: tt.componentName,
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				}
				fkWatch.Modify(podStatus)
			}()
			fkclientset.Kubernetes.PrependWatchReactor("pods", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch, nil
			})

			// DoesComponentExist requires an already started component, so start it.
			componentAdapter := New(adapterCtx, *fkclient)
			err := componentAdapter.Start()

			// Checks for unexpected error cases
			if err != nil {
				t.Errorf("component adapter start unexpected error %v", err)
			}

			// Verify that a comopnent with the specified name exists
			componentExists := componentAdapter.DoesComponentExist(tt.componentName)
			if !componentExists {
				t.Errorf("unable to find component with name %s", tt.componentName)
			}

			// Verify that a component with some fake name doesn't exist
			componentExists = componentAdapter.DoesComponentExist(fakeComponentName)
			if componentExists {
				t.Errorf("found non-existent component with name %s", fakeComponentName)
			}

		})
	}

}

func TestGetFirstContainerWithSourceVolume(t *testing.T) {
	tests := []struct {
		name       string
		containers []corev1.Container
		want       string
		wantErr    bool
	}{
		{
			name: "Case: One container, no volumes",
			containers: []corev1.Container{
				{
					Name: "test",
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Case: One container, no source volume",
			containers: []corev1.Container{
				{
					Name: "test",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name: "test",
						},
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Case: One container, source volume",
			containers: []corev1.Container{
				{
					Name: "test",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name: kclient.OdoSourceVolume,
						},
					},
				},
			},
			want:    "test",
			wantErr: false,
		},
		{
			name: "Case: One container, multiple volumes",
			containers: []corev1.Container{
				{
					Name: "test",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name: "test",
						},
						{
							Name: kclient.OdoSourceVolume,
						},
					},
				},
			},
			want:    "test",
			wantErr: false,
		},
		{
			name: "Case: Multiple containers, no source volumes",
			containers: []corev1.Container{
				{
					Name: "test",
				},
				{
					Name: "test",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name: "test",
						},
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Case: Multiple containers, multiple volumes",
			containers: []corev1.Container{
				{
					Name: "test",
				},
				{
					Name: "container-two",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name: "test",
						},
						{
							Name: kclient.OdoSourceVolume,
						},
					},
				},
			},
			want:    "container-two",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		container, err := getFirstContainerWithSourceVolume(tt.containers)
		if container != tt.want {
			t.Errorf("expected %s, actual %s", tt.want, container)
		}

		if !tt.wantErr == (err != nil) {
			t.Errorf("expected %v, actual %v", tt.wantErr, err)
		}
	}
}
