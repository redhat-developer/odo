package devstate

import (
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/pointer"

	. "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
	"github.com/redhat-developer/odo/pkg/testingutil"
)

func TestDevfileState_GetContent(t *testing.T) {
	tests := []struct {
		name    string
		state   func() DevfileState
		want    DevfileContent
		wantErr bool
	}{
		{
			state: func() DevfileState {
				return NewDevfileState()
			},
			want: DevfileContent{
				Content:    "metadata: {}\nschemaVersion: 2.2.0\n",
				Version:    "2.2.0",
				Commands:   []Command{},
				Containers: []Container{},
				Images:     []Image{},
				Resources:  []Resource{},
				Volumes:    []Volume{},
				Events:     Events{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state()
			got, err := o.GetContent()
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.GetContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.GetContent() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDevfileState_getVolumes(t *testing.T) {
	var (
		volWithNoEphemeral = testingutil.GetFakeVolumeComponent("vol-ephemeral-not-set", "1Gi")
		volEphemeral       = testingutil.GetFakeVolumeComponent("vol-ephemeral-true", "2Gi")
		volEphemeralFalse  = testingutil.GetFakeVolumeComponent("vol-ephemeral-false", "3Gi")
	)
	volWithNoEphemeral.Volume.Ephemeral = nil
	volEphemeral.Volume.Ephemeral = pointer.Bool(true)
	volEphemeralFalse.Volume.Ephemeral = pointer.Bool(false)

	tests := []struct {
		name    string
		state   func() (DevfileState, error)
		want    []Volume
		wantErr bool
	}{
		{
			name: "should not panic if 'ephemeral' is not set on the Devfile volume component",
			state: func() (DevfileState, error) {
				devfileData, err := data.NewDevfileData(string(data.APISchemaVersion220))
				if err != nil {
					return DevfileState{}, err
				}
				err = devfileData.AddComponents([]v1alpha2.Component{
					volEphemeral,
					volWithNoEphemeral,
					volEphemeralFalse,
				})
				if err != nil {
					return DevfileState{}, err
				}
				s := NewDevfileState()
				s.Devfile.Data = devfileData
				return s, nil
			},
			want: []Volume{
				{Name: "vol-ephemeral-true", Ephemeral: true, Size: "2Gi"},
				{Name: "vol-ephemeral-not-set", Ephemeral: false, Size: "1Gi"},
				{Name: "vol-ephemeral-false", Ephemeral: false, Size: "3Gi"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, err := tt.state()
			if err != nil {
				t.Fatalf("DevfileState.getVolumes() error preparing Devfile state: %v", err)
				return
			}
			got, err := o.getVolumes()
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.getVolumes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.getVolumes() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
