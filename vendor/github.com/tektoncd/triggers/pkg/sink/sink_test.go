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

package sink

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/logging"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	faketriggersclientset "github.com/tektoncd/triggers/pkg/client/clientset/versioned/fake"
	dynamicclientset "github.com/tektoncd/triggers/pkg/client/dynamic/clientset"
	"github.com/tektoncd/triggers/pkg/client/dynamic/clientset/tekton"
	"github.com/tektoncd/triggers/pkg/template"
	"github.com/tektoncd/triggers/test"
	bldr "github.com/tektoncd/triggers/test/builder"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

const (
	resourceLabel = triggersv1.GroupName + triggersv1.EventListenerLabelKey
	triggerLabel  = triggersv1.GroupName + triggersv1.TriggerLabelKey
	eventIDLabel  = triggersv1.GroupName + triggersv1.EventIDLabelKey

	eventID = "12345"
)

func init() {
	// Override UID generator for consistent test results.
	template.UID = func() string { return eventID }
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}

func TestHandleEvent(t *testing.T) {
	namespace := "foo"
	eventBody := json.RawMessage(`{"head_commit": {"id": "testrevision"}, "repository": {"url": "testurl"}, "foo": "bar\t\r\nbaz昨"}`)
	numTriggers := 10

	pipelineResource := pipelinev1.PipelineResource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1alpha1",
			Kind:       "PipelineResource",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "$(params.name)",
			Namespace: namespace,
			Labels:    map[string]string{"app": "$(params.appLabel)"},
		},
		Spec: pipelinev1.PipelineResourceSpec{
			Type: pipelinev1.PipelineResourceTypeGit,
			Params: []pipelinev1.ResourceParam{
				{Name: "url", Value: "$(params.url)"},
				{Name: "revision", Value: "$(params.revision)"},
				{Name: "contenttype", Value: "$(params.contenttype)"},
				{Name: "foo", Value: "$(params.foo)"},
			},
		},
	}
	pipelineResourceBytes, err := json.Marshal(pipelineResource)
	if err != nil {
		t.Fatalf("Error unmarshalling pipelineResource: %s", err)
	}

	tt := bldr.TriggerTemplate("my-triggertemplate", namespace,
		bldr.TriggerTemplateSpec(
			bldr.TriggerTemplateParam("name", "", ""),
			bldr.TriggerTemplateParam("url", "", ""),
			bldr.TriggerTemplateParam("revision", "", ""),
			bldr.TriggerTemplateParam("appLabel", "", "foo"),
			bldr.TriggerTemplateParam("contenttype", "", ""),
			bldr.TriggerTemplateParam("foo", "", ""),
			bldr.TriggerResourceTemplate(json.RawMessage(pipelineResourceBytes)),
		))
	var tbs []*triggersv1.TriggerBinding
	var triggers []bldr.EventListenerSpecOp
	for i := 0; i < numTriggers; i++ {
		// Create TriggerBinding
		tbName := fmt.Sprintf("my-triggerbinding-%d", i)
		tb := bldr.TriggerBinding(tbName, namespace,
			bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("name", fmt.Sprintf("my-pipelineresource-%d", i)),
				bldr.TriggerBindingParam("url", "$(body.repository.url)"),
				bldr.TriggerBindingParam("revision", "$(body.head_commit.id)"),
				bldr.TriggerBindingParam("contenttype", "$(header.Content-Type)"),
				bldr.TriggerBindingParam("foo", "$(body.foo)"),
			))
		tbs = append(tbs, tb)
		// Add TriggerBinding to trigger in EventListener
		trigger := bldr.EventListenerTrigger(tbName, "my-triggertemplate", "v1alpha1")
		triggers = append(triggers, trigger)
	}
	el := bldr.EventListener("my-eventlistener", namespace, bldr.EventListenerSpec(triggers...))

	kubeClient := fakekubeclientset.NewSimpleClientset()
	test.AddTektonResources(kubeClient)

	triggersClient := faketriggersclientset.NewSimpleClientset()
	if _, err := triggersClient.TektonV1alpha1().TriggerTemplates(namespace).Create(tt); err != nil {
		t.Fatalf("Error creating TriggerTemplate: %s", err)
	}
	// for _, tb := range []*triggersv1.TriggerBinding{tb, tb2, tb3} {
	for _, tb := range tbs {
		if _, err := triggersClient.TektonV1alpha1().TriggerBindings(namespace).Create(tb); err != nil {
			t.Fatalf("Error creating TriggerBinding %s: %s", tb.GetName(), err)
		}
	}
	el, err = triggersClient.TektonV1alpha1().EventListeners(namespace).Create(el)
	if err != nil {
		t.Fatalf("Error creating EventListener: %s", err)
	}

	logger, _ := logging.NewLogger("", "")

	dynamicClient := fakedynamic.NewSimpleDynamicClient(runtime.NewScheme())
	dynamicSet := dynamicclientset.New(tekton.WithClient(dynamicClient))

	r := Sink{
		EventListenerName:      el.Name,
		EventListenerNamespace: namespace,
		DynamicClient:          dynamicSet,
		DiscoveryClient:        kubeClient.Discovery(),
		TriggersClient:         triggersClient,
		Logger:                 logger,
	}
	ts := httptest.NewServer(http.HandlerFunc(r.HandleEvent))
	defer ts.Close()

	var wg sync.WaitGroup
	wg.Add(numTriggers)

	dynamicClient.PrependReactor("*", "*", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		defer wg.Done()
		return false, nil, nil
	})

	resp, err := http.Post(ts.URL, "application/json", bytes.NewReader(eventBody))
	if err != nil {
		t.Fatalf("Error creating Post request: %s", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Response code doesn't match: %v", resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %s", err)
	}

	wantBody := Response{
		EventListener: el.Name,
		Namespace:     el.Namespace,
		EventID:       eventID,
	}

	got := Response{}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Error unmarshalling response body: %s", err)
	}
	if diff := cmp.Diff(wantBody, got); diff != "" {
		t.Errorf("did not get expected response back -want,+got: %s", diff)
	}

	// We expect that the EventListener will be able to immediately handle the event so we
	// can use a very short timeout
	if waitTimeout(&wg, time.Second) {
		t.Fatalf("timed out waiting for reactor to fire")
	}
	// var wantResources []pipelinev1.PipelineResource
	gvr := schema.GroupVersionResource{
		Group:    "tekton.dev",
		Version:  "v1alpha1",
		Resource: "pipelineresources",
	}
	var wantActions []ktesting.Action
	for i := 0; i < numTriggers; i++ {
		wantResource := pipelinev1.PipelineResource{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "tekton.dev/v1alpha1",
				Kind:       "PipelineResource",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("my-pipelineresource-%d", i),
				Namespace: namespace,
				Labels: map[string]string{
					"app":         "foo",
					resourceLabel: "my-eventlistener",
					triggerLabel:  el.Spec.Triggers[0].Name,
					eventIDLabel:  eventID,
				},
			},
			Spec: pipelinev1.PipelineResourceSpec{
				Type: pipelinev1.PipelineResourceTypeGit,
				Params: []pipelinev1.ResourceParam{
					{Name: "url", Value: "testurl"},
					{Name: "revision", Value: "testrevision"},
					{Name: "contenttype", Value: "application/json"},
					{Name: "foo", Value: "bar\t\r\nbaz昨"},
				},
			},
		}
		action := ktesting.NewCreateAction(gvr, "foo", test.ToUnstructured(t, wantResource))
		wantActions = append(wantActions, action)
	}
	// Sort actions (we do not know what order they executed in)
	gotActions := sortCreateActions(t, dynamicClient.Actions())
	if diff := cmp.Diff(wantActions, gotActions); diff != "" {
		t.Errorf("Actions mismatch (-want +got): %s", diff)
	}
}

