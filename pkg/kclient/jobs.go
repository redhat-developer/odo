package kclient

import (
	"context"
	"fmt"
	"io"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
)

// constants for volumes
const (
	JobsKind       = "Job"
	JobsAPIVersion = "batch/v1"
	// JobNameOdoMaxLength is the max length of a job name
	// To be on the safe side, we keep the max length less than the original(k8s) max length;
	// we do this because k8s job in odo is created to run exec commands in Deploy mode and this is not a user created resource,
	// so we do not want to break because of any error with job
	JobNameOdoMaxLength = 60
)

func (c *Client) ListJobs(selector string) (*batchv1.JobList, error) {
	return c.KubeClient.BatchV1().Jobs(c.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: selector})
}

// CreateJobs creates a K8s job to execute task
func (c *Client) CreateJob(job batchv1.Job, namespace string) (*batchv1.Job, error) {
	if namespace == "" {
		namespace = c.Namespace
	}
	createdJob, err := c.KubeClient.BatchV1().Jobs(namespace).Create(context.TODO(), &job, metav1.CreateOptions{FieldManager: FieldManager})
	if err != nil {
		return nil, fmt.Errorf("unable to create Jobs: %w", err)
	}
	return createdJob, nil
}

// WaitForJobToComplete to wait until a job completes or fails; it starts printing log or error if the job does not complete execution after 2 minutes
func (c *Client) WaitForJobToComplete(job *batchv1.Job) (*batchv1.Job, error) {
	klog.V(3).Infof("Waiting for Job %s to complete successfully", job.Name)

	w, err := c.KubeClient.BatchV1().Jobs(c.Namespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: fields.Set{"metadata.name": job.Name}.AsSelector().String(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to watch job: %w", err)
	}
	defer w.Stop()

	for {
		val, ok := <-w.ResultChan()
		if !ok {
			break
		}

		wJob, ok := val.Object.(*batchv1.Job)
		if !ok {
			klog.V(4).Infof("did not receive job object, received: %v", val)
			continue
		}
		for _, condition := range wJob.Status.Conditions {
			if condition.Type == batchv1.JobFailed {
				klog.V(4).Infof("Failed to execute the job, reason: %s", condition.String())
				// we return the job as it is in case the caller requires it for further investigation.
				return wJob, fmt.Errorf("failed to execute the job")
			}
			if condition.Type == batchv1.JobComplete {
				return wJob, nil
			}
		}
	}
	return nil, nil
}

// GetJobLogs retrieves pod logs of a job
func (c *Client) GetJobLogs(job *batchv1.Job, containerName string) (io.ReadCloser, error) {
	// Set standard log options
	// RESTClient call to kubernetes
	selector := labels.Set{"controller-uid": string(job.UID), "job-name": job.Name}.AsSelector().String()
	pods, err := c.GetPodsMatchingSelector(selector)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pod found for job %q", job.Name)
	}
	pod := pods.Items[0]
	return c.GetPodLogs(pod.Name, containerName, false)
}

func (c *Client) DeleteJob(jobName string) error {
	propagationPolicy := metav1.DeletePropagationBackground
	return c.KubeClient.BatchV1().Jobs(c.Namespace).Delete(context.Background(), jobName, metav1.DeleteOptions{PropagationPolicy: &propagationPolicy})
}
