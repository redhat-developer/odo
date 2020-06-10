package adapters

import (
	"github.com/openshift/odo/pkg/devfile/adapters/common"
)

type PlatformAdapter interface {
	Push(parameters common.PushParameters) error
	Build(parameters common.BuildParameters) error
	Deploy(parameters common.DeployParameters) error
	DoesComponentExist(cmpName string) bool
	Delete(labels map[string]string) error
}
