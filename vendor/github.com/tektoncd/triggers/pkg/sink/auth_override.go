/*
Copyright 2020 The Tekton Authors

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
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	dynamicClientset "github.com/tektoncd/triggers/pkg/client/dynamic/clientset"
	"github.com/tektoncd/triggers/pkg/client/dynamic/clientset/tekton"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	discoveryclient "k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

//AuthOverride is an interface that constructs a discovery client for the ServerResourceInterface
//and a dynamic client for the Tekton Resources, using the token provide as the bearer token in the
//REST config used to build those client.  The other non-credential related parameters for the
//REST client used are copied from the in cluster config of the event sink.
type AuthOverride interface {
	OverrideAuthentication(token string,
		log *zap.SugaredLogger,
		defaultDiscoveryClient discoveryclient.ServerResourcesInterface,
		defaultDynamicClient dynamic.Interface) (discoveryClient discoveryclient.ServerResourcesInterface,
		dynamicClient dynamic.Interface,
		err error)
}

func isServiceAccountToken(secret *corev1.Secret, sa *corev1.ServiceAccount) bool {
	if secret.Type != corev1.SecretTypeServiceAccountToken {
		return false
	}

	name := secret.Annotations[corev1.ServiceAccountNameKey]
	uid := secret.Annotations[corev1.ServiceAccountUIDKey]
	if name != sa.Name {
		// Name must match
		return false
	}
	if len(uid) > 0 && uid != string(sa.UID) {
		// If UID is specified, it must match
		return false
	}

	return true
}

func (r Sink) retrieveAuthToken(saRef *corev1.ObjectReference, eventLog *zap.SugaredLogger) (string, error) {
	if saRef == nil {
		return "", nil
	}

	if len(saRef.Name) == 0 || len(saRef.Namespace) == 0 {
		return "", nil
	}

	var log *zap.SugaredLogger
	if eventLog != nil {
		log = eventLog.With(zap.String(triggersv1.TriggerLabelKey, "retriveAuthToken"))
	}

	sa, err := r.KubeClientSet.CoreV1().ServiceAccounts(saRef.Namespace).Get(saRef.Name, metav1.GetOptions{})
	if err != nil {
		if log != nil {
			log.Error(err)
		}
		return "", err
	}
	var savedErr error
	for _, ref := range sa.Secrets {
		// secret ref namespace most likely won't be set, as the secret can only reside in
		// sa's namespace, so use that
		secret, err := r.KubeClientSet.CoreV1().Secrets(saRef.Namespace).Get(ref.Name, metav1.GetOptions{})
		if err != nil {
			if log != nil {
				log.Error(err)
			}
			savedErr = err
			continue
		}
		if isServiceAccountToken(secret, sa) {
			token, exists := secret.Data[corev1.ServiceAccountTokenKey]
			if log != nil {
				log.Debugf("retrieveAuthToken found SA auth: %v", exists)
			}
			if !exists {
				continue
			}

			return string(token), nil
		}
	}
	return "", savedErr
}

func newConfig(newToken string, config *rest.Config) *rest.Config {
	// first clean out all user credentials from the pods' in cluster config
	newConfig := rest.AnonymousClientConfig(config)
	// add the token from our interceptors
	newConfig.BearerToken = newToken
	return newConfig
}

type DefaultAuthOverride struct {
}

func (r DefaultAuthOverride) OverrideAuthentication(token string,
	log *zap.SugaredLogger,
	defaultDiscoverClient discoveryclient.ServerResourcesInterface,
	defaultDynamicClient dynamic.Interface) (discoveryClient discoveryclient.ServerResourcesInterface,
	dynamicClient dynamic.Interface,
	err error) {
	dynamicClient = defaultDynamicClient
	discoveryClient = defaultDiscoverClient
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Errorf("overrideAuthentication: problem getting in cluster config: %#v\n", err)
		return
	}
	clusterConfig = newConfig(token, clusterConfig)
	dc, err := dynamic.NewForConfig(clusterConfig)
	if err != nil {
		log.Errorf("overrideAuthentication: problem getting dynamic client set: %#v\n", err)
		return
	}
	kubeClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		log.Errorf("overrideAuthentication: problem getting kube client: %#v\n", err)
		return
	}
	dynamicClient = dynamicClientset.New(tekton.WithClient(dc))
	discoveryClient = kubeClient.Discovery()

	return
}
