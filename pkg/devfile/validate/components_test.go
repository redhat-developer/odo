package validate

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

func TestValidateComponents(t *testing.T) {

	t.Run("No components present", func(t *testing.T) {

		// Empty components
		components := []common.DevfileComponent{}

		got := ValidateComponents(components)
		want := fmt.Errorf(ErrorNoComponents)

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got: '%v', want: '%v'", got, want)
		}
	})

	t.Run("DockerImage type of component present", func(t *testing.T) {

		components := []common.DevfileComponent{
			{
				Type: common.ContainerComponentType,
			},
		}

		got := ValidateComponents(components)

		if got != nil {
			t.Errorf("Not expecting an error: '%v'", got)
		}
	})

	t.Run("DockerImage type of component NOT present", func(t *testing.T) {

		components := []common.DevfileComponent{
			{
				Type: common.PluginComponentType,
			},
			{
				Type: common.KubernetesComponentType,
			},
		}

		got := ValidateComponents(components)
		want := fmt.Errorf(ErrorNoContainerComponent)

		if !reflect.DeepEqual(got, want) {
			t.Errorf("Incorrect error; want: '%v', got: '%v'", want, got)
		}
	})
}
