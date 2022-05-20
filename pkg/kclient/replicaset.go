package kclient

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetReplicaSetByName gets a replicaset matching the given name
func (c *Client) GetReplicaSetByName(name string) (*appsv1.ReplicaSet, error) {
	replicaSet, err := c.KubeClient.AppsV1().ReplicaSets(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return &appsv1.ReplicaSet{}, err
	}
	return replicaSet, nil
}
