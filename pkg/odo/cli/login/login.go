package login

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/auth"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
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
func NewLoginOptions() *LoginOptions {
	return &LoginOptions{}
}

// Complete completes LoginOptions after they've been created
func (o *LoginOptions) Complete(name string, cmdline cmdline.Cmdline, args []string) (err error) {
	if len(args) == 1 {
		o.server = args[0]
	}
	return
}

// Validate validates the LoginOptions based on completed values
func (o *LoginOptions) Validate() (err error) {
	return
}

// Run contains the logic for the odo command
func (o *LoginOptions) Run(cmd *cobra.Command) (err error) {
	return auth.Login(o.server, o.userNameFlag, o.passwordFlag, o.tokenFlag, o.caAuthFlag, o.skipTlsFlag)
}

// NewCmdLogin implements the odo command
func NewCmdLogin(name, fullName string) *cobra.Command {
	o := NewLoginOptions()
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
	loginCmd.Annotations = map[string]string{"command": "utility"}
	loginCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	loginCmd.Flags().StringVarP(&o.userNameFlag, "username", "u", "", "username, will prompt if not provided")
	loginCmd.Flags().StringVarP(&o.passwordFlag, "password", "p", "", "password, will prompt if not provided")
	loginCmd.Flags().StringVarP(&o.tokenFlag, "token", "t", "", "token, will prompt if not provided")
	loginCmd.Flags().BoolVar(&o.skipTlsFlag, "insecure-skip-tls-verify", false, "If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure")
	loginCmd.Flags().StringVar(&o.caAuthFlag, "certificate-authority", "", "Path to a cert file for the certificate authority")
	return loginCmd
}
