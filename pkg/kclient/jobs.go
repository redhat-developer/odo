package kclient

import (
	"context"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) CreateJob(job batchv1.Job) error {
	_, err := c.KubeClient.BatchV1().Jobs(c.Namespace).Create(context.TODO(), &job, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
