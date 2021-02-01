package adapters

import (
	"github.com/devfile/library/pkg/devfile/parser/data"
	"reflect"
	"testing"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/occlient"
)

func TestNewPlatformAdapter(t *testing.T) {
	tests := []struct {
		adapterType   string
		name          string
		componentName string
		componentType devfilev1.ComponentType
		wantErr       bool
	}{
		{
			adapterType:   "kubernetes.Adapter",
			name:          "get platform adapter",
			componentName: "test",
			componentType: devfilev1.ContainerComponentType,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run("get platform adapter", func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, _ := data.NewDevfileData(string(data.APIVersion200))
					_ = devfileData.AddComponents([]devfilev1.Component{})
					return devfileData
				}(),
			}

			adapterContext := adaptersCommon.AdapterContext{
				ComponentName: tt.componentName,
				Devfile:       devObj,
			}
			fkclient, _ := occlient.FakeNew()
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
