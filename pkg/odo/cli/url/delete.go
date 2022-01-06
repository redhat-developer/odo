package url

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	clicomponent "github.com/redhat-developer/odo/pkg/odo/cli/component"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/project"
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
	// Push context
	*clicomponent.PushOptions

	// Parameters
	urlName string

	// Flags
	forceFlag bool
	nowFlag   bool
}

// NewURLDeleteOptions creates a new DeleteOptions instance
func NewURLDeleteOptions(prjClient project.Client, prefClient preference.Client) *DeleteOptions {
	return &DeleteOptions{PushOptions: clicomponent.NewPushOptions(prjClient, prefClient)}
}

// Complete completes DeleteOptions after they've been Deleted
func (o *DeleteOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(o.GetComponentContext()))
	if err != nil {
		return err
	}

	o.urlName = args[0]

	if o.nowFlag {
		prjName := o.LocalConfigProvider.GetNamespace()
		o.ResolveSrcAndConfigFlags()
		err = o.ResolveProject(prjName)
		if err != nil {
			return err
		}
	}

	return nil
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
	return nil
}

// Run contains the logic for the odo url delete command
func (o *DeleteOptions) Run() (err error) {
	if o.forceFlag || ui.Proceed(fmt.Sprintf("Are you sure you want to delete the url %v", o.urlName)) {
		err := o.LocalConfigProvider.DeleteURL(o.urlName)
		if err != nil {
			return nil

		}

		log.Successf("URL %s removed from component %s", o.urlName, o.LocalConfigProvider.GetName())

		if o.nowFlag {
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
	// The error is not handled at this point, it will be handled during Context creation
	kubclient, _ := kclient.New()
	prefClient, err := preference.NewClient()
	if err != nil {
		odoutil.LogErrorAndExit(err, "unable to set preference, something is wrong with odo, kindly raise an issue at https://github.com/redhat-developer/odo/issues/new?template=Bug.md")
	}
	o := NewURLDeleteOptions(project.NewClient(kubclient), prefClient)
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
	urlDeleteCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Delete url without prompting")

	o.AddContextFlag(urlDeleteCmd)
	odoutil.AddNowFlag(urlDeleteCmd, &o.nowFlag)
	completion.RegisterCommandHandler(urlDeleteCmd, completion.URLCompletionHandler)
	completion.RegisterCommandFlagHandler(urlDeleteCmd, "context", completion.FileCompletionHandler)

	return urlDeleteCmd
}
