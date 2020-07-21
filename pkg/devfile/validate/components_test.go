package validate

import (
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
		want := &NoComponentsError{}

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

		got := validateComponents(components)
		want := &DuplicateVolumeComponentsError{}

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

		got := validateComponents(components)

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

		got := validateComponents(components)
		want := "size randomgarbage for volume component myvol is invalid"

		if !strings.Contains(got.Error(), want) {
			t.Errorf("TestValidateComponents error - got: '%v', want substring: '%v'", got.Error(), want)
		}
	})

	t.Run("Invalid volume mount", func(t *testing.T) {

		components := []common.DevfileComponent{
			{
				Volume: &common.Volume{
					Name: "myvol",
					Size: "2Gi",
				},
			},
			{
				Container: &common.Container{
					Name: "container",
					VolumeMounts: []common.VolumeMount{
						{
							Name: "myinvalidvol",
						},
						{
							Name: "myinvalidvol2",
						},
					},
				},
			},
		}

		got := validateComponents(components)
		want := "unable to find volume mount"

		if !strings.Contains(got.Error(), want) {
			t.Errorf("TestValidateComponents error - got: '%v', want substr: '%v'", got.Error(), want)
		}
	})
}
