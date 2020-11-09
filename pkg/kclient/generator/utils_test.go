package generator

import (
	"path/filepath"
	"reflect"
	"testing"

	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/testingutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestConvertEnvs(t *testing.T) {
	envVarsNames := []string{"test", "sample-var", "myvar"}
	envVarsValues := []string{"value1", "value2", "value3"}
	tests := []struct {
		name    string
		envVars []common.Env
		want    []corev1.EnvVar
	}{
		{
			name: "Case 1: One env var",
			envVars: []common.Env{
				{
					Name:  envVarsNames[0],
					Value: envVarsValues[0],
				},
			},
			want: []corev1.EnvVar{
				{
					Name:  envVarsNames[0],
					Value: envVarsValues[0],
				},
			},
		},
		{
			name: "Case 2: Multiple env vars",
			envVars: []common.Env{
				{
					Name:  envVarsNames[0],
					Value: envVarsValues[0],
				},
				{
					Name:  envVarsNames[1],
					Value: envVarsValues[1],
				},
				{
					Name:  envVarsNames[2],
					Value: envVarsValues[2],
				},
			},
			want: []corev1.EnvVar{
				{
					Name:  envVarsNames[0],
					Value: envVarsValues[0],
				},
				{
					Name:  envVarsNames[1],
					Value: envVarsValues[1],
				},
				{
					Name:  envVarsNames[2],
					Value: envVarsValues[2],
				},
			},
		},
		{
			name:    "Case 3: No env vars",
			envVars: []common.Env{},
			want:    []corev1.EnvVar{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars := convertEnvs(tt.envVars)
			if !reflect.DeepEqual(tt.want, envVars) {
				t.Errorf("expected %v, wanted %v", envVars, tt.want)
			}
		})
	}
}

func TestConvertPorts(t *testing.T) {
	endpointsNames := []string{"endpoint1", "endpoint2"}
	endpointsPorts := []int32{8080, 9090}
	tests := []struct {
		name      string
		endpoints []common.Endpoint
		want      []corev1.ContainerPort
	}{
		{
			name: "Case 1: One Endpoint",
			endpoints: []common.Endpoint{
				{
					Name:       endpointsNames[0],
					TargetPort: endpointsPorts[0],
				},
			},
			want: []corev1.ContainerPort{
				{
					Name:          endpointsNames[0],
					ContainerPort: endpointsPorts[0],
				},
			},
		},
		{
			name: "Case 2: Multiple env vars",
			endpoints: []common.Endpoint{
				{
					Name:       endpointsNames[0],
					TargetPort: endpointsPorts[0],
				},
				{
					Name:       endpointsNames[1],
					TargetPort: endpointsPorts[1],
				},
			},
			want: []corev1.ContainerPort{
				{
					Name:          endpointsNames[0],
					ContainerPort: endpointsPorts[0],
				},
				{
					Name:          endpointsNames[1],
					ContainerPort: endpointsPorts[1],
				},
			},
		},
		{
			name:      "Case 3: No endpoints",
			endpoints: []common.Endpoint{},
			want:      []corev1.ContainerPort{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ports := convertPorts(tt.endpoints)
			if !reflect.DeepEqual(tt.want, ports) {
				t.Errorf("expected %v, wanted %v", ports, tt.want)
			}
		})
	}
}

func TestGetResourceReqs(t *testing.T) {
	limit := "1024Mi"
	quantity, err := resource.ParseQuantity(limit)
	if err != nil {
		t.Errorf("expected %v", err)
	}
	tests := []struct {
		name      string
		component common.DevfileComponent
		want      corev1.ResourceRequirements
	}{
		{
			name: "Case 1: One Endpoint",
			component: common.DevfileComponent{
				Name: "testcomponent",
				Container: &common.Container{
					MemoryLimit: "1024Mi",
				},
			},
			want: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: quantity,
				},
			},
		},
		{
			name:      "Case 2: Empty DevfileComponent",
			component: common.DevfileComponent{},
			want:      corev1.ResourceRequirements{},
		},
		{
			name: "Case 3: Valid container, but empty memoryLimit",
			component: common.DevfileComponent{
				Name: "testcomponent",
				Container: &common.Container{
					Image: "testimage",
				}},
			want: corev1.ResourceRequirements{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := getResourceReqs(tt.component)
			if !reflect.DeepEqual(tt.want, req) {
				t.Errorf("expected %v, wanted %v", req, tt.want)
			}
		})
	}
}

