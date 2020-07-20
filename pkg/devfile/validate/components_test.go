package validate

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

func TestValidateComponents(t *testing.T) {

	t.Run("No components present", func(t *testing.T) {

		// Empty components
		components := []common.DevfileComponent{}

		got := validateComponents(components)
		want := fmt.Errorf(ErrorNoComponents)

		if !reflect.DeepEqual(got, want) {
			t.Errorf("TestValidateComponents error - got: '%v', want: '%v'", got, want)
		}
	})

	t.Run("Container type of component present", func(t *testing.T) {

		components := []common.DevfileComponent{
			{
				Container: &common.Container{
					Name: "container",
				},
			},
		}

		got := validateComponents(components)

		if got != nil {
			t.Errorf("TestValidateComponents error - Not expecting an error: '%v'", got)
		}
	})

	t.Run("Duplicate volume components present", func(t *testing.T) {

		components := []common.DevfileComponent{
			{
				Volume: &common.Volume{
					Name: "myvol",
				},
			},
			{
				Volume: &common.Volume{
					Name: "myvol",
				},
			},
		}

		got := ValidateComponents(components)
		want := fmt.Errorf(ErrorDuplicateVolumeComponents)

		if !reflect.DeepEqual(got, want) {
			t.Errorf("TestValidateComponents error - got: '%v', want: '%v'", got, want)
		}
	})

	t.Run("Valid container and volume component", func(t *testing.T) {

		components := []common.DevfileComponent{
			{
				Volume: &common.Volume{
					Name: "myvol",
				},
			},
			{
				Container: &common.Container{
					Name: "container",
					VolumeMounts: []common.VolumeMount{
						{
							Name: "myvol",
							Path: "/some/path/",
						},
					},
				},
			},
			{
				Container: &common.Container{
					Name: "container2",
					VolumeMounts: []common.VolumeMount{
						{
							Name: "myvol",
						},
					},
				},
			},
		}

		got := ValidateComponents(components)

		if got != nil {
			t.Errorf("TestValidateComponents error - got: '%v'", got)
		}
	})

	t.Run("Invalid volume component size", func(t *testing.T) {

		components := []common.DevfileComponent{
			{
				Volume: &common.Volume{
					Name: "myvol",
					Size: "randomgarbage",
				},
			},
			{
				Container: &common.Container{
					Name: "container",
					VolumeMounts: []common.VolumeMount{
						{
							Name: "myvol",
							Path: "/some/path/",
						},
					},
				},
			},
		}

		got := ValidateComponents(components)
		want := "size randomgarbage for volume component myvol is invalid"

		if !strings.Contains(got.Error(), want) {
			t.Errorf("TestValidateComponents error - got: '%v', want substring: '%v'", got.Error(), want)
		}
	})
}
