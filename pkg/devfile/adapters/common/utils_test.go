package common

import (
	"os"
	"testing"

	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/openshift/odo/pkg/util"
)

func TestGetSupportedComponents(t *testing.T) {

	tests := []struct {
		name                 string
		componentType        versionsCommon.DevfileComponentType
		alias                []string
		expectedMatchesCount int
	}{
		{
			name:                 "Case: Invalid devfile",
			componentType:        "",
			expectedMatchesCount: 0,
		},
		{
			name:                 "Case: Valid devfile with wrong component type (CheEditor)",
			componentType:        versionsCommon.DevfileComponentTypeCheEditor,
			alias:                []string{"alias1", "alias2"},
			expectedMatchesCount: 0,
		},
		{
			name:                 "Case: Valid devfile with wrong component type (ChePlugin)",
			componentType:        versionsCommon.DevfileComponentTypeChePlugin,
			alias:                []string{"alias1", "alias2"},
			expectedMatchesCount: 0,
		},
		{
			name:                 "Case: Valid devfile with wrong component type (Kubernetes)",
			componentType:        versionsCommon.DevfileComponentTypeKubernetes,
			alias:                []string{"alias1", "alias2"},
			expectedMatchesCount: 0,
		},
		{
			name:                 "Case: Valid devfile with wrong component type (Openshift)",
			componentType:        versionsCommon.DevfileComponentTypeOpenshift,
			alias:                []string{"alias1", "alias2"},
			expectedMatchesCount: 0,
		},
		{
			name:                 "Case: Valid devfile with correct component type (Dockerimage)",
			componentType:        versionsCommon.DevfileComponentTypeDockerimage,
			alias:                []string{"alias1", "alias2"},
			expectedMatchesCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}

			devfileComponents := GetSupportedComponents(devObj.Data)

			componentsMatched := 0
			for _, component := range devfileComponents {
				if component.Type != versionsCommon.DevfileComponentTypeDockerimage {
					t.Errorf("TestGetSupportedComponents error: wrong component type expected %v, actual %v", versionsCommon.DevfileComponentTypeDockerimage, component.Type)
				}
				if util.In(tt.alias, *component.Alias) {
					componentsMatched++
				}
			}

			if componentsMatched != tt.expectedMatchesCount {
				t.Errorf("TestGetSupportedComponents error: wrong number of components matched: expected %v, actual %v", tt.expectedMatchesCount, componentsMatched)
			}
		})
	}

}

func TestIsEnvPresent(t *testing.T) {

	envName := "myenv"
	envValue := "myenvvalue"

	envVars := []common.DockerimageEnv{
		{
			Name:  &envName,
			Value: &envValue,
		},
	}

	tests := []struct {
		name          string
		envVarName    string
		wantIsPresent bool
	}{
		{
			name:          "Case 1: Env var present",
			envVarName:    envName,
			wantIsPresent: true,
		},
		{
			name:          "Case 2: Env var absent",
			envVarName:    "someenv",
			wantIsPresent: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isPresent := IsEnvPresent(envVars, tt.envVarName)
			if isPresent != tt.wantIsPresent {
				t.Errorf("TestIsEnvPresent error: env var expectation mismatch, want: %v got: %v", tt.wantIsPresent, isPresent)
			}
		})
	}

}

func TestIsPortPresent(t *testing.T) {

	endpointName := "8080/tcp"
	var endpointPort int32 = 8080

	endpoints := []common.DockerimageEndpoint{
		{
			Name: &endpointName,
			Port: &endpointPort,
		},
	}

	tests := []struct {
		name          string
		port          int
		wantIsPresent bool
	}{
		{
			name:          "Case 1: Endpoint port present",
			port:          8080,
			wantIsPresent: true,
		},
		{
			name:          "Case 2: Endpoint port absent",
			port:          1234,
			wantIsPresent: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isPresent := IsPortPresent(endpoints, tt.port)
			if isPresent != tt.wantIsPresent {
				t.Errorf("TestIsPortPresent error: endpoint port expectation mismatch, want: %v got: %v", tt.wantIsPresent, isPresent)
			}
		})
	}

}

func TestGetBootstrapperImage(t *testing.T) {

	customImage := "customimage:customtag"

	tests := []struct {
		name        string
		customImage bool
		wantImage   string
	}{
		{
			name:        "Case 1: Default bootstrap image",
			customImage: false,
			wantImage:   defaultBootstrapperImage,
		},
		{
			name:        "Case 2: Custom bootstrap image",
			customImage: true,
			wantImage:   customImage,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.customImage {
				os.Setenv(bootstrapperImageEnvName, customImage)
			}
			image := GetBootstrapperImage()

			if image != tt.wantImage {
				t.Errorf("TestGetBootstrapperImage error: bootstrap image mismatch, expected: %v got: %v", tt.wantImage, image)
			}
		})
	}

}

func TestIsComponentSupported(t *testing.T) {

	tests := []struct {
		name            string
		component       common.DevfileComponent
		wantIsSupported bool
	}{
		{
			name: "Case 1: Supported component",
			component: common.DevfileComponent{
				Type: versionsCommon.DevfileComponentTypeDockerimage,
			},
			wantIsSupported: true,
		},
		{
			name: "Case 2: Unsupported component",
			component: common.DevfileComponent{
				Type: versionsCommon.DevfileComponentTypeCheEditor,
			},
			wantIsSupported: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isSupported := isComponentSupported(tt.component)
			if isSupported != tt.wantIsSupported {
				t.Errorf("TestIsComponentSupported error: component support mismatch, expected: %v got: %v", tt.wantIsSupported, isSupported)
			}
		})
	}

}
