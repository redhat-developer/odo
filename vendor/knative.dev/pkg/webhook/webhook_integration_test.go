/*
Copyright 2018 The Knative Authors

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
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	kubeclient "knative.dev/pkg/client/injection/kube/client/fake"
	_ "knative.dev/pkg/injection/clients/namespacedkube/informers/core/v1/secret/fake"
	"knative.dev/pkg/system"

	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/metrics/metricstest"
	pkgtest "knative.dev/pkg/testing"
)

// createResource creates a testing.Resource with the given name in the system namespace.
func createResource(name string) *pkgtest.Resource {
	return &pkgtest.Resource{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: system.Namespace(),
			Name:      name,
		},
		Spec: pkgtest.ResourceSpec{
			FieldWithValidation: "magic value",
		},
	}
}

const testTimeout = 10 * time.Second

func TestMissingContentType(t *testing.T) {
	wh, serverURL, ctx, cancel, err := testSetup(t)
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

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s", serverURL), nil)
	if err != nil {
		t.Fatalf("http.NewRequest() = %v", err)
	}

	response, err := tlsClient.Do(req)
	if err != nil {
		t.Fatalf("Received %v error from server %s", err, serverURL)
	}

	if got, want := response.StatusCode, http.StatusUnsupportedMediaType; got != want {
		t.Errorf("Response status code = %v, wanted %v", got, want)
	}

	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Failed to read response body %v", err)
	}

	if !strings.Contains(string(responseBody), "invalid Content-Type") {
		t.Errorf("Response body to contain 'invalid Content-Type' , got = '%s'", string(responseBody))
	}

	// Stats are not reported for internal server errors
	metricstest.CheckStatsNotReported(t, requestCountName, requestLatenciesName)
}

func testEmptyRequestBody(t *testing.T, controller interface{}) {
	wh, serverURL, ctx, cancel, err := testSetup(t, controller)
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

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/bazinga", serverURL), nil)
	if err != nil {
		t.Fatalf("http.NewRequest() = %v", err)
	}

	req.Header.Add("Content-Type", "application/json")

	response, err := tlsClient.Do(req)
	if err != nil {
		t.Fatalf("failed to get resp %v", err)
	}

	if got, want := response.StatusCode, http.StatusBadRequest; got != want {
		t.Errorf("Response status code = %v, wanted %v", got, want)
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Failed to read response body %v", err)
	}

	if !strings.Contains(string(responseBody), "could not decode body") {
		t.Errorf("Response body to contain 'decode failure information' , got = %q", string(responseBody))
	}
}

func TestSetupWebhookHTTPServerError(t *testing.T) {
	defaultOpts := newDefaultOptions()
	defaultOpts.Port = -1 // invalid port
	ctx, wh, cancel := newNonRunningTestWebhook(t, defaultOpts)
	defer cancel()
	kubeClient := kubeclient.Get(ctx)

	nsErr := createNamespace(t, kubeClient, metav1.NamespaceSystem)
	if nsErr != nil {
		t.Fatalf("createNamespace() = %v", nsErr)
	}
	cMapsErr := createTestConfigMap(t, kubeClient)
	if cMapsErr != nil {
		t.Fatalf("createTestConfigMap() = %v", cMapsErr)
	}

	stopCh := make(chan struct{})
	errCh := make(chan error)
	go func() {
		if err := wh.Run(stopCh); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-time.After(6 * time.Second):
		t.Error("Timeout in testing bootstrap webhook http server failed")
	case errItem := <-errCh:
		if !strings.Contains(errItem.Error(), "bootstrap failed") {
			t.Error("Expected bootstrap webhook http server failed")
		}
	}
}

func testSetup(t *testing.T, acs ...interface{}) (*Webhook, string, context.Context, context.CancelFunc, error) {
	t.Helper()
	port, err := newTestPort()
	if err != nil {
		return nil, "", nil, nil, err
	}

	defaultOpts := newDefaultOptions()
	defaultOpts.Port = port
	ctx, wh, cancel := newNonRunningTestWebhook(t, defaultOpts, acs...)

	resetMetrics()
	return wh, fmt.Sprintf("0.0.0.0:%d", port), ctx, cancel, nil
}
