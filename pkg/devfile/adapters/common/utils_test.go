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
		name          string
		componentType versionsCommon.DevfileComponentType
		alias         []string
	}{
		{
			name:          "Case: Invalid devfile",
			componentType: "",
		},
		{
			name:          "Case: Valid devfile",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			alias:         []string{"alias1", "alias2"},
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

			if componentsMatched != len(tt.alias) {
				t.Errorf("TestGetSupportedComponents error: wrong number of components matched: expected %v, actual %v", len(tt.alias), componentsMatched)
			}
		})
	}

}
