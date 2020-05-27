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
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"

	fakekubeclientset "k8s.io/client-go/kubernetes/fake"

	"go.uber.org/zap"
	discoveryclient "k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/gorilla/mux"
	pipelinev1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	dynamicclientset "github.com/tektoncd/triggers/pkg/client/dynamic/clientset"
	"github.com/tektoncd/triggers/pkg/client/dynamic/clientset/tekton"
	"github.com/tektoncd/triggers/pkg/template"
	"github.com/tektoncd/triggers/test"
	bldr "github.com/tektoncd/triggers/test/builder"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	ktesting "k8s.io/client-go/testing"
	rtesting "knative.dev/pkg/reconciler/testing"
)

const (
	resourceLabel = triggersv1.GroupName + triggersv1.EventListenerLabelKey
	triggerLabel  = triggersv1.GroupName + triggersv1.TriggerLabelKey
	eventIDLabel  = triggersv1.GroupName + triggersv1.EventIDLabelKey

	eventID   = "12345"
	namespace = "foo"
)

func init() {
	// Override UID generator for consistent test results.
	template.UID = func() string { return eventID }
}

// Compare two PipelineResources for sorting purposes
func comparePR(x, y pipelinev1alpha1.PipelineResource) bool {
	return x.GetName() < y.GetName()
}

// getSinkAssets seeds test resources and returns a testable Sink and a dynamic
// client. The returned client is used to creating the fake resources and can be
// used to check if the correct resources were created.
func getSinkAssets(t *testing.T, resources test.Resources, elName string, auth AuthOverride) (Sink, *fakedynamic.FakeDynamicClient) {
	t.Helper()
	ctx, _ := rtesting.SetupFakeContext(t)
	clients := test.SeedResources(t, ctx, resources)
	test.AddTektonResources(clients.Kube)

	logger, _ := zap.NewProduction()

	dynamicClient := fakedynamic.NewSimpleDynamicClient(runtime.NewScheme())
	dynamicSet := dynamicclientset.New(tekton.WithClient(dynamicClient))

	r := Sink{
		EventListenerName:      elName,
		EventListenerNamespace: namespace,
		DynamicClient:          dynamicSet,
		DiscoveryClient:        clients.Kube.Discovery(),
		KubeClientSet:          clients.Kube,
		TriggersClient:         clients.Triggers,
		Logger:                 logger.Sugar(),
		Auth:                   auth,
	}
	return r, dynamicClient
}

// getCreatedPipelineResources returns the pipeline resources that were created from the given actions
func getCreatedPipelineResources(t *testing.T, actions []ktesting.Action) []pipelinev1alpha1.PipelineResource {
	t.Helper()
	prs := []pipelinev1alpha1.PipelineResource{}
	for i := range actions {
		obj := actions[i].(ktesting.CreateAction).GetObject()
		// Since we use dynamic client, we cannot directly get the concrete type
		uns := obj.(*unstructured.Unstructured).Object
		pr := pipelinev1alpha1.PipelineResource{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(uns, &pr); err != nil {
			t.Errorf("failed to get created pipeline resource: %v", err)
		}
		prs = append(prs, pr)
	}
	return prs
}

// checkSinkResponse checks that the sink response status code is 201
// and that the body returns the EventListener, namespace, and eventID.
func checkSinkResponse(t *testing.T, resp *http.Response, elName string) {
	t.Helper()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected response code 201 but got: %v", resp.Status)
	}
	var gotBody Response
	if err := json.NewDecoder(resp.Body).Decode(&gotBody); err != nil {
		t.Fatalf("Error reading response body: %s", err)
	}
	wantBody := Response{
		EventListener: elName,
		Namespace:     namespace,
		EventID:       eventID,
	}
	if diff := cmp.Diff(wantBody, gotBody); diff != "" {
		t.Errorf("did not get expected response back -want,+got: %s", diff)
	}
}

