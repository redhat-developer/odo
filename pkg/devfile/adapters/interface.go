package adapters

import (
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/envinfo"
)

type PlatformAdapter interface {
	Push(parameters common.PushParameters) error
	DoesComponentExist(cmpName string) bool
	Delete(labels map[string]string) error
	ValidateURL(url []envinfo.EnvInfoURL)
}
