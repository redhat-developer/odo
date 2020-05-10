package url

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/openshift/odo/pkg/odo/util/experimental"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/url"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
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
	componentContext string
	*genericclioptions.Context
}

// NewURLListOptions creates a new URLCreateOptions instance
func NewURLListOptions() *URLListOptions {
	return &URLListOptions{}
}

// Complete completes URLListOptions after they've been Listed
func (o *URLListOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	if experimental.IsExperimentalModeEnabled() {
		o.Context = genericclioptions.NewDevfileContext(cmd)
	} else {
		o.Context = genericclioptions.NewContext(cmd)
	}
	o.LocalConfigInfo, err = config.NewLocalConfigInfo(o.componentContext)
	if err != nil {
		return errors.Wrap(err, "failed intiating local config")
	}
	return
}

// Validate validates the URLListOptions based on completed values
func (o *URLListOptions) Validate() (err error) {
	return util.CheckOutputFlag(o.OutputFlag)
}

// Run contains the logic for the odo url list command
func (o *URLListOptions) Run() (err error) {
	if experimental.IsExperimentalModeEnabled() {
		componentName := o.EnvSpecificInfo.GetName()
		// TODO: Need to list all local and pushed ingresses
		//		 issue to track: https://github.com/openshift/odo/issues/2787
		urls, err := url.ListPushedIngress(o.KClient, componentName)
		if err != nil {
			return err
		}
		localUrls := o.EnvSpecificInfo.GetURL()
		if log.IsJSON() {
			machineoutput.OutputSuccess(urls)
		} else {
			if len(urls.Items) == 0 {
				return fmt.Errorf("no URLs found for component %v", componentName)
			}

			log.Infof("Found the following URLs for component %v", componentName)
			tabWriterURL := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
			fmt.Fprintln(tabWriterURL, "NAME", "\t", "URL", "\t", "PORT", "\t", "SECURE")

			// are there changes between local and cluster states?
			outOfSync := false
			for _, i := range localUrls {
				var present bool
				for _, u := range urls.Items {
					if i.Name == u.Name {
						fmt.Fprintln(tabWriterURL, u.Name, "\t", url.GetURLString(url.GetProtocol(routev1.Route{}, u, experimental.IsExperimentalModeEnabled()), "", u.Spec.Rules[0].Host, experimental.IsExperimentalModeEnabled()), "\t", u.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServicePort.IntVal, "\t", u.Spec.TLS != nil)
						present = true
					}
				}
				if !present {
					fmt.Fprintln(tabWriterURL, i.Name, "\t", "<not created on cluster>", "\t", i.Port)
				}
			}
			tabWriterURL.Flush()
			if outOfSync {
				fmt.Fprintf(os.Stdout, "\n")
				fmt.Fprintf(os.Stdout, "There are local changes. Please run 'odo push'.\n")
			}
		}
	} else {
		urls, err := url.List(o.Client, o.LocalConfigInfo, o.Component(), o.Application)
		if err != nil {
			return err
		}
		if log.IsJSON() {
			machineoutput.OutputSuccess(urls)
		} else {
			if len(urls.Items) == 0 {
				return fmt.Errorf("no URLs found for component %v in application %v", o.Component(), o.Application)
			}

			log.Infof("Found the following URLs for component %v in application %v:", o.Component(), o.Application)
			tabWriterURL := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
			fmt.Fprintln(tabWriterURL, "NAME", "\t", "STATE", "\t", "URL", "\t", "PORT", "\t", "SECURE")

			// are there changes between local and cluster states?
			outOfSync := false
			for _, u := range urls.Items {
				fmt.Fprintln(tabWriterURL, u.Name, "\t", u.Status.State, "\t", url.GetURLString(u.Spec.Protocol, u.Spec.Host, "", experimental.IsExperimentalModeEnabled()), "\t", u.Spec.Port, "\t", u.Spec.Secure)
				if u.Status.State != url.StateTypePushed {
					outOfSync = true
				}
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

// NewCmdURLList implements the odo url list command.
func NewCmdURLList(name, fullName string) *cobra.Command {
	o := NewURLListOptions()
	urlListCmd := &cobra.Command{
		Use:         name,
		Short:       urlListShortDesc,
		Long:        urlListLongDesc,
		Example:     fmt.Sprintf(urlListExample, fullName),
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	genericclioptions.AddContextFlag(urlListCmd, &o.componentContext)
	completion.RegisterCommandFlagHandler(urlListCmd, "context", completion.FileCompletionHandler)

	return urlListCmd
}