func TestHandleEventWithInterceptors(t *testing.T) {
	namespace := "foo"
	eventBody := json.RawMessage(`{"head_commit": {"id": "testrevision"}, "repository": {"url": "testurl"}, "foo": "bar\t\r\nbaz昨"}`)

	pipelineResource := pipelinev1.PipelineResource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1alpha1",
			Kind:       "PipelineResource",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pipelineresource",
			Namespace: namespace,
		},
		Spec: pipelinev1.PipelineResourceSpec{
			Type: pipelinev1.PipelineResourceTypeGit,
			Params: []pipelinev1.ResourceParam{{
				Name:  "url",
				Value: "$(params.url)",
			}},
		},
	}
	pipelineResourceBytes, err := json.Marshal(pipelineResource)
	if err != nil {
		t.Fatalf("Error unmarshalling pipelineResource: %s", err)
	}

	tt := bldr.TriggerTemplate("tt", namespace,
		bldr.TriggerTemplateSpec(
			bldr.TriggerTemplateParam("url", "", ""),
			bldr.TriggerResourceTemplate(pipelineResourceBytes),
		))
	tb := bldr.TriggerBinding("tb", namespace,
		bldr.TriggerBindingSpec(
			bldr.TriggerBindingParam("url", "$(body.repository.url)"),
		))

	el := &triggersv1.EventListener{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "el",
			Namespace: namespace,
		},
		Spec: triggersv1.EventListenerSpec{
			Triggers: []triggersv1.EventListenerTrigger{{
				Bindings: []*triggersv1.EventListenerBinding{{Name: "tb"}},
				Template: triggersv1.EventListenerTemplate{Name: "tt"},
				Interceptors: []*triggersv1.EventInterceptor{{
					GitHub: &triggersv1.GitHubInterceptor{
						SecretRef: &triggersv1.SecretRef{
							SecretKey:  "secretKey",
							SecretName: "secret",
							Namespace:  namespace,
						},
						EventTypes: []string{"pull_request"},
					},
				}},
			}},
		},
	}

	kubeClient := fakekubeclientset.NewSimpleClientset()
	test.AddTektonResources(kubeClient)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "secret",
		},
		Data: map[string][]byte{
			"secretKey": []byte("secret"),
		},
	}
	if _, err := kubeClient.CoreV1().Secrets(namespace).Create(secret); err != nil {
		t.Fatalf("error creating secret: %v", secret)
	}
	triggersClient := faketriggersclientset.NewSimpleClientset()
	if _, err := triggersClient.TektonV1alpha1().TriggerTemplates(namespace).Create(tt); err != nil {
		t.Fatalf("Error creating TriggerTemplate: %s", err)
	}
	if _, err := triggersClient.TektonV1alpha1().TriggerBindings(namespace).Create(tb); err != nil {
		t.Fatalf("Error creating TriggerBinding: %s", err)
	}
	el, err = triggersClient.TektonV1alpha1().EventListeners(namespace).Create(el)
	if err != nil {
		t.Fatalf("Error creating EventListener: %s", err)
	}

	logger, _ := logging.NewLogger("", "")

	dynamicClient := fakedynamic.NewSimpleDynamicClient(runtime.NewScheme())
	dynamicSet := dynamicclientset.New(tekton.WithClient(dynamicClient))

	r := Sink{
		EventListenerName:      el.Name,
		EventListenerNamespace: namespace,
		DynamicClient:          dynamicSet,
		DiscoveryClient:        kubeClient.Discovery(),
		TriggersClient:         triggersClient,
		KubeClientSet:          kubeClient,
		Logger:                 logger,
	}
	ts := httptest.NewServer(http.HandlerFunc(r.HandleEvent))
	defer ts.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	dynamicClient.PrependReactor("*", "*", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		defer wg.Done()
		return false, nil, nil
	})

	req, err := http.NewRequest("POST", ts.URL, bytes.NewReader(eventBody))
	if err != nil {
		t.Fatalf("Error creating Post request: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Github-Event", "pull_request")
	// This was generated by using SHA1 and hmac from go stdlib on secret and payload.
	// https://play.golang.org/p/8D2E-Yz3zWf for a sample.
	req.Header.Add("X-Hub-Signature", "sha1=c0f3a2bbd1cdb062ba4f54b2a1cad3d171b7a129")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error sending Post request: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Response code doesn't match: %v", resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %s", err)
	}

	wantBody := Response{
		EventListener: el.Name,
		Namespace:     el.Namespace,
		EventID:       eventID,
	}

	got := Response{}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Error unmarshalling response body: %s", err)
	}
	if diff := cmp.Diff(wantBody, got); diff != "" {
		t.Errorf("did not get expected response back -want,+got: %s", diff)
	}

	// We expect that the EventListener will be able to immediately handle the event so we
	// can use a very short timeout
	if waitTimeout(&wg, time.Second) {
		t.Fatalf("timed out waiting for reactor to fire")
	}
	wantResource := pipelinev1.PipelineResource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1alpha1",
			Kind:       "PipelineResource",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pipelineresource",
			Namespace: namespace,
			Labels: map[string]string{
				resourceLabel: "el",
				triggerLabel:  el.Spec.Triggers[0].Name,
				eventIDLabel:  eventID,
			},
		},
		Spec: pipelinev1.PipelineResourceSpec{
			Type: pipelinev1.PipelineResourceTypeGit,
			Params: []pipelinev1.ResourceParam{
				{Name: "url", Value: "testurl"},
			},
		},
	}
	gvr := schema.GroupVersionResource{
		Group:    "tekton.dev",
		Version:  "v1alpha1",
		Resource: "pipelineresources",
	}
	want := []ktesting.Action{ktesting.NewCreateAction(gvr, "foo", test.ToUnstructured(t, wantResource))}
	if diff := cmp.Diff(want, dynamicClient.Actions()); diff != "" {
		t.Error(diff)
	}
}

