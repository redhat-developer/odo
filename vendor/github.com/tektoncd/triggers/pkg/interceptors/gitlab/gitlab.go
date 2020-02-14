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

package gitlab

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"

	"github.com/tektoncd/triggers/pkg/interceptors"

	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

type Interceptor struct {
	KubeClientSet          kubernetes.Interface
	Logger                 *zap.SugaredLogger
	GitLab                 *triggersv1.GitLabInterceptor
	EventListenerNamespace string
}

func NewInterceptor(gl *triggersv1.GitLabInterceptor, k kubernetes.Interface, ns string, l *zap.SugaredLogger) interceptors.Interceptor {
	return &Interceptor{
		Logger:                 l,
		GitLab:                 gl,
		KubeClientSet:          k,
		EventListenerNamespace: ns,
	}
}

func (w *Interceptor) ExecuteTrigger(request *http.Request) (*http.Response, error) {
	// Validate the secret first, if set.
	if w.GitLab.SecretRef != nil {
		header := request.Header.Get("X-GitLab-Token")
		if header == "" {
			return nil, errors.New("no X-GitLab-Token header set")
		}

		secretToken, err := interceptors.GetSecretToken(w.KubeClientSet, w.GitLab.SecretRef, w.EventListenerNamespace)
		if err != nil {
			return nil, err
		}

		// Make sure to use a constant time comparison here.
		if subtle.ConstantTimeCompare([]byte(header), secretToken) == 0 {
			return nil, errors.New("Invalid X-GitLab-Token")
		}
	}
	if w.GitLab.EventTypes != nil {
		actualEvent := request.Header.Get("X-GitLab-Event")
		isAllowed := false
		for _, allowedEvent := range w.GitLab.EventTypes {
			if actualEvent == allowedEvent {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			return nil, fmt.Errorf("event type %s is not allowed", actualEvent)
		}
	}

	return &http.Response{
		Header: request.Header,
		Body:   request.Body,
	}, nil
}
