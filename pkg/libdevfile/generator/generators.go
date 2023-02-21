package generator

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JobParams struct {
	TypeMeta        metav1.TypeMeta
	ObjectMeta      metav1.ObjectMeta
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

func GetJob(jobParams JobParams) batchv1.Job {
	return batchv1.Job{
		TypeMeta:   jobParams.TypeMeta,
		ObjectMeta: jobParams.ObjectMeta,
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
