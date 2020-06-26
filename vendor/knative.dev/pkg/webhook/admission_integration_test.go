/*
Copyright 2020 The Knative Authors

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
package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/sync/errgroup"
	jsonpatch "gomodules.xyz/jsonpatch/v2"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/metrics/metricstest"
	_ "knative.dev/pkg/metrics/testing"
)

type fixedAdmissionController struct {
	path     string
	response *admissionv1.AdmissionResponse
}

var _ AdmissionController = (*fixedAdmissionController)(nil)

func (fac *fixedAdmissionController) Path() string {
	return fac.path
}

func (fac *fixedAdmissionController) Admit(ctx context.Context, req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	return fac.response
}

func TestAdmissionEmptyRequestBody(t *testing.T) {
	c := &fixedAdmissionController{
		path:     "/bazinga",
		response: &admissionv1.AdmissionResponse{},
	}

	testEmptyRequestBody(t, c)
}

func TestAdmissionValidResponseForResource(t *testing.T) {
	ac := &fixedAdmissionController{
		path:     "/bazinga",
		response: &admissionv1.AdmissionResponse{},
	}
	wh, serverURL, ctx, cancel, err := testSetup(t, ac)
	if err != nil {
		t.Fatalf("testSetup() = %v", err)
	}

	eg, _ := errgroup.WithContext(ctx)
	eg.Go(func() error { return wh.Run(ctx.Done()) })
	defer func() {
		cancel()
		if err := eg.Wait(); err != nil {
			t.Errorf("Unable to run controller: %s", err)
		}
	}()

	pollErr := waitForServerAvailable(t, serverURL, testTimeout)
	if pollErr != nil {
		t.Fatalf("waitForServerAvailable() = %v", err)
	}
	tlsClient, err := createSecureTLSClient(t, wh.Client, &wh.Options)
	if err != nil {
		t.Fatalf("createSecureTLSClient() = %v", err)
	}

	admissionreq := &admissionv1.AdmissionRequest{
		Operation: admissionv1.Create,
		Kind: metav1.GroupVersionKind{
			Group:   "pkg.knative.dev",
			Version: "v1alpha1",
			Kind:    "Resource",
		},
	}
	testRev := createResource("testrev")
	marshaled, err := json.Marshal(testRev)
	if err != nil {
		t.Fatalf("Failed to marshal resource: %s", err)
	}

	admissionreq.Resource.Group = "pkg.knative.dev"
	admissionreq.Object.Raw = marshaled
	rev := &admissionv1.AdmissionReview{
		Request: admissionreq,
	}

	reqBuf := new(bytes.Buffer)
	err = json.NewEncoder(reqBuf).Encode(&rev)
	if err != nil {
		t.Fatalf("Failed to marshal admission review: %v", err)
	}

	u, err := url.Parse(fmt.Sprintf("https://%s", serverURL))
	if err != nil {
		t.Fatalf("bad url %v", err)
	}

	u.Path = path.Join(u.Path, ac.Path())

	req, err := http.NewRequest("GET", u.String(), reqBuf)
	if err != nil {
		t.Fatalf("http.NewRequest() = %v", err)
	}
	req.Header.Add("Content-Type", "application/json")

	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		response, err := tlsClient.Do(req)
		if err != nil {
			t.Errorf("Failed to get response %v", err)
			return
		}

		if got, want := response.StatusCode, http.StatusOK; got != want {
			t.Errorf("Response status code = %v, wanted %v", got, want)
			return
		}

		defer response.Body.Close()
		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			t.Errorf("Failed to read response body %v", err)
			return
		}

		reviewResponse := admissionv1.AdmissionReview{}

		err = json.NewDecoder(bytes.NewReader(responseBody)).Decode(&reviewResponse)
		if err != nil {
			t.Errorf("Failed to decode response: %v", err)
			return
		}

		if diff := cmp.Diff(rev.TypeMeta, reviewResponse.TypeMeta); diff != "" {
			t.Errorf("expected the response typeMeta to be the same as the request (-want, +got)\n%s", diff)
			return
		}
	}()

	// Check that Admit calls block when they are initiated before informers sync.
	select {
	case <-time.After(5 * time.Second):
	case <-doneCh:
		t.Fatal("Admit was called before informers had synced.")
	}

	// Signal the webhook that informers have synced.
	wh.InformersHaveSynced()

	// Check that after informers have synced that things start completing immediately (including outstanding requests).
	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		t.Error("Timed out waiting on Admit to complete after informers synced.")
	}

	metricstest.CheckStatsReported(t, requestCountName, requestLatenciesName)
}

func TestAdmissionInvalidResponseForResource(t *testing.T) {
	expectedError := "everything is fine."
	ac := &fixedAdmissionController{
		path:     "/booger",
		response: MakeErrorStatus(expectedError),
	}
	wh, serverURL, ctx, cancel, err := testSetup(t, ac)
	if err != nil {
		t.Fatalf("testSetup() = %v", err)
	}

	eg, _ := errgroup.WithContext(ctx)
	eg.Go(func() error { return wh.Run(ctx.Done()) })
	wh.InformersHaveSynced()
	defer func() {
		cancel()
		if err := eg.Wait(); err != nil {
			t.Errorf("Unable to run controller: %s", err)
		}
	}()

	pollErr := waitForServerAvailable(t, serverURL, testTimeout)
	if pollErr != nil {
		t.Fatalf("waitForServerAvailable() = %v", err)
	}
	tlsClient, err := createSecureTLSClient(t, wh.Client, &wh.Options)
	if err != nil {
		t.Fatalf("createSecureTLSClient() = %v", err)
	}

	resource := createResource(testResourceName)

	resource.Spec.FieldWithValidation = "not the right value"
	marshaled, err := json.Marshal(resource)
	if err != nil {
		t.Fatalf("Failed to marshal resource: %s", err)
	}

	admissionreq := &admissionv1.AdmissionRequest{
		Operation: admissionv1.Create,
		Kind: metav1.GroupVersionKind{
			Group:   "pkg.knative.dev",
			Version: "v1alpha1",
			Kind:    "Resource",
		},
		UserInfo: authenticationv1.UserInfo{
			Username: user1,
		},
	}

	admissionreq.Resource.Group = "pkg.knative.dev"
	admissionreq.Object.Raw = marshaled

	rev := &admissionv1.AdmissionReview{
		Request: admissionreq,
	}
	reqBuf := new(bytes.Buffer)
	err = json.NewEncoder(reqBuf).Encode(&rev)
	if err != nil {
		t.Fatalf("Failed to marshal admission review: %v", err)
	}

	u, err := url.Parse(fmt.Sprintf("https://%s", serverURL))
	if err != nil {
		t.Fatalf("bad url %v", err)
	}

	u.Path = path.Join(u.Path, ac.Path())

	req, err := http.NewRequest("GET", u.String(), reqBuf)
	if err != nil {
		t.Fatalf("http.NewRequest() = %v", err)
	}

	req.Header.Add("Content-Type", "application/json")

	response, err := tlsClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to receive response %v", err)
	}

	if got, want := response.StatusCode, http.StatusOK; got != want {
		t.Errorf("Response status code = %v, wanted %v", got, want)
	}

	defer response.Body.Close()
	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Failed to read response body %v", err)
	}

	reviewResponse := admissionv1.AdmissionReview{}

	err = json.NewDecoder(bytes.NewReader(respBody)).Decode(&reviewResponse)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	var respPatch []jsonpatch.JsonPatchOperation
	err = json.Unmarshal(reviewResponse.Response.Patch, &respPatch)
	if err == nil {
		t.Fatalf("Expected to fail JSON unmarshal of resposnse")
	}

	if got, want := reviewResponse.Response.Result.Status, "Failure"; got != want {
		t.Errorf("Response status = %v, wanted %v", got, want)
	}

	if !strings.Contains(reviewResponse.Response.Result.Message, expectedError) {
		t.Errorf("Received unexpected response status message %s", reviewResponse.Response.Result.Message)
	}

	// Stats should be reported for requests that have admission disallowed
	metricstest.CheckStatsReported(t, requestCountName, requestLatenciesName)
}
