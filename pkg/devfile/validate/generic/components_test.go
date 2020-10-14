package generic

import (
	"testing"

	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

func TestValidateComponents(t *testing.T) {

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

		if _, ok := got.(*InvalidVolumeError); got != nil && !ok {
			t.Errorf("TestValidateComponents duplicate vol component error - got: '%v' but wanted a different err type", got)
		} else if got == nil {
			t.Errorf("TestValidateComponents reserved env error - expected an err but got nil")
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

	t.Run("Invalid container using reserved env", func(t *testing.T) {

		envName := []string{adaptersCommon.EnvProjectsSrc, adaptersCommon.EnvProjectsRoot}

		for _, env := range envName {
			components := []common.DevfileComponent{
				{
					Name: "container",
					Container: &common.Container{
						Env: []common.Env{
							{
								Name:  env,
								Value: "/some/path/",
							},
						},
					},
				},
			}

			got := validateComponents(components)
			if _, ok := got.(*ReservedEnvError); got != nil && !ok {
				t.Errorf("TestValidateComponents reserved env error - got: '%v' but wanted a different err type", got)
			} else if got == nil {
				t.Errorf("TestValidateComponents reserved env error - expected an err but got nil")
			}
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
		if _, ok := got.(*InvalidVolumeError); got != nil && !ok {
			t.Errorf("TestValidateComponents vol size error - got: '%v' but wanted a different err type", got)
		} else if got == nil {
			t.Errorf("TestValidateComponents vol size error - expected an err but got nil")
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
		if _, ok := got.(*MissingVolumeMountError); got != nil && !ok {
			t.Errorf("TestValidateComponents vol mount error - got: '%v' but wanted a different err type", got)
		} else if got == nil {
			t.Errorf("TestValidateComponents vol mount error - expected an err but got nil")
		}
	})
}
