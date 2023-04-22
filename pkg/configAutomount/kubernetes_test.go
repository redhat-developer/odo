package configAutomount

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/redhat-developer/odo/pkg/kclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
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

	roPVC := corev1.PersistentVolumeClaim{}
	roPVC.SetName("roPVC")
	roPVC.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})
	roPVC.SetAnnotations(map[string]string{
		annotationReadOnlyName: "true",
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

	secretMountAsSubpath := corev1.Secret{}
	secretMountAsSubpath.Data = map[string][]byte{
		"secretKey1": []byte(""),
		"secretKey2": []byte(""),
	}
	secretMountAsSubpath.SetName("secretMountAsSubpath")
	secretMountAsSubpath.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})
	secretMountAsSubpath.SetAnnotations(map[string]string{
		annotationMountAsName: "subpath",
	})

	secretMountAsEnv := corev1.Secret{}
	secretMountAsEnv.SetName("secretMountAsEnv")
	secretMountAsEnv.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})
	secretMountAsEnv.SetAnnotations(map[string]string{
		annotationMountAsName: "env",
	})

	roSecret := corev1.Secret{}
	roSecret.SetName("roSecret")
	roSecret.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})
	roSecret.SetAnnotations(map[string]string{
		annotationReadOnlyName: "true",
	})

	secretMountAccessMode := corev1.Secret{}
	secretMountAccessMode.SetName("secretMountAccessMode")
	secretMountAccessMode.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})
	secretMountAccessMode.SetAnnotations(map[string]string{
		annotationMountAccessMode: "0400",
	})

	secretMountAccessModeInvalid := corev1.Secret{}
	secretMountAccessModeInvalid.SetName("secretMountAccessModeInvalid")
	secretMountAccessModeInvalid.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})
	secretMountAccessModeInvalid.SetAnnotations(map[string]string{
		annotationMountAccessMode: "01444",
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

	cmMountAsSubpath := corev1.ConfigMap{}
	cmMountAsSubpath.Data = map[string]string{
		"cmKey1": "",
		"cmKey2": "",
	}
	cmMountAsSubpath.SetName("cmMountAsSubpath")
	cmMountAsSubpath.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})
	cmMountAsSubpath.SetAnnotations(map[string]string{
		annotationMountAsName: "subpath",
	})

	cmMountAsEnv := corev1.ConfigMap{}
	cmMountAsEnv.SetName("cmMountAsEnv")
	cmMountAsEnv.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})
	cmMountAsEnv.SetAnnotations(map[string]string{
		annotationMountAsName: "env",
	})

	roCM := corev1.ConfigMap{}
	roCM.SetName("roCM")
	roCM.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})
	roCM.SetAnnotations(map[string]string{
		annotationReadOnlyName: "true",
	})

	cmMountAccessMode := corev1.ConfigMap{}
	cmMountAccessMode.SetName("cmMountAccessMode")
	cmMountAccessMode.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})
	cmMountAccessMode.SetAnnotations(map[string]string{
		annotationMountAccessMode: "0444",
	})

	cmMountAccessModeInvalid := corev1.ConfigMap{}
	cmMountAccessModeInvalid.SetName("cmMountAccessModeInvalid")
	cmMountAccessModeInvalid.SetLabels(map[string]string{
		labelMountName: labelMountValue,
	})
	cmMountAccessModeInvalid.SetAnnotations(map[string]string{
		annotationMountAccessMode: "01444",
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
		{
			name: "Secret and ConfigMap with mount-as annotations",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().ListPVCs(gomock.Any()).Return([]corev1.PersistentVolumeClaim{}, nil).AnyTimes()
					client.EXPECT().ListSecrets(gomock.Any()).Return([]corev1.Secret{secretMountAsSubpath, secretMountAsEnv}, nil).AnyTimes()
					client.EXPECT().ListConfigMaps(gomock.Any()).Return([]corev1.ConfigMap{cmMountAsSubpath, cmMountAsEnv}, nil).AnyTimes()
					return client
				},
			},
			want: []AutomountInfo{
				{
					VolumeType: VolumeTypeSecret,
					VolumeName: "secretMountAsSubpath",
					MountPath:  "/etc/secret/secretMountAsSubpath",
					MountAs:    MountAsSubpath,
					Keys:       []string{"secretKey1", "secretKey2"},
				},
				{
					VolumeType: VolumeTypeSecret,
					VolumeName: "secretMountAsEnv",
					MountPath:  "",
					MountAs:    MountAsEnv,
				},
				{
					VolumeType: VolumeTypeConfigmap,
					VolumeName: "cmMountAsSubpath",
					MountPath:  "/etc/config/cmMountAsSubpath",
					MountAs:    MountAsSubpath,
					Keys:       []string{"cmKey1", "cmKey2"},
				},
				{
					VolumeType: VolumeTypeConfigmap,
					VolumeName: "cmMountAsEnv",
					MountPath:  "",
					MountAs:    MountAsEnv,
				},
			},
			wantErr: false,
		},
		{
			name: "Secret and ConfigMap with mount-access-mode annotations",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().ListPVCs(gomock.Any()).Return([]corev1.PersistentVolumeClaim{}, nil).AnyTimes()
					client.EXPECT().ListSecrets(gomock.Any()).Return([]corev1.Secret{secretMountAccessMode}, nil).AnyTimes()
					client.EXPECT().ListConfigMaps(gomock.Any()).Return([]corev1.ConfigMap{cmMountAccessMode}, nil).AnyTimes()
					return client
				},
			},
			want: []AutomountInfo{
				{
					VolumeType:      VolumeTypeSecret,
					VolumeName:      "secretMountAccessMode",
					MountPath:       "/etc/secret/secretMountAccessMode",
					MountAs:         MountAsFile,
					MountAccessMode: pointer.Int32(0400),
				},
				{
					VolumeType:      VolumeTypeConfigmap,
					VolumeName:      "cmMountAccessMode",
					MountPath:       "/etc/config/cmMountAccessMode",
					MountAs:         MountAsFile,
					MountAccessMode: pointer.Int32(0444),
				},
			},
			wantErr: false,
		},
		{
			name: "Secret with invalid mount-access-mode annotation",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().ListPVCs(gomock.Any()).Return([]corev1.PersistentVolumeClaim{}, nil).AnyTimes()
					client.EXPECT().ListSecrets(gomock.Any()).Return([]corev1.Secret{secretMountAccessModeInvalid}, nil).AnyTimes()
					client.EXPECT().ListConfigMaps(gomock.Any()).Return([]corev1.ConfigMap{}, nil).AnyTimes()
					return client
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Configmap with invalid mount-access-mode annotation",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().ListPVCs(gomock.Any()).Return([]corev1.PersistentVolumeClaim{}, nil).AnyTimes()
					client.EXPECT().ListSecrets(gomock.Any()).Return([]corev1.Secret{}, nil).AnyTimes()
					client.EXPECT().ListConfigMaps(gomock.Any()).Return([]corev1.ConfigMap{cmMountAccessModeInvalid}, nil).AnyTimes()
					return client
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "PVC, Secret and ConfigMap read-only",
			fields: fields{
				kubeClient: func(ctrl *gomock.Controller) kclient.ClientInterface {
					client := kclient.NewMockClientInterface(ctrl)
					client.EXPECT().ListPVCs(gomock.Any()).Return([]corev1.PersistentVolumeClaim{roPVC}, nil).AnyTimes()
					client.EXPECT().ListSecrets(gomock.Any()).Return([]corev1.Secret{roSecret}, nil).AnyTimes()
					client.EXPECT().ListConfigMaps(gomock.Any()).Return([]corev1.ConfigMap{roCM}, nil).AnyTimes()
					return client
				},
			},
			want: []AutomountInfo{
				{
					VolumeType: VolumeTypePVC,
					VolumeName: "roPVC",
					MountPath:  "/tmp/roPVC",
					MountAs:    MountAsFile,
					ReadOnly:   true,
				},
				{
					VolumeType: VolumeTypeSecret,
					VolumeName: "roSecret",
					MountPath:  "/etc/secret/roSecret",
					MountAs:    MountAsFile,
					ReadOnly:   true,
				},
				{
					VolumeType: VolumeTypeConfigmap,
					VolumeName: "roCM",
					MountPath:  "/etc/config/roCM",
					MountAs:    MountAsFile,
					ReadOnly:   true,
				},
			},
			wantErr: false,
		}}
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
