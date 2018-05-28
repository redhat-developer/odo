package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/redhat-developer/odo/pkg/url"
	"github.com/spf13/cobra"
)

var (
	urlListComponent   string
	urlListApplication string
	urlForceDeleteflag bool
)

var urlCmd = &cobra.Command{
	Use:   "url",
	Short: "Expose component to the outside world",
	Long: `Expose component to the outside world.

The URLs that are generated using this command, can be used to access the deployed components from outside the cluster.`,
	Example: fmt.Sprintf("%s\n%s\n%s",
		urlCreateCmd.Example,
		urlDeleteCmd.Example,
		urlListCmd.Example),
}

var urlCreateCmd = &cobra.Command{
	Use:   "create [component name]",
	Short: "Create a URL for a component",
	Long: `Create a URL for a component.

The created URL can be used to access the specified component from outside the OpenShift cluster.
`,
	Example: `  # Create a URL for the current component.
  odo url create

  # Create a URL for a specific component
  odo url create mycomponent
	`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)

		var cmp string
		switch len(args) {
		case 0:
			var err error
			cmp, err = component.GetCurrent(client, applicationName, projectName)
			checkError(err, "")
		case 1:
			cmp = args[0]
		default:
			fmt.Println("unable to get component")
			os.Exit(1)
		}

		fmt.Printf("Adding URL to component: %v\n", cmp)
		u, err := url.Create(client, cmp, applicationName)
		checkError(err, "")
		fmt.Printf("URL created for component: %v\n\n"+
			"%v - %v\n", cmp, u.Name, url.GetUrlString(*u))
	},
}

var urlDeleteCmd = &cobra.Command{
	Use:   "delete <URL>",
	Short: "Delete a URL",
	Long:  `Delete the given URL, hence making the service inaccessible.`,
	Example: `  # Delete a URL to a component
  odo url delete myurl
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		u := args[0]
		var confirmDeletion string
		if urlForceDeleteflag {
			confirmDeletion = "y"
		} else {
			fmt.Printf("Are you sure you want to delete the url %v? [y/N] ", u)
			fmt.Scanln(&confirmDeletion)
		}

		if strings.ToLower(confirmDeletion) == "y" {
			err := url.Delete(client, u)
			checkError(err, "")
			fmt.Printf("Deleted URL: %v\n", u)
		} else {
			fmt.Printf("Aborting deletion of url: %v\n", u)
		}

	},
}

var urlListCmd = &cobra.Command{
	Use:   "list",
	Short: "List URLs",
	Long:  `Lists all the available URLs which can be used to access the components.`,
	Example: ` # List the available URLs
  odo url list
	`,
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()

		var app string
		if urlListApplication == "" {
			var err error
			app, err = application.GetCurrent(client)
			checkError(err, "")
		} else {
			app = urlListApplication
		}

		cmp := urlListComponent
		urls, err := url.List(client, cmp, app)
		checkError(err, "")

		if len(urls) == 0 {
			fmt.Printf("No URLs found for component %v in application %v\n", cmp, app)
		} else {
			fmt.Printf("Found the following URLs for component %v in application %v:\n", cmp, app)
			for _, u := range urls {
				fmt.Printf("%v - %v\n", u.Name, url.GetUrlString(u))
			}
		}
	},
}

func init() {
	urlDeleteCmd.Flags().BoolVarP(&urlForceDeleteflag, "force", "f", false, "Delete url without prompting")
	urlListCmd.Flags().StringVarP(&urlListApplication, "application", "a", "", "list URLs for application")
	urlListCmd.Flags().StringVarP(&urlListComponent, "component", "c", "", "list URLs for component")

	urlCmd.AddCommand(urlListCmd)
	urlCmd.AddCommand(urlDeleteCmd)
	urlCmd.AddCommand(urlCreateCmd)
	rootCmd.AddCommand(urlCmd)
}
