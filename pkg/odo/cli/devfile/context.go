package devfile

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/odo/pkg/component/devfile/adapters/common"
	"github.com/spf13/cobra"
)

// NewDevfileContext creates a new Context struct populated with the current state based on flags specified for the provided command
func NewDevfileContext(command *cobra.Command) (*Context, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	componentName := strings.ToLower(filepath.Base(workingDir))

	devfileComponent := common.DevfileComponent{
		Name: componentName,
	}

	context := &Context{
		DevfileComponent: devfileComponent,
	}
	return context, nil
}

// Context contains contextual information for Devfile commands
type Context struct {
	common.DevfileComponent
}
