package adapters

import (
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	parsercommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
)

type PlatformAdapter interface {
	Push(parameters common.PushParameters) error
	DoesComponentExist(cmpName string) bool
	Delete(labels map[string]string) error
	Test(testCmd parsercommon.DevfileCommand, show bool) error
}
