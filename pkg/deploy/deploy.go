package deploy

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/generator"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/util"
)

type DeployClient struct {
	kubeClient kclient.ClientInterface
	fs         filesystem.Filesystem
}

var _ Client = (*DeployClient)(nil)

func NewDeployClient(kubeClient kclient.ClientInterface, fs filesystem.Filesystem) *DeployClient {
	return &DeployClient{
		kubeClient: kubeClient,
		fs:         fs,
	}
}

func (o *DeployClient) Deploy(ctx context.Context) error {
	var (
		devfileObj    = odocontext.GetDevfileObj(ctx)
		devfilePath   = odocontext.GetDevfilePath(ctx)
		path          = filepath.Dir(devfilePath)
		componentName = odocontext.GetComponentName(ctx)
		appName       = odocontext.GetApplication(ctx)
	)
	deployHandler := newDeployHandler(ctx, o.fs, *devfileObj, path, o.kubeClient, appName, componentName)
	return libdevfile.Deploy(*devfileObj, deployHandler)
}

type deployHandler struct {
	ctx           context.Context
	fs            filesystem.Filesystem
	devfileObj    parser.DevfileObj
	path          string
	kubeClient    kclient.ClientInterface
	appName       string
	componentName string
}

var _ libdevfile.Handler = (*deployHandler)(nil)

func newDeployHandler(ctx context.Context, fs filesystem.Filesystem, devfileObj parser.DevfileObj, path string, kubeClient kclient.ClientInterface, appName string, componentName string) *deployHandler {
	return &deployHandler{
		ctx:           ctx,
		fs:            fs,
		devfileObj:    devfileObj,
		path:          path,
		kubeClient:    kubeClient,
		appName:       appName,
		componentName: componentName,
	}
}

// ApplyImage builds and pushes the OCI image to be used on Kubernetes
func (o *deployHandler) ApplyImage(img v1alpha2.Component) error {
	return image.BuildPushSpecificImage(o.ctx, o.fs, img, true)
}

// ApplyKubernetes applies inline Kubernetes YAML from the devfile.yaml file
func (o *deployHandler) ApplyKubernetes(kubernetes v1alpha2.Component) error {
	return component.ApplyKubernetes(odolabels.ComponentDeployMode, o.appName, o.componentName, o.devfileObj, kubernetes, o.kubeClient, o.path)
}

// ApplyOpenShift applies inline OpenShift YAML from the devfile.yaml file
func (o *deployHandler) ApplyOpenShift(openshift v1alpha2.Component) error {
	return component.ApplyKubernetes(odolabels.ComponentDeployMode, o.appName, o.componentName, o.devfileObj, openshift, o.kubeClient, o.path)
}

// Execute will deploy the listed information in the `exec` section of devfile.yaml
// We currently do NOT support this in `odo deploy`.
func (o *deployHandler) Execute(command v1alpha2.Command) error {
	// TODO: Handle cases where `odo deploy` is run again and the job still exists: ✗  unable to create Jobs: jobs.batch "demo-app" already exists
	containerComps, err := generator.GetContainers(o.devfileObj, common.DevfileOptions{FilterByName: command.Exec.Component})
	if err != nil {
		return err
	}
	if len(containerComps) != 1 {
		return fmt.Errorf("could not find the component")
	}
	containerComp := containerComps[0]

	// Create a Kubernetes Job and use the container image referenced by command.Exec.Component
	// Get the component for the command with command.Exec.Component
	completionMode := v1.CompletionMode("Indexed")
	job := v1.Job{
		TypeMeta: generator.GetTypeMeta(kclient.JobsKind, kclient.JobsAPIVersion),
		ObjectMeta: metav1.ObjectMeta{
			// TODO: Check if the name is K8s valid.
			Name: o.componentName + "-" + o.appName + "-" + command.Id, // TODO: Is there a function to return the standard odo names?
		},
		Spec: v1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  containerComp.Name,
							Image: containerComp.Image,
							// TODO: Should we use the command and args already defined inside the container component?
							// TODO: What if the 'sh' binary is not always available?
							Command:         []string{"sh"},
							Args:            []string{"-c", command.Exec.CommandLine},
							ImagePullPolicy: "IfNotPresent",
						},
					},
					// Set the policy to `Never` so that it keeps the pod around and they can be used to debug.
					RestartPolicy: "Never",
				},
			},
			// TODO: change this to 4.
			BackoffLimit:            pointer.Int32(2),
			CompletionMode:          &completionMode,
			TTLSecondsAfterFinished: pointer.Int32(60),
		},
		Status: v1.JobStatus{},
	}

	// Set labels and annotations
	job.SetLabels(odolabels.GetLabels(o.componentName, o.appName, component.GetComponentRuntimeFromDevfileMetadata(o.devfileObj.Data.GetMetadata()), odolabels.ComponentDeployMode, false))
	job.Annotations = map[string]string{}
	odolabels.AddCommonAnnotations(job.Annotations)
	odolabels.SetProjectType(job.Annotations, component.GetComponentTypeFromDevfileMetadata(o.devfileObj.Data.GetMetadata()))

	log.Sectionf("Executing command in container (command: %s)", command.Id)
	spinner := log.Spinnerf("Executing %s\n", command.Exec.CommandLine)
	defer spinner.End(err == nil)

	var createdJob *v1.Job
	createdJob, err = o.kubeClient.CreateJobs(job, "")
	if err != nil {
		return err
	}

	_, err = o.kubeClient.WaitForJobToComplete(job.Name)
	if err != nil {
		// If the job fails, then we fetch pod logs and print them
		jobLogs, err := o.kubeClient.GetJobLogs(createdJob, command.Exec.Component)
		if err != nil {
			return fmt.Errorf("failed to fetch the logs required to identify the reason for execution failure; cause: %w", err)
		}
		err = util.DisplayLog(false, jobLogs, log.GetStderr(), o.componentName, -1)
		if err != nil {
			return fmt.Errorf("unable to log error; cause: %w", err)
		}
		return fmt.Errorf("failed to execute the command")
	}
	// Figure out why pods are not deleted along with the job: Because they do not have the same labels as the job and there's no owner reference set on them.
	return err
}
