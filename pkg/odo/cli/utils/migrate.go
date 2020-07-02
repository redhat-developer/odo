package utils

import (
	"fmt"
	"os"

	"github.com/openshift/odo/pkg/devfile/parser"
	devfileCtx "github.com/openshift/odo/pkg/devfile/parser/context"
	"github.com/openshift/odo/pkg/devfile/parser/data"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	migrateCommandName = "migrate-to-devfile"
)

var migrateLongDesc = ktemplates.LongDesc(`Migrate S2I components to devfile components`)

var migrateExample = ktemplates.Examples(`odo utils migrate-to-devfile`)

// MigrateOptions encapsulates the options for the command
type MigrateOptions struct {
}

// NewMigrateOptions creates a new MigrateOptions instance
func NewMigrateOptions() *MigrateOptions {
	return &MigrateOptions{}
}

// Complete completes MigrateOptions after they've been created
func (o *MigrateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {

	// create env.yaml from config.yaml
	// create devfile.yaml from config.yaml

	// Set the correct context, which also sets the LocalConfigInfo
	context := genericclioptions.NewContextCreatingAppIfNeeded(cmd)
	// sourceType := context.LocalConfigInfo.GetSourceType()
	compType := context.LocalConfigInfo.GetType()
	// sourcePath, err := context.LocalConfigInfo.GetOSSourcePath()
	// urls := context.LocalConfigInfo.GetURL()
	// namespace := context.LocalConfigInfo.GetProject()
	name = context.LocalConfigInfo.GetName()
	if err != nil {
		return errors.Wrap(err, "unable to set source information")
	}

	imageNS, imageName, imageTag, _, err := occlient.ParseImageName(compType)

	if err != nil {
		return errors.Wrap(err, "unable to create new s2i git build ")
	}
	imageStream, err := context.Client.GetImageStream(imageNS, imageName, imageTag)
	if err != nil {
		return errors.Wrap(err, "Failed to bootstrap supervisored")
	}

	imageNS = imageStream.ObjectMeta.Namespace

	_, err = context.Client.GetImageStreamImage(imageStream, imageTag)
	if err != nil {
		return errors.Wrap(err, "unable to bootstrap supervisord")
	}

	wd, _ := os.Getwd()

	s2iDevfile, _ := data.NewDevfileData("2.0.0")
	s2iDevfile.SetMetadata(name, "2.0.0")

	container := common.Container{
		Image:        "openshift/java:latest",
		MountSources: true,
	}

	c := common.DevfileComponent{Container: &container}
	s2iDevfile.SetComponent(c)

	devObj := parser.DevfileObj{
		Ctx:  devfileCtx.NewDevfileCtx(wd), //component context needs to be passed here
		Data: s2iDevfile,
	}

	devObj.WriteYamlDevfile()
	return nil

}

// Validate validates the MigrateOptions based on completed values
func (o *MigrateOptions) Validate() (err error) {
	return nil
}

// Run contains the logic for the command
func (o *MigrateOptions) Run() (err error) {
	return nil
}

// NewCmdMigrate implements the odo utils migrate-to-devfile command
func NewCmdMigrate(name, fullName string) *cobra.Command {
	o := NewMigrateOptions()
	migrateCmd := &cobra.Command{
		Use:     name,
		Short:   "migrates s2i based components to devfile based components",
		Long:    migrateLongDesc,
		Example: fmt.Sprintf(migrateExample, fullName),
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	return migrateCmd
}