func TestGetDevfileContainerComponents(t *testing.T) {

	tests := []struct {
		name                 string
		component            []common.DevfileComponent
		alias                []string
		expectedMatchesCount int
	}{
		{
			name:                 "Case 1: Invalid devfile",
			component:            []common.DevfileComponent{},
			expectedMatchesCount: 0,
		},
		{
			name:                 "Case 2: Valid devfile with wrong component type (Openshift)",
			component:            []common.DevfileComponent{{Openshift: &common.Openshift{}}},
			expectedMatchesCount: 0,
		},
		{
			name:                 "Case 3: Valid devfile with wrong component type (Kubernetes)",
			component:            []common.DevfileComponent{{Kubernetes: &common.Kubernetes{}}},
			expectedMatchesCount: 0,
		},

		{
			name:                 "Case 4 : Valid devfile with correct component type (Container)",
			component:            []common.DevfileComponent{testingutil.GetFakeContainerComponent("comp1"), testingutil.GetFakeContainerComponent("comp2")},
			expectedMatchesCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: tt.component,
				},
			}

			devfileComponents := GetDevfileContainerComponents(devObj.Data)

			if len(devfileComponents) != tt.expectedMatchesCount {
				t.Errorf("TestGetDevfileContainerComponents error: wrong number of components matched: expected %v, actual %v", tt.expectedMatchesCount, len(devfileComponents))
			}
		})
	}

}

func TestGetDevfileVolumeComponents(t *testing.T) {

	tests := []struct {
		name                 string
		component            []common.DevfileComponent
		alias                []string
		expectedMatchesCount int
	}{
		{
			name:                 "Case 1: Invalid devfile",
			component:            []common.DevfileComponent{},
			expectedMatchesCount: 0,
		},
		{
			name:                 "Case 2: Valid devfile with wrong component type (Openshift)",
			component:            []common.DevfileComponent{{Openshift: &common.Openshift{}}},
			expectedMatchesCount: 0,
		},
		{
			name:                 "Case 3: Valid devfile with wrong component type (Kubernetes)",
			component:            []common.DevfileComponent{{Kubernetes: &common.Kubernetes{}}},
			expectedMatchesCount: 0,
		},

		{
			name:                 "Case 4 : Valid devfile with wrong component type (Container)",
			component:            []common.DevfileComponent{testingutil.GetFakeContainerComponent("comp1"), testingutil.GetFakeContainerComponent("comp2")},
			expectedMatchesCount: 0,
		},

		{
			name:                 "Case 5: Valid devfile with correct component type (Volume)",
			component:            []common.DevfileComponent{testingutil.GetFakeContainerComponent("comp1"), testingutil.GetFakeVolumeComponent("myvol", "4Gi")},
			expectedMatchesCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: tt.component,
				},
			}

			devfileComponents := GetDevfileVolumeComponents(devObj.Data)

			if len(devfileComponents) != tt.expectedMatchesCount {
				t.Errorf("TestGetDevfileVolumeComponents error: wrong number of components matched: expected %v, actual %v", tt.expectedMatchesCount, len(devfileComponents))
			}
		})
	}

}

