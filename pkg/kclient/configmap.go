package kclient

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListConfigMaps lists all the configmaps based on the given label selector
func (c *Client) ListConfigMaps(labelSelector string) ([]corev1.ConfigMap, error) {
	listOptions := metav1.ListOptions{}
	if len(labelSelector) > 0 {
		listOptions = metav1.ListOptions{
			LabelSelector: labelSelector,
		}
	}

	cmList, err := c.KubeClient.CoreV1().ConfigMaps(c.Namespace).List(context.TODO(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to get configmap list: %w", err)
	}

	return cmList.Items, nil
}
