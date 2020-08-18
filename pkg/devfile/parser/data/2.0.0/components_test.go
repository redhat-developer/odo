package version200

import (
	"github.com/kylelemons/godebug/pretty"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/testingutil"
	"reflect"
	"testing"
)

func TestDevfile200_AddVolume(t *testing.T) {
	image0 := "some-image-0"
	container0 := "container0"

	image1 := "some-image-1"
	container1 := "container1"

	volume0 := "volume0"
	volume1 := "volume1"

	type args struct {
		volume common.Volume
		path   string
	}
	tests := []struct {
		name              string
		currentComponents []common.DevfileComponent
		wantComponents    []common.DevfileComponent
		args              args
		wantErr           bool
	}{
		{
			name: "case 1: it should add the volume to all the containers",
			currentComponents: []common.DevfileComponent{
				{
					Container: &common.Container{
						Name:  container0,
						Image: image0,
					},
				},
				{
					Container: &common.Container{
						Name:  container1,
						Image: image1,
					},
				},
			},
			args: args{
				volume: testingutil.GetFakeVolume(volume0, "5Gi"),
				path:   "/path",
			},
			wantComponents: []common.DevfileComponent{
				{
					Container: &common.Container{
						Name:  container0,
						Image: image0,
						VolumeMounts: []common.VolumeMount{
							testingutil.GetFakeVolumeMount(volume0, "/path"),
						},
					},
				},
				{
					Container: &common.Container{
						Name:  container1,
						Image: image1,
						VolumeMounts: []common.VolumeMount{
							testingutil.GetFakeVolumeMount(volume0, "/path"),
						},
					},
				},
				testingutil.GetFakeVolumeComponent(volume0, "5Gi"),
			},
		},
		{
			name: "case 2: it should add the volume when other volumes are present",
			currentComponents: []common.DevfileComponent{
				{
					Container: &common.Container{
						Name:  container0,
						Image: image0,
						VolumeMounts: []common.VolumeMount{
							testingutil.GetFakeVolumeMount(volume1, "/data"),
						},
					},
				},
			},
			args: args{
				volume: testingutil.GetFakeVolume(volume0, "5Gi"),
				path:   "/path",
			},
			wantComponents: []common.DevfileComponent{
				{
					Container: &common.Container{
						Name:  container0,
						Image: image0,
						VolumeMounts: []common.VolumeMount{
							testingutil.GetFakeVolumeMount(volume1, "/data"),
							testingutil.GetFakeVolumeMount(volume0, "/path"),
						},
					},
				},
				testingutil.GetFakeVolumeComponent(volume0, "5Gi"),
			},
		},
		{
			name: "case 3: error out when same volume is present",
			currentComponents: []common.DevfileComponent{
				testingutil.GetFakeVolumeComponent(volume0, "1Gi"),
			},
			args: args{
				volume: testingutil.GetFakeVolume(volume0, "5Gi"),
				path:   "/path",
			},
			wantErr: true,
		},
		{
			name: "case 4: it should error out when another volume is mounted to the same path",
			currentComponents: []common.DevfileComponent{
				{
					Container: &common.Container{
						Name:  container0,
						Image: image0,
						VolumeMounts: []common.VolumeMount{
							testingutil.GetFakeVolumeMount(volume1, "/path"),
						},
					},
				},
				testingutil.GetFakeVolumeComponent(volume1, "5Gi"),
			},
			args: args{
				volume: testingutil.GetFakeVolume(volume0, "5Gi"),
				path:   "/path",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Devfile200{
				Components: tt.currentComponents,
			}

			err := d.AddVolume(tt.args.volume, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddVolume() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr {
				return
			}

			if !reflect.DeepEqual(d.Components, tt.wantComponents) {
				t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantComponents, d.Components, pretty.Compare(tt.wantComponents, d.Components))
			}
		})
	}
}

