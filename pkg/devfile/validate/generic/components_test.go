package generic

import (
	"testing"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
)

func TestValidateComponents(t *testing.T) {

	tests := []struct {
		name        string
		components  []devfilev1.Component
		wantErr     bool
		wantErrType error
	}{
		{
			name: "Case 1: Duplicate volume components present",
			components: []devfilev1.Component{
				{
					Name: "myvol",
					ComponentUnion: devfilev1.ComponentUnion{
						Volume: &devfilev1.VolumeComponent{
							Volume: devfilev1.Volume{
								Size: "1Gi",
							},
						},
					},
				},
				{
					Name: "myvol",
					ComponentUnion: devfilev1.ComponentUnion{
						Volume: &devfilev1.VolumeComponent{
							Volume: devfilev1.Volume{
								Size: "1Gi",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 2: Long component name",
			components: []devfilev1.Component{
				{
					Name: "myvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvolmyvol",
				},
			},
			wantErr: true,
		},
		{
			name: "Case 3: Valid container and volume component",
			components: []devfilev1.Component{
				{
					Name: "myvol",
					ComponentUnion: devfilev1.ComponentUnion{
						Volume: &devfilev1.VolumeComponent{
							Volume: devfilev1.Volume{
								Size: "1Gi",
							},
						},
					},
				},
				{
					Name: "container",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								VolumeMounts: []devfilev1.VolumeMount{
									{
										Name: "myvol",
										Path: "/some/path/",
									},
								},
							},
						},
					},
				},
				{
					Name: "container2",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								VolumeMounts: []devfilev1.VolumeMount{
									{
										Name: "myvol",
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 4: Invalid container using reserved env PROJECT_SOURCE",
			components: []devfilev1.Component{
				{
					Name: "container",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								Env: []devfilev1.EnvVar{
									{
										Name:  adaptersCommon.EnvProjectsSrc,
										Value: "/some/path/",
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 5: Invalid container using reserved env PROJECTS_ROOT",
			components: []devfilev1.Component{
				{
					Name: "container",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								Env: []devfilev1.EnvVar{
									{
										Name:  adaptersCommon.EnvProjectsRoot,
										Value: "/some/path/",
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 6: Invalid volume component size",
			components: []devfilev1.Component{
				{
					Name: "myvol",
					ComponentUnion: devfilev1.ComponentUnion{
						Volume: &devfilev1.VolumeComponent{
							Volume: devfilev1.Volume{
								Size: "randomgarbage",
							},
						},
					},
				},
				{
					Name: "container",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								VolumeMounts: []devfilev1.VolumeMount{
									{
										Name: "myvol",
										Path: "/some/path/",
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 7: Invalid volume mount",
			components: []devfilev1.Component{
				{
					Name: "myvol",
					ComponentUnion: devfilev1.ComponentUnion{
						Volume: &devfilev1.VolumeComponent{
							Volume: devfilev1.Volume{
								Size: "2Gi",
							},
						},
					},
				},
				{
					Name: "container",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								VolumeMounts: []devfilev1.VolumeMount{
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
				},
			},
			wantErr: true,
		},
		{
			name: "Case 8: Special character in container name",
			components: []devfilev1.Component{
				{
					Name: "run@time",
				},
			},
			wantErr: true,
		},
		{
			name: "Case 9: Numeric container name",
			components: []devfilev1.Component{
				{
					Name: "12345",
				},
			},
			wantErr: true,
		},
		{
			name: "Case 10: Container name with capitalised character",
			components: []devfilev1.Component{
				{
					Name: "runTime",
				},
			},
			wantErr: true,
		},
		{
			name: "Case 11: Invalid container with same endpoint names",
			components: []devfilev1.Component{
				{
					Name: "name1",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Endpoints: []devfilev1.Endpoint{
								{
									Name:       "url1",
									TargetPort: 8080,
									Exposure:   devfilev1.PublicEndpointExposure,
								},
							},
							Container: devfilev1.Container{
								Image: "image1",
							},
						},
					},
				},
				{
					Name: "name2",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Endpoints: []devfilev1.Endpoint{
								{
									Name:       "url1",
									TargetPort: 8081,
									Exposure:   devfilev1.PublicEndpointExposure,
								},
							},
							Container: devfilev1.Container{
								Image: "image2",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 12: Invalid container with same endpoint target ports",
			components: []devfilev1.Component{
				{
					Name: "name1",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								Image: "image1",
							},
							Endpoints: []devfilev1.Endpoint{
								{
									Name:       "url1",
									TargetPort: 8080,
								},
							},
						},
					},
				},
				{
					Name: "name2",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								Image: "image2",
							},
							Endpoints: []devfilev1.Endpoint{
								{
									Name:       "url2",
									TargetPort: 8080,
								},
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
