package adapters

import (
	"reflect"
	"testing"

	"github.com/devfile/library/pkg/devfile/parser/data"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	adaptersCommon "github.com/openshift/odo/v2/pkg/devfile/adapters/common"
	"github.com/openshift/odo/v2/pkg/occlient"
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
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{})
					if err != nil {
						t.Error(err)
					}
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
