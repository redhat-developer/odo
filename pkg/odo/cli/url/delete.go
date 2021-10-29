package url

import (
	"fmt"

	"github.com/openshift/odo/pkg/log"
	clicomponent "github.com/openshift/odo/pkg/odo/cli/component"
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const deleteRecommendedCommandName = "delete"

var (
	urlDeleteShortDesc = `Delete a URL`
	urlDeleteLongDesc  = ktemplates.LongDesc(`Delete the given URL, hence making the service inaccessible.`)
	urlDeleteExample   = ktemplates.Examples(`  # Delete a URL to a component
 %[1]s myurl
	`)
)

// DeleteOptions encapsulates the options for the odo url delete command
type DeleteOptions struct {
	*clicomponent.PushOptions
	urlName            string
	urlForceDeleteFlag bool
	now                bool
}

// NewURLDeleteOptions creates a new DeleteOptions instance
func NewURLDeleteOptions() *DeleteOptions {
	return &DeleteOptions{PushOptions: clicomponent.NewPushOptions()}
}

// Complete completes DeleteOptions after they've been Deleted
func (o *DeleteOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.CreateParameters{
		Cmd:              cmd,
		Devfile:          true,
		DevfilePath:      o.DevfilePath,
		ComponentContext: o.GetComponentContext(),
	})

	o.urlName = args[0]

	if o.now {
		prjName := o.LocalConfigProvider.GetNamespace()
		o.ResolveSrcAndConfigFlags()
		err = o.ResolveProject(prjName)
		if err != nil {
			return err
		}
	}

	return

}

// Validate validates the DeleteOptions based on completed values
func (o *DeleteOptions) Validate() (err error) {
	url, err := o.Context.LocalConfigProvider.GetURL(o.urlName)
	if err != nil {
		return err
	}
	if url == nil {
		return fmt.Errorf("the URL %s does not exist within the component %s", o.urlName, o.LocalConfigProvider.GetName())
	}
	return
}

// Run contains the logic for the odo url delete command
func (o *DeleteOptions) Run(cmd *cobra.Command) (err error) {
	if o.urlForceDeleteFlag || ui.Proceed(fmt.Sprintf("Are you sure you want to delete the url %v", o.urlName)) {
		err := o.LocalConfigProvider.DeleteURL(o.urlName)
		if err != nil {
			return nil

		}

		log.Successf("URL %s removed from component %s", o.urlName, o.LocalConfigProvider.GetName())

		if o.now {
			o.CompleteDevfilePath()
			o.EnvSpecificInfo = o.Context.EnvSpecificInfo
			err = o.DevfilePush()
			if err != nil {
				return err
			}
			log.Italic("\nTo delete the URL on the cluster, please use `odo push`")
		}

	} else {
		return fmt.Errorf("aborting deletion of URL: %v", o.urlName)
	}
	return
}

// NewCmdURLDelete implements the odo url delete command.
func NewCmdURLDelete(name, fullName string) *cobra.Command {
	o := NewURLDeleteOptions()
	urlDeleteCmd := &cobra.Command{
		Use:   name + " [url name]",
		Short: urlDeleteShortDesc,
		Long:  urlDeleteLongDesc,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
		Example: fmt.Sprintf(urlDeleteExample, fullName),
	}
	urlDeleteCmd.Flags().BoolVarP(&o.urlForceDeleteFlag, "force", "f", false, "Delete url without prompting")

	o.AddContextFlag(urlDeleteCmd)
	genericclioptions.AddNowFlag(urlDeleteCmd, &o.now)
	completion.RegisterCommandHandler(urlDeleteCmd, completion.URLCompletionHandler)
	completion.RegisterCommandFlagHandler(urlDeleteCmd, "context", completion.FileCompletionHandler)

	return urlDeleteCmd
}
