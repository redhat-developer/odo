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

package webhook

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/triggers/pkg/interceptors"
	corev1 "k8s.io/api/core/v1"

	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"

	"go.uber.org/zap"
)

// Timeout for outgoing requests to interceptor services
const interceptorTimeout = 5 * time.Second

type Interceptor struct {
	HTTPClient             *http.Client
	EventListenerNamespace string
	Logger                 *zap.SugaredLogger
	Webhook                *triggersv1.WebhookInterceptor
}

func NewInterceptor(wh *triggersv1.WebhookInterceptor, c *http.Client, ns string, l *zap.SugaredLogger) interceptors.Interceptor {
	timeoutClient := &http.Client{
		Transport: c.Transport,
		Timeout:   interceptorTimeout,
	}
	return &Interceptor{
		HTTPClient:             timeoutClient,
		EventListenerNamespace: ns,
		Logger:                 l,
		Webhook:                wh,
	}
}

func (w *Interceptor) ExecuteTrigger(request *http.Request) (*http.Response, error) {
	u, err := getURI(w.Webhook.ObjectRef, w.EventListenerNamespace) // TODO: Cache this result or do this on initialization
	if err != nil {
		return nil, err
	}
	request.URL = u
	request.Host = u.Host
	addInterceptorHeaders(request.Header, w.Webhook.Header)

	resp, err := w.HTTPClient.Do(request)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode != http.StatusOK {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return resp, errors.New("failed to parse response body")
		}
		return resp, fmt.Errorf("request rejected; status: %s; message: %s", resp.Status, respBody)
	}
	return resp, err
}

// getURI retrieves the ObjectReference to URI.
func getURI(objRef *corev1.ObjectReference, ns string) (*url.URL, error) {
	// TODO: This should work for any Addressable.
	// Use something like https://github.com/knative/eventing-contrib/blob/7c0fc5cfa8bd44da0767d9e7b250264ea6eb7d8d/pkg/controller/sinks/sinks.go#L32
	if objRef.Kind == "Service" && objRef.APIVersion == "v1" {
		// TODO: Also assuming port 80 and http here. Use DNS/or the env vars?
		if objRef.Namespace != "" {
			ns = objRef.Namespace
		}
		return url.Parse(fmt.Sprintf("http://%s.%s.svc/", objRef.Name, ns))
	}
	return nil, errors.New("Invalid objRef")
}

func addInterceptorHeaders(header http.Header, headerParams []pipelinev1.Param) {
	// This clobbers any matching headers
	for _, param := range headerParams {
		if param.Value.Type == pipelinev1.ParamTypeString {
			header.Set(param.Name, param.Value.StringVal)
		} else {
			header.Del(param.Name)
			for _, v := range param.Value.ArrayVal {
				header.Add(param.Name, v)
			}
		}
	}
}
