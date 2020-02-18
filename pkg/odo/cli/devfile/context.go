package devfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewDevfileContext creates a new Context struct populated with the current state based on flags specified for the provided command
func NewDevfileContext(command *cobra.Command, devfilePath string) (*Context, error) {

	componentName, err := GetComponentName()

	if err != nil || len(componentName) == 0 {
		return nil, fmt.Errorf("Unable to generate a valid component name. Component names are based on the component directory name and must consist of lower case alphanumeric characters, '-' or '.' and follow DNS-1123 subdomain rules")
	}

	adapterMetadata := common.AdapterMetadata{
		Name: componentName,
	}

	// Parse devfile and add it to the context
	devObj, err := devfile.Parse(devfilePath)
	if err == nil {
		adapterMetadata.Devfile = devObj
	}

	context := &Context{
		AdapterMetadata: adapterMetadata,
	}

	return context, err
}

// GetComponentName returns component name
// Returns: directory name or error
func GetComponentName() (string, error) {
	retVal := ""
	currDir, err := os.Getwd()
	if err != nil {
		return "", errors.Wrapf(err, "unable to get component because getting current directory failed")
	}
	retVal = filepath.Base(currDir)
	retVal = strings.TrimSpace(util.GetDNS1123Name(strings.ToLower(retVal)))
	return retVal, nil
}

// Context contains contextual information for Devfile commands
type Context struct {
	common.AdapterMetadata
}
