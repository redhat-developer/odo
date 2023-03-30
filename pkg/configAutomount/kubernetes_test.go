package configAutomount

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/redhat-developer/odo/pkg/kclient"
	corev1 "k8s.io/api/core/v1"
)

func TestKubernetesClient_GetAutomountingVolumes(t *testing.T) {

	defaultPVC1 := corev1.PersistentVolumeClaim{}
	defaultPVC1.SetName("defaultPVC1")
	defaultPVC1.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})

	defaultPVC2 := corev1.PersistentVolumeClaim{}
	defaultPVC2.SetName("defaultPVC2")
	defaultPVC2.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})

	pvcMountPath := corev1.PersistentVolumeClaim{}
	pvcMountPath.SetName("pvcMountPath")
	pvcMountPath.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})
	pvcMountPath.SetAnnotations(map[string]string{
		annotationMountPathName: "/specific/pvc/mount/path",
	})

	defaultSecret1 := corev1.Secret{}
	defaultSecret1.SetName("defaultSecret1")
	defaultSecret1.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})

	defaultSecret2 := corev1.Secret{}
	defaultSecret2.SetName("defaultSecret2")
	defaultSecret2.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})

	secretMountPath := corev1.Secret{}
	secretMountPath.SetName("secretMountPath")
	secretMountPath.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})
	secretMountPath.SetAnnotations(map[string]string{
		annotationMountPathName: "/specific/secret/mount/path",
	})

	defaultCM1 := corev1.ConfigMap{}
	defaultCM1.SetName("defaultCM1")
	defaultCM1.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})

	defaultCM2 := corev1.ConfigMap{}
	defaultCM2.SetName("defaultCM2")
	defaultCM2.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})

	cmMountPath := corev1.ConfigMap{}
	cmMountPath.SetName("cmMountPath")
	cmMountPath.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})
	cmMountPath.SetAnnotations(map[string]string{
		annotationMountPathName: "/specific/configmap/mount/path",
	})

	type fields struct {
		kubeClient func(ctrl *gomock.Controller) kclient.ClientInterface
	}
	tests := []struct {
		name    string
		fields  fields
		want    []AutomountInfo
		wantErr bool
	}{
		{
			name: "Single default PVC",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().ListPVCs(gomock.Any()).Return([]corev1.PersistentVolumeClaim{defaultPVC1}, nil).AnyTimes()
					client.EXPECT().ListSecrets(gomock.Any()).Return([]corev1.Secret{}, nil).AnyTimes()
					client.EXPECT().ListConfigMaps(gomock.Any()).Return([]corev1.ConfigMap{}, nil).AnyTimes()
					return client
				},
			},
			want: []AutomountInfo{
				{
					VolumeType: VolumeTypePVC,
					VolumeName: "defaultPVC1",
					MountPath:  "/tmp/defaultPVC1",
					MountAs:    MountAsFile,
				},
			},
			wantErr: false,
		},
		{
			name: "Two default PVCs",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().ListPVCs(gomock.Any()).Return([]corev1.PersistentVolumeClaim{defaultPVC1, defaultPVC2}, nil).AnyTimes()
					client.EXPECT().ListSecrets(gomock.Any()).Return([]corev1.Secret{}, nil).AnyTimes()
					client.EXPECT().ListConfigMaps(gomock.Any()).Return([]corev1.ConfigMap{}, nil).AnyTimes()
					return client
				},
			},
			want: []AutomountInfo{
				{
					VolumeType: VolumeTypePVC,
					VolumeName: "defaultPVC1",
					MountPath:  "/tmp/defaultPVC1",
					MountAs:    MountAsFile,
				},
				{
					VolumeType: VolumeTypePVC,
					VolumeName: "defaultPVC2",
					MountPath:  "/tmp/defaultPVC2",
					MountAs:    MountAsFile,
				},
			},
			wantErr: false,
		},
		{
			name: "Two default secrets",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().ListPVCs(gomock.Any()).Return([]corev1.PersistentVolumeClaim{}, nil).AnyTimes()
					client.EXPECT().ListSecrets(gomock.Any()).Return([]corev1.Secret{defaultSecret1, defaultSecret2}, nil).AnyTimes()
					client.EXPECT().ListConfigMaps(gomock.Any()).Return([]corev1.ConfigMap{}, nil).AnyTimes()
					return client
				},
			},
			want: []AutomountInfo{
				{
					VolumeType: VolumeTypeSecret,
					VolumeName: "defaultSecret1",
					MountPath:  "/etc/secret/defaultSecret1",
					MountAs:    MountAsFile,
				},
				{
					VolumeType: VolumeTypeSecret,
					VolumeName: "defaultSecret2",
					MountPath:  "/etc/secret/defaultSecret2",
					MountAs:    MountAsFile,
				},
			},
			wantErr: false,
		},
		{
			name: "Two default configmaps",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().ListPVCs(gomock.Any()).Return([]corev1.PersistentVolumeClaim{}, nil).AnyTimes()
					client.EXPECT().ListSecrets(gomock.Any()).Return([]corev1.Secret{}, nil).AnyTimes()
					client.EXPECT().ListConfigMaps(gomock.Any()).Return([]corev1.ConfigMap{defaultCM1, defaultCM2}, nil).AnyTimes()
					return client
				},
			},
			want: []AutomountInfo{
				{
					VolumeType: VolumeTypeConfigmap,
					VolumeName: "defaultCM1",
					MountPath:  "/etc/config/defaultCM1",
					MountAs:    MountAsFile,
				},
				{
					VolumeType: VolumeTypeConfigmap,
					VolumeName: "defaultCM2",
					MountPath:  "/etc/config/defaultCM2",
					MountAs:    MountAsFile,
				},
			},
			wantErr: false,
		},
		{
			name: "PVC, Secret and ConfigMap with non default mount paths",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().ListPVCs(gomock.Any()).Return([]corev1.PersistentVolumeClaim{pvcMountPath}, nil).AnyTimes()
					client.EXPECT().ListSecrets(gomock.Any()).Return([]corev1.Secret{secretMountPath}, nil).AnyTimes()
					client.EXPECT().ListConfigMaps(gomock.Any()).Return([]corev1.ConfigMap{cmMountPath}, nil).AnyTimes()
					return client
				},
			},
			want: []AutomountInfo{
				{
					VolumeType: VolumeTypePVC,
					VolumeName: "pvcMountPath",
					MountPath:  "/specific/pvc/mount/path",
					MountAs:    MountAsFile,
				},
				{
					VolumeType: VolumeTypeSecret,
					VolumeName: "secretMountPath",
					MountPath:  "/specific/secret/mount/path",
					MountAs:    MountAsFile,
				},
				{
					VolumeType: VolumeTypeConfigmap,
					VolumeName: "cmMountPath",
					MountPath:  "/specific/configmap/mount/path",
					MountAs:    MountAsFile,
				},
			},
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			o := KubernetesClient{
				kubeClient: tt.fields.kubeClient(ctrl),
			}
			got, err := o.GetAutomountingVolumes()
			if (err != nil) != tt.wantErr {
				t.Errorf("KubernetesClient.GetAutomountingVolumes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("KubernetesClient.GetAutomountingVolumes() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
