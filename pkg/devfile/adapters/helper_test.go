package adapters

import (
	"reflect"
	"testing"

	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
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
			componentType: versionsCommon.ContainerComponentType,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run("get platform adapter", func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: []versionsCommon.DevfileComponent{},
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
