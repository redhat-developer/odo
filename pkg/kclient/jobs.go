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
	JobsResource   = "jobs"
	JobsAPIVersion = "batch/v1"
)

func (c *Client) CreateJobs(job batchv1.Job, namespace string) (*batchv1.Job, error) {
	if namespace == "" {
		namespace = c.Namespace
	}
	createdJob, err := c.KubeClient.BatchV1().Jobs(namespace).Create(context.TODO(), &job, metav1.CreateOptions{FieldManager: FieldManager})
	if err != nil {
		return nil, fmt.Errorf("unable to create Jobs: %w", err)
	}
	return createdJob, nil
}

func (c *Client) WaitForJobToComplete(jobName string) (*batchv1.Job, error) {
	klog.V(3).Infof("Waiting for Job %s to complete successfully", jobName)

	w, err := c.KubeClient.BatchV1().Jobs(c.Namespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: fields.Set{"metadata.name": jobName}.AsSelector().String(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to watch job: %w", err)
	}
	defer w.Stop()
	for {
		select {
		case val, ok := <-w.ResultChan():
			if !ok {
				break
			}
			wJob := val.Object.(*batchv1.Job)
			for _, condition := range wJob.Status.Conditions {
				if condition.Type == batchv1.JobFailed {
					klog.V(4).Infof("Failed to run the job, reason: %s", condition.String())
					// we return the job as it is in case the caller requires it for further investigation.
					return wJob, fmt.Errorf("failed to run the job")
				}
				if condition.Type == batchv1.JobComplete {
					return wJob, nil
				}
			}
		}
	}
	return nil, nil
}

func (c *Client) DeleteJob(jobName string) error {
	return c.KubeClient.BatchV1().Jobs(c.Namespace).Delete(context.TODO(), jobName, metav1.DeleteOptions{})
}

func (c *Client) GetJobLogs(job *batchv1.Job, containerName string) (io.ReadCloser, error) {
	// Set standard log options
	// RESTClient call to kubernetes
	selector := labels.Set{"controller-uid": string(job.UID), "job-name": job.Name}.AsSelector().String()
	pods, err := c.GetPodsMatchingSelector(selector)
	if err != nil {
		return nil, err
	}
	pod := pods.Items[0]
	return c.GetPodLogs(pod.Name, containerName, false)
}
