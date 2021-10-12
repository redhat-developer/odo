package devfile

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/validate"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/util"
)

func parseDevfile(args parser.ParserArgs) (parser.DevfileObj, error) {
	devObj, varWarnings, err := devfile.ParseDevfileAndValidate(args)
	if err != nil {
		return parser.DevfileObj{}, err
	}

	// odo specific validations
	err = validate.ValidateDevfileData(devObj.Data)
	if err != nil {
		return parser.DevfileObj{}, err
	}

	// display warnings related to variable substitution
	for variable, messages := range varWarnings.Commands {
		log.Warningf(variableWarning("commands", variable, messages))
	}
	for variable, messages := range varWarnings.Components {
		log.Warningf(variableWarning("components", variable, messages))
	}
	for variable, messages := range varWarnings.Projects {
		log.Warningf(variableWarning("projects", variable, messages))
	}
	for variable, messages := range varWarnings.StarterProjects {
		log.Warningf(variableWarning("starterProjects", variable, messages))
	}

	return devObj, nil
}

// ParseFromFile reads, parses and validates  devfile from a file
// if there are warning it logs them on stdout
func ParseFromFile(devfilePath string) (parser.DevfileObj, error) {
	return parseDevfile(parser.ParserArgs{Path: devfilePath})
}

// ParseFromData parses devfile from []byte and does all the validation
// if there are warning it logs them on stdout
func ParseFromData(data []byte) (parser.DevfileObj, error) {
	return parseDevfile(parser.ParserArgs{Data: data})

}

// ParseFromURL parses devfile from given url and does all the validation
// if there are warning it logs them on stdout
func ParseFromURL(url string) (parser.DevfileObj, error) {
	return parseDevfile(parser.ParserArgs{URL: url})
}

func variableWarning(section string, variable string, messages []string) string {
	quotedVars := []string{}
	for _, v := range messages {
		quotedVars = append(quotedVars, fmt.Sprintf("%q", v))
	}
	return fmt.Sprintf("Invalid variable(s) %s in %q section with name %q. ", strings.Join(quotedVars, ","), section, variable)

}

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
