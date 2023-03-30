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
