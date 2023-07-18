package describe

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	fcontext "github.com/redhat-developer/odo/pkg/odo/commonflags/context"
)

type testType struct {
	value    string
	platform string
}

var _ platformDependent = testType{}

func (c testType) GetPlatform() string {
	return c.platform
}

func Test_filterByPlatform(t *testing.T) {
	type args struct {
		ctx           context.Context
		isFeatEnabled bool
	}
	type testCase struct {
		name       string
		args       args
		wantResult []testType
	}
	allValues := []testType{
		{value: "value without platform"},
		{value: "value11 (cluster)", platform: "cluster"},
		{value: "value12 (cluster)", platform: "cluster"},
		{value: "value21 (podman)", platform: "podman"},
		{value: "value22 (podman)", platform: "podman"},
	}
	tests := []testCase{
		{
			name: "feature disabled",
			args: args{
				ctx:           context.Background(),
				isFeatEnabled: false,
			},
			wantResult: nil,
		},
		{
			name: "feature enabled and platform unset in context",
			args: args{
				ctx:           context.Background(),
				isFeatEnabled: true,
			},
			wantResult: allValues,
		},
		{
			name: "feature enabled and platform set to cluster in context",
			args: args{
				ctx:           fcontext.WithPlatform(context.Background(), "cluster"),
				isFeatEnabled: true,
			},
			wantResult: []testType{
				{"value without platform", ""},
				{"value11 (cluster)", "cluster"},
				{"value12 (cluster)", "cluster"},
			},
		},
		{
			name: "feature enabled and platform set to podman in context",
			args: args{
				ctx:           fcontext.WithPlatform(context.Background(), "podman"),
				isFeatEnabled: true,
			},
			wantResult: []testType{
				{"value21 (podman)", "podman"},
				{"value22 (podman)", "podman"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := filterByPlatform(tt.args.ctx, tt.args.isFeatEnabled, allValues)
			if diff := cmp.Diff(tt.wantResult, gotResult, cmp.AllowUnexported(testType{})); diff != "" {
				t.Errorf("filterByPlatform() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
