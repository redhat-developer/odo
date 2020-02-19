package component

import (
	"testing"

	"github.com/openshift/odo/pkg/devfile"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	versionsCommon "github.com/openshift/odo/pkg/devfile/versions/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/testingutil"
)

func TestStart(t *testing.T) {

	tests := []struct {
		name          string
		componentName string
		componentType versionsCommon.DevfileComponentType
		wantErr       bool
	}{
		{
			name:          "Case: Invalid devfile",
			componentName: "",
			componentType: "",
			wantErr:       true,
		},
		{
			name:          "Case: Valid devfile",
			componentName: "test",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfile.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}

			adapterMeta := adaptersCommon.AdapterMetadata{
				ComponentName: tt.componentName,
				Devfile:       devObj,
			}

			client, _ := kclient.FakeNew()
			componentAdapter := New(adapterMeta, *client)

			err := componentAdapter.Start()

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter start unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}
