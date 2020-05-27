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

// This file contains initialization logic for the tests, such as special magical global state that needs to be initialized.

package test

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/tektoncd/pipeline/pkg/names"
	"golang.org/x/xerrors"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	knativetest "knative.dev/pkg/test"
	"knative.dev/pkg/test/logging"

	// Mysteriously by k8s libs, or they fail to create `KubeClient`s from config. Apparently just importing it is enough. @_@ side effects @_@. https://github.com/kubernetes/client-go/issues/242
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// Mysteriously by k8s libs, or they fail to create `KubeClient`s when using oidc authentication. Apparently just importing it is enough. @_@ side effects @_@. https://github.com/kubernetes/client-go/issues/345
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

const triggersNamespace = "tekton-pipelines"

var initMetrics sync.Once

func setup(t *testing.T) (*clients, string) {
	t.Helper()
	namespace := names.SimpleNameGenerator.RestrictLengthWithRandomSuffix("arrakis")

	initializeLogsAndMetrics(t)

	c := newClients(t, knativetest.Flags.Kubeconfig, knativetest.Flags.Cluster)
	createNamespace(t, namespace, c.KubeClient)
	verifyDefaultServiceAccountExists(t, namespace, c.KubeClient)
	return c, namespace
}

func tearDown(t *testing.T, cs *clients, namespace string) {
	t.Helper()
	if cs.KubeClient == nil {
		return
	}
	if t.Failed() {
		header(t.Logf, fmt.Sprintf("Dumping objects from %s", namespace))
		bs, err := getCRDYaml(cs, namespace)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(string(bs))
		}

		header(t.Logf, fmt.Sprintf("Dumping logs from tekton-triggers-controller in namespace %s", triggersNamespace))
		controllerLogs, err := CollectPodLogsWithLabel(cs.KubeClient, triggersNamespace, "app=tekton-triggers-controller")
		if err != nil {
			t.Logf("Could not get logs for tekton-triggers-controller Pod: %s", err)
		} else {
			t.Log(controllerLogs)
		}

		header(t.Logf, fmt.Sprintf("Dumping logs from tekton-triggers-webhook in namespace %s", triggersNamespace))
		webhookLogs, err := CollectPodLogsWithLabel(cs.KubeClient, triggersNamespace, "app=tekton-triggers-webhook")
		if err != nil {
			t.Logf("Could not get logs for tekton-triggers-webhook Pod: %s", err)
		} else {
			t.Log(webhookLogs)
		}

		header(t.Logf, fmt.Sprintf("Dumping logs from EventListener sinks in namespace %s", namespace))
		elSinkLogs, err := CollectPodLogsWithLabel(cs.KubeClient, namespace, "triggers=eventlistener")
		if err != nil {
			t.Logf("Could not get logs for EventListener sink Pods: %s", err)
		} else {
			t.Log(elSinkLogs)
		}
	}

	if os.Getenv("TEST_KEEP_NAMESPACES") == "" {
		t.Logf("Deleting namespace %s", namespace)
		if err := cs.KubeClient.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{}); err != nil {
			t.Errorf("Failed to delete namespace %s: %s", namespace, err)
		}
	}
}

func header(logf logging.FormatLogger, text string) {
	left := "### "
	right := " ###"
	txt := left + text + right
	bar := strings.Repeat("#", len(txt))
	logf(bar)
	logf(txt)
	logf(bar)
}

func initializeLogsAndMetrics(t *testing.T) {
	t.Helper()
	initMetrics.Do(func() {
		flag.Parse()
		if err := flag.Set("alsologtostderr", "true"); err != nil {
			t.Fatalf("Failed to set 'alsologtostderr' flag to 'true': %s", err)
		}
		logging.InitializeLogger(knativetest.Flags.LogVerbose)
	})
}

func createNamespace(t *testing.T, namespace string, kubeClient kubernetes.Interface) {
	t.Helper()
	t.Logf("Create namespace %s to deploy to", namespace)
	if _, err := kubeClient.CoreV1().Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}); err != nil {
		t.Fatalf("Failed to create namespace %s for tests: %s", namespace, err)
	}
}

func verifyDefaultServiceAccountExists(t *testing.T, namespace string, kubeClient kubernetes.Interface) {
	t.Helper()
	defaultSA := "default"
	t.Logf("Verify SA %s is created in namespace %s", defaultSA, namespace)

	if err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		_, err := kubeClient.CoreV1().ServiceAccounts(namespace).Get(defaultSA, metav1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			return false, nil
		}
		return true, err
	}); err != nil {
		t.Fatalf("Failed to get SA %q in namespace %q for tests: %s", defaultSA, namespace, err)
	}
}

func getCRDYaml(cs *clients, ns string) ([]byte, error) {
	var output []byte
	printOrAdd := func(kind, name string, i interface{}) {
		bs, err := yaml.Marshal(i)
		if err != nil {
			return
		}
		output = append(output, []byte("\n---\n")...)
		output = append(output, bs...)
	}

	ctbs, err := cs.TriggersClient.TriggersV1alpha1().ClusterTriggerBindings().List(metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("could not get ClusterTriggerBindings: %w", err)
	}
	for _, i := range ctbs.Items {
		printOrAdd("ClusterTriggerBinding", i.Name, i)
	}

	els, err := cs.TriggersClient.TriggersV1alpha1().EventListeners(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("could not get EventListeners: %w", err)
	}
	for _, i := range els.Items {
		printOrAdd("EventListener", i.Name, i)
	}

	tbs, err := cs.TriggersClient.TriggersV1alpha1().TriggerBindings(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("could not get TriggerBindings: %w", err)
	}
	for _, i := range tbs.Items {
		printOrAdd("TriggerBindings", i.Name, i)
	}
	// TODO: Update TriggerTemplates Marshalling so it isn't a byte array in debug log
	tts, err := cs.TriggersClient.TriggersV1alpha1().TriggerTemplates(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("could not get TriggerTemplates: %w", err)
	}
	for _, i := range tts.Items {
		printOrAdd("TriggerTemplate", i.Name, i)
	}

	pods, err := cs.KubeClient.CoreV1().Pods(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("could not get Pods: %w", err)
	}
	for _, i := range pods.Items {
		printOrAdd("Pod", i.Name, i)
	}

	services, err := cs.KubeClient.CoreV1().Services(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("could not get Services: %w", err)
	}
	for _, i := range services.Items {
		printOrAdd("Service", i.Name, i)
	}

	roles, err := cs.KubeClient.RbacV1().Roles(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("could not get Roles: %w", err)
	}
	for _, i := range roles.Items {
		printOrAdd("Role", i.Name, i)
	}

	roleBindings, err := cs.KubeClient.RbacV1().RoleBindings(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("could not get RoleBindings: %w", err)
	}
	for _, i := range roleBindings.Items {
		printOrAdd("Role", i.Name, i)
	}

	return output, nil
}
