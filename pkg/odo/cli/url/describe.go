package url

import (
	"fmt"
	"os"
	"text/tabwriter"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/odo/util/experimental"
	"github.com/openshift/odo/pkg/url"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// DescribeRecommendedCommandName is the recommended describe command name
const describeRecommendedCommandName = "describe"

var describeExample = ktemplates.Examples(`  # Describe a URL
%[1]s myurl
`)

// URLListOptions encapsulates the options for the odo url list command
type URLDescribeOptions struct {
	localConfigInfo  *config.LocalConfigInfo
	componentContext string
	url              string
	*genericclioptions.Context
}

// NewURLDescribeOptions creates a new URLCreateOptions instance
func NewURLDescribeOptions() *URLDescribeOptions {
	return &URLDescribeOptions{&config.LocalConfigInfo{}, "", "", &genericclioptions.Context{}}
}

// Complete completes URLDescribeOptions after they've been Listed
func (o *URLDescribeOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	if experimental.IsExperimentalModeEnabled() {
		o.Context = genericclioptions.NewDevfileContext(cmd)
	} else {
		o.Context = genericclioptions.NewContext(cmd)
	}
	o.localConfigInfo, err = config.NewLocalConfigInfo(o.componentContext)
	if err != nil {
		return errors.Wrap(err, "failed initialize local config")
	}
	o.url = args[0]
	return
}

// Validate validates the URLDescribeOptions based on completed values
func (o *URLDescribeOptions) Validate() (err error) {
	return util.CheckOutputFlag(o.OutputFlag)
}

// Run contains the logic for the odo url list command
func (o *URLDescribeOptions) Run() (err error) {
	if experimental.IsExperimentalModeEnabled() {
		u, err := url.GetIngress(o.KClient, o.EnvSpecificInfo, o.url)
		if err != nil {
			return err
		}
		if log.IsJSON() {
			machineoutput.OutputSuccess(u)
		} else {

			tabWriterURL := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
			fmt.Fprintln(tabWriterURL, "NAME", "\t", "URL", "\t", "PORT")

			fmt.Fprintln(tabWriterURL, u.Name, "\t", url.GetURLString(url.GetProtocol(routev1.Route{}, u, experimental.IsExperimentalModeEnabled()), "", u.Spec.Rules[0].Host, experimental.IsExperimentalModeEnabled()), "\t", u.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServicePort.IntVal)
			tabWriterURL.Flush()
		}
	} else {
		u, err := url.Get(o.Client, o.localConfigInfo, o.url, o.Application)
		if err != nil {
			return err
		}

		if log.IsJSON() {
			machineoutput.OutputSuccess(u)
		} else {

			tabWriterURL := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
			fmt.Fprintln(tabWriterURL, "NAME", "\t", "STATE", "\t", "URL", "\t", "PORT")

			// are there changes between local and cluster states?
			outOfSync := false
			fmt.Fprintln(tabWriterURL, u.Name, "\t", u.Status.State, "\t", url.GetURLString(u.Spec.Protocol, u.Spec.Host, "", experimental.IsExperimentalModeEnabled()), "\t", u.Spec.Port)
			if u.Status.State != url.StateTypePushed {
				outOfSync = true
			}
			tabWriterURL.Flush()
			if outOfSync {
				fmt.Fprintf(os.Stdout, "\n")
				fmt.Fprintf(os.Stdout, "There are local changes. Please run 'odo push'.\n")
			}
		}
	}

	return
}

// NewCmdURLDescribe implements the odo url describe command.
func NewCmdURLDescribe(name, fullName string) *cobra.Command {
	o := NewURLDescribeOptions()
	urlDescribeCmd := &cobra.Command{
		Use:         name + " [url name]",
		Short:       "Describe a URL",
		Long:        `Describe a URL`,
		Example:     fmt.Sprintf(describeExample, fullName),
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{"machineoutput": "json", "command": "url"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	urlDescribeCmd.SetUsageTemplate(util.CmdUsageTemplate)
	genericclioptions.AddContextFlag(urlDescribeCmd, &o.componentContext)
	completion.RegisterCommandHandler(urlDescribeCmd, completion.URLCompletionHandler)
	completion.RegisterCommandFlagHandler(urlDescribeCmd, "context", completion.FileCompletionHandler)

	return urlDescribeCmd
}
