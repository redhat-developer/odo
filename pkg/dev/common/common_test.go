package common

import (
	"testing"

	"github.com/devfile/library/v2/pkg/devfile/generator"

	corev1 "k8s.io/api/core/v1"
)

func TestGetFirstContainerWithSourceVolume(t *testing.T) {
	tests := []struct {
		name           string
		containers     []corev1.Container
		want           string
		wantSourcePath string
		wantErr        bool
	}{
		{
			name: "Case: One container, Project Source Env",
			containers: []corev1.Container{
				{
					Name: "test",
					Env: []corev1.EnvVar{
						{
							Name:  "RANDOMENV",
							Value: "/mypath2",
						},
						{
							Name:  generator.EnvProjectsSrc,
							Value: "/mypath",
						},
					},
				},
			},
			want:           "test",
			wantSourcePath: "/mypath",
			wantErr:        false,
		},
		{
			name: "Case: Multiple containers, multiple Project Source Env",
			containers: []corev1.Container{
				{
					Name: "test1",
					Env: []corev1.EnvVar{
						{
							Name:  "RANDOMENV",
							Value: "/mypath1",
						},
						{
							Name:  generator.EnvProjectsSrc,
							Value: "/mypath1",
						},
					},
				},
				{
					Name: "test2",
					Env: []corev1.EnvVar{
						{
							Name:  "RANDOMENV",
							Value: "/mypath2",
						},
						{
							Name:  generator.EnvProjectsSrc,
							Value: "/mypath2",
						},
					},
				},
			},
			want:           "test1",
			wantSourcePath: "/mypath1",
			wantErr:        false,
		},
		{
			name: "Case: Multiple containers, no Project Source Env",
			containers: []corev1.Container{
				{
					Name: "test1",
					Env: []corev1.EnvVar{
						{
							Name:  "RANDOMENV",
							Value: "/mypath1",
						},
					},
				},
				{
					Name: "test2",
					Env: []corev1.EnvVar{
						{
							Name:  "RANDOMENV",
							Value: "/mypath2",
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, syncFolder, err := GetFirstContainerWithSourceVolume(tt.containers)
			if container != tt.want {
				t.Errorf("expected %s, actual %s", tt.want, container)
			}
			if syncFolder != tt.wantSourcePath {
				t.Errorf("expected %s, actual %s", tt.wantSourcePath, syncFolder)
			}
			if !tt.wantErr == (err != nil) {
				t.Errorf("expected %v, actual %v", tt.wantErr, err)
			}
		})
	}
}
