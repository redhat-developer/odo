package pipelines

import (
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/tools/clientcmd"
)

// getCrdInterface returns CustomResourceDefinitionInterface
func getCrdInterface() (v1beta1.CustomResourceDefinitionInterface, error) {
	// obtain kubeconfig
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	clientConfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	// get client set from client config
	cs, err := clientset.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return cs.ApiextensionsV1beta1().CustomResourceDefinitions(), nil
}

// isCrdFound retuns true if crdName is found.  Otherwise, it returns false.
func isCrdFound(crdInterface v1beta1.CustomResourceDefinitionInterface, crdName string) (bool, error) {
	crd, err := crdInterface.Get(crdName, metav1.GetOptions{})
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

// areCrdsInstalled returns true if all given CRDs are fouund.   Otherwise, it returns false.
// If crdNames is empty, it returns false.
func areCrdsInstalled(crdNames ...string) (bool, error) {
	// If crdNames is empty, it returns false.
	if len(crdNames) == 0 {
		return false, nil
	}

	// get CRD interface
	crdInterface, err := getCrdInterface()
	if err != nil {
		return false, err
	}

	// check each CRD name
	for _, name := range crdNames {
		found, err := isCrdFound(crdInterface, name)
		if err != nil {
			// return false if crd is not found
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		// return false immmediately if a CRD is not found
		if !found {
			return found, nil
		}
	}

	// crds are found
	return true, nil
}

// isTektonPipelinesInstalled returns true if Tekton Pipeline CRD is installed.   Otherwiise, it returns false.
func isTektonPipelinesInstalled() (bool, error) {
	return areCrdsInstalled("pipelineresources.tekton.dev", "pipelineresources.tekton.dev",
		"pipelineruns.tekton.dev", "triggerbindings.tekton.dev", "triggertemplates.tekton.dev",
		"clustertasks.tekton.dev", "taskruns.tekton.dev", "tasks.tekton.dev")

}
