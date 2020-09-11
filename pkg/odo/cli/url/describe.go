package url

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/devfile"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/lclient"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/cli/component"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/odo/util/pushtarget"
	"github.com/openshift/odo/pkg/url"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// DescribeRecommendedCommandName is the recommended describe command name
const describeRecommendedCommandName = "describe"

var describeExample = ktemplates.Examples(`  # Describe a URL
%[1]s myurl
`)

// URLDescribeOptions encapsulates the options for the odo url describe command
type URLDescribeOptions struct {
	localConfigInfo  *config.LocalConfigInfo
	componentContext string
	url              string
	*genericclioptions.Context
	devfilePath string
	isDevFile   bool
}

// NewURLDescribeOptions creates a new NewURLDescribeOptions instance
func NewURLDescribeOptions() *URLDescribeOptions {
	return &URLDescribeOptions{&config.LocalConfigInfo{}, "", "", &genericclioptions.Context{}, "", false}
}

// Complete completes URLDescribeOptions after they've been Listed
func (o *URLDescribeOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.devfilePath = filepath.Join(o.componentContext, component.DevfilePath)

	if util.CheckPathExists(o.devfilePath) {
		o.Context = genericclioptions.NewDevfileContext(cmd)
		o.EnvSpecificInfo, err = envinfo.NewEnvSpecificInfo(o.componentContext)
	} else {
		o.Context = genericclioptions.NewContext(cmd)
		o.localConfigInfo, err = config.NewLocalConfigInfo(o.componentContext)
	}
	if err != nil {
		return errors.Wrap(err, "failed intiating local config")
	}
	o.url = args[0]
	return
}

// Validate validates the URLDescribeOptions based on completed values
func (o *URLDescribeOptions) Validate() (err error) {
	return odoutil.CheckOutputFlag(o.OutputFlag)
}

// Run contains the logic for the odo url describe command
func (o *URLDescribeOptions) Run() (err error) {
	if util.CheckPathExists(o.devfilePath) {
		if pushtarget.IsPushTargetDocker() {
			client, err := lclient.New()
			if err != nil {
				return err
			}
			u, err := url.GetContainerURL(client, o.EnvSpecificInfo, o.url, o.EnvSpecificInfo.GetName())
			if err != nil {
				return err
			}

			if log.IsJSON() {
				machineoutput.OutputSuccess(u)
			} else {
				tabWriterURL := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
				fmt.Fprintln(tabWriterURL, "NAME", "\t", "STATE", "\t", "URL", "\t", "PORT")
				var urlString string
				if u.Status.State == url.StateTypeNotPushed {
					// to be consistent with URL for ingress and routes
					// if not pushed, display URl as ://
					urlString = "://"
				} else {
					urlString = fmt.Sprintf("%s:%s", u.Spec.Host, strconv.Itoa(u.Spec.ExternalPort))
				}
				// are there changes between local and cluster states?
				outOfSync := false
				fmt.Fprintln(tabWriterURL, u.Name, "\t", u.Status.State, "\t", urlString, "\t", u.Spec.Port)
				if u.Status.State != url.StateTypePushed {
					outOfSync = true
				}
				tabWriterURL.Flush()
				if outOfSync {
					log.Info("There are local changes. Please run 'odo push'.")
				}
			}
		} else {
			componentName := o.EnvSpecificInfo.GetName()
			oclient, err := occlient.New()
			if err != nil {
				return err
			}
			oclient.Namespace = o.KClient.Namespace
			routeSupported, err := oclient.IsRouteSupported()
			if err != nil {
				return err
			}

			devObj, err := devfile.ParseAndValidate(o.devfilePath)
			if err != nil {
				return errors.Wrap(err, "fail to parse the devfile")
			}
			containerComponents := adaptersCommon.GetDevfileContainerComponents(devObj.Data)
			u, err := url.GetIngressOrRoute(oclient, o.KClient, o.EnvSpecificInfo, o.url, containerComponents, componentName, routeSupported)
			if err != nil {
				return err
			}
			if log.IsJSON() {
				machineoutput.OutputSuccess(u)
			} else {
				tabWriterURL := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
				fmt.Fprintln(tabWriterURL, "NAME", "\t", "STATE", "\t", "URL", "\t", "PORT", "\t", "SECURE", "\t", "KIND")

				// are there changes between local and cluster states?
				outOfSync := false
				if u.Spec.Kind == envinfo.ROUTE {
					fmt.Fprintln(tabWriterURL, u.Name, "\t", u.Status.State, "\t", url.GetURLString(u.Spec.Protocol, u.Spec.Host, "", false), "\t", u.Spec.Port, "\t", u.Spec.Secure, "\t", u.Spec.Kind)
				} else {
					fmt.Fprintln(tabWriterURL, u.Name, "\t", u.Status.State, "\t", url.GetURLString(url.GetProtocol(routev1.Route{}, url.ConvertIngressURLToIngress(u, componentName)), "", u.Spec.Host, false), "\t", u.Spec.Port, "\t", u.Spec.Secure, "\t", u.Spec.Kind)
				}
				if u.Status.State != url.StateTypePushed {
					outOfSync = true
				}
				tabWriterURL.Flush()
				if outOfSync {
					log.Info("There are local changes. Please run 'odo push'.")
				}
			}
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
			fmt.Fprintln(tabWriterURL, u.Name, "\t", u.Status.State, "\t", url.GetURLString(u.Spec.Protocol, u.Spec.Host, "", false), "\t", u.Spec.Port)
			if u.Status.State != url.StateTypePushed {
				outOfSync = true
			}
			tabWriterURL.Flush()
			if outOfSync {
				log.Info("There are local changes. Please run 'odo push'.")
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
	urlDescribeCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	genericclioptions.AddContextFlag(urlDescribeCmd, &o.componentContext)
	completion.RegisterCommandHandler(urlDescribeCmd, completion.URLCompletionHandler)
	completion.RegisterCommandFlagHandler(urlDescribeCmd, "context", completion.FileCompletionHandler)

	return urlDescribeCmd
}
