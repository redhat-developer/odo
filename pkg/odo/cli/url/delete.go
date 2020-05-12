package url

import (
	"fmt"

	"github.com/openshift/odo/pkg/log"
	clicomponent "github.com/openshift/odo/pkg/odo/cli/component"
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/odo/util/experimental"
	"github.com/pkg/errors"
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

// URLDeleteOptions encapsulates the options for the odo url delete command
type URLDeleteOptions struct {
	*clicomponent.CommonPushOptions
	urlName            string
	urlForceDeleteFlag bool
	now                bool
}

// NewURLDeleteOptions creates a new URLDeleteOptions instance
func NewURLDeleteOptions() *URLDeleteOptions {
	return &URLDeleteOptions{CommonPushOptions: clicomponent.NewCommonPushOptions()}
}

// Complete completes URLDeleteOptions after they've been Deleted
func (o *URLDeleteOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	if experimental.IsExperimentalModeEnabled() {
		o.Context = genericclioptions.NewDevfileContext(cmd)
		o.urlName = args[0]
		err = o.InitEnvInfoFromContext()
		if err != nil {
			return err
		}
	} else {
		if o.now {
			o.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)
		} else {
			o.Context = genericclioptions.NewContext(cmd)
		}
		o.urlName = args[0]
		err = o.InitConfigFromContext()
		if err != nil {
			return err
		}
		if o.now {
			prjName := o.LocalConfigInfo.GetProject()
			o.ResolveSrcAndConfigFlags()
			err = o.ResolveProject(prjName)
			if err != nil {
				return err
			}
		}
	}
	return

}

// Validate validates the URLDeleteOptions based on completed values
func (o *URLDeleteOptions) Validate() (err error) {
	var exists bool
	if experimental.IsExperimentalModeEnabled() {
		urls := o.EnvSpecificInfo.GetURL()
		componentName := o.EnvSpecificInfo.GetName()
		for _, url := range urls {
			if url.Name == o.urlName {
				exists = true
			}
		}
		if !exists {
			return fmt.Errorf("the URL %s does not exist within the component %s", o.urlName, componentName)
		}
	} else {
		urls := o.LocalConfigInfo.GetURL()

		for _, url := range urls {
			if url.Name == o.urlName {
				exists = true
			}
		}
		if o.now {
			err = o.ValidateComponentCreate()
			if err != nil {
				return err
			}
		}
		if !exists {
			return fmt.Errorf("the URL %s does not exist within the component %s", o.urlName, o.Component())
		}
	}

	return
}

// Run contains the logic for the odo url delete command
func (o *URLDeleteOptions) Run() (err error) {
	if o.urlForceDeleteFlag || ui.Proceed(fmt.Sprintf("Are you sure you want to delete the url %v", o.urlName)) {
		if experimental.IsExperimentalModeEnabled() {
			err = o.EnvSpecificInfo.DeleteURL(o.urlName)
			if err != nil {
				return err
			}
			log.Successf("URL %s removed from the env file", o.urlName)
			log.Italic("\nTo delete the URL on the cluster, please use `odo push`")
		} else {
			err = o.LocalConfigInfo.DeleteURL(o.urlName)
			if err != nil {
				return err
			}
			log.Successf("URL %s removed from the config file", o.urlName)
			if o.now {
				err = o.Push()
				if err != nil {
					return errors.Wrap(err, "failed to push changes")
				}
			} else {
				log.Italic("\nTo delete the URL on the cluster, please use `odo push`")
			}
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
