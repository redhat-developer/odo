package login

import (
	"context"
	"fmt"

	"github.com/redhat-developer/odo/pkg/auth"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "login"

// LoginOptions encapsulates the options for the odo command
type LoginOptions struct {
	// Parameters
	server string

	// Flags
	userNameFlag string
	passwordFlag string
	tokenFlag    string
	caAuthFlag   string
	skipTlsFlag  bool
	serverFlag   string

	// client
	loginClient auth.Client
}

var loginExample = templates.Examples(`
  # Log in interactively
  %[1]s

  # Log in to the given server with the given certificate authority file
  %[1]s localhost:8443 --certificate-authority=/path/to/cert.crt

  # Log in to the given server with the given credentials (basic auth)
  %[1]s localhost:8443 --username=myuser --password=mypass

  # Log in to the given server with the given credentials (token)
  %[1]s localhost:8443 --token=xxxxxxxxxxxxxxxxxxxxxxx
`)

// NewLoginOptions creates a new LoginOptions instance
func NewLoginOptions(client auth.Client) *LoginOptions {
	return &LoginOptions{
		loginClient: client,
	}
}

func (o *LoginOptions) SetClientset(clientset *clientset.Clientset) {
}

// Complete completes LoginOptions after they've been created
func (o *LoginOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	if len(args) == 1 {
		// if the user specifies server without --server flag. Example:
		// odo login -u developer -p developer https://api.crc.testing:6443
		// odo login --token=<some-token> https://api.crc.testing:6443
		o.server = args[0]
	}
	return
}

// Validate validates the LoginOptions based on completed values
func (o *LoginOptions) Validate() (err error) {
	if o.server != "" && o.serverFlag != "" && o.server != o.serverFlag {
		// if user has passed server value as parameter as well as used --server flag:
		// * odo errors *if* the values are different
		// * odo silently continues if the values are same
		return fmt.Errorf("either use --server flag or pass server link as a paremeter, don't use both")
	} else if o.serverFlag == "" {
		o.serverFlag = o.server //	set o.serverFlag to same as o.server if there was no error
	}

	return
}

// Run contains the logic for the odo command
func (o *LoginOptions) Run(ctx context.Context) (err error) {
	return o.loginClient.Login(o.serverFlag, o.userNameFlag, o.passwordFlag, o.tokenFlag, o.caAuthFlag, o.skipTlsFlag)
}

// NewCmdLogin implements the odo command
func NewCmdLogin(name, fullName string) *cobra.Command {
	loginClient := auth.NewKubernetesClient()
	o := NewLoginOptions(loginClient)

	loginCmd := &cobra.Command{
		Use:     name,
		Short:   "Login to cluster",
		Long:    "Login to cluster",
		Example: fmt.Sprintf(loginExample, fullName),
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	// Add a defined annotation in order to appear in the help menu
	loginCmd.Annotations = map[string]string{"command": "openshift"}
	loginCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	loginCmd.Flags().StringVarP(&o.userNameFlag, "username", "u", "", "username, will prompt if not provided")
	loginCmd.Flags().StringVarP(&o.passwordFlag, "password", "p", "", "password, will prompt if not provided")
	loginCmd.Flags().StringVarP(&o.tokenFlag, "token", "t", "", "token, will prompt if not provided")
	loginCmd.Flags().BoolVar(&o.skipTlsFlag, "insecure-skip-tls-verify", false, "If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure")
	loginCmd.Flags().StringVar(&o.caAuthFlag, "certificate-authority", "", "Path to a cert file for the certificate authority")
	loginCmd.Flags().StringVar(&o.serverFlag, "server", "", "OpenShift server to log into")
	return loginCmd
}
