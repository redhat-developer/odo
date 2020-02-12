package pipelines

import (
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//"k8s.io/client-go/kubernetes/typed/scheduling/v1beta1"
	"k8s.io/client-go/tools/clientcmd"
)

// IsTektonPipelineInstalled returns true if Tekton Pipeline CRD is installed.   Otherwiise, it returns false.
func IsTektonPipelineInstalled() (bool, error) {

	// obtain kubeconfig
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	clientConfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return false, err
	}

	// get client set
	clients, err := clientset.NewForConfig(clientConfig)
	if err != nil {
		return false, err
	}

	// get pipeline CRD
	crd, err := clients.ApiextensionsV1beta1().CustomResourceDefinitions().Get("pipelineresources.tekton.dev", metav1.GetOptions{})
	if err != nil {
		// return false if crd is not found
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	if crd == nil {
		return false, nil
	} else {
		return true, nil
	}
}
