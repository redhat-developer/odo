/*
Copyright 2019 The Knative Authors

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

package metrics

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"go.opencensus.io/stats/view"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/metrics"

	"knative.dev/pkg/metrics/metricstest"
)

// clientFunc lets us implement rest.HTTPClient with a function matching
// the signature of its Do method.
type clientFunc struct {
	do func(*http.Request) (*http.Response, error)
}

var _ rest.HTTPClient = (*clientFunc)(nil)

// Do implements rest.HTTPClient
func (cf *clientFunc) Do(req *http.Request) (*http.Response, error) {
	return cf.do(req)
}

// ClientFunc turns a method matching the signature of rest.HTTPClient's
// Do() method into an implementation of rest.HTTPClient.
func ClientFunc(f func(*http.Request) (*http.Response, error)) rest.HTTPClient {
	return &clientFunc{do: f}
}

func TestClientMetrics(t *testing.T) {
	cp := &ClientProvider{
		Latency: newFloat64("latency"),
		Result:  newInt64("result"),
	}
	metrics.Register(cp.NewLatencyMetric(), cp.NewResultMetric())

	// Reset the metrics configuration to avoid leaked state from other tests.
	setCurMetricsConfig(nil)

	views := cp.DefaultViews()
	if got, want := len(views), 2; got != want {
		t.Errorf("len(DefaultViews()) = %d, want %d", got, want)
	}
	if err := view.Register(views...); err != nil {
		t.Errorf("view.Register() = %v", err)
	}
	defer view.Unregister(views...)

	// No stats have been reported yet.
	metricstest.CheckStatsNotReported(t, "latency", "result")

	base := &url.URL{
		Scheme: "http",
		Host:   "api.mattmoor.dev",
	}
	config := rest.ContentConfig{
		ContentType: "application/json",
		GroupVersion: &schema.GroupVersion{
			Group:   "testing.knative.dev",
			Version: "v1alpha1",
		},
	}
	client := ClientFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString("hi")),
		}, nil
	})

	// When we send rest requests, we should trigger the metrics setup above.
	req := rest.NewRequest(client, http.MethodGet, base, "/testing.knative.dev/v1alpha1",
		config, rest.Serializers{}, nil, nil, 0)
	result := req.Do()
	if err := result.Error(); err != nil {
		t.Errorf("Do() = %v", err)
	}

	// Now we have stats reported!
	metricstest.CheckStatsReported(t, "latency", "result")
}