func TestHandleEvent(t *testing.T) {
	eventBody := json.RawMessage(`{"head_commit": {"id": "testrevision"}, "repository": {"url": "testurl"}, "foo": "bar\t\r\nbaz昨"}`)
	numTriggers := 10

	pipelineResource := pipelinev1alpha1.PipelineResource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1alpha1",
			Kind:       "PipelineResource",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "$(params.name)",
			Namespace: namespace,
			Labels: map[string]string{
				"app":  "$(params.foo)",
				"type": "$(params.type)",
			},
		},
		Spec: pipelinev1alpha1.PipelineResourceSpec{
			Type: pipelinev1alpha1.PipelineResourceTypeGit,
			Params: []pipelinev1alpha1.ResourceParam{
				{Name: "url", Value: "$(params.url)"},
				{Name: "revision", Value: "$(params.revision)"},
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
			bldr.TriggerTemplateParam("foo", "", ""),
			bldr.TriggerTemplateParam("type", "", ""),
			bldr.TriggerResourceTemplate(runtime.RawExtension{Raw: pipelineResourceBytes}),
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
				bldr.TriggerBindingParam("foo", "$(body.foo)"),
				bldr.TriggerBindingParam("type", "$(header.Content-Type)"),
			))
		tbs = append(tbs, tb)
		// Add TriggerBinding to trigger in EventListener
		trigger := bldr.EventListenerTrigger("my-triggertemplate", "v1alpha1",
			bldr.EventListenerTriggerBinding(tbName, "", tbName, "v1alpha1"),
		)
		triggers = append(triggers, trigger)
	}
	el := bldr.EventListener("my-eventlistener", namespace, bldr.EventListenerSpec(triggers...))

	resources := test.Resources{
		TriggerBindings:  tbs,
		TriggerTemplates: []*triggersv1.TriggerTemplate{tt},
		EventListeners:   []*triggersv1.EventListener{el},
	}

	sink, dynamicClient := getSinkAssets(t, resources, el.Name, DefaultAuthOverride{})

	ts := httptest.NewServer(http.HandlerFunc(sink.HandleEvent))
	defer ts.Close()

	resp, err := http.Post(ts.URL, "application/json", bytes.NewReader(eventBody))
	if err != nil {
		t.Fatalf("Error creating Post request: %s", err)
	}

	checkSinkResponse(t, resp, el.Name)
	// Check right resources were created.
	var wantPrs []pipelinev1alpha1.PipelineResource
	for i := 0; i < numTriggers; i++ {
		wantResource := pipelinev1alpha1.PipelineResource{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "tekton.dev/v1alpha1",
				Kind:       "PipelineResource",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("my-pipelineresource-%d", i),
				Namespace: namespace,
				Labels: map[string]string{
					"app":         "bar\t\r\nbaz昨",
					"type":        "application/json",
					resourceLabel: "my-eventlistener",
					triggerLabel:  el.Spec.Triggers[0].Name,
					eventIDLabel:  eventID,
				},
			},
			Spec: pipelinev1alpha1.PipelineResourceSpec{
				Type: pipelinev1alpha1.PipelineResourceTypeGit,
				Params: []pipelinev1alpha1.ResourceParam{
					{Name: "url", Value: "testurl"},
					{Name: "revision", Value: "testrevision"},
				},
			},
		}
		wantPrs = append(wantPrs, wantResource)
	}
	// Sort actions (we do not know what order they executed in)
	gotPrs := getCreatedPipelineResources(t, dynamicClient.Actions())
	if diff := cmp.Diff(wantPrs, gotPrs, cmpopts.SortSlices(comparePR)); diff != "" {
		t.Errorf("Created resources mismatch (-want + got): %s", diff)
	}
}

