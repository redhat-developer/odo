package url

import (
	"fmt"

	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	componentCmd "github.com/openshift/odo/pkg/odo/cli/component"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"

	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended url command name
const RecommendedCommandName = "url"

var (
	urlShortDesc = `Expose component to the outside world`
	urlLongDesc  = ktemplates.LongDesc(`Expose component to the outside world.
		
		The URLs that are generated using this command, can be used to access the deployed components from outside the cluster.`)
)

// NewCmdURL returns the top-level url command
func NewCmdURL(name, fullName string) *cobra.Command {
	urlCreateCmd := NewCmdURLCreate(createRecommendedCommandName, odoutil.GetFullName(fullName, createRecommendedCommandName))
	urlDeleteCmd := NewCmdURLDelete(deleteRecommendedCommandName, odoutil.GetFullName(fullName, deleteRecommendedCommandName))
	urlListCmd := NewCmdURLList(listRecommendedCommandName, odoutil.GetFullName(fullName, listRecommendedCommandName))
	urlCmd := &cobra.Command{
		Use:   name,
		Short: urlShortDesc,
		Long:  urlLongDesc,
		Example: fmt.Sprintf("%s\n%s\n%s",
			urlCreateCmd.Example,
			urlDeleteCmd.Example,
			urlListCmd.Example),
	}

	// Add a defined annotation in order to appear in the help menu
	urlCmd.Annotations = map[string]string{"command": "main"}
	urlCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	urlCmd.AddCommand(urlCreateCmd, urlDeleteCmd, urlListCmd)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(urlListCmd)
	projectCmd.AddProjectFlag(urlCreateCmd)
	projectCmd.AddProjectFlag(urlDeleteCmd)

	//Adding `--application` flag
	appCmd.AddApplicationFlag(urlListCmd)
	appCmd.AddApplicationFlag(urlDeleteCmd)
	appCmd.AddApplicationFlag(urlCreateCmd)

	//Adding `--component` flag
	componentCmd.AddComponentFlag(urlDeleteCmd)
	componentCmd.AddComponentFlag(urlListCmd)
	componentCmd.AddComponentFlag(urlCreateCmd)

	return urlCmd
}