// nameInterceptor is an HTTP server that reads a "Name" from the header, and
// writes the name in its body as {"name": "VALUE"}.
// It expects a request with the header "Name".
// The response body will always return with {"name": "VALUE"} where VALUE is
// the value of the first element in the header "Name".
type nameInterceptor struct{}

func (f *nameInterceptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Copy over all headers
	for k := range r.Header {
		for _, v := range r.Header[k] {
			w.Header().Add(k, v)
		}
	}
	// Read the Name header
	var name string
	if nameValue, ok := r.Header["Name"]; ok {
		name = nameValue[0]
	}
	// Write the name to the body
	body := fmt.Sprintf(`{"name": "%s"}`, name)
	_, _ = w.Write([]byte(body))
}

func TestHandleEventWithWebhookInterceptors(t *testing.T) {
	namespace := "foo"
	eventBody := json.RawMessage(`{}`)
	numTriggers := 10

	resourceTemplate := pipelinev1.PipelineResource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1alpha1",
			Kind:       "PipelineResource",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "$(params.name)",
			Namespace: namespace,
		},
		Spec: pipelinev1.PipelineResourceSpec{
			Type: pipelinev1.PipelineResourceTypeGit,
			Params: []pipelinev1.ResourceParam{{
				Name:  "url",
				Value: "testurl",
			}},
		},
	}
	resourceTemplateBytes, err := json.Marshal(resourceTemplate)
	if err != nil {
		t.Fatalf("Error unmarshalling pipelineResource: %s", err)
	}

	tt := bldr.TriggerTemplate("tt", namespace,
		bldr.TriggerTemplateSpec(
			bldr.TriggerTemplateParam("name", "", ""),
			bldr.TriggerResourceTemplate(resourceTemplateBytes),
		))
	tb := bldr.TriggerBinding("tb", namespace,
		bldr.TriggerBindingSpec(
			bldr.TriggerBindingParam("name", "$(body.name)"),
		))

	interceptorObjectRef := &corev1.ObjectReference{
		APIVersion: "v1",
		Kind:       "Service",
		Name:       "foo",
	}
	var triggers []triggersv1.EventListenerTrigger
	for i := 0; i < numTriggers; i++ {
		trigger := triggersv1.EventListenerTrigger{
			Bindings: []*triggersv1.EventListenerBinding{{Name: "tb"}},
			Template: triggersv1.EventListenerTemplate{Name: "tt"},
			Interceptors: []*triggersv1.EventInterceptor{{
				Webhook: &triggersv1.WebhookInterceptor{
					ObjectRef: interceptorObjectRef,
					Header:    []pipelinev1.Param{bldr.Param("Name", fmt.Sprintf("my-resource-%d", i))},
				},
			}},
		}
		triggers = append(triggers, trigger)
	}
	el := &triggersv1.EventListener{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "el",
			Namespace: namespace,
		},
		Spec: triggersv1.EventListenerSpec{
			Triggers: triggers,
		},
	}

	kubeClient := fakekubeclientset.NewSimpleClientset()
	test.AddTektonResources(kubeClient)

	triggersClient := faketriggersclientset.NewSimpleClientset()
	if _, err := triggersClient.TektonV1alpha1().TriggerTemplates(namespace).Create(tt); err != nil {
		t.Fatalf("Error creating TriggerTemplate: %s", err)
	}
	if _, err := triggersClient.TektonV1alpha1().TriggerBindings(namespace).Create(tb); err != nil {
		t.Fatalf("Error creating TriggerBinding: %s", err)
	}
	if _, err = triggersClient.TektonV1alpha1().EventListeners(namespace).Create(el); err != nil {
		t.Fatalf("Error creating EventListener: %s", err)
	}

	logger, _ := logging.NewLogger("", "")

	dynamicClient := fakedynamic.NewSimpleDynamicClient(runtime.NewScheme())
	dynamicSet := dynamicclientset.New(tekton.WithClient(dynamicClient))

	// Redirect all requests to the fake server.
	srv := httptest.NewServer(&nameInterceptor{})
	defer srv.Close()
	client := srv.Client()
	u, _ := url.Parse(srv.URL)
	client.Transport = &http.Transport{
		Proxy: http.ProxyURL(u),
	}

	r := Sink{
		HTTPClient:             srv.Client(),
		EventListenerName:      el.Name,
		EventListenerNamespace: namespace,
		DynamicClient:          dynamicSet,
		DiscoveryClient:        kubeClient.Discovery(),
		TriggersClient:         triggersClient,
		KubeClientSet:          kubeClient,
		Logger:                 logger,
	}
	ts := httptest.NewServer(http.HandlerFunc(r.HandleEvent))
	defer ts.Close()

	var wg sync.WaitGroup
	wg.Add(numTriggers)
	dynamicClient.PrependReactor("*", "*", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		defer wg.Done()
		return false, nil, nil
	})

	resp, err := http.Post(ts.URL, "application/json", bytes.NewReader(eventBody))
	if err != nil {
		t.Fatalf("Error creating Post request: %s", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Response code doesn't match: %v", resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %s", err)
	}

	wantBody := Response{
		EventListener: el.Name,
		Namespace:     el.Namespace,
		EventID:       eventID,
	}

	got := Response{}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Error unmarshalling response body: %s", err)
	}
	if diff := cmp.Diff(wantBody, got); diff != "" {
		t.Errorf("did not get expected response back -want,+got: %s", diff)
	}

	// We expect that the EventListener will be able to immediately handle the event so we
	// can use a very short timeout
	if waitTimeout(&wg, time.Second) {
		t.Fatalf("timed out waiting for reactor to fire")
	}
	gvr := schema.GroupVersionResource{
		Group:    "tekton.dev",
		Version:  "v1alpha1",
		Resource: "pipelineresources",
	}
	var wantActions []ktesting.Action
	for i := 0; i < numTriggers; i++ {
		wantResource := resourceTemplate.DeepCopy()
		wantResource.ObjectMeta.Name = fmt.Sprintf("my-resource-%d", i)
		wantResource.ObjectMeta.Labels = map[string]string{
			resourceLabel: "el",
			triggerLabel:  "",
			eventIDLabel:  eventID,
		}
		action := ktesting.NewCreateAction(gvr, "foo", test.ToUnstructured(t, wantResource))
		wantActions = append(wantActions, action)
	}
	// Sort actions (we do not know what order they executed in)
	gotActions := sortCreateActions(t, dynamicClient.Actions())
	if diff := cmp.Diff(wantActions, gotActions); diff != "" {
		t.Errorf("Actions mismatch (-want +got): %s", diff)
	}
}

