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
				Name: "container",
				Container: &common.Container{
					Image: "image",
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
				Name: "myvol",
				Volume: &common.Volume{
					Size: "1Gi",
				},
			},
			{
				Name: "myvol",
				Volume: &common.Volume{
					Size: "1Gi",
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
				Name: "myvol",
				Volume: &common.Volume{
					Size: "1Gi",
				},
			},
			{
				Name: "container",
				Container: &common.Container{
					VolumeMounts: []common.VolumeMount{
						{
							Name: "myvol",
							Path: "/some/path/",
						},
					},
				},
			},
			{
				Name: "container2",
				Container: &common.Container{
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
				Name: "myvol",
				Volume: &common.Volume{
					Size: "randomgarbage",
				},
			},
			{
				Name: "container",
				Container: &common.Container{
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

		if got != nil && !strings.Contains(got.Error(), want) {
			t.Errorf("TestValidateComponents error - got: '%v', want substring: '%v'", got.Error(), want)
		}
	})

	t.Run("Invalid volume mount", func(t *testing.T) {

		components := []common.DevfileComponent{
			{
				Name: "myvol",
				Volume: &common.Volume{
					Size: "2Gi",
				},
			},
			{
				Name: "container",
				Container: &common.Container{
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
