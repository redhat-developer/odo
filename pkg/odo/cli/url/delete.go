package url

import (
	"fmt"
	"strings"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/redhat-developer/odo/pkg/url"
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
	return

}

// Validate validates the URLDeleteOptions based on completed values
func (o *URLDeleteOptions) Validate() (err error) {
	exists, err := url.Exists(o.Client, o.urlName, o.Component(), o.Application)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("the URL %s does not exist within the component %s", o.urlName, o.Component())
	}
	return
}

// Run contains the logic for the odo url delete command
func (o *URLDeleteOptions) Run() (err error) {
	var confirmDeletion string
	if o.urlForceDeleteFlag {
		confirmDeletion = "y"
	} else {
		log.Askf("Are you sure you want to delete the url %v? [y/N]: ", o.urlName)
		fmt.Scanln(&confirmDeletion)
	}

	if strings.ToLower(confirmDeletion) == "y" {

		err = url.Delete(o.Client, o.urlName, o.Application)
		if err != nil {
			return err
		}
		log.Infof("Deleted URL: %v", o.urlName)
	} else {
		return fmt.Errorf("Aborting deletion of url: %v", o.urlName)
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
			odoutil.LogErrorAndExit(o.Complete(name, cmd, args), "")
			odoutil.LogErrorAndExit(o.Validate(), "")
			odoutil.LogErrorAndExit(o.Run(), "")
		},
		Example: fmt.Sprintf(urlDeleteExample, fullName),
	}
	urlDeleteCmd.Flags().BoolVarP(&o.urlForceDeleteFlag, "force", "f", false, "Delete url without prompting")

	completion.RegisterCommandHandler(urlDeleteCmd, completion.URLCompletionHandler)
	return urlDeleteCmd
}
