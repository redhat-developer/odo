package versions

import (
	"github.com/openshift/odo/pkg/devfile/versions/common"
)

type DevfileData interface {
	GetComponents() []common.DevfileComponent
	GetAliasedComponents() []common.DevfileComponent
}
