/*
Copyright 2019 The Tekton Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Poll Kubernetes resources

After creating Kubernetes resources or making changes to them, you will need to
wait for the system to realize those changes. You can use polling methods to
check the resources reach the desired state.

The `WaitFor*` functions use the Kubernetes
[`wait` package](https://godoc.org/k8s.io/apimachinery/pkg/util/wait). For
polling they use
[`PollImmediate`](https://godoc.org/k8s.io/apimachinery/pkg/util/wait#PollImmediate)
with a [`ConditionFunc`](https://godoc.org/k8s.io/apimachinery/pkg/util/wait#ConditionFunc)
callback function, which returns a `bool` to indicate if the polling should stop
and an `error` to indicate if there was an error.
*/

package test

import (
	"time"

	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"knative.dev/pkg/apis"
)

const (
	interval = 1 * time.Second
	timeout  = 10 * time.Minute
)

// WaitFor waits for the specified ConditionFunc every internal until the timeout.
func WaitFor(waitFunc wait.ConditionFunc) error {
	return wait.PollImmediate(interval, timeout, waitFunc)
}

// eventListenerReady returns a function that checks if all conditions on the
// specified EventListener are true and that the deployment available condition
// is within this set
func eventListenerReady(t *testing.T, c *clients, namespace, name string) wait.ConditionFunc {
	return func() (bool, error) {
		el, err := c.TriggersClient.TriggersV1alpha1().EventListeners(namespace).Get(name, metav1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			t.Log("EventListener not found")
			return false, nil
		}
		t.Log("EventListenerStatus:", el.Status)
		// No conditions have been set yet
		if len(el.Status.Conditions) == 0 {
			return false, nil
		}
		if el.Status.GetCondition(apis.ConditionType(appsv1.DeploymentAvailable)) == nil {
			return false, nil
		}
		for _, cond := range el.Status.Conditions {
			if cond.Status != corev1.ConditionTrue {
				return false, nil
			}
		}
		return true, nil
	}
}

// deploymentNotExist returns a function that checks if the specified Deployment does not exist
func deploymentNotExist(t *testing.T, c *clients, namespace, name string) wait.ConditionFunc {
	return func() (bool, error) {
		_, err := c.KubeClient.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			return true, nil
		}
		return false, nil
	}
}

// serviceNotExist returns a function that checks if the specified Service does not exist
func serviceNotExist(t *testing.T, c *clients, namespace, name string) wait.ConditionFunc {
	return func() (bool, error) {
		_, err := c.KubeClient.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			return true, nil
		}
		return false, nil
	}
}

// pipelineResourceExist returns a function that checks if the specified PipelineResource exists
func pipelineResourceExist(t *testing.T, c *clients, namespace, name string) wait.ConditionFunc {
	return func() (bool, error) {
		_, err := c.ResourceClient.TektonV1alpha1().PipelineResources(namespace).Get(name, metav1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			return false, nil
		}
		return true, err
	}
}
