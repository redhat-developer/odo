package validate

import (
	"reflect"
	"testing"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

func TestValidateComponents(t *testing.T) {

	t.Run("No components present", func(t *testing.T) {

		// Empty components
		components := []devfilev1.Component{}

		got := validateComponents(components)
		want := &NoComponentsError{}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("TestValidateComponents error - got: '%v', want: '%v'", got, want)
		}
	})

	t.Run("Container type component present", func(t *testing.T) {

		components := []devfilev1.Component{
			{
				Name: "container",
				ComponentUnion: devfilev1.ComponentUnion{
					Container: &devfilev1.ContainerComponent{
						Container: devfilev1.Container{
							Image: "image",
						},
					},
				},
			},
		}

		got := validateComponents(components)

		if got != nil {
			t.Errorf("TestValidateComponents error - Not expecting an error: '%v'", got)
		}
	})
}
