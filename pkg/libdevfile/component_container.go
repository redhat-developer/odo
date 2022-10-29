package libdevfile

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
)

// containerComponent implements the component interface
type containerComponent struct {
	component  v1alpha2.Component
	devfileObj parser.DevfileObj
}

var _ component = (*containerComponent)(nil)

func newContainerComponent(devfileObj parser.DevfileObj, component v1alpha2.Component) *containerComponent {
	return &containerComponent{
		component:  component,
		devfileObj: devfileObj,
	}
}

func (e *containerComponent) CheckValidity() error {
	return nil
}

func (e *containerComponent) Apply(_ Handler) error {
	return nil
}

func GetJobFromContainerWithCommand(devfileContainer v1alpha2.Component, command v1alpha2.Command) (batchv1.Job, error) {
	var job batchv1.Job
	var container corev1.Container

	job.GenerateName = devfileContainer.Name
	container.Image = devfileContainer.Container.Image
	container.Name = devfileContainer.Name
	container.Command = []string{"/bin/sh", "-c", "(" + command.Exec.CommandLine + ") "}
	container.ImagePullPolicy = corev1.PullIfNotPresent

	job.Spec.Template.Spec.Containers = []corev1.Container{container}
	job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyNever

	// delete the job from the cluster after 100 seconds.
	job.Spec.TTLSecondsAfterFinished = pointer.Int32Ptr(100)

	return job, nil
}
