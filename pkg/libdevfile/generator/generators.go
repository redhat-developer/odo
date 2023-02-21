package generator

import (
	"github.com/devfile/library/v2/pkg/devfile/generator"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	
	"github.com/redhat-developer/odo/pkg/kclient"
)

type JobParams struct {
	PodTemplateSpec corev1.PodTemplateSpec
	SpecParams      JobSpecParams
}
type JobSpecParams struct {
	CompletionMode          *batchv1.CompletionMode
	TTLSecondsAfterFinished *int32
	BackOffLimit            *int32
	Parallelism             *int32
	Completion              *int32
	ActiveDeadlineSeconds   *int64
}

func GetJob(jobName string, jobParams JobParams) batchv1.Job {
	return batchv1.Job{
		TypeMeta: generator.GetTypeMeta(kclient.JobsKind, kclient.JobsAPIVersion),
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: batchv1.JobSpec{
			Template:              jobParams.PodTemplateSpec,
			Parallelism:           jobParams.SpecParams.Parallelism,
			Completions:           jobParams.SpecParams.Completion,
			ActiveDeadlineSeconds: jobParams.SpecParams.ActiveDeadlineSeconds,
			BackoffLimit:          jobParams.SpecParams.BackOffLimit,
			// we delete jobs before exiting this function but setting this as a backup in case DeleteJob fails
			TTLSecondsAfterFinished: jobParams.SpecParams.TTLSecondsAfterFinished,
			CompletionMode:          jobParams.SpecParams.CompletionMode,
		},
	}

}
