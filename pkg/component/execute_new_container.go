package component

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/generator"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"

	"github.com/redhat-developer/odo/pkg/configAutomount"
	"github.com/redhat-developer/odo/pkg/dev/kubedev/storage"
	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	odogenerator "github.com/redhat-developer/odo/pkg/libdevfile/generator"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/util"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/utils/pointer"
)

func ExecuteInNewContainer(
	ctx context.Context,
	kubeClient kclient.ClientInterface,
	configAutomountClient configAutomount.Client,
	devfileObj parser.DevfileObj,
	componentName string,
	appName string,
	command v1alpha2.Command,
) error {
	policy, err := kubeClient.GetCurrentNamespacePolicy()
	if err != nil {
		return err
	}
	podTemplateSpec, err := generator.GetPodTemplateSpec(devfileObj, generator.PodTemplateParams{
		Options: common.DevfileOptions{
			FilterByName: command.Exec.Component,
		},
		PodSecurityAdmissionPolicy: policy,
	})
	if err != nil {
		return err
	}
	// Setting the restart policy to "never" so that pods are kept around after the job finishes execution; this is helpful in obtaining logs to debug.
	podTemplateSpec.Spec.RestartPolicy = "Never"

	if len(podTemplateSpec.Spec.Containers) != 1 {
		return fmt.Errorf("could not find the component")
	}

	podTemplateSpec.Spec.Containers[0].Command = []string{"/bin/sh"}
	podTemplateSpec.Spec.Containers[0].Args = getCmdline(command)

	volumes, err := storage.GetAutomountVolumes(configAutomountClient, podTemplateSpec.Spec.Containers, podTemplateSpec.Spec.InitContainers)
	if err != nil {
		return err
	}

	podTemplateSpec.Spec.Volumes = volumes

	// Create a Kubernetes Job and use the container image referenced by command.Exec.Component
	// Get the component for the command with command.Exec.Component
	getJobName := func() string {
		maxLen := kclient.JobNameOdoMaxLength - len(command.Id)
		// We ignore the error here because our component name or app name will never be empty; which are the only cases when an error might be raised.
		name, _ := util.NamespaceKubernetesObjectWithTrim(componentName, appName, maxLen)
		name += "-" + command.Id
		return name
	}
	completionMode := batchv1.CompletionMode("Indexed")
	jobParams := odogenerator.JobParams{
		TypeMeta: generator.GetTypeMeta(kclient.JobsKind, kclient.JobsAPIVersion),
		ObjectMeta: metav1.ObjectMeta{
			Name: getJobName(),
		},
		PodTemplateSpec: *podTemplateSpec,
		SpecParams: odogenerator.JobSpecParams{
			CompletionMode:          &completionMode,
			TTLSecondsAfterFinished: pointer.Int32(60),
			BackOffLimit:            pointer.Int32(1),
		},
	}
	job := odogenerator.GetJob(jobParams)
	// Set labels and annotations
	job.SetLabels(odolabels.GetLabels(componentName, appName, GetComponentRuntimeFromDevfileMetadata(devfileObj.Data.GetMetadata()), odolabels.ComponentDeployMode, false))
	job.Annotations = map[string]string{}
	odolabels.AddCommonAnnotations(job.Annotations)
	odolabels.SetProjectType(job.Annotations, GetComponentTypeFromDevfileMetadata(devfileObj.Data.GetMetadata()))

	//	Make sure there are no existing jobs
	checkAndDeleteExistingJob := func() {
		items, dErr := kubeClient.ListJobs(odolabels.GetSelector(componentName, appName, odolabels.ComponentDeployMode, false))
		if dErr != nil {
			klog.V(4).Infof("failed to list jobs; cause: %s", dErr.Error())
			return
		}
		jobName := getJobName()
		for _, item := range items.Items {
			if strings.Contains(item.Name, jobName) {
				dErr = kubeClient.DeleteJob(item.Name)
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
	createdJob, err = kubeClient.CreateJob(job, "")
	if err != nil {
		return err
	}
	defer func() {
		err = kubeClient.DeleteJob(createdJob.Name)
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
	_, err = kubeClient.WaitForJobToComplete(createdJob)
	done <- struct{}{}

	spinner.End(err == nil)

	if err != nil {
		err = fmt.Errorf("failed to execute (command: %s)", command.Id)
		// Print the job logs if the job failed
		jobLogs, logErr := kubeClient.GetJobLogs(createdJob, command.Exec.Component)
		if logErr != nil {
			log.Warningf("failed to fetch the logs of execution; cause: %s", logErr)
		}
		fmt.Println("Execution output:")
		_ = util.DisplayLog(false, jobLogs, log.GetStderr(), componentName, 100)
	}

	return err
}
