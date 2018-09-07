package cmd

import (
	"github.com/spf13/cobra"
)

var componentLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of the active session out by clearing saved tokens",
	Long:  "Log out of the active session out by clearing saved tokens",
	Example: `  # Logout
  odo logout
	`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		client.RunLogout()
	},
}

func init() {
	// Add a defined annotation in order to appear in the help menu
	componentListCmd.Annotations = map[string]string{"command": "component"}

	rootCmd.AddCommand(componentLogoutCmd)
}
