package helper

import (
	"path/filepath"

	. "github.com/onsi/gomega"
	"github.com/openshift/odo/v2/pkg/envinfo"
)

const configFileDirectory = ".odo"
const envInfoFile = "env.yaml"

func LocalEnvInfo(context string) *envinfo.EnvSpecificInfo {
	info, err := envinfo.NewEnvSpecificInfo(filepath.Join(context, configFileDirectory, envInfoFile))
	if err != nil {
		Expect(err).To(Equal(nil))
	}
	return info
}