func TestHandleEventWithInterceptors(t *testing.T) {
	eventBody := json.RawMessage(`{"head_commit": {"id": "testrevision"}, "repository": {"url": "testurl"}, "foo": "bar\t\r\nbaz昨"}`)

	pipelineResource := pipelinev1alpha1.PipelineResource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1alpha1",
			Kind:       "PipelineResource",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pipelineresource",
			Namespace: namespace,
		},
		Spec: pipelinev1alpha1.PipelineResourceSpec{
			Type: pipelinev1alpha1.PipelineResourceTypeGit,
			Params: []pipelinev1alpha1.ResourceParam{{
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
			bldr.TriggerResourceTemplate(runtime.RawExtension{Raw: pipelineResourceBytes}),
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
				Bindings: []*triggersv1.EventListenerBinding{{Name: "tb", Kind: "TriggerBinding"}},
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
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret",
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"secretKey": []byte("secret"),
		},
	}

	resources := test.Resources{
		TriggerBindings:  []*triggersv1.TriggerBinding{tb},
		TriggerTemplates: []*triggersv1.TriggerTemplate{tt},
		EventListeners:   []*triggersv1.EventListener{el},
		Secrets:          []*corev1.Secret{secret},
	}

	sink, dynamicClient := getSinkAssets(t, resources, el.Name, DefaultAuthOverride{})
	ts := httptest.NewServer(http.HandlerFunc(sink.HandleEvent))
	defer ts.Close()

	req, err := http.NewRequest("POST", ts.URL, bytes.NewReader(eventBody))
	if err != nil {
		t.Fatalf("Error creating Post request: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Github-Event", "pull_request")
	// This was generated by using SHA1 and hmac from go stdlib on secret and payload.
	// https://play.golang.org/p/8D2E-Yz3zWf for a sample.
	// TODO(dibyom): Add helper method that does this instead of link above
	req.Header.Add("X-Hub-Signature", "sha1=c0f3a2bbd1cdb062ba4f54b2a1cad3d171b7a129")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error sending Post request: %v", err)
	}
	checkSinkResponse(t, resp, el.Name)

	wantResource := []pipelinev1alpha1.PipelineResource{{
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
		Spec: pipelinev1alpha1.PipelineResourceSpec{
			Type: pipelinev1alpha1.PipelineResourceTypeGit,
			Params: []pipelinev1alpha1.ResourceParam{
				{Name: "url", Value: "testurl"},
			},
		},
	}}
	gotPrs := getCreatedPipelineResources(t, dynamicClient.Actions())
	if diff := cmp.Diff(wantResource, gotPrs); diff != "" {
		t.Errorf("Created resources mismatch (-want + got): %s", diff)
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
	eventBody := json.RawMessage(`{}`)
	numTriggers := 10

	resourceTemplate := pipelinev1alpha1.PipelineResource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1alpha1",
			Kind:       "PipelineResource",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "$(params.name)",
			Namespace: namespace,
		},
		Spec: pipelinev1alpha1.PipelineResourceSpec{
			Type: pipelinev1alpha1.PipelineResourceTypeGit,
		},
	}
	resourceTemplateBytes, err := json.Marshal(resourceTemplate)
	if err != nil {
		t.Fatalf("Error unmarshalling pipelineResource: %s", err)
	}

	tt := bldr.TriggerTemplate("tt", namespace,
		bldr.TriggerTemplateSpec(
			bldr.TriggerTemplateParam("name", "", ""),
			bldr.TriggerResourceTemplate(runtime.RawExtension{Raw: resourceTemplateBytes}),
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
			Bindings: []*triggersv1.EventListenerBinding{{Name: "tb", Kind: "TriggerBinding"}},
			Template: triggersv1.EventListenerTemplate{Name: "tt"},
			Interceptors: []*triggersv1.EventInterceptor{{
				Webhook: &triggersv1.WebhookInterceptor{
					ObjectRef: interceptorObjectRef,
					Header: []v1beta1.Param{{
						Name: "Name",
						Value: v1beta1.ArrayOrString{
							Type:      v1beta1.ParamTypeString,
							StringVal: fmt.Sprintf("my-resource-%d", i),
						},
					}},
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

	resources := test.Resources{
		TriggerBindings:  []*triggersv1.TriggerBinding{tb},
		TriggerTemplates: []*triggersv1.TriggerTemplate{tt},
		EventListeners:   []*triggersv1.EventListener{el},
	}

	// Redirect all requests to the fake server.
	srv := httptest.NewServer(&nameInterceptor{})
	defer srv.Close()
	client := srv.Client()
	u, _ := url.Parse(srv.URL)
	client.Transport = &http.Transport{
		Proxy: http.ProxyURL(u),
	}

	sink, dynamicClient := getSinkAssets(t, resources, el.Name, DefaultAuthOverride{})
	sink.HTTPClient = srv.Client()

	ts := httptest.NewServer(http.HandlerFunc(sink.HandleEvent))
	defer ts.Close()

	resp, err := http.Post(ts.URL, "application/json", bytes.NewReader(eventBody))
	if err != nil {
		t.Fatalf("Error creating Post request: %s", err)
	}
	checkSinkResponse(t, resp, el.Name)

	var wantPRs []pipelinev1alpha1.PipelineResource
	for i := 0; i < numTriggers; i++ {
		wantResource := resourceTemplate.DeepCopy()
		wantResource.ObjectMeta.Name = fmt.Sprintf("my-resource-%d", i)
		wantResource.ObjectMeta.Labels = map[string]string{
			resourceLabel: "el",
			triggerLabel:  "",
			eventIDLabel:  eventID,
		}
		wantPRs = append(wantPRs, *wantResource)
	}
	gotPrs := getCreatedPipelineResources(t, dynamicClient.Actions())
	if diff := cmp.Diff(wantPRs, gotPrs, cmpopts.SortSlices(comparePR)); diff != "" {
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
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
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

	logger, _ := zap.NewProduction()

	r := Sink{
		HTTPClient: srv.Client(),
		Logger:     logger.Sugar(),
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

	for _, method := range []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	} {
		t.Run(method, func(t *testing.T) {
			req, err := http.NewRequest(method, "/", nil)
			if err != nil {
				t.Fatalf("http.NewRequest: %v", err)
			}
			resp, header, err := r.executeInterceptors(trigger, req, []byte(`{}`), logger.Sugar())
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
		})
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

	logger, _ := zap.NewProduction()
	s := Sink{
		HTTPClient: client,
		Logger:     logger.Sugar(),
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
	if resp, _, err := s.executeInterceptors(trigger, req, nil, logger.Sugar()); err == nil {
		t.Errorf("expected error, got: %+v, %v", string(resp), err)
	}

	if si.called {
		t.Error("expected sequential interceptor to not be called")
	}
}

const userWithPermissions = "user-with-permissions"
const userWithoutPermissions = "user-with-no-permissions"

func TestRetriveveAuthToken(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: userWithoutPermissions,
			UID:  types.UID(userWithoutPermissions),
			Annotations: map[string]string{
				corev1.ServiceAccountNameKey: userWithoutPermissions,
				corev1.ServiceAccountUIDKey:  userWithoutPermissions,
			},
		},
		Type: corev1.SecretTypeServiceAccountToken,
		Data: map[string][]byte{
			corev1.ServiceAccountTokenKey: []byte(userWithoutPermissions),
		},
	}
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      userWithoutPermissions,
			Namespace: userWithoutPermissions,
			UID:       types.UID(userWithoutPermissions),
		},
		Secrets: []corev1.ObjectReference{
			{
				Name:      userWithoutPermissions,
				Namespace: userWithoutPermissions,
			},
		},
	}

	kubeClient := fakekubeclientset.NewSimpleClientset()
	test.AddTektonResources(kubeClient)
	if err := kubeClient.CoreV1().Secrets(userWithoutPermissions).Delete(userWithoutPermissions, &metav1.DeleteOptions{}); err != nil && !kerrors.IsNotFound(err) {
		t.Fatalf("Error deleting secret %v: %s", secret, err.Error())
	}
	if _, err := kubeClient.CoreV1().Secrets(userWithoutPermissions).Create(secret); err != nil {
		t.Fatalf("Error creating secret %v: %s", secret, err.Error())
	}
	if err := kubeClient.CoreV1().ServiceAccounts(userWithoutPermissions).Delete(userWithoutPermissions, &metav1.DeleteOptions{}); err != nil && !kerrors.IsNotFound(err) {
		t.Fatalf("Error delete sa %v: %s", sa, err.Error())
	}
	if _, err := kubeClient.CoreV1().ServiceAccounts(userWithoutPermissions).Create(sa); err != nil {
		t.Fatalf("Error creating sa %v: %s", sa, err.Error())
	}

	r := Sink{
		KubeClientSet: kubeClient,
	}

	token, err := r.retrieveAuthToken(&corev1.ObjectReference{Name: userWithoutPermissions, Namespace: userWithoutPermissions}, nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if token != userWithoutPermissions {
		t.Fatalf("got token %s instead of %s", token, userWithoutPermissions)
	}
}

type fakeAuth struct {
}

var triggerAuthWG sync.WaitGroup

func (r fakeAuth) OverrideAuthentication(token string,
	log *zap.SugaredLogger,
	defaultDiscoverClient discoveryclient.ServerResourcesInterface,
	defaultDynamicClient dynamic.Interface) (discoveryClient discoveryclient.ServerResourcesInterface,
	dynamicClient dynamic.Interface,
	err error) {

	if token == userWithoutPermissions {
		dynamicClient := fakedynamic.NewSimpleDynamicClient(runtime.NewScheme())
		dynamicClient.PrependReactor("*", "*", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			defer triggerAuthWG.Done()
			return true, nil, kerrors.NewUnauthorized(token + " unauthorized")
		})
		dynamicSet := dynamicclientset.New(tekton.WithClient(dynamicClient))
		return defaultDiscoverClient, dynamicSet, nil
	}

	return defaultDiscoverClient, defaultDynamicClient, nil
}

func TestHandleEventWithInterceptorsAndTriggerAuth(t *testing.T) {
	for _, testCase := range []struct {
		userVal    string
		statusCode int
	}{
		{
			userVal:    userWithoutPermissions,
			statusCode: http.StatusUnauthorized,
		},
		{
			userVal:    userWithPermissions,
			statusCode: http.StatusCreated,
		},
	} {
		eventBody := json.RawMessage(`{"head_commit": {"id": "testrevision"}, "repository": {"url": "testurl"}, "foo": "bar\t\r\nbaz昨"}`)

		pipelineResource := pipelinev1alpha1.PipelineResource{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "tekton.dev/v1alpha1",
				Kind:       "PipelineResource",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pipelineresource",
				Namespace: namespace,
			},
			Spec: pipelinev1alpha1.PipelineResourceSpec{
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
				bldr.TriggerResourceTemplate(runtime.RawExtension{Raw: pipelineResourceBytes}),
			))
		tb := bldr.TriggerBinding("tb", namespace,
			bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("url", "$(body.repository.url)"),
			))

		authSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testCase.userVal,
				Namespace: testCase.userVal,
				UID:       types.UID(testCase.userVal),
				Annotations: map[string]string{
					corev1.ServiceAccountNameKey: testCase.userVal,
					corev1.ServiceAccountUIDKey:  testCase.userVal,
				},
			},
			Type: corev1.SecretTypeServiceAccountToken,
			Data: map[string][]byte{
				corev1.ServiceAccountTokenKey: []byte(testCase.userVal),
			},
		}
		authSA := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testCase.userVal,
				Namespace: testCase.userVal,
				UID:       types.UID(testCase.userVal),
			},
			Secrets: []corev1.ObjectReference{
				{
					Name:      testCase.userVal,
					Namespace: testCase.userVal,
				},
			},
		}

		el := &triggersv1.EventListener{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "el",
				Namespace: namespace,
			},
			Spec: triggersv1.EventListenerSpec{
				Triggers: []triggersv1.EventListenerTrigger{{
					ServiceAccount: &corev1.ObjectReference{
						Namespace: testCase.userVal,
						Name:      testCase.userVal,
					},
					Bindings: []*triggersv1.EventListenerBinding{{Name: "tb", Kind: "TriggerBinding"}},
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
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret",
				Namespace: namespace,
			},
			Data: map[string][]byte{
				"secretKey": []byte("secret"),
			},
		}

		resources := test.Resources{
			TriggerBindings:  []*triggersv1.TriggerBinding{tb},
			TriggerTemplates: []*triggersv1.TriggerTemplate{tt},
			EventListeners:   []*triggersv1.EventListener{el},
			Secrets:          []*corev1.Secret{secret, authSecret},
			ServiceAccounts:  []*corev1.ServiceAccount{authSA},
		}
		sink, dynamicClient := getSinkAssets(t, resources, el.Name, fakeAuth{})
		ts := httptest.NewServer(http.HandlerFunc(sink.HandleEvent))
		defer ts.Close()

		triggerAuthWG.Add(1)

		dynamicClient.PrependReactor("*", "*", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			defer triggerAuthWG.Done()
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

		if resp.StatusCode != testCase.statusCode {
			t.Fatalf("Response code doesn't match: expected status code %d vs. actual %d, entire statutes %v",
				testCase.statusCode,
				resp.StatusCode,
				resp.Status)
		}
	}

}