func TestDevfile200_DeleteVolume(t *testing.T) {
	image0 := "some-image-0"
	container0 := "container0"

	image1 := "some-image-1"
	container1 := "container1"

	volume0 := "volume0"
	volume1 := "volume1"

	type args struct {
		name string
	}
	tests := []struct {
		name              string
		currentComponents []common.DevfileComponent
		wantComponents    []common.DevfileComponent
		args              args
		wantErr           bool
	}{
		{
			name: "case 1: volume is present and mounted to multiple components",
			currentComponents: []common.DevfileComponent{
				{
					Container: &common.Container{
						Name:  container0,
						Image: image0,
						VolumeMounts: []common.VolumeMount{
							testingutil.GetFakeVolumeMount(volume0, "/path"),
						},
					},
				},
				{
					Container: &common.Container{
						Name:  container1,
						Image: image1,
						VolumeMounts: []common.VolumeMount{
							{
								Name: volume0,
								Path: "/path",
							},
						},
					},
				},
				{
					Volume: &common.Volume{
						Name: volume0,
						Size: "5Gi",
					},
				},
			},
			wantComponents: []common.DevfileComponent{
				{
					Container: &common.Container{
						Name:  container0,
						Image: image0,
					},
				},
				{
					Container: &common.Container{
						Name:  container1,
						Image: image1,
					},
				},
			},
			args: args{
				name: volume0,
			},
			wantErr: false,
		},
		{
			name: "case 2: delete only the required volume in case of multiples",
			currentComponents: []common.DevfileComponent{
				{
					Container: &common.Container{
						Name:  container0,
						Image: image0,
						VolumeMounts: []common.VolumeMount{
							testingutil.GetFakeVolumeMount(volume0, "/path"),
							testingutil.GetFakeVolumeMount(volume1, "/data"),
						},
					},
				},
				{
					Container: &common.Container{
						Name:  container1,
						Image: image1,
						VolumeMounts: []common.VolumeMount{
							testingutil.GetFakeVolumeMount(volume1, "/data"),
						},
					},
				},
				testingutil.GetFakeVolumeComponent(volume0, "5Gi"),
				testingutil.GetFakeVolumeComponent(volume1, "5Gi"),
			},
			wantComponents: []common.DevfileComponent{
				{
					Container: &common.Container{
						Name:  container0,
						Image: image0,
						VolumeMounts: []common.VolumeMount{
							testingutil.GetFakeVolumeMount(volume1, "/data"),
						},
					},
				},
				{
					Container: &common.Container{
						Name:  container1,
						Image: image1,
						VolumeMounts: []common.VolumeMount{
							testingutil.GetFakeVolumeMount(volume1, "/data"),
						},
					},
				},
				testingutil.GetFakeVolumeComponent(volume1, "5Gi"),
			},
			args: args{
				name: volume0,
			},
			wantErr: false,
		},
		{
			name: "case 3: volume is not present",
			currentComponents: []common.DevfileComponent{
				{
					Container: &common.Container{
						Name:  container0,
						Image: image0,
						VolumeMounts: []common.VolumeMount{
							testingutil.GetFakeVolumeMount(volume1, "/data"),
						},
					},
				},
				testingutil.GetFakeVolumeComponent(volume1, "5Gi"),
			},
			wantComponents: []common.DevfileComponent{},
			args: args{
				name: volume0,
			},
			wantErr: true,
		},
		{
			name: "case 4: volume is present but not mounted to any component",
			currentComponents: []common.DevfileComponent{
				testingutil.GetFakeVolumeComponent(volume0, "5Gi"),
			},
			wantComponents: []common.DevfileComponent{},
			args: args{
				name: volume0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Devfile200{
				Components: tt.currentComponents,
			}
			err := d.DeleteVolume(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteVolume() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr {
				return
			}

			if !reflect.DeepEqual(d.Components, tt.wantComponents) {
				t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantComponents, d.Components, pretty.Compare(tt.wantComponents, d.Components))
			}
		})
	}
}

func TestDevfile200_GetVolumeMountPath(t *testing.T) {
	volume1 := "volume1"

	type args struct {
		name string
	}
	tests := []struct {
		name              string
		currentComponents []common.DevfileComponent
		wantPath          string
		args              args
		wantErr           bool
	}{
		{
			name: "case 1: volume is present and mounted on a component",
			currentComponents: []common.DevfileComponent{
				{
					Container: &common.Container{
						VolumeMounts: []common.VolumeMount{
							testingutil.GetFakeVolumeMount(volume1, "/path"),
						},
					},
				},
				testingutil.GetFakeVolumeComponent(volume1, "5Gi"),
			},
			wantPath: "/path",
			args: args{
				name: volume1,
			},
			wantErr: false,
		},
		{
			name: "case 2: volume is not present but mounted on a component",
			currentComponents: []common.DevfileComponent{
				{
					Container: &common.Container{
						VolumeMounts: []common.VolumeMount{
							testingutil.GetFakeVolumeMount(volume1, "/path"),
						},
					},
				},
			},
			args: args{
				name: volume1,
			},
			wantErr: true,
		},
		{
			name:              "case 3: volume is not present and not mounted on a component",
			currentComponents: []common.DevfileComponent{},
			args: args{
				name: volume1,
			},
			wantErr: true,
		},
		{
			name: "case 4: volume is present but not mounted",
			currentComponents: []common.DevfileComponent{
				testingutil.GetFakeVolumeComponent(volume1, "5Gi"),
			},
			args: args{
				name: volume1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Devfile200{
				Components: tt.currentComponents,
			}
			got, err := d.GetVolumeMountPath(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVolumeMountPath() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr {
				return
			}

			if !reflect.DeepEqual(got, tt.wantPath) {
				t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantPath, got, pretty.Compare(tt.wantPath, got))
			}
		})
	}
}
