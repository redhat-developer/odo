/*
Copyright 2019 The Tekton Authors

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

package test

import (
	"fmt"
	"io/ioutil"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CollectPodLogsWithLabel will get the logs of all the Pods given a labelSelector
func CollectPodLogsWithLabel(c kubernetes.Interface, namespace, labelSelector string) (string, error) {
	pods, err := c.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return "", err
	}
	sb := strings.Builder{}
	for _, pod := range pods.Items {
		req := c.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})
		rc, err := req.Stream()
		if err != nil {
			return "", err
		}
		bs, err := ioutil.ReadAll(rc)
		if err != nil {
			return "", err
		}
		sb.WriteString(fmt.Sprintf("\n>>> Pod %s:\n", pod.Name))
		sb.Write(bs)
	}
	return sb.String(), nil
}
