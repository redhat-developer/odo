package backend

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/alizer"
	"github.com/redhat-developer/odo/pkg/init/asker"
	"github.com/redhat-developer/odo/pkg/registry"
	"github.com/redhat-developer/odo/pkg/testingutil"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	parsercontext "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/devfile/library/pkg/testingutil/filesystem"
)

func TestInteractiveBackend_SelectDevfile(t *testing.T) {
	type fields struct {
		buildAsker         func(ctrl *gomock.Controller) asker.Asker
		buildCatalogClient func(ctrl *gomock.Controller) registry.Client
	}
	tests := []struct {
		name    string
		fields  fields
		want    *alizer.DevfileLocation
		wantErr bool
	}{
		{
			name: "direct selection",
			fields: fields{
				buildAsker: func(ctrl *gomock.Controller) asker.Asker {
					client := asker.NewMockAsker(ctrl)
					client.EXPECT().AskLanguage(gomock.Any()).Return("java", nil)
					client.EXPECT().AskType(gomock.Any()).Return(false, registry.DevfileStack{
						Name: "a-devfile-name",
						Registry: registry.Registry{
							Name: "MyRegistry1",
						},
					}, nil)
					return client
				},
				buildCatalogClient: func(ctrl *gomock.Controller) registry.Client {
					client := registry.NewMockClient(ctrl)
					client.EXPECT().ListDevfileStacks(gomock.Any())
					return client
				},
			},
			want: &alizer.DevfileLocation{
				Devfile:         "a-devfile-name",
				DevfileRegistry: "MyRegistry1",
			},
		},
		{
			name: "selection with back",
			fields: fields{
				buildAsker: func(ctrl *gomock.Controller) asker.Asker {
					client := asker.NewMockAsker(ctrl)
					client.EXPECT().AskLanguage(gomock.Any()).Return("java", nil)
					client.EXPECT().AskType(gomock.Any()).Return(true, registry.DevfileStack{}, nil)
					client.EXPECT().AskLanguage(gomock.Any()).Return("go", nil)
					client.EXPECT().AskType(gomock.Any()).Return(false, registry.DevfileStack{
						Name: "a-devfile-name",
						Registry: registry.Registry{
							Name: "MyRegistry1",
						},
					}, nil)
					return client
				},
				buildCatalogClient: func(ctrl *gomock.Controller) registry.Client {
					client := registry.NewMockClient(ctrl)
					client.EXPECT().ListDevfileStacks(gomock.Any())
					return client
				},
			},
			want: &alizer.DevfileLocation{
				Devfile:         "a-devfile-name",
				DevfileRegistry: "MyRegistry1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			o := &InteractiveBackend{
				askerClient:    tt.fields.buildAsker(ctrl),
				registryClient: tt.fields.buildCatalogClient(ctrl),
			}
			got, err := o.SelectDevfile(map[string]string{}, nil, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("InteractiveBuilder.ParamsBuild() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InteractiveBuilder.ParamsBuild() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInteractiveBackend_SelectStarterProject(t *testing.T) {
	type fields struct {
		asker          func(ctrl *gomock.Controller) asker.Asker
		registryClient registry.Client
	}
	type args struct {
		devfile func() parser.DevfileObj
		flags   map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *v1alpha2.StarterProject
		wantErr bool
	}{
		{
			name: "no flags, no starter selected",
			fields: fields{
				asker: func(ctrl *gomock.Controller) asker.Asker {
					client := asker.NewMockAsker(ctrl)
					client.EXPECT().AskStarterProject(gomock.Any()).Return(false, 0, nil)
					return client
				},
			},
			args: args{
				devfile: func() parser.DevfileObj {
					devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					return parser.DevfileObj{
						Data: devfileData,
					}
				},
				flags: map[string]string{},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "no flags, starter selected",
			fields: fields{
				asker: func(ctrl *gomock.Controller) asker.Asker {
					client := asker.NewMockAsker(ctrl)
					client.EXPECT().AskStarterProject(gomock.Any()).Return(true, 1, nil)
					return client
				},
			},
			args: args{
				devfile: func() parser.DevfileObj {
					devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = devfileData.AddStarterProjects([]v1alpha2.StarterProject{
						{
							Name: "starter1",
						},
						{
							Name: "starter2",
						},
						{
							Name: "starter3",
						},
					})
					return parser.DevfileObj{
						Data: devfileData,
					}
				},
				flags: map[string]string{},
			},
			want: &v1alpha2.StarterProject{
				Name: "starter2",
			},
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			var askerClient asker.Asker
			if tt.fields.asker != nil {
				askerClient = tt.fields.asker(ctrl)
			}
			o := &InteractiveBackend{
				askerClient:    askerClient,
				registryClient: tt.fields.registryClient,
			}
			got1, err := o.SelectStarterProject(tt.args.devfile(), tt.args.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("InteractiveBackend.SelectStarterProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got1, tt.want) {
				t.Errorf("InteractiveBackend.SelectStarterProject() got1 = %v, want %v", got1, tt.want)
			}
		})
	}
}

func TestInteractiveBackend_PersonalizeName(t *testing.T) {
	type fields struct {
		asker          func(ctrl *gomock.Controller) asker.Asker
		registryClient registry.Client
	}
	type args struct {
		devfile func(fs filesystem.Filesystem) parser.DevfileObj
		flags   map[string]string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		checkResult func(newName string, args args) bool
	}{
		{
			name: "no flag",
			fields: fields{
				asker: func(ctrl *gomock.Controller) asker.Asker {
					client := asker.NewMockAsker(ctrl)
					client.EXPECT().AskName(gomock.Any()).Return("aname", nil)
					return client
				},
			},
			args: args{
				devfile: func(fs filesystem.Filesystem) parser.DevfileObj {
					devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					obj := parser.DevfileObj{
						Ctx:  parsercontext.FakeContext(fs, "/tmp/devfile.yaml"),
						Data: devfileData,
					}
					return obj
				},
				flags: map[string]string{},
			},
			wantErr: false,
			checkResult: func(newName string, args args) bool {
				return newName == "aname"
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			var askerClient asker.Asker
			if tt.fields.asker != nil {
				askerClient = tt.fields.asker(ctrl)
			}
			o := &InteractiveBackend{
				askerClient:    askerClient,
				registryClient: tt.fields.registryClient,
			}
			fs := filesystem.NewFakeFs()
			newName, err := o.PersonalizeName(tt.args.devfile(fs), tt.args.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("InteractiveBackend.PersonalizeName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkResult != nil && !tt.checkResult(newName, tt.args) {
				t.Errorf("InteractiveBackend.PersonalizeName(), checking result failed")
			}
		})
	}
}

func TestInteractiveBackend_PersonalizeDevfileconfig(t *testing.T) {
	container1 := "runtime"

	type fields struct {
		asker          func(ctrl *gomock.Controller, configuration asker.DevfileConfiguration) asker.Asker
		registryClient registry.Client
	}
	type args struct {
		devfileobj func(fs filesystem.Filesystem) parser.DevfileObj
		key        string
		value      string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		checkResult func(config asker.ContainerConfiguration, key string, value string) bool
	}{
		// TODO: Add test cases.
		{
			name: "Add new port",
			fields: fields{
				asker: func(ctrl *gomock.Controller, configuration asker.DevfileConfiguration) asker.Asker {
					client := asker.NewMockAsker(ctrl)
					client.EXPECT().AskContainerName(append(configuration.GetContainers(), "NONE - configuration is correct")).Return(container1, nil)
					containerConfig := configuration[container1]
					selectContainer := client.EXPECT().AskPersonalizeConfiguration(containerConfig).Return(asker.OperationOnContainer{
						Ops:  "Add",
						Kind: "Port",
					}, nil).MaxTimes(1)
					addPort := client.EXPECT().AskAddPort().Return("5000", nil).After(selectContainer)
					containerConfig.Ports = append(containerConfig.Ports, "5000")
					containerConfigDone := client.EXPECT().AskPersonalizeConfiguration(containerConfig).Return(asker.OperationOnContainer{Ops: "Nothing"}, nil).After(addPort)
					client.EXPECT().AskContainerName(append(configuration.GetContainers(), "NONE - configuration is correct")).Return("NONE - configuration is correct", nil).After(containerConfigDone)
					return client
				},
				registryClient: nil,
			},
			args: args{
				key: "5000",
				devfileobj: func(fs filesystem.Filesystem) parser.DevfileObj {
					ports := []string{"7000", "8000"}
					envVars := []v1alpha2.EnvVar{{Name: "env1", Value: "val1"}, {Name: "env2", Value: "val2"}}
					return getDevfileObj(fs, container1, ports, envVars)
				},
			},
			wantErr: false,
			checkResult: func(config asker.ContainerConfiguration, key string, value string) bool {
				for _, port := range config.Ports {
					if port == key {
						return true
					}
				}
				return false
			},
		},
		{
			name: "Add new environment variable",
			fields: fields{
				asker: func(ctrl *gomock.Controller, configuration asker.DevfileConfiguration) asker.Asker {
					client := asker.NewMockAsker(ctrl)
					askContainerName := client.EXPECT().AskContainerName(append(configuration.GetContainers(), "NONE - configuration is correct")).Return(container1, nil)
					containerConfig := configuration[container1]
					selectContainer := client.EXPECT().AskPersonalizeConfiguration(containerConfig).Return(asker.OperationOnContainer{
						Ops:  "Add",
						Kind: "EnvVar",
					}, nil).After(askContainerName)
					key, val := "env3", "val3"
					addEnvVar := client.EXPECT().AskAddEnvVar().Return(key, val, nil).After(selectContainer)
					// containerConfig.Envs[key] = val
					containerConfigDone := client.EXPECT().AskPersonalizeConfiguration(gomock.Any()).Return(asker.OperationOnContainer{Ops: "Nothing"}, nil).After(addEnvVar)
					client.EXPECT().AskContainerName(append(configuration.GetContainers(), "NONE - configuration is correct")).Return("NONE - configuration is correct", nil).After(containerConfigDone)
					return client
				},
				registryClient: nil,
			},
			args: args{
				devfileobj: func(fs filesystem.Filesystem) parser.DevfileObj {
					ports := []string{"7000", "8000"}
					envVars := []v1alpha2.EnvVar{{Name: "env1", Value: "val1"}, {Name: "env2", Value: "val2"}}
					return getDevfileObj(fs, container1, ports, envVars)
				},
				key:   "env3",
				value: "val3",
			},
			wantErr: false,
			checkResult: func(config asker.ContainerConfiguration, key string, value string) bool {
				if val, ok := config.Envs[key]; ok && val == value {
					return true
				}
				return false
			},
		},
		{
			name: "Delete port",
			fields: fields{
				asker: func(ctrl *gomock.Controller, configuration asker.DevfileConfiguration) asker.Asker {
					client := asker.NewMockAsker(ctrl)
					client.EXPECT().AskContainerName(append(configuration.GetContainers(), "NONE - configuration is correct")).Return(container1, nil)
					containerConfig := configuration[container1]
					selectContainer := client.EXPECT().AskPersonalizeConfiguration(containerConfig).Return(asker.OperationOnContainer{
						Ops:  "Delete",
						Kind: "Port",
						Key:  "7000",
					}, nil).MaxTimes(1)
					containerConfig.Ports = []string{"8000"}
					containerConfigDone := client.EXPECT().AskPersonalizeConfiguration(containerConfig).Return(asker.OperationOnContainer{Ops: "Nothing"}, nil).After(selectContainer)
					client.EXPECT().AskContainerName(append(configuration.GetContainers(), "NONE - configuration is correct")).Return("NONE - configuration is correct", nil).After(containerConfigDone)
					return client
				},
				registryClient: nil,
			},
			args: args{
				devfileobj: func(fs filesystem.Filesystem) parser.DevfileObj {
					ports := []string{"7000", "8000"}
					envVars := []v1alpha2.EnvVar{{Name: "env1", Value: "val1"}, {Name: "env2", Value: "val2"}}
					return getDevfileObj(fs, container1, ports, envVars)
				},
				key:   "7000",
				value: "",
			},
			checkResult: func(config asker.ContainerConfiguration, key string, value string) bool {
				for _, port := range config.Ports {
					if port == key {
						return false
					}
				}
				return true
			},
			wantErr: false,
		},
		{
			name: "Delete environment variable",
			fields: fields{
				asker: func(ctrl *gomock.Controller, configuration asker.DevfileConfiguration) asker.Asker {
					client := asker.NewMockAsker(ctrl)
					client.EXPECT().AskContainerName(append(configuration.GetContainers(), "NONE - configuration is correct")).Return(container1, nil)
					containerConfig := configuration[container1]
					selectContainer := client.EXPECT().AskPersonalizeConfiguration(containerConfig).Return(asker.OperationOnContainer{
						Ops:  "Delete",
						Kind: "EnvVar",
						Key:  "env2",
					}, nil).MaxTimes(1)

					// delete(containerConfig.Envs, "env2")
					containerConfigDone := client.EXPECT().AskPersonalizeConfiguration(gomock.Any()).Return(asker.OperationOnContainer{Ops: "Nothing"}, nil).After(selectContainer)
					client.EXPECT().AskContainerName(append(configuration.GetContainers(), "NONE - configuration is correct")).Return("NONE - configuration is correct", nil).After(containerConfigDone)
					return client
				},
				registryClient: nil,
			},
			args: args{
				devfileobj: func(fs filesystem.Filesystem) parser.DevfileObj {
					ports := []string{"7000", "8000"}
					envVars := []v1alpha2.EnvVar{{Name: "env1", Value: "val1"}, {Name: "env2", Value: "val2"}}
					return getDevfileObj(fs, container1, ports, envVars)
				},
				key:   "env2",
				value: "",
			},
			wantErr: false,
			checkResult: func(config asker.ContainerConfiguration, key string, value string) bool {
				if _, ok := config.Envs[key]; ok {
					return false
				}
				return true
			},
		},
		{
			name: "None - Configuration is correct",
			fields: fields{
				asker: func(ctrl *gomock.Controller, configuration asker.DevfileConfiguration) asker.Asker {
					client := asker.NewMockAsker(ctrl)
					client.EXPECT().AskContainerName(append(configuration.GetContainers(), "NONE - configuration is correct")).Return("NONE - configuration is correct", nil)
					containerConfig := configuration[container1]
					client.EXPECT().AskPersonalizeConfiguration(containerConfig).Return(asker.OperationOnContainer{
						Ops: "Nothing",
					}, nil).MaxTimes(1)
					return client
				},
				registryClient: nil,
			},
			args: args{
				devfileobj: func(fs filesystem.Filesystem) parser.DevfileObj {
					ports := []string{"7000", "8000"}
					envVars := []v1alpha2.EnvVar{{Name: "env1", Value: "val1"}, {Name: "env2", Value: "val2"}}
					return getDevfileObj(fs, container1, ports, envVars)
				},
				key:   "",
				value: "",
			},
			wantErr: false,
			checkResult: func(config asker.ContainerConfiguration, key string, value string) bool {
				checkConfig := asker.ContainerConfiguration{
					Ports: []string{"7000", "8000"},
					Envs:  map[string]string{"env1": "val1", "env2": "val2"},
				}
				return reflect.DeepEqual(config, checkConfig)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := filesystem.NewFakeFs()
			devfile := tt.args.devfileobj(fs)
			config, err := getPortsAndEnvVar(devfile)
			if err != nil {
				t.Errorf("getPortsAndEnvVar() error = %v", err)
			}

			ctrl := gomock.NewController(t)
			var askerClient asker.Asker
			if tt.fields.asker != nil {
				askerClient = tt.fields.asker(ctrl, config)
			}

			o := &InteractiveBackend{
				askerClient:    askerClient,
				registryClient: tt.fields.registryClient,
			}
			devfile, err = o.PersonalizeDevfileConfig(devfile)
			if (err != nil) != tt.wantErr {
				t.Errorf("PersonalizeDevfileConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			config, err = getPortsAndEnvVar(devfile)
			if err != nil {
				t.Errorf("getPortsAndEnvVar() error = %v", err)
			}
			if tt.checkResult != nil && !tt.checkResult(config[container1], tt.args.key, tt.args.value) {
				t.Errorf("InteractiveBackend.PersonalizeName(), checking result failed")
			}
		})
	}
}

func getDevfileObj(fs filesystem.Filesystem, containerName string, ports []string, envVars []v1alpha2.EnvVar) parser.DevfileObj {
	devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
	_ = devfileData.AddComponents([]v1alpha2.Component{
		testingutil.GetFakeContainerComponent(containerName),
	})
	obj := parser.DevfileObj{
		Ctx:  parsercontext.FakeContext(fs, "/tmp/devfile.yaml"),
		Data: devfileData,
	}
	_ = obj.SetPorts(map[string][]string{containerName: ports})
	_ = obj.AddEnvVars(map[string][]v1alpha2.EnvVar{containerName: envVars})
	return obj
}
