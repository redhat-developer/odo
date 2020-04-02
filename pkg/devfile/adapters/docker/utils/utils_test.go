package utils

import (
	"reflect"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/openshift/odo/pkg/devfile/versions/common"
	"github.com/openshift/odo/pkg/lclient"
)

func TestComponentExists(t *testing.T) {
	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name          string
		componentName string
		client        *lclient.Client
		want          bool
	}{
		{
			name:          "Case 1: Component exists",
			componentName: "golang",
			client:        fakeClient,
			want:          true,
		},
		{
			name:          "Case 2: Component doesn't exist",
			componentName: "fakecomponent",
			client:        fakeClient,
			want:          false,
		},
		{
			name:          "Case 3: Error with docker client",
			componentName: "golang",
			client:        fakeErrorClient,
			want:          false,
		},
	}

	for _, tt := range tests {
		cmpExists := ComponentExists(*tt.client, tt.componentName)
		if tt.want != cmpExists {
			t.Errorf("expected %v, wanted %v", cmpExists, tt.want)
		}
	}
}

func TestConvertEnvs(t *testing.T) {
	envVarsNames := []string{"test", "sample-var", "myvar"}
	envVarsValues := []string{"value1", "value2", "value3"}
	tests := []struct {
		name    string
		envVars []common.DockerimageEnv
		want    []string
	}{
		{
			name: "Case 1: One env var",
			envVars: []common.DockerimageEnv{
				{
					Name:  &envVarsNames[0],
					Value: &envVarsValues[0],
				},
			},
			want: []string{"test=value1"},
		},
		{
			name: "Case 2: Multiple env vars",
			envVars: []common.DockerimageEnv{
				{
					Name:  &envVarsNames[0],
					Value: &envVarsValues[0],
				},
				{
					Name:  &envVarsNames[1],
					Value: &envVarsValues[1],
				},
				{
					Name:  &envVarsNames[2],
					Value: &envVarsValues[2],
				},
			},
			want: []string{"test=value1", "sample-var=value2", "myvar=value3"},
		},
		{
			name:    "Case 3: No env vars",
			envVars: []common.DockerimageEnv{},
			want:    []string{},
		},
	}

	for _, tt := range tests {
		envVars := ConvertEnvs(tt.envVars)
		if !reflect.DeepEqual(tt.want, envVars) {
			t.Errorf("expected %v, wanted %v", envVars, tt.want)
		}
	}
}

func TestDoesContainerNeedUpdating(t *testing.T) {
	envVarsNames := []string{"test", "sample-var", "myvar"}
	envVarsValues := []string{"value1", "value2", "value3"}
	tests := []struct {
		name            string
		envVars         []common.DockerimageEnv
		image           string
		containerConfig container.Config
		want            bool
	}{
		{
			name: "Case 1: No changes",
			envVars: []common.DockerimageEnv{
				{
					Name:  &envVarsNames[0],
					Value: &envVarsValues[0],
				},
				{
					Name:  &envVarsNames[1],
					Value: &envVarsValues[1],
				},
			},
			image: "golang",
			containerConfig: container.Config{
				Image: "golang",
				Env:   []string{"test=value1", "sample-var=value2"},
			},
			want: false,
		},
		{
			name: "Case 2: Update required, env var changed",
			envVars: []common.DockerimageEnv{
				{
					Name:  &envVarsNames[2],
					Value: &envVarsValues[2],
				},
			},
			image: "golang",
			containerConfig: container.Config{
				Image: "golang",
				Env:   []string{"test=value1", "sample-var=value2"},
			},
			want: true,
		},
		{
			name: "Case 2: Update required, image changed",
			envVars: []common.DockerimageEnv{
				{
					Name:  &envVarsNames[2],
					Value: &envVarsValues[2],
				},
			},
			image: "node",
			containerConfig: container.Config{
				Image: "golang",
				Env:   []string{"test=value1", "sample-var=value2"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		component := common.DevfileComponent{
			DevfileComponentDockerimage: common.DevfileComponentDockerimage{
				Image: &tt.image,
				Env:   tt.envVars,
			},
		}
		needsUpdating := DoesContainerNeedUpdating(component, &tt.containerConfig)
		if needsUpdating != tt.want {
			t.Errorf("expected %v, wanted %v", needsUpdating, tt.want)
		}
	}
}
