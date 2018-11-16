package cli

import (
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	"os"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of the current OpenShift session",
	Long:  "Log out of the current OpenShift session",
	Example: `  # Logout
  odo logout
	`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		err := client.RunLogout(os.Stdout)
		util.CheckError(err, "")
	},
}

func init() {
	// Add a defined annotation in order to appear in the help menu
	logoutCmd.Annotations = map[string]string{"command": "utility"}
	logoutCmd.SetUsageTemplate(cmdUsageTemplate)

	rootCmd.AddCommand(logoutCmd)
}
