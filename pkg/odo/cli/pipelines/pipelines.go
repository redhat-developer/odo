package pipelines

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/openshift/odo/pkg/odo/cli/pipelines/environment"
	"github.com/openshift/odo/pkg/odo/cli/pipelines/service"
	"github.com/openshift/odo/pkg/odo/cli/pipelines/webhook"
	odoutil "github.com/openshift/odo/pkg/odo/util"
)

// RecommendedCommandName is the recommended pipelines command name.
const RecommendedCommandName = "pipelines"

// NewCmdPipelines implements the component odo command
func NewCmdPipelines(name, fullName string) *cobra.Command {
	var pipelinesCmd = &cobra.Command{
		Use:   name,
		Short: "Pipeline operations",
		Example: fmt.Sprintf("%s\n%s\n\n  See sub-commands individually for more examples",
			fullName, InitRecommendedCommandName),
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	initCmd := NewCmdInit(InitRecommendedCommandName, odoutil.GetFullName(fullName, InitRecommendedCommandName))
	bootstrapCmd := NewCmdBootstrap(BootstrapRecommendedCommandName, odoutil.GetFullName(fullName, BootstrapRecommendedCommandName))
	envCmd := environment.NewCmdEnv(environment.EnvRecommendedCommandName, odoutil.GetFullName(fullName, environment.EnvRecommendedCommandName))
	serviceCmd := service.NewCmd(service.RecommendedCommandName, odoutil.GetFullName(fullName, service.RecommendedCommandName))

	webhookCmd := webhook.NewCmdWebhook(webhook.RecommendedCommandName, odoutil.GetFullName(fullName, webhook.RecommendedCommandName))

	pipelinesCmd.AddCommand(initCmd)
	pipelinesCmd.AddCommand(bootstrapCmd)
	pipelinesCmd.AddCommand(envCmd)
	pipelinesCmd.AddCommand(serviceCmd)
	pipelinesCmd.AddCommand(webhookCmd)

	buildCmd := NewCmdBuild(BuildRecommendedCommandName, odoutil.GetFullName(fullName, BuildRecommendedCommandName))
	pipelinesCmd.AddCommand(buildCmd)

	pipelinesCmd.Annotations = map[string]string{"command": "main"}
	pipelinesCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return pipelinesCmd
}
