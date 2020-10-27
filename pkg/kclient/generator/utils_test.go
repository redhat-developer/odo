package generator

import (
	"reflect"
	"testing"

	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/testingutil"
)

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

		{
			name:                 "Case 5: Valid devfile with correct component type (Container) without name",
			component:            []common.DevfileComponent{testingutil.GetFakeContainerComponent("comp1"), testingutil.GetFakeContainerComponent("")},
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

			devfileComponents := GetDevfileContainerComponents(devObj.Data)

			if len(devfileComponents) != tt.expectedMatchesCount {
				t.Errorf("TestGetDevfileContainerComponents error: wrong number of components matched: expected %v, actual %v", tt.expectedMatchesCount, len(devfileComponents))
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
