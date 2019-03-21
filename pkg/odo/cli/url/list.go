package url

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/url"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const listRecommendedCommandName = "list"

var (
	urlListShortDesc = `List URLs`
	urlListLongDesc  = ktemplates.LongDesc(`Lists all the available URLs which can be used to access the components.`)
	urlListExample   = ktemplates.Examples(` # List the available URLs
  %[1]s
	`)
)

// URLListOptions encapsulates the options for the odo url list command
type URLListOptions struct {
	outputFlag string
	*genericclioptions.Context
}

// NewURLListOptions creates a new UrlCreateOptions instance
func NewURLListOptions() *URLListOptions {
	return &URLListOptions{}
}

// Complete completes UrlListOptions after they've been Listed
func (o *URLListOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	return
}

// Validate validates the UrlListOptions based on completed values
func (o *URLListOptions) Validate() (err error) {
	return util.CheckOutputFlag(o.outputFlag)
}

// Run contains the logic for the odo url list command
func (o *URLListOptions) Run() (err error) {

	urls, err := url.List(o.Client, o.Component(), o.Application)
	if err != nil {
		return err
	}

	if len(urls.Items) == 0 {
		return fmt.Errorf("no URLs found for component %v in application %v", o.Component(), o.Application)
	} else {
		if o.outputFlag == "json" {
			out, err := json.Marshal(urls)
			if err != nil {
				return err
			}
			fmt.Println(string(out))

		} else {
			log.Infof("Found the following URLs for component %v in application %v:", o.Component(), o.Application)

			tabWriterURL := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)

			//create headers
			fmt.Fprintln(tabWriterURL, "NAME", "\t", "URL", "\t", "PORT")

			for _, u := range urls.Items {
				fmt.Fprintln(tabWriterURL, u.Name, "\t", url.GetURLString(u.Spec.Protocol, u.Spec.Host), "\t", u.Spec.Port)
			}
			tabWriterURL.Flush()

		}
	}

	return
}

// NewCmdURLList implements the odo url list command.
func NewCmdURLList(name, fullName string) *cobra.Command {
	o := NewURLListOptions()
	urlListCmd := &cobra.Command{
		Use:     name,
		Short:   urlListShortDesc,
		Long:    urlListLongDesc,
		Example: fmt.Sprintf(urlListExample, fullName),
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	urlListCmd.Flags().StringVarP(&o.outputFlag, "output", "o", "", "gives output in the form of json")

	return urlListCmd
}
