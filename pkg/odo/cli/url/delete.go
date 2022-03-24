package url

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
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
	*genericclioptions.Context

	contextFlag string

	// Parameters
	urlName string

	// Flags
	forceFlag bool
}

// NewURLDeleteOptions creates a new DeleteOptions instance
func NewURLDeleteOptions() *DeleteOptions {
	return &DeleteOptions{}
}

func (o *DeleteOptions) SetClientset(clientset *clientset.Clientset) {}

// Complete completes DeleteOptions after they've been Deleted
func (o *DeleteOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(o.contextFlag))
	if err != nil {
		return err
	}

	o.urlName = args[0]

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
func (o *DeleteOptions) Run(cmdline cmdline.Cmdline) (err error) {
	if o.forceFlag || ui.Proceed(fmt.Sprintf("Are you sure you want to delete the url %v", o.urlName)) {
		err := o.LocalConfigProvider.DeleteURL(o.urlName)
		if err != nil {
			return nil

		}

		log.Successf("URL %s removed from component %s", o.urlName, o.LocalConfigProvider.GetName())

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
	clientset.Add(urlDeleteCmd, clientset.PROJECT, clientset.PREFERENCE)

	urlDeleteCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Delete url without prompting")

	odoutil.AddContextFlag(urlDeleteCmd, &o.contextFlag)
	completion.RegisterCommandHandler(urlDeleteCmd, completion.URLCompletionHandler)
	completion.RegisterCommandFlagHandler(urlDeleteCmd, "context", completion.FileCompletionHandler)

	return urlDeleteCmd
}
