package common

import (
	"testing"

	"github.com/openshift/odo/pkg/devfile"
	versionsCommon "github.com/openshift/odo/pkg/devfile/versions/common"
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
			devObj := devfile.DevfileObj{
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
