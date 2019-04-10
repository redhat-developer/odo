package url

import (
	"fmt"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
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
	localConfigInfo    *config.LocalConfigInfo
	componentContext   string
	urlName            string
	urlForceDeleteFlag bool
	*genericclioptions.Context
}

// NewURLDeleteOptions creates a new UrlDeleteOptions instance
func NewURLDeleteOptions() *URLDeleteOptions {
	return &URLDeleteOptions{}
}

// Complete completes URLDeleteOptions after they've been Deleted
func (o *URLDeleteOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	o.urlName = args[0]
	o.localConfigInfo, err = config.NewLocalConfigInfo(o.componentContext)
	return

}

// Validate validates the URLDeleteOptions based on completed values
func (o *URLDeleteOptions) Validate() (err error) {
	//exists, err := url.Exists(o.Client, o.localConfigInfo, o.urlName, o.Component(), o.Application)
	//if err != nil {
	//	return err
	//}
	var exists bool
	urls := o.localConfigInfo.GetUrl()

	for _, url := range urls {
		if url.Name == o.urlName {
			exists = true
		}
	}
	if !exists {
		return fmt.Errorf("the URL %s does not exist within the component %s", o.urlName, o.Component())
	}
	return
}

// Run contains the logic for the odo url delete command
func (o *URLDeleteOptions) Run() (err error) {

	if o.urlForceDeleteFlag || ui.Proceed(fmt.Sprintf("Are you sure you want to delete the url %v", o.urlName)) {
		err = o.localConfigInfo.DeleteUrl(o.urlName)
		if err != nil {
			return err
		}
		log.Info("URL removed from the config file")
		log.Info("To delete URL from the OpenShift cluster, please run `odo push`")
	} else {
		return fmt.Errorf("aborting deletion of url: %v", o.urlName)
	}
	return
}

// NewCmdURLDelete implements the odo url delete command.
func NewCmdURLDelete(name, fullName string) *cobra.Command {
	o := NewURLDeleteOptions()
	urlDeleteCmd := &cobra.Command{
		Use:   name + " [component name]",
		Short: urlDeleteShortDesc,
		Long:  urlDeleteLongDesc,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
		Example: fmt.Sprintf(urlDeleteExample, fullName),
	}
	urlDeleteCmd.Flags().BoolVarP(&o.urlForceDeleteFlag, "force", "f", false, "Delete url without prompting")
	genericclioptions.AddContextFlag(urlDeleteCmd, &o.componentContext)
	completion.RegisterCommandHandler(urlDeleteCmd, completion.URLCompletionHandler)
	completion.RegisterCommandFlagHandler(urlDeleteCmd, "context", completion.FileCompletionHandler)
	return urlDeleteCmd
}
