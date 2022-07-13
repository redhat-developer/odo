package kclient

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func (c *Client) PodWarningEventWatcher(ctx context.Context) (watch.Interface, error) {
	selector := "involvedObject.kind=Pod,involvedObject.apiVersion=v1,type=Warning"
	ns := c.GetCurrentNamespace()
	return c.GetClient().CoreV1().Events(ns).
		Watch(ctx, metav1.ListOptions{
			FieldSelector: selector,
		})
}
