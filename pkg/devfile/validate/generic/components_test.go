package generic

import (
	"testing"

	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

func TestValidateComponents(t *testing.T) {

	tests := []struct {
		name        string
		components  []common.DevfileComponent
		wantErr     bool
		wantErrType error
	}{
		{
			name: "Case 1: Duplicate volume components present",
			components: []common.DevfileComponent{
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
			},
			wantErr: true,
		},
		{
			name: "Case 2: Long component name",
			components: []common.DevfileComponent{
				{
					Name: "myvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvol",
				},
			},
			wantErr: true,
		},
		{
			name: "Case 3: Valid container and volume component",
			components: []common.DevfileComponent{
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
			},
			wantErr: false,
		},
		{
			name: "Case 4: Invalid container using reserved env PROJECT_SOURCE",
			components: []common.DevfileComponent{
				{
					Name: "container",
					Container: &common.Container{
						Env: []common.Env{
							{
								Name:  adaptersCommon.EnvProjectsSrc,
								Value: "/some/path/",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 5: Invalid container using reserved env PROJECTS_ROOT",
			components: []common.DevfileComponent{
				{
					Name: "container",
					Container: &common.Container{
						Env: []common.Env{
							{
								Name:  adaptersCommon.EnvProjectsRoot,
								Value: "/some/path/",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 6: Invalid volume component size",
			components: []common.DevfileComponent{
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
			},
			wantErr: true,
		},
		{
			name: "Case 7: Invalid volume mount",
			components: []common.DevfileComponent{
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
			},
			wantErr: true,
		},
		{
			name: "Case 8: Special character in container name",
			components: []common.DevfileComponent{
				{
					Name: "run@time",
				},
			},
			wantErr: true,
		},
		{
			name: "Case 9: Numeric container name",
			components: []common.DevfileComponent{
				{
					Name: "12345",
				},
			},
			wantErr: true,
		},
		{
			name: "Case 10: Container name with capitalised character",
			components: []common.DevfileComponent{
				{
					Name: "runTime",
				},
			},
			wantErr: true,
		},
		{
			name: "Case 11: Invalid container with same endpoint names",
			components: []common.DevfileComponent{
				{
					Name: "name1",
					Container: &common.Container{
						Image: "image1",

						Endpoints: []common.Endpoint{
							{
								Name:       "url1",
								TargetPort: 8080,
								Exposure:   common.Public,
							},
						},
					},
				},
				{
					Name: "name2",
					Container: &common.Container{
						Image: "image2",

						Endpoints: []common.Endpoint{
							{
								Name:       "url1",
								TargetPort: 8081,
								Exposure:   common.Public,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 12: Invalid container with same endpoint target ports",
			components: []common.DevfileComponent{
				{
					Name: "name1",
					Container: &common.Container{
						Image: "image1",
						Endpoints: []common.Endpoint{
							{
								Name:       "url1",
								TargetPort: 8080,
							},
						},
					},
				},
				{
					Name: "name2",
					Container: &common.Container{
						Image: "image2",
						Endpoints: []common.Endpoint{
							{
								Name:       "url2",
								TargetPort: 8080,
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateComponents(tt.components)

			if tt.wantErr && got == nil {
				t.Errorf("TestValidateComponents error - expected an err but got nil")
			} else if !tt.wantErr && got != nil {
				t.Errorf("TestValidateComponents error - unexpected err %v", got)
			}
		})
	}

}
