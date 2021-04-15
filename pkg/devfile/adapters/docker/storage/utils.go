package storage

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/util"
)

const volNameMaxLength = 45

// GenerateVolName generates a Docker volume name from the Devfile volume name and component name
func GenerateVolName(volName, componentName string) (string, error) {

	if volName == "" {
		err := errors.New("unable to generate volume name with an empty name")
		return "", err
	}

	dockerVolName := fmt.Sprintf("%v-%v", volName, componentName)
	dockerVolName = util.TruncateString(dockerVolName, volNameMaxLength)
	randomChars := util.GenerateRandomString(4)
	dockerVolName, err := util.NamespaceOpenShiftObject(dockerVolName, randomChars)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create namespaced name for volume %s", volName)
	}

	return dockerVolName, nil
}
