package helper

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/pkg/envinfo"
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

// CreateLocalEnv creates a .odo/env/env.yaml file
// Useful for commands that require this file and cannot create one on their own, for e.g. url, list
func CreateLocalEnv(context, compName, projectName string) {
	var config = fmt.Sprintf(`
ComponentSettings:
  Name: %s
  Project: %s
  AppName: app
`, compName, projectName)
	dir := filepath.Join(context, ".odo", "env")
	MakeDir(dir)
	Expect(ioutil.WriteFile(filepath.Join(dir, "env.yaml"), []byte(config), 0600)).To(BeNil())
}
