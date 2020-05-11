package logs

import (
	"fmt"

	"k8s.io/kubectl/pkg/cmd/logs"

	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"

	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	buildv1client "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	buildhelpers "github.com/openshift/oc/pkg/helpers/build"
)

var (
	logsLong = templates.LongDesc(`
		Print the logs for a resource

		Supported resources are builds, build configs (bc), deployment configs (dc), and pods.
		When a pod is specified and has more than one container, the container name should be
		specified via -c. When a build config or deployment config is specified, you can view
		the logs for a particular version of it via --version.

		If your pod is failing to start, you may need to use the --previous option to see the
		logs of the last attempt.`)

	logsExample = templates.Examples(`
		# Start streaming the logs of the most recent build of the openldap build config.
	  %[1]s logs -f bc/openldap

	  # Start streaming the logs of the latest deployment of the mysql deployment config.
	  %[1]s logs -f dc/mysql

	  # Get the logs of the first deployment for the mysql deployment config. Note that logs
	  # from older deployments may not exist either because the deployment was successful
	  # or due to deployment pruning or manual deletion of the deployment.
	  %[1]s logs --version=1 dc/mysql

	  # Return a snapshot of ruby-container logs from pod backend.
	  %[1]s logs backend -c ruby-container

	  # Start streaming of ruby-container logs from pod backend.
	  %[1]s logs -f pod/backend -c ruby-container`)
)

// LogsOptions holds all the necessary options for running oc logs.
type LogsOptions struct {
	// Client enables access to the Build object when processing
	// build logs for Jenkins Pipeline Strategy builds
	Client buildv1client.BuildV1Interface

	Version int64

	// Embed kubectl's LogsOptions directly.
	*logs.LogsOptions
}

func NewLogsOptions(streams genericclioptions.IOStreams) *LogsOptions {
	return &LogsOptions{
		LogsOptions: logs.NewLogsOptions(streams, false),
	}
}

// NewCmdLogs creates a new logs command that supports OpenShift resources.
func NewCmdLogs(baseName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewLogsOptions(streams)
	cmd := &cobra.Command{
		Use:        "logs [-f] [-p] (POD | TYPE/NAME) [-c CONTAINER]",
		Short:      "Print the logs for a container in a pod",
		Long:       logsLong,
		Example:    fmt.Sprintf(logsExample, baseName),
		SuggestFor: []string{"builds", "deployments"},
		Run: func(cmd *cobra.Command, args []string) {
			kcmdutil.CheckErr(o.Complete(f, cmd, args))
			kcmdutil.CheckErr(o.Validate(args))
			kcmdutil.CheckErr(o.RunLog())
		},
	}

	o.LogsOptions.AddFlags(cmd)
	cmd.Flags().Int64Var(&o.Version, "version", o.Version, "View the logs of a particular build or deployment by version if greater than zero")

	return cmd
}

// Complete calls the upstream Complete for the logs command and then resolves the
// resource a user requested to view its logs and creates the appropriate logOptions
// object for it.
func (o *LogsOptions) Complete(f kcmdutil.Factory, cmd *cobra.Command, args []string) error {
	return o.LogsOptions.Complete(f, cmd, args)
}

// Validate runs the upstream validation for the logs command and then it
// will validate any OpenShift-specific log options.
func (o *LogsOptions) Validate(args []string) error {
	return o.LogsOptions.Validate()
}

// RunLog will run the upstream logs command and may use an OpenShift
// logOptions object.
func (o *LogsOptions) RunLog() error {
	podLogOptions := o.LogsOptions.Options.(*corev1.PodLogOptions)
	var (
		isPipeline bool
		build      *buildv1.Build
	)
	switch t := o.LogsOptions.Object.(type) {
	case *buildv1.Build:
		build = t
		isPipeline = t.Spec.CommonSpec.Strategy.JenkinsPipelineStrategy != nil
		o.LogsOptions.Options = o.buildLogOptions(podLogOptions)

	case *buildv1.BuildConfig:
		buildName := buildhelpers.BuildNameForConfigVersion(t.ObjectMeta.Name, int(t.Status.LastVersion))
		isPipeline = t.Spec.CommonSpec.Strategy.JenkinsPipelineStrategy != nil
		if isPipeline {
			build, _ = o.Client.Builds(o.LogsOptions.Namespace).Get(buildName, metav1.GetOptions{})
			if build == nil {
				return fmt.Errorf("the build %s for build config %s was not found", buildName, t.Name)
			}
		}
		o.LogsOptions.Options = o.buildLogOptions(podLogOptions)

	case *appsv1.DeploymentConfig:
		o.LogsOptions.Options = o.deployLogOptions(podLogOptions)
	}

	if !isPipeline {
		return o.LogsOptions.RunLogs()
	}

	urlString, _ := build.Annotations[buildv1.BuildJenkinsBlueOceanLogURLAnnotation]
	if len(urlString) == 0 {
		return fmt.Errorf("the pipeline strategy build %s does not yet contain the log URL; wait a few moments, then try again", build.Name)
	}
	fmt.Fprintf(o.LogsOptions.Out, "info: logs available at %s\n", urlString)

	return nil
}

func (o *LogsOptions) buildLogOptions(podLogOptions *corev1.PodLogOptions) *buildv1.BuildLogOptions {
	bopts := &buildv1.BuildLogOptions{
		Container:    podLogOptions.Container,
		Follow:       podLogOptions.Follow,
		Previous:     podLogOptions.Previous,
		SinceSeconds: podLogOptions.SinceSeconds,
		SinceTime:    podLogOptions.SinceTime,
		Timestamps:   podLogOptions.Timestamps,
		TailLines:    podLogOptions.TailLines,
		LimitBytes:   podLogOptions.LimitBytes,
	}
	if o.Version != 0 {
		bopts.Version = &o.Version
	}
	return bopts
}

func (o *LogsOptions) deployLogOptions(podLogOptions *corev1.PodLogOptions) *appsv1.DeploymentLogOptions {
	dopts := &appsv1.DeploymentLogOptions{
		Container:    podLogOptions.Container,
		Follow:       podLogOptions.Follow,
		Previous:     podLogOptions.Previous,
		SinceSeconds: podLogOptions.SinceSeconds,
		SinceTime:    podLogOptions.SinceTime,
		Timestamps:   podLogOptions.Timestamps,
		TailLines:    podLogOptions.TailLines,
		LimitBytes:   podLogOptions.LimitBytes,
	}
	if o.Version != 0 {
		dopts.Version = &o.Version
	}
	return dopts
}
