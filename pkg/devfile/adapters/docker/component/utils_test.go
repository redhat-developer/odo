package component

import (
	"testing"

	"github.com/docker/go-connections/nat"

	"github.com/docker/docker/api/types/mount"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	envinfo "github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/lclient"
	"github.com/openshift/odo/pkg/testingutil"
)

func TestCreateComponent(t *testing.T) {

	testComponentName := "test"
	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name          string
		componentType versionsCommon.DevfileComponentType
		client        *lclient.Client
		wantErr       bool
	}{
		{
			name:          "Case 1: Invalid devfile",
			componentType: "",
			client:        fakeClient,
			wantErr:       true,
		},
		{
			name:          "Case 2: Valid devfile",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			client:        fakeClient,
			wantErr:       false,
		},
		{
			name:          "Case 3: Valid devfile, docker client error",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			client:        fakeErrorClient,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			componentAdapter := New(adapterCtx, *tt.client)
			err := componentAdapter.createComponent()

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter create unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestUpdateComponent(t *testing.T) {

	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name          string
		componentType versionsCommon.DevfileComponentType
		componentName string
		client        *lclient.Client
		wantErr       bool
	}{
		{
			name:          "Case 1: Invalid devfile",
			componentType: "",
			componentName: "",
			client:        fakeClient,
			wantErr:       true,
		},
		{
			name:          "Case 2: Valid devfile",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			componentName: "test",
			client:        fakeClient,
			wantErr:       false,
		},
		{
			name:          "Case 3: Valid devfile, docker client error",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			componentName: "",
			client:        fakeErrorClient,
			wantErr:       true,
		},
		{
			name:          "Case 3: Valid devfile, missing component",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			componentName: "fakecomponent",
			client:        fakeClient,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: tt.componentName,
				Devfile:       devObj,
			}

			componentAdapter := New(adapterCtx, *tt.client)
			err := componentAdapter.updateComponent()

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter update unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestPullAndStartContainer(t *testing.T) {

	testComponentName := "test"
	testVolumeName := "projects"

	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name          string
		componentType versionsCommon.DevfileComponentType
		client        *lclient.Client
		mounts        []mount.Mount
		wantErr       bool
	}{
		{
			name:          "Case 1: Successfully start container, no mount",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			client:        fakeClient,
			mounts:        []mount.Mount{},
			wantErr:       false,
		},
		{
			name:          "Case 2: Docker client error",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			client:        fakeErrorClient,
			mounts:        []mount.Mount{},
			wantErr:       true,
		},
		{
			name:          "Case 3: Successfully start container, one mount",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			client:        fakeClient,
			mounts: []mount.Mount{
				{
					Source: "test-vol",
					Target: "/path",
				},
			},
			wantErr: false,
		},
		{
			name:          "Case 4: Successfully start container, multiple mounts",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			client:        fakeClient,
			mounts: []mount.Mount{
				{
					Source: "test-vol",
					Target: "/path",
				},
				{
					Source: "test-vol-two",
					Target: "/path-two",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			componentAdapter := New(adapterCtx, *tt.client)
			err := componentAdapter.pullAndStartContainer(tt.mounts, testVolumeName, adapterCtx.Devfile.Data.GetAliasedComponents()[0])

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter create unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestStartContainer(t *testing.T) {

	testComponentName := "test"
	testVolumeName := "projects"

	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name          string
		componentType versionsCommon.DevfileComponentType
		client        *lclient.Client
		mounts        []mount.Mount
		wantErr       bool
	}{
		{
			name:          "Case 1: Successfully start container, no mount",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			client:        fakeClient,
			mounts:        []mount.Mount{},
			wantErr:       false,
		},
		{
			name:          "Case 2: Docker client error",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			client:        fakeErrorClient,
			mounts:        []mount.Mount{},
			wantErr:       true,
		},
		{
			name:          "Case 3: Successfully start container, one mount",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			client:        fakeClient,
			mounts: []mount.Mount{
				{
					Source: "test-vol",
					Target: "/path",
				},
			},
			wantErr: false,
		},
		{
			name:          "Case 4: Successfully start container, multiple mount",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			client:        fakeClient,
			mounts: []mount.Mount{
				{
					Source: "test-vol",
					Target: "/path",
				},
				{
					Source: "test-vol-two",
					Target: "/path-two",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			componentAdapter := New(adapterCtx, *tt.client)
			err := componentAdapter.startComponent(tt.mounts, testVolumeName, adapterCtx.Devfile.Data.GetAliasedComponents()[0])

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter create unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestGenerateAndGetHostConfig(t *testing.T) {
	fakeClient := lclient.FakeNew()
	testComponentName := "test"
	componentType := versionsCommon.DevfileComponentTypeDockerimage

	tests := []struct {
		name         string
		urlValue     []envinfo.EnvInfoURL
		expectResult nat.PortMap
		client       *lclient.Client
	}{
		{
			name:         "Case 1: no port mappings",
			urlValue:     []envinfo.EnvInfoURL{},
			expectResult: nil,
			client:       fakeClient,
		},
		{
			name: "Case 2: only one port mapping",
			urlValue: []envinfo.EnvInfoURL{
				{Port: 8080, ExposedPort: 65432},
			},
			expectResult: nat.PortMap{
				"8080/tcp": []nat.PortBinding{
					{
						HostIP:   LocalhostIP,
						HostPort: "65432",
					},
				},
			},
			client: fakeClient,
		},
		{
			name: "Case 3: multiple port mappings",
			urlValue: []envinfo.EnvInfoURL{
				{Port: 8080, ExposedPort: 65432},
				{Port: 9090, ExposedPort: 54321},
				{Port: 9080, ExposedPort: 45678},
			},
			expectResult: nat.PortMap{
				"8080/tcp": []nat.PortBinding{
					{
						HostIP:   LocalhostIP,
						HostPort: "65432",
					},
				},
				"9090/tcp": []nat.PortBinding{
					{
						HostIP:   LocalhostIP,
						HostPort: "54321",
					},
				},
				"9080/tcp": []nat.PortBinding{
					{
						HostIP:   LocalhostIP,
						HostPort: "45678",
					},
				},
			},
			client: fakeClient,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: componentType,
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			esi, err := envinfo.NewEnvSpecificInfo("")
			if err != nil {
				t.Error(err)
			}
			for _, element := range tt.urlValue {
				err = esi.SetConfiguration("URL", element)
				if err != nil {
					t.Error(err)
				}
			}
			componentAdapter := New(adapterCtx, *tt.client)
			hostConfig, err := componentAdapter.generateAndGetHostConfig()
			if err != nil {
				t.Error(err)
			}

			if len(hostConfig.PortBindings) != len(tt.expectResult) {
				t.Errorf("host config PortBindings length mismatch: actual value %v, expected value %v", len(hostConfig.PortBindings), len(tt.expectResult))
			}
			if len(hostConfig.PortBindings) != 0 {
				for key, value := range hostConfig.PortBindings {
					if tt.expectResult[key][0].HostIP != value[0].HostIP || tt.expectResult[key][0].HostPort != value[0].HostPort {
						t.Errorf("host config PortBindings mismatch: actual value %v, expected value %v", hostConfig.PortBindings, tt.expectResult)
					}
				}
			}
			err = esi.DeleteEnvInfoFile()
			if err != nil {
				t.Error(err)
			}
		})
	}
}
