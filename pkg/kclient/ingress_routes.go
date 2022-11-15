package kclient

import (
	"context"

	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	IngressGVK = schema.GroupVersionKind{
		Group:   "networking.k8s.io",
		Version: "v1",
		Kind:    "Ingress",
	}
	RouteGVK = schema.GroupVersionKind{
		Group:   "route.openshift.io",
		Version: "v1",
		Kind:    "Route",
	}
)

func (c *Client) ListIngresses(namespace, selector string) (*v1.IngressList, error) {
	if namespace == "" {
		namespace = c.Namespace
	}
	return c.KubeClient.NetworkingV1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: selector})
}
