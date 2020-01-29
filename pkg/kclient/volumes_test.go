package kclient

import (
	"testing"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktesting "k8s.io/client-go/testing"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCreatePVC(t *testing.T) {

	tests := []struct {
		name      string
		pvcName   string
		size      string
		namespace string
		labels    map[string]string
		wantErr   bool
	}{
		{
			name:      "Case: Valid pvc name",
			pvcName:   "mypvc",
			size:      "1Gi",
			namespace: "default",
			labels: map[string]string{
				"testpvc": "testpvc",
			},
			wantErr: false,
		},
		{
			name:      "Case: Invalid pvc name",
			pvcName:   "",
			size:      "1Gi",
			namespace: "default",
			labels: map[string]string{
				"testpvc": "testpvc",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = tt.namespace

			quantity, err := resource.ParseQuantity(tt.size)
			if err != nil {
				t.Errorf("resource.ParseQuantity unexpected error %v", err)
			}
			pvcSpec := GeneratePVCSpec(quantity)

			objectMeta := CreateObjectMeta(tt.pvcName, tt.namespace, tt.labels, nil)

			fkclientset.Kubernetes.PrependReactor("create", "persistentvolumeclaims", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.pvcName == "" {
					return true, nil, errors.Errorf("pvc name is empty")
				}
				pvc := corev1.PersistentVolumeClaim{
					TypeMeta: metav1.TypeMeta{
						Kind:       PersistentVolumeClaimKind,
						APIVersion: PersistentVolumeClaimAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: tt.pvcName,
					},
				}
				return true, &pvc, nil
			})

			createdPVC, err := fkclient.CreatePVC(objectMeta, *pvcSpec)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.CreatePVC unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 action in StartPVC got: %v", fkclientset.Kubernetes.Actions())
				} else {
					if createdPVC.Name != tt.pvcName {
						t.Errorf("deployment name does not match the expected name, expected: %s, got %s", tt.pvcName, createdPVC.Name)
					}
				}
			}
		})
	}
}

func TestAddPVCToPodTemplateSpec(t *testing.T) {

	container := &corev1.Container{
		Name:            "container1",
		Image:           "image1",
		ImagePullPolicy: corev1.PullAlways,

		Command: []string{"tail"},
		Args:    []string{"-f", "/dev/null"},
		Env:     []corev1.EnvVar{},
	}

	tests := []struct {
		podName        string
		namespace      string
		serviceAccount string
		pvc            string
		volumeName     string
		labels         map[string]string
	}{
		{
			podName:        "podSpecTest",
			namespace:      "default",
			serviceAccount: "default",
			pvc:            "mypvc",
			volumeName:     "myvolume",
			labels: map[string]string{
				"app":       "app",
				"component": "frontend",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.podName, func(t *testing.T) {

			objectMeta := CreateObjectMeta(tt.podName, tt.namespace, tt.labels, nil)

			podTemplateSpec := GeneratePodTemplateSpec(objectMeta, tt.serviceAccount, []corev1.Container{*container})

			AddPVCToPodTemplateSpec(podTemplateSpec, tt.pvc, tt.volumeName)

			pvcMatched := false
			for _, volume := range podTemplateSpec.Spec.Volumes {
				if volume.Name == tt.volumeName && volume.VolumeSource.PersistentVolumeClaim != nil && volume.VolumeSource.PersistentVolumeClaim.ClaimName == tt.pvc {
					pvcMatched = true
				}
			}

			if !pvcMatched {
				t.Errorf("Volume does not exist with Volume Name %s and PVC claim name %s", tt.volumeName, tt.pvc)
			}

		})
	}
}

func TestAddVolumeMountToPodTemplateSpec(t *testing.T) {

	container := &corev1.Container{
		Name:            "container1",
		Image:           "image1",
		ImagePullPolicy: corev1.PullAlways,

		Command: []string{"tail"},
		Args:    []string{"-f", "/dev/null"},
		Env:     []corev1.EnvVar{},
	}

	tests := []struct {
		podName                string
		namespace              string
		serviceAccount         string
		pvc                    string
		volumeName             string
		containerMountPathsMap map[string][]string
		labels                 map[string]string
	}{
		{
			podName:        "podSpecTest",
			namespace:      "default",
			serviceAccount: "default",
			pvc:            "mypvc",
			volumeName:     "myvolume",
			containerMountPathsMap: map[string][]string{
				container.Name: {"/tmp/path1", "/tmp/path2"},
			},
			labels: map[string]string{
				"app":       "app",
				"component": "frontend",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.podName, func(t *testing.T) {

			objectMeta := CreateObjectMeta(tt.podName, tt.namespace, tt.labels, nil)

			podTemplateSpec := GeneratePodTemplateSpec(objectMeta, tt.serviceAccount, []corev1.Container{*container})

			AddVolumeMountToPodTemplateSpec(podTemplateSpec, tt.volumeName, tt.pvc, tt.containerMountPathsMap)
			t.Logf("podTemplateSpec is %v", podTemplateSpec)
			mountPathCount := 0
			for _, podTempSpecContainer := range podTemplateSpec.Spec.Containers {
				if podTempSpecContainer.Name == container.Name {
					for _, volumeMount := range podTempSpecContainer.VolumeMounts {
						if volumeMount.Name == tt.volumeName {
							for _, mountPath := range tt.containerMountPathsMap[container.Name] {
								if volumeMount.MountPath == mountPath {
									mountPathCount++
								}
							}
						}
					}
				}
			}

			if mountPathCount != len(tt.containerMountPathsMap[container.Name]) {
				t.Errorf("Volume Mounts for %s have not been properly mounted to the podTemplateSpec", tt.volumeName)
			}
		})
	}
}
