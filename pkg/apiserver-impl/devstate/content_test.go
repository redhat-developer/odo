package devstate

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
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
				Commands:   []Command{},
				Containers: []Container{},
				Images:     []Image{},
				Resources:  []Resource{},
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
