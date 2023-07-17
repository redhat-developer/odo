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
		ctx                   context.Context
		isFeatEnabled         bool
		includeIfFeatDisabled bool
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
			name: "platform unset in context, isFeatEnabled=true, includeIfFeatDisabled=false",
			args: args{
				ctx:                   context.Background(),
				isFeatEnabled:         true,
				includeIfFeatDisabled: false,
			},
			wantResult: allValues,
		},
		{
			name: "platform unset in context, isFeatEnabled=true, includeIfFeatDisabled=true",
			args: args{
				ctx:                   context.Background(),
				isFeatEnabled:         true,
				includeIfFeatDisabled: true,
			},
			wantResult: allValues,
		},
		{
			name: "platform unset in context, isFeatEnabled=false, includeIfFeatDisabled=true",
			args: args{
				ctx:                   context.Background(),
				isFeatEnabled:         false,
				includeIfFeatDisabled: true,
			},
			wantResult: []testType{
				{"value without platform", ""},
				{"value11 (cluster)", "cluster"},
				{"value12 (cluster)", "cluster"},
			},
		},
		{
			name: "platform unset in context, isFeatEnabled=false, includeIfFeatDisabled=false",
			args: args{
				ctx:                   context.Background(),
				isFeatEnabled:         false,
				includeIfFeatDisabled: false,
			},
			wantResult: nil,
		},
		{
			name: "platform set to cluster in context, isFeatEnabled=true, includeIfFeatDisabled=false",
			args: args{
				ctx:                   fcontext.WithPlatform(context.Background(), "cluster"),
				isFeatEnabled:         true,
				includeIfFeatDisabled: false,
			},
			wantResult: []testType{
				{"value without platform", ""},
				{"value11 (cluster)", "cluster"},
				{"value12 (cluster)", "cluster"},
			},
		},
		{
			name: "platform set to cluster in context, isFeatEnabled=true, includeIfFeatDisabled=true",
			args: args{
				ctx:                   fcontext.WithPlatform(context.Background(), "cluster"),
				isFeatEnabled:         true,
				includeIfFeatDisabled: true,
			},
			wantResult: []testType{
				{"value without platform", ""},
				{"value11 (cluster)", "cluster"},
				{"value12 (cluster)", "cluster"},
			},
		},
		{
			name: "platform set to cluster in context, isFeatEnabled=false, includeIfFeatDisabled=false",
			args: args{
				ctx:                   fcontext.WithPlatform(context.Background(), "cluster"),
				isFeatEnabled:         false,
				includeIfFeatDisabled: false,
			},
			wantResult: nil,
		},
		{
			name: "platform set to cluster in context, isFeatEnabled=false, includeIfFeatDisabled=true",
			args: args{
				ctx:                   fcontext.WithPlatform(context.Background(), "cluster"),
				isFeatEnabled:         false,
				includeIfFeatDisabled: true,
			},
			wantResult: []testType{
				{"value without platform", ""},
				{"value11 (cluster)", "cluster"},
				{"value12 (cluster)", "cluster"},
			},
		},
		{
			name: "platform set to podman in context, isFeatEnabled=true, includeIfFeatDisabled=false",
			args: args{
				ctx:                   fcontext.WithPlatform(context.Background(), "podman"),
				isFeatEnabled:         true,
				includeIfFeatDisabled: false,
			},
			wantResult: []testType{
				{"value21 (podman)", "podman"},
				{"value22 (podman)", "podman"},
			},
		},
		{
			name: "platform set to podman in context, isFeatEnabled=true, includeIfFeatDisabled=true",
			args: args{
				ctx:                   fcontext.WithPlatform(context.Background(), "podman"),
				isFeatEnabled:         true,
				includeIfFeatDisabled: true,
			},
			wantResult: []testType{
				{"value21 (podman)", "podman"},
				{"value22 (podman)", "podman"},
			},
		},
		{
			name: "platform set to podman in context, isFeatEnabled=false, includeIfFeatDisabled=false",
			args: args{
				ctx:                   fcontext.WithPlatform(context.Background(), "podman"),
				isFeatEnabled:         false,
				includeIfFeatDisabled: false,
			},
			wantResult: nil,
		},
		{
			name: "platform set to podman in context, isFeatEnabled=false, includeIfFeatDisabled=true",
			args: args{
				ctx:                   fcontext.WithPlatform(context.Background(), "podman"),
				isFeatEnabled:         false,
				includeIfFeatDisabled: true,
			},
			wantResult: []testType{
				{"value21 (podman)", "podman"},
				{"value22 (podman)", "podman"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := filterByPlatform(tt.args.ctx, tt.args.isFeatEnabled, allValues, tt.args.includeIfFeatDisabled)
			if diff := cmp.Diff(tt.wantResult, gotResult, cmp.AllowUnexported(testType{})); diff != "" {
				t.Errorf("filterByPlatform() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
