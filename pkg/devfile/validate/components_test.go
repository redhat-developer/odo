package validate

import (
	"testing"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/google/go-cmp/cmp"
)

func TestValidateComponents(t *testing.T) {

	t.Run("No components present", func(t *testing.T) {

		// Empty components
		components := []devfilev1.Component{}

		got := validateComponents(components)
		want := &NoComponentsError{}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("validateComponents() mismatch (-want +got):\n%s", diff)
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