// sequentialInterceptor is a HTTP server that will return sequential responses.
// It expects a request of the form `{"i": n}`.
// The response body will always return with the next value set, whereas the
// headers will append new values in the sequence.
type sequentialInterceptor struct {
	called bool
}

func (f *sequentialInterceptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.called = true
	var data map[string]int
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	defer r.Body.Close()
	data["i"]++

	// Copy over all old headers, then set new value.
	key := "Foo"
	for _, v := range r.Header[key] {
		w.Header().Add(key, v)
	}
	w.Header().Add(key, strconv.Itoa(int(data["i"])))
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
	}
}

// TestExecuteInterceptor tests that two interceptors can be called
// sequentially. It uses a HTTP server that returns a sequential response
// and two webhook interceptors pointing at the test server, validating
// that the last response is as expected.
func TestExecuteInterceptor(t *testing.T) {
	srv := httptest.NewServer(&sequentialInterceptor{})
	defer srv.Close()
	client := srv.Client()
	// Redirect all requests to the fake server.
	u, _ := url.Parse(srv.URL)
	client.Transport = &http.Transport{
		Proxy: http.ProxyURL(u),
	}

	logger, _ := logging.NewLogger("", "")
	r := Sink{
		HTTPClient: srv.Client(),
		Logger:     logger,
	}

	a := &triggersv1.EventInterceptor{
		Webhook: &triggersv1.WebhookInterceptor{
			ObjectRef: &corev1.ObjectReference{
				APIVersion: "v1",
				Kind:       "Service",
				Name:       "foo",
			},
		},
	}
	trigger := &triggersv1.EventListenerTrigger{
		Interceptors: []*triggersv1.EventInterceptor{a, a},
	}

	req, err := http.NewRequest(http.MethodPost, "/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	resp, header, err := r.executeInterceptors(trigger, req, []byte(`{}`), "", logger)
	if err != nil {
		t.Fatalf("executeInterceptors: %v", err)
	}

	var got map[string]int
	if err := json.Unmarshal(resp, &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	want := map[string]int{"i": 2}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Body: -want +got: %s", diff)
	}
	if diff := cmp.Diff([]string{"1", "2"}, header["Foo"]); diff != "" {
		t.Errorf("Header: -want +got: %s", diff)
	}
}

// errorInterceptor is a HTTP server that will always return an error response.
type errorInterceptor struct{}

func (e *errorInterceptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func TestExecuteInterceptor_error(t *testing.T) {
	// Route requests to either the error interceptor or sequential interceptor
	// based on the host.
	errHost := "error"
	match := func(r *http.Request, _ *mux.RouteMatch) bool {
		return strings.Contains(r.Host, errHost)
	}
	r := mux.NewRouter()
	r.MatcherFunc(match).Handler(&errorInterceptor{})
	si := &sequentialInterceptor{}
	r.Handle("/", si)

	srv := httptest.NewServer(r)
	defer srv.Close()
	client := srv.Client()
	u, _ := url.Parse(srv.URL)
	// Redirect all requests to the fake server.
	client.Transport = &http.Transport{
		Proxy: http.ProxyURL(u),
	}

	logger, _ := logging.NewLogger("", "")
	s := Sink{
		HTTPClient: client,
		Logger:     logger,
	}

	trigger := &triggersv1.EventListenerTrigger{
		Interceptors: []*triggersv1.EventInterceptor{
			// Error interceptor needs to come first.
			{
				Webhook: &triggersv1.WebhookInterceptor{
					ObjectRef: &corev1.ObjectReference{
						APIVersion: "v1",
						Kind:       "Service",
						Name:       errHost,
					},
				},
			},
			{
				Webhook: &triggersv1.WebhookInterceptor{
					ObjectRef: &corev1.ObjectReference{
						APIVersion: "v1",
						Kind:       "Service",
						Name:       "foo",
					},
				},
			},
		},
	}
	req, err := http.NewRequest(http.MethodPost, "/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	if resp, _, err := s.executeInterceptors(trigger, req, nil, "", logger); err == nil {
		t.Errorf("expected error, got: %+v, %v", string(resp), err)
	}

	if si.called {
		t.Error("expected sequential interceptor to not be called")
	}
}

// Sort CreateActions by the name of their resource.
// The Actions must be CreateActions, and they must have an Object that has a
// name.
func sortCreateActions(t *testing.T, actions []ktesting.Action) []ktesting.Action {
	sort.SliceStable(actions, func(i int, j int) bool {
		objectI := actions[i].(ktesting.CreateAction).GetObject()
		objectJ := actions[j].(ktesting.CreateAction).GetObject()
		unstructuredI, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(objectI)
		unstructuredJ, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(objectJ)
		nameI := (&unstructured.Unstructured{Object: unstructuredI}).GetName()
		nameJ := (&unstructured.Unstructured{Object: unstructuredJ}).GetName()
		if nameI == "" || nameJ == "" {
			t.Errorf("Error getting resource name from action; names are empty")
		}
		return nameI < nameJ
	})
	return actions
}
