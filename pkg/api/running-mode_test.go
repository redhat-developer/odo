package api

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMergeRunningModes(t *testing.T) {
	type args struct {
		m map[string]RunningModes
	}
	tests := []struct {
		name string
		args args
		want RunningModes
	}{
		{
			name: "nil map",
			want: nil,
		},
		{
			name: "all false with some unknown modes",
			args: args{
				m: map[string]RunningModes{
					"podman": map[RunningMode]bool{
						RunningModeDev:    false,
						RunningModeDeploy: false,
					},
					"cluster": map[RunningMode]bool{
						RunningModeDev:    false,
						RunningModeDeploy: false,
					},
					"unknown-platform": map[RunningMode]bool{
						"unknown-mode": true,
						"another-mode": true,
					},
				},
			},
			want: NewRunningModes(),
		},
		{
			name: "true for one platform and false for another",
			args: args{
				m: map[string]RunningModes{
					"podman": map[RunningMode]bool{
						RunningModeDev:    true,
						RunningModeDeploy: false,
					},
					"cluster": map[RunningMode]bool{
						RunningModeDev:    false,
						RunningModeDeploy: true,
					},
				},
			},
			want: map[RunningMode]bool{
				RunningModeDev:    true,
				RunningModeDeploy: true,
			},
		},
		{
			name: "true for all platforms",
			args: args{
				m: map[string]RunningModes{
					"podman": map[RunningMode]bool{
						RunningModeDev:    true,
						RunningModeDeploy: true,
					},
					"cluster": map[RunningMode]bool{
						RunningModeDev:    true,
						RunningModeDeploy: true,
					},
				},
			},
			want: map[RunningMode]bool{
				RunningModeDev:    true,
				RunningModeDeploy: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeRunningModes(tt.args.m)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("MergeRunningModes() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
