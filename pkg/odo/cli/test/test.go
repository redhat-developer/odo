package test

import (
	"fmt"

	"github.com/openshift/odo/pkg/auth"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "test"

// TestOptions encapsulates the options for the odo command
type TestOptions struct {
	userName string
	password string
	token    string
	caAuth   string
	skipTLS  bool
	server   string
}

var testExample = templates.Examples(`
  # Log in interactively
  %[1]s

  # Log in to the given server with the given certificate authority file
  %[1]s localhost:8443 --certificate-authority=/path/to/cert.crt

  # Log in to the given server with the given credentials (basic auth)
  %[1]s localhost:8443 --username=myuser --password=mypass

  # Log in to the given server with the given credentials (token)
  %[1]s localhost:8443 --token=xxxxxxxxxxxxxxxxxxxxxxx
`)

// NewTestOptions creates a new TestOptions instance
func NewTestOptions() *TestOptions {
	return &TestOptions{}
}

// Complete completes TestOptions after they've been created
func (o *TestOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	if len(args) == 1 {
		o.server = args[0]
	}
	return
}

// Validate validates the TestOptions based on completed values
func (o *TestOptions) Validate() (err error) {
	return
}

// Run contains the logic for the odo command
func (o *TestOptions) Run() (err error) {
	return auth.Login(o.server, o.userName, o.password, o.token, o.caAuth, o.skipTLS)
}

// NewCmdTest implements the odo tets command
func NewCmdTest(name, fullName string) *cobra.Command {
	o := NewTestOptions()
	testCmd := &cobra.Command{
		Use:     name,
		Short:   "Login to cluster",
		Long:    "Login to cluster",
		Example: fmt.Sprintf(testExample, fullName),
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	// Add a defined annotation in order to appear in the help menu
	testCmd.Annotations = map[string]string{"command": "utility"}
	testCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	testCmd.Flags().StringVarP(&o.userName, "username", "u", "", "username, will prompt if not provided")
	testCmd.Flags().StringVarP(&o.password, "password", "p", "", "password, will prompt if not provided")
	testCmd.Flags().StringVarP(&o.token, "token", "t", "", "token, will prompt if not provided")
	testCmd.Flags().BoolVar(&o.skipTLS, "insecure-skip-tls-verify", false, "If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure")
	testCmd.Flags().StringVar(&o.caAuth, "certificate-authority", "", "Path to a cert file for the certificate authority")
	return testCmd
}
