package cli

import (
	"github.com/redhat-developer/odo/pkg/auth"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

var (
	userName string
	password string
	token    string
	caAuth   string
	skipTLS  bool
)

// versionCmd represents the version command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to cluster",
	Long:  "Login to cluster",
	Example: `
  # Log in interactively
  odo login

  # Log in to the given server with the given certificate authority file
  odo login localhost:8443 --certificate-authority=/path/to/cert.crt

  # Log in to the given server with the given credentials (basic auth)
  odo login localhost:8443 --username=myuser --password=mypass

  # Log in to the given server with the given credentials (token)
  odo login localhost:8443 --token=xxxxxxxxxxxxxxxxxxxxxxx
	`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var server string
		if len(args) == 1 {
			server = args[0]
		}
		err := auth.Login(server, userName, password, token, caAuth, skipTLS)
		if err != nil {
			util.CheckError(err, "")
		}
	},
}

func init() {
	// Add a defined annotation in order to appear in the help menu
	loginCmd.Annotations = map[string]string{"command": "utility"}
	loginCmd.SetUsageTemplate(CmdUsageTemplate)
	loginCmd.Flags().StringVarP(&userName, "username", "u", userName, "username, will prompt if not provided")
	loginCmd.Flags().StringVarP(&password, "password", "p", password, "password, will prompt if not provided")
	loginCmd.Flags().StringVarP(&token, "token", "t", token, "token, will prompt if not provided")
	loginCmd.Flags().BoolVar(&skipTLS, "insecure-skip-tls-verify", false, "If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure")
	loginCmd.Flags().StringVar(&caAuth, "certificate-authority", userName, "Path to a cert file for the certificate authority")
	rootCmd.AddCommand(loginCmd)
}
