package common

import (
	"fmt"
	"reflect"
	"testing"
)

func TestValidateComponents(t *testing.T) {

	t.Run("No components present", func(t *testing.T) {

		// Empty components
		components := []DevfileComponent{}

		got := ValidateComponents(components)
		want := fmt.Errorf(ErrorNoComponents)

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got: '%v', want: '%v'", got, want)
		}
	})

	t.Run("DockerImage type of component present", func(t *testing.T) {

		components := []DevfileComponent{
			{
				Type: DevfileComponentTypeDockerimage,
			},
		}

		got := ValidateComponents(components)

		if got != nil {
			t.Errorf("Not expecting an error: '%v'", got)
		}
	})

	t.Run("DockerImage type of component NOT present", func(t *testing.T) {

		components := []DevfileComponent{
			{
				Type: DevfileComponentTypeCheEditor,
			},
			{
				Type: DevfileComponentTypeChePlugin,
			},
		}

		got := ValidateComponents(components)
		want := fmt.Errorf(ErrorNoDockerImageComponent)

		if !reflect.DeepEqual(got, want) {
			t.Errorf("Incorrect error; want: '%v', got: '%v'", want, got)
		}
	})
}
