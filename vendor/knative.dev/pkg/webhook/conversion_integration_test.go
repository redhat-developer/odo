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
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/sync/errgroup"
	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type fixedConversionController struct {
	path     string
	response *apixv1.ConversionResponse
}

var _ ConversionController = (*fixedConversionController)(nil)

func (fcc *fixedConversionController) Path() string {
	return fcc.path
}

func (fcc *fixedConversionController) Convert(context.Context, *apixv1.ConversionRequest) *apixv1.ConversionResponse {
	return fcc.response
}

func TestConversionEmptyRequestBody(t *testing.T) {
	c := &fixedConversionController{
		path:     "/bazinga",
		response: &apixv1.ConversionResponse{},
	}

	testEmptyRequestBody(t, c)
}

func TestConversionValidResponse(t *testing.T) {
	cc := &fixedConversionController{
		path: "/bazinga",
		response: &apixv1.ConversionResponse{
			UID: types.UID("some-uid"),
		},
	}
	wh, serverURL, ctx, cancel, err := testSetup(t, cc)
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

	review := apixv1.ConversionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "ConversionReview",
		},
		Request: &apixv1.ConversionRequest{
			UID:               types.UID("some-uid"),
			DesiredAPIVersion: "example.com/v1",
			Objects:           []runtime.RawExtension{},
		},
	}

	reqBuf := new(bytes.Buffer)
	err = json.NewEncoder(reqBuf).Encode(&review)
	if err != nil {
		t.Fatalf("Failed to marshal conversion review: %v", err)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s%s", serverURL, cc.Path()), reqBuf)
	if err != nil {
		t.Fatalf("http.NewRequest() = %v", err)
	}
	req.Header.Add("Content-Type", "application/json")

	response, err := tlsClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to get response %v", err)
	}

	defer response.Body.Close()

	if got, want := response.StatusCode, http.StatusOK; got != want {
		t.Errorf("Response status code = %v, wanted %v", got, want)
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Failed to read response body %v", err)
	}

	reviewResponse := apixv1.ConversionReview{}

	err = json.NewDecoder(bytes.NewReader(responseBody)).Decode(&reviewResponse)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if reviewResponse.Response.UID != "some-uid" {
		t.Errorf("expected the response uid to be the stubbed version")
	}

	if diff := cmp.Diff(review.TypeMeta, reviewResponse.TypeMeta); diff != "" {
		t.Errorf("expected the response typeMeta to be the same as the request (-want, +got)\n%s", diff)
	}
}

func TestConversionInvalidResponse(t *testing.T) {
	cc := &fixedConversionController{
		path: "/bazinga",
		response: &apixv1.ConversionResponse{
			UID: types.UID("some-uid"),
			Result: metav1.Status{
				Status: metav1.StatusFailure,
			},
		},
	}
	wh, serverURL, ctx, cancel, err := testSetup(t, cc)
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

	review := apixv1.ConversionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "ConversionReview",
		},
		Request: &apixv1.ConversionRequest{
			UID:               types.UID("some-uid"),
			DesiredAPIVersion: "example.com/v1",
			Objects:           []runtime.RawExtension{},
		},
	}

	reqBuf := new(bytes.Buffer)
	err = json.NewEncoder(reqBuf).Encode(&review)
	if err != nil {
		t.Fatalf("Failed to marshal conversion review: %v", err)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s%s", serverURL, cc.Path()), reqBuf)
	if err != nil {
		t.Fatalf("http.NewRequest() = %v", err)
	}
	req.Header.Add("Content-Type", "application/json")

	response, err := tlsClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to get response %v", err)
	}

	defer response.Body.Close()

	if got, want := response.StatusCode, http.StatusOK; got != want {
		t.Errorf("Response status code = %v, wanted %v", got, want)
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Failed to read response body %v", err)
	}

	reviewResponse := apixv1.ConversionReview{}

	err = json.NewDecoder(bytes.NewReader(responseBody)).Decode(&reviewResponse)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if reviewResponse.Response.UID != "some-uid" {
		t.Errorf("expected the response uid to be the stubbed version")
	}

	if reviewResponse.Response.Result.Status != metav1.StatusFailure {
		t.Errorf("expected the response uid to be the stubbed version")
	}
}
