package deploy

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	
	dfutil "github.com/devfile/library/v2/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
	
	odogenerator "github.com/redhat-developer/odo/pkg/libdevfile/generator"
	
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/generator"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/klog/v2"
	
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
	containerComps, err := generator.GetContainers(o.devfileObj, common.DevfileOptions{FilterByName: command.Exec.Component})
	if err != nil {
		return err
	}
	if len(containerComps) != 1 {
		return fmt.Errorf("could not find the component")
	}
	
	containerComp := containerComps[0]
	containerComp.Command = []string{"/bin/sh"}
	containerComp.Args = getCmdline(command)
	
	// Create a Kubernetes Job and use the container image referenced by command.Exec.Component
	// Get the component for the command with command.Exec.Component
	jobName := o.componentName + "-" + o.appName + "-" + command.Id + "-" + dfutil.GenerateRandomString(3) // TODO: Is there a function to return the standard odo names?
	completionMode := batchv1.CompletionMode("Indexed")
	jobParams := odogenerator.JobParams{
		PodTemplateSpec: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{containerComp},
				// Setting the restart policy to "never" so that pods are kept around after the job finishes execution; this is helpful in obtaining logs to debug.
				RestartPolicy: "Never",
			},
		},
		SpecParams: odogenerator.JobSpecParams{
			CompletionMode:          &completionMode,
			TTLSecondsAfterFinished: pointer.Int32(60),
			BackOffLimit:            pointer.Int32(1),
		},
	}
	job := odogenerator.GetJob(jobName, jobParams)
	// Set labels and annotations
	job.SetLabels(odolabels.GetLabels(o.componentName, o.appName, component.GetComponentRuntimeFromDevfileMetadata(o.devfileObj.Data.GetMetadata()), odolabels.ComponentDeployMode, false))
	job.Annotations = map[string]string{}
	odolabels.AddCommonAnnotations(job.Annotations)
	odolabels.SetProjectType(job.Annotations, component.GetComponentTypeFromDevfileMetadata(o.devfileObj.Data.GetMetadata()))
	
	//	Make sure there are no existing jobs
	checkAndDeleteExistingJob := func() {
		items, dErr := o.kubeClient.ListJobs(odolabels.GetSelector(o.componentName, o.appName, odolabels.ComponentDeployMode, false))
		if dErr != nil {
			klog.V(4).Infof("failed to list jobs; cause: %s", dErr.Error())
			return
		}
		partialJobName := o.componentName + "-" + o.appName + "-" + command.Id
		for _, item := range items.Items {
			if strings.Contains(item.Name, partialJobName) {
				dErr = o.kubeClient.DeleteJob(item.Name)
				if dErr != nil {
					klog.V(4).Infof("failed to delete job %q; cause: %s", item.Name, dErr.Error())
				}
			}
		}
	}
	checkAndDeleteExistingJob()
	
	log.Sectionf("Executing command:")
	spinner := log.Spinnerf("Executing command in container (command: %s)", command.Id)
	defer spinner.End(false)
	
	var createdJob *batchv1.Job
	createdJob, err = o.kubeClient.CreateJob(job, "")
	if err != nil {
		return err
	}
	defer func() {
		err = o.kubeClient.DeleteJob(createdJob.Name)
		if err != nil {
			klog.V(4).Infof("failed to delete job %q; cause: %s", createdJob.Name, err)
		}
	}()

	var done = make(chan struct{}, 1)
	// Print the tip to use `odo logs` if the command is still running after 1 minute
	go func() {
		select {
		case <-time.After(1 * time.Minute):
			log.Info("\nTip: Run `odo logs --deploy --follow` to get the logs of the command output.")
		case <-done:
			return
		}
	}()

	// Wait for the command to complete execution
	_, err = o.kubeClient.WaitForJobToComplete(createdJob)
	done <- struct{}{}
	if err != nil {
		err = fmt.Errorf("failed to execute (command: %s)", command.Id)
		// Print the job logs if the job failed
		jobLogs, logErr := o.kubeClient.GetJobLogs(createdJob, command.Exec.Component)
		if logErr != nil {
			log.Warningf("failed to fetch the logs of execution; cause: %s", logErr)
		}
		fmt.Println("Execution output:")
		_ = util.DisplayLog(false, jobLogs, log.GetStderr(), o.componentName, 100)
	}
	
	spinner.End(err == nil)

	return err
}

func getCmdline(command v1alpha2.Command) []string {
	// deal with environment variables
	var cmdLine string
	setEnvVariable := util.GetCommandStringFromEnvs(command.Exec.Env)

	if setEnvVariable == "" {
		cmdLine = command.Exec.CommandLine
	} else {
		cmdLine = setEnvVariable + " && " + command.Exec.CommandLine
	}
	var args []string
	if command.Exec.WorkingDir != "" {
		// since we are using /bin/sh -c, the command needs to be within a single double quote instance, for example "cd /tmp && pwd"
		args = []string{"-c", "cd " + command.Exec.WorkingDir + " && " + cmdLine}
	} else {
		args = []string{"-c", cmdLine}
	}
	return args
}
