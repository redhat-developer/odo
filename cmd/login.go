package cmd

import (
	"github.com/redhat-developer/odo/pkg/login"
	"github.com/spf13/cobra"
)

var Username string
var Password string
var Token string

// versionCmd represents the version command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to cluster",
	Long:  "Login to cluster",
	Example: `  # Print the client version of Odo
  odo version
	`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var server string
		if len(args) == 1 {
			server = args[0]
		}
		client := getOcClient()
		err := login.Login(client, server, Username, Password, Token)
		if err != nil {
			checkError(err, "LOGIN FAILED")
		}

	},
}

func init() {
	// Add a defined annotation in order to appear in the help menu
	loginCmd.Annotations = map[string]string{"command": "utility"}
	loginCmd.SetUsageTemplate(cmdUsageTemplate)
	loginCmd.Flags().StringVarP(&Username, "username", "u", Username, "Username, will prompt if not provided")
	loginCmd.Flags().StringVarP(&Password, "password", "p", Password, "Password, will prompt if not provided")
	loginCmd.Flags().StringVarP(&Token, "token", "t", Token, "Token, will prompt if not provided")

	rootCmd.AddCommand(loginCmd)
}
