package adapters

import (
	"reflect"
	"testing"

	"github.com/openshift/odo/pkg/devfile"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	versionsCommon "github.com/openshift/odo/pkg/devfile/versions/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/testingutil"
)

func TestNewPlatformAdapter(t *testing.T) {
	tests := []struct {
		adapterType   string
		name          string
		componentName string
		componentType versionsCommon.DevfileComponentType
		wantErr       bool
	}{
		{
			adapterType:   "kubernetes.Adapter",
			name:          "get platform adapter",
			componentName: "test",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run("get platform adapter", func(t *testing.T) {
			devObj := devfile.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}

			adapterContext := adaptersCommon.AdapterContext{
				ComponentName: tt.componentName,
				Devfile:       devObj,
			}
			fkclient, _ := kclient.FakeNew()
			adapter, err := newKubernetesAdapter(adapterContext, *fkclient)
			if err != nil {
				t.Errorf("unexpected error: '%v'", err)
			}

			// test that the returned adapter is of the right type
			if !tt.wantErr == (reflect.TypeOf(adapter).String() != tt.adapterType) {
				t.Errorf("incorrect adapter type: '%v'", err)
			}
		})
	}
}