func TestGetPortExposure(t *testing.T) {
	urlName := "testurl"
	urlName2 := "testurl2"
	tests := []struct {
		name                string
		containerComponents []common.DevfileComponent
		wantMap             map[int32]common.ExposureType
		wantErr             bool
	}{
		{
			name: "Case 1: devfile has single container with single endpoint",
			wantMap: map[int32]common.ExposureType{
				8080: common.Public,
			},
			containerComponents: []common.DevfileComponent{
				{
					Name: "testcontainer1",
					Container: &common.Container{
						Image: "quay.io/nodejs-12",
						Endpoints: []common.Endpoint{
							{
								Name:       urlName,
								TargetPort: 8080,
								Exposure:   common.Public,
							},
						},
					},
				},
			},
		},
		{
			name:    "Case 2: devfile no endpoints",
			wantMap: map[int32]common.ExposureType{},
			containerComponents: []common.DevfileComponent{
				{
					Name: "testcontainer1",
					Container: &common.Container{
						Image: "quay.io/nodejs-12",
					},
				},
			},
		},
		{
			name: "Case 3: devfile has multiple endpoints with same port, 1 public and 1 internal, should assign public",
			wantMap: map[int32]common.ExposureType{
				8080: common.Public,
			},
			containerComponents: []common.DevfileComponent{
				{
					Name: "testcontainer1",
					Container: &common.Container{
						Image: "quay.io/nodejs-12",
						Endpoints: []common.Endpoint{
							{
								Name:       urlName,
								TargetPort: 8080,
								Exposure:   common.Public,
							},
							{
								Name:       urlName,
								TargetPort: 8080,
								Exposure:   common.Internal,
							},
						},
					},
				},
			},
		},
		{
			name: "Case 4: devfile has multiple endpoints with same port, 1 public and 1 none, should assign public",
			wantMap: map[int32]common.ExposureType{
				8080: common.Public,
			},
			containerComponents: []common.DevfileComponent{
				{
					Name: "testcontainer1",
					Container: &common.Container{
						Image: "quay.io/nodejs-12",
						Endpoints: []common.Endpoint{
							{
								Name:       urlName,
								TargetPort: 8080,
								Exposure:   common.Public,
							},
							{
								Name:       urlName,
								TargetPort: 8080,
								Exposure:   common.None,
							},
						},
					},
				},
			},
		},
		{
			name: "Case 5: devfile has multiple endpoints with same port, 1 internal and 1 none, should assign internal",
			wantMap: map[int32]common.ExposureType{
				8080: common.Internal,
			},
			containerComponents: []common.DevfileComponent{
				{
					Name: "testcontainer1",
					Container: &common.Container{
						Image: "quay.io/nodejs-12",
						Endpoints: []common.Endpoint{
							{
								Name:       urlName,
								TargetPort: 8080,
								Exposure:   common.Internal,
							},
							{
								Name:       urlName,
								TargetPort: 8080,
								Exposure:   common.None,
							},
						},
					},
				},
			},
		},
		{
			name: "Case 6: devfile has multiple endpoints with different port",
			wantMap: map[int32]common.ExposureType{
				8080: common.Public,
				9090: common.Internal,
				3000: common.None,
			},
			containerComponents: []common.DevfileComponent{
				{
					Name: "testcontainer1",
					Container: &common.Container{
						Image: "quay.io/nodejs-12",
						Endpoints: []common.Endpoint{
							{
								Name:       urlName,
								TargetPort: 8080,
							},
							{
								Name:       urlName,
								TargetPort: 3000,
								Exposure:   common.None,
							},
						},
					},
				},
				{
					Name: "testcontainer2",
					Container: &common.Container{
						Endpoints: []common.Endpoint{
							{
								Name:       urlName2,
								TargetPort: 9090,
								Secure:     true,
								Path:       "/testpath",
								Exposure:   common.Internal,
								Protocol:   common.HTTPS,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapCreated := GetPortExposure(tt.containerComponents)
			if !reflect.DeepEqual(mapCreated, tt.wantMap) {
				t.Errorf("Expected: %v, got %v", tt.wantMap, mapCreated)
			}

		})
	}

}

func TestAddSyncRootFolder(t *testing.T) {

	tests := []struct {
		name               string
		sourceMapping      string
		wantSyncRootFolder string
	}{
		{
			name:               "Case 1: Valid Source Mapping",
			sourceMapping:      "/mypath",
			wantSyncRootFolder: "/mypath",
		},
		{
			name:               "Case 2: No Source Mapping",
			sourceMapping:      "",
			wantSyncRootFolder: DevfileSourceVolumeMount,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := testingutil.CreateFakeContainer("container1")

			syncRootFolder := addSyncRootFolder(&container, tt.sourceMapping)

			if syncRootFolder != tt.wantSyncRootFolder {
				t.Errorf("TestAddSyncRootFolder sync root folder error - expected %v got %v", tt.wantSyncRootFolder, syncRootFolder)
			}

			for _, env := range container.Env {
				if env.Name == EnvProjectsRoot && env.Value != tt.wantSyncRootFolder {
					t.Errorf("PROJECT_ROOT error expected %s, actual %s", tt.wantSyncRootFolder, env.Value)
				}
			}
		})
	}
}

func TestAddSyncFolder(t *testing.T) {
	projectNames := []string{"some-name", "another-name"}
	projectRepos := []string{"https://github.com/some/repo.git", "https://github.com/another/repo.git"}
	projectClonePath := "src/github.com/golang/example/"
	invalidClonePaths := []string{"/var", "../var", "pkg/../../var"}
	sourceVolumePath := "/projects/app"

	tests := []struct {
		name     string
		projects []common.DevfileProject
		want     string
		wantErr  bool
	}{
		{
			name:     "Case 1: No projects",
			projects: []common.DevfileProject{},
			want:     sourceVolumePath,
			wantErr:  false,
		},
		{
			name: "Case 2: One project",
			projects: []common.DevfileProject{
				{
					Name: projectNames[0],
					Git: &common.Git{
						GitLikeProjectSource: common.GitLikeProjectSource{
							Remotes: map[string]string{"origin": projectRepos[0]},
						},
					},
				},
			},
			want:    filepath.ToSlash(filepath.Join(sourceVolumePath, projectNames[0])),
			wantErr: false,
		},
		{
			name: "Case 3: Multiple projects",
			projects: []common.DevfileProject{
				{
					Name: projectNames[0],
					Git: &common.Git{
						GitLikeProjectSource: common.GitLikeProjectSource{
							Remotes: map[string]string{"origin": projectRepos[0]},
						},
					},
				},
				{
					Name: projectNames[1],
					Github: &common.Github{
						GitLikeProjectSource: common.GitLikeProjectSource{
							Remotes: map[string]string{"origin": projectRepos[1]},
						},
					},
				},
				{
					Name: projectNames[1],
					Zip: &common.Zip{
						Location: projectRepos[1],
					},
				},
			},
			want:    filepath.ToSlash(filepath.Join(sourceVolumePath, projectNames[0])),
			wantErr: false,
		},
		{
			name: "Case 4: Clone path set",
			projects: []common.DevfileProject{
				{
					ClonePath: projectClonePath,
					Name:      projectNames[0],
					Zip: &common.Zip{
						Location: projectRepos[0],
					},
				},
			},
			want:    filepath.ToSlash(filepath.Join(sourceVolumePath, projectClonePath)),
			wantErr: false,
		},
		{
			name: "Case 5: Invalid clone path, set with absolute path",
			projects: []common.DevfileProject{
				{
					ClonePath: invalidClonePaths[0],
					Name:      projectNames[0],
					Github: &common.Github{
						GitLikeProjectSource: common.GitLikeProjectSource{
							Remotes: map[string]string{"origin": projectRepos[0]},
						},
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Case 6: Invalid clone path, starts with ..",
			projects: []common.DevfileProject{
				{
					ClonePath: invalidClonePaths[1],
					Name:      projectNames[0],
					Git: &common.Git{
						GitLikeProjectSource: common.GitLikeProjectSource{
							Remotes: map[string]string{"origin": projectRepos[0]},
						},
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Case 7: Invalid clone path, contains ..",
			projects: []common.DevfileProject{
				{
					ClonePath: invalidClonePaths[2],
					Name:      projectNames[0],
					Zip: &common.Zip{
						Location: projectRepos[0],
					},
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := testingutil.CreateFakeContainer("container1")

			err := addSyncFolder(&container, sourceVolumePath, tt.projects)

			if !tt.wantErr == (err != nil) {
				t.Errorf("expected %v, actual %v", tt.wantErr, err)
			}

			for _, env := range container.Env {
				if env.Name == EnvProjectsSrc && env.Value != tt.want {
					t.Errorf("expected %s, actual %s", tt.want, env.Value)
				}
			}
		})
	}
}

func TestGetContainer(t *testing.T) {

	tests := []struct {
		name          string
		containerName string
		image         string
		isPrivileged  bool
		command       []string
		args          []string
		envVars       []corev1.EnvVar
		resourceReqs  corev1.ResourceRequirements
		ports         []corev1.ContainerPort
	}{
		{
			name:          "Case 1: Empty container params",
			containerName: "",
			image:         "",
			isPrivileged:  false,
			command:       []string{},
			args:          []string{},
			envVars:       []corev1.EnvVar{},
			resourceReqs:  corev1.ResourceRequirements{},
			ports:         []corev1.ContainerPort{},
		},
		{
			name:          "Case 2: Valid container params",
			containerName: "container1",
			image:         "quay.io/eclipse/che-java8-maven:nightly",
			isPrivileged:  true,
			command:       []string{"tail"},
			args:          []string{"-f", "/dev/null"},
			envVars: []corev1.EnvVar{
				{
					Name:  "test",
					Value: "123",
				},
			},
			resourceReqs: fakeResources,
			ports: []corev1.ContainerPort{
				{
					Name:          "port-9090",
					ContainerPort: 9090,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containerParams := ContainerParams{
				Name:         tt.containerName,
				Image:        tt.image,
				IsPrivileged: tt.isPrivileged,
				Command:      tt.command,
				Args:         tt.args,
				EnvVars:      tt.envVars,
				ResourceReqs: tt.resourceReqs,
				Ports:        tt.ports,
			}
			container := getContainer(containerParams)

			if container.Name != tt.containerName {
				t.Errorf("expected %s, actual %s", tt.containerName, container.Name)
			}

			if container.Image != tt.image {
				t.Errorf("expected %s, actual %s", tt.image, container.Image)
			}

			if tt.isPrivileged {
				if *container.SecurityContext.Privileged != tt.isPrivileged {
					t.Errorf("expected %t, actual %t", tt.isPrivileged, *container.SecurityContext.Privileged)
				}
			} else if tt.isPrivileged == false && container.SecurityContext != nil {
				t.Errorf("expected security context to be nil but it was defined")
			}

			if len(container.Command) != len(tt.command) {
				t.Errorf("expected %d, actual %d", len(tt.command), len(container.Command))
			} else {
				for i := range container.Command {
					if container.Command[i] != tt.command[i] {
						t.Errorf("expected %s, actual %s", tt.command[i], container.Command[i])
					}
				}
			}

			if len(container.Args) != len(tt.args) {
				t.Errorf("expected %d, actual %d", len(tt.args), len(container.Args))
			} else {
				for i := range container.Args {
					if container.Args[i] != tt.args[i] {
						t.Errorf("expected %s, actual %s", tt.args[i], container.Args[i])
					}
				}
			}

			if len(container.Env) != len(tt.envVars) {
				t.Errorf("expected %d, actual %d", len(tt.envVars), len(container.Env))
			} else {
				for i := range container.Env {
					if container.Env[i].Name != tt.envVars[i].Name {
						t.Errorf("expected name %s, actual name %s", tt.envVars[i].Name, container.Env[i].Name)
					}
					if container.Env[i].Value != tt.envVars[i].Value {
						t.Errorf("expected value %s, actual value %s", tt.envVars[i].Value, container.Env[i].Value)
					}
				}
			}

			if len(container.Ports) != len(tt.ports) {
				t.Errorf("expected %d, actual %d", len(tt.ports), len(container.Ports))
			} else {
				for i := range container.Ports {
					if container.Ports[i].Name != tt.ports[i].Name {
						t.Errorf("expected name %s, actual name %s", tt.ports[i].Name, container.Ports[i].Name)
					}
					if container.Ports[i].ContainerPort != tt.ports[i].ContainerPort {
						t.Errorf("expected port number is %v, actual %v", tt.ports[i].ContainerPort, container.Ports[i].ContainerPort)
					}
				}
			}

		})
	}
}

func TestGenerateServiceSpec(t *testing.T) {
	port1 := corev1.ContainerPort{
		Name:          "port-9090",
		ContainerPort: 9090,
	}
	port2 := corev1.ContainerPort{
		Name:          "port-8080",
		ContainerPort: 8080,
	}

	tests := []struct {
		name  string
		ports []corev1.ContainerPort
	}{
		{
			name:  "singlePort",
			ports: []corev1.ContainerPort{port1},
		},
		{
			name:  "multiplePorts",
			ports: []corev1.ContainerPort{port1, port2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceSpecParams := ServiceSpecParams{
				ContainerPorts: tt.ports,
				SelectorLabels: map[string]string{
					"component": tt.name,
				},
			}
			serviceSpec := getServiceSpec(serviceSpecParams)

			if len(serviceSpec.Ports) != len(tt.ports) {
				t.Errorf("expected service ports length is %v, actual %v", len(tt.ports), len(serviceSpec.Ports))
			} else {
				for i := range serviceSpec.Ports {
					if serviceSpec.Ports[i].Name != tt.ports[i].Name {
						t.Errorf("expected name %s, actual name %s", tt.ports[i].Name, serviceSpec.Ports[i].Name)
					}
					if serviceSpec.Ports[i].Port != tt.ports[i].ContainerPort {
						t.Errorf("expected port number is %v, actual %v", tt.ports[i].ContainerPort, serviceSpec.Ports[i].Port)
					}
				}
			}

		})
	}
}
