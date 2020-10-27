package generator

import (
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
