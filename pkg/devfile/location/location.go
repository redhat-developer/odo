package location

import (
	"path/filepath"

	"github.com/openshift/odo/pkg/util"
)

//  possibleDevfileNames contains possivle devfile name that should be checked in the context dir.
var possibleDevfileNames = [...]string{"devfile.yaml", ".devfile.yaml"}

// DevfileFilenamesProvider checks if the context dir contains devfile with possible names from possibleDevfileNames variable,
// else it returns "devfile.yaml" as default value.
func DevfileFilenamesProvider(contexDir string) string {
	for _, devFile := range possibleDevfileNames {
		if util.CheckPathExists(filepath.Join(contexDir, devFile)) {
			return devFile
		}
	}
	return "devfile.yaml"
}

// DevfileLocation returns path to devfile File with approprite devfile File name from possibleDevfileNames variable,
// if contextDir value is provided as argument then DevfileLocation returns devfile path
// else it assumes current dir as contextDir and returns appropriate value.
func DevfileLocation(contexDir string) string {
	if contexDir == "" {
		contexDir = "./"
	}
	devFile := DevfileFilenamesProvider(contexDir)
	return filepath.Join(contexDir, devFile)
}
