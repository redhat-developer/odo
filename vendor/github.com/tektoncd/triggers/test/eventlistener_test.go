// +build e2e

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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	eventReconciler "github.com/tektoncd/triggers/pkg/reconciler/v1alpha1/eventlistener"
	"github.com/tektoncd/triggers/pkg/sink"
	bldr "github.com/tektoncd/triggers/test/builder"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	knativetest "knative.dev/pkg/test"
)

const (
	resourceLabel = triggersv1.GroupName + triggersv1.EventListenerLabelKey
	triggerLabel  = triggersv1.GroupName + triggersv1.TriggerLabelKey
	eventIDLabel  = triggersv1.GroupName + triggersv1.EventIDLabelKey

	examplePRJsonFilename = "pr.json"
)

func loadExamplePREventBytes() ([]byte, error) {
	path := filepath.Join("testdata", examplePRJsonFilename)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load testdata example PullRequest event data: %v", err)
	}
	return bytes, nil
}

func TestEventListenerCreate(t *testing.T) {
	c, namespace := setup(t)
	t.Parallel()

	defer tearDown(t, c, namespace)
	knativetest.CleanupOnInterrupt(func() { tearDown(t, c, namespace) }, t.Logf)

	t.Log("Start EventListener e2e test")

	// TemplatedPipelineResources
	pr1 := v1alpha1.PipelineResource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PipelineResource",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr1",
			Namespace: namespace,
			Labels: map[string]string{
				"$(params.oneparam)": "$(params.oneparam)",
			},
		},
		Spec: v1alpha1.PipelineResourceSpec{
			Type: "git",
		},
	}
	pr1Bytes, err := json.Marshal(pr1)
	if err != nil {
		t.Fatalf("Error marshalling PipelineResource 1: %s", err)
	}

	// This is a templated resource, which does not have a namespace.
	// This is defaulted to the EventListener namespace.
	pr2 := v1alpha1.PipelineResource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PipelineResource",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "pr2",
			Labels: map[string]string{
				"$(params.twoparamname)": "$(params.twoparamvalue)",
			},
		},
		Spec: v1alpha1.PipelineResourceSpec{
			Type: "git",
			Params: []v1alpha1.ResourceParam{
				{Name: "license", Value: "$(params.license)"},
				{Name: "header", Value: "$(params.header)"},
				{Name: "prmessage", Value: "$(params.prmessage)"},
			},
		},
	}

	pr2Bytes, err := json.Marshal(pr2)
	if err != nil {
		t.Fatalf("Error marshalling ResourceTemplate PipelineResource 2: %s", err)
	}

	// TriggerTemplate
	tt, err := c.TriggersClient.TektonV1alpha1().TriggerTemplates(namespace).Create(
		bldr.TriggerTemplate("my-triggertemplate", "",
			bldr.TriggerTemplateSpec(
				bldr.TriggerTemplateParam("oneparam", "", ""),
				bldr.TriggerTemplateParam("twoparamname", "", ""),
				bldr.TriggerTemplateParam("twoparamvalue", "", "defaultvalue"),
				bldr.TriggerTemplateParam("license", "", ""),
				bldr.TriggerTemplateParam("header", "", ""),
				bldr.TriggerTemplateParam("prmessage", "", ""),
				bldr.TriggerResourceTemplate(pr1Bytes),
				bldr.TriggerResourceTemplate(pr2Bytes),
			),
		),
	)
	if err != nil {
		t.Fatalf("Error creating TriggerTemplate: %s", err)
	}

	// TriggerBinding
	tb, err := c.TriggersClient.TektonV1alpha1().TriggerBindings(namespace).Create(
		bldr.TriggerBinding("my-triggerbinding", "",
			bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("oneparam", "$(body.action)"),
				bldr.TriggerBindingParam("twoparamname", "$(body.pull_request.state)"),
				bldr.TriggerBindingParam("license", "$(body.repository.license)"),
				bldr.TriggerBindingParam("header", "$(header)"),
				bldr.TriggerBindingParam("prmessage", "$(body.pull_request.body)"),
			),
		),
	)
	if err != nil {
		t.Fatalf("Error creating TriggerBinding: %s", err)
	}

	// ServiceAccount + Role + RoleBinding to authorize the creation of our
	// templated resources
	sa, err := c.KubeClient.CoreV1().ServiceAccounts(namespace).Create(
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: "my-serviceaccount"},
		},
	)
	if err != nil {
		t.Fatalf("Error creating ServiceAccount: %s", err)
	}
	_, err = c.KubeClient.RbacV1().Roles(namespace).Create(
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{"tekton.dev"},
					Resources: []string{"eventlisteners", "triggerbindings", "triggertemplates", "pipelineresources"},
					Verbs:     []string{"create", "get"},
				},
				{
					APIGroups: []string{""},
					Resources: []string{"configmaps"},
					Verbs:     []string{"get", "list", "watch"},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("Error creating Role: %s", err)
	}
	_, err = c.KubeClient.RbacV1().RoleBindings(namespace).Create(
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: "my-rolebinding"},
			Subjects: []rbacv1.Subject{{
				Kind:      "ServiceAccount",
				Name:      sa.Name,
				Namespace: namespace,
			}},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     "my-role",
			},
		},
	)
	if err != nil {
		t.Fatalf("Error creating RoleBinding: %s", err)
	}

	// EventListener
	el, err := c.TriggersClient.TektonV1alpha1().EventListeners(namespace).Create(
		bldr.EventListener("my-eventlistener", namespace,
			bldr.EventListenerMeta(
				bldr.Label("triggers", "eventlistener"),
			),
			bldr.EventListenerSpec(
				bldr.EventListenerServiceAccount(sa.Name),
				bldr.EventListenerTrigger(tb.Name, tt.Name, ""),
			),
		))
	if err != nil {
		t.Fatalf("Failed to create EventListener: %s", err)
	}

	// Verify the EventListener to be ready
	if err := WaitFor(eventListenerReady(t, c, namespace, el.Name)); err != nil {
		t.Fatalf("EventListener not ready: %s", err)
	}
	t.Log("EventListener is ready")

	// Load the example pull request event data
	eventBodyJSON, err := loadExamplePREventBytes()
	if err != nil {
		t.Fatalf("Couldn't load test data: %v", err)
	}

	// Event body & Expected ResourceTemplates after instantiation
	wantPr1 := v1alpha1.PipelineResource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PipelineResource",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr1",
			Namespace: namespace,
			Labels: map[string]string{
				resourceLabel: "my-eventlistener",
				triggerLabel:  el.Spec.Triggers[0].Name,
				"edited":      "edited",
			},
		},
		Spec: v1alpha1.PipelineResourceSpec{
			Type: "git",
		},
	}
	wantPr2 := v1alpha1.PipelineResource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PipelineResource",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr2",
			Namespace: namespace,
			Labels: map[string]string{
				resourceLabel: "my-eventlistener",
				triggerLabel:  el.Spec.Triggers[0].Name,
				"open":        "defaultvalue",
			},
		},
		Spec: v1alpha1.PipelineResourceSpec{
			Type: "git",
			Params: []v1alpha1.ResourceParam{
				{Name: "license", Value: `{"key":"apache-2.0","name":"Apache License 2.0","spdx_id":"Apache-2.0","url":"https://api.github.com/licenses/apache-2.0","node_id":"MDc6TGljZW5zZTI="}`},
				{Name: "header", Value: `{"Accept-Encoding":"gzip","Content-Length":"2154","Content-Type":"application/json","User-Agent":"Go-http-client/1.1"}`},
				{Name: "prmessage", Value: "Git admission control\r\n\r\nNow with new lines!\r\n\r\n# :sunglasses: \r\n\r\naw yis"},
			},
		},
	}

	labelSelector := fields.SelectorFromSet(eventReconciler.GenerateResourceLabels(el.Name)).String()
	// Grab EventListener sink pods
	sinkPods, err := c.KubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		t.Fatalf("Error listing EventListener sink pods: %s", err)
	}

	// ElPort forward sink pod for http request
	portString := strconv.Itoa(*eventReconciler.ElPort)
	podName := sinkPods.Items[0].Name
	cmd := exec.Command("kubectl", "port-forward", podName, "-n", namespace, fmt.Sprintf("%s:%s", portString, portString))
	err = cmd.Start()
	if err != nil {
		t.Fatalf("Error starting port-forward command: %s", err)
	}
	if cmd.Process == nil {
		t.Fatalf("Error starting command. Process is nil")
	}
	defer func() {
		if err = cmd.Process.Kill(); err != nil {
			t.Fatalf("Error killing port-forward process: %s", err)
		}
	}()
	// Wait for port forward to take effect
	time.Sleep(5 * time.Second)

	// Send POST request to EventListener sink
	req, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:%s", portString), bytes.NewBuffer(eventBodyJSON))
	if err != nil {
		t.Fatalf("Error creating POST request: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error sending POST request: %s", err)
	}

	if resp.StatusCode > http.StatusAccepted {
		t.Errorf("sink did not return 2xx response. Got status code: %d", resp.StatusCode)
	}
	wantBody := sink.Response{
		EventListener: "my-eventlistener",
		Namespace:     namespace,
	}
	var gotBody sink.Response
	if err := json.NewDecoder(resp.Body).Decode(&gotBody); err != nil {
		t.Fatalf("failed to read/decode sink response: %v", err)
	}
	if diff := cmp.Diff(wantBody, gotBody, cmpopts.IgnoreFields(sink.Response{}, "EventID")); diff != "" {
		t.Errorf("unexpected sink response -want/+got: %s", diff)
	}
	if gotBody.EventID == "" {
		t.Errorf("sink response no eventID")
	}

	for _, wantPr := range []v1alpha1.PipelineResource{wantPr1, wantPr2} {
		if err = WaitFor(pipelineResourceExist(t, c, namespace, wantPr.Name)); err != nil {
			t.Fatalf("Failed to create ResourceTemplate %s: %s", wantPr.Name, err)
		}
		gotPr, err := c.PipelineClient.TektonV1alpha1().PipelineResources(namespace).Get(wantPr.Name, metav1.GetOptions{})
		if err != nil {
			t.Errorf("Error getting ResourceTemplate: %s: %s", wantPr.Name, err)
		}
		if gotPr.Labels[eventIDLabel] == "" {
			t.Errorf("Instantiated ResourceTemplate missing EventId")
		} else {
			delete(gotPr.Labels, eventIDLabel)
		}
		if diff := cmp.Diff(wantPr.Labels, gotPr.Labels); diff != "" {
			t.Errorf("Diff instantiated ResourceTemplate labels %s: -want +got: %s", wantPr.Name, diff)
		}
		if diff := cmp.Diff(wantPr.Spec, gotPr.Spec, cmp.Comparer(compareParamsWithLicenseJSON)); diff != "" {
			t.Errorf("Diff instantiated ResourceTemplate spec %s: -want +got: %s", wantPr.Name, diff)
		}
	}

	// Delete EventListener
	err = c.TriggersClient.TektonV1alpha1().EventListeners(namespace).Delete(el.Name, &metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Failed to delete EventListener: %s", err)
	}
	t.Log("Deleted EventListener")

	// Verify the EventListener's Deployment is deleted
	if err = WaitFor(deploymentNotExist(t, c, namespace, fmt.Sprintf("%s-%s", eventReconciler.GeneratedResourcePrefix, el.Name))); err != nil {
		t.Fatalf("Failed to delete EventListener Deployment: %s", err)
	}
	t.Log("EventListener's Deployment was deleted")

	// Verify the EventListener's Service is deleted
	if err = WaitFor(serviceNotExist(t, c, namespace, fmt.Sprintf("%s-%s", eventReconciler.GeneratedResourcePrefix, el.Name))); err != nil {
		t.Fatalf("Failed to delete EventListener Service: %s", err)
	}
	t.Log("EventListener's Service was deleted")
}

// The structure of this field corresponds to values for the `license` key in
// testdata/pr.json, and can be used to unmarshal the dat.
type license struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	SpdxID string `json:"spdx_id"`
	URL    string `json:"url"`
	NodeID string `json:"node_id"`
}

// compareParamsWithLicenseJSON will compare the passed in ResourceParams by further checking
// when the values aren't equal if they can be unmarshalled into the license object and if they are
// then equal. This is because the order of values in a dictionary is not deterministic and dictionary
// values passed through an event listener may change order.
func compareParamsWithLicenseJSON(x, y v1alpha1.ResourceParam) bool {
	xData := license{}
	yData := license{}
	if x.Name == y.Name {
		if x.Value != y.Value {
			// In order to compare these values, we are first unmarshalling them into the expected
			// structures because differences in the dictionary order of keys can cause
			// a string comparison to fail.
			if err := json.Unmarshal([]byte(x.Value), &xData); err != nil {
				return false
			}
			if err := json.Unmarshal([]byte(y.Value), &yData); err != nil {
				return false
			}
			if diff := cmp.Diff(xData, yData); diff != "" {
				return false
			}
		}
		return true
	}
	return false
}
