package pipelines

import (
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// These CRDs names are checked to confirm that Tekton Pipelines/Triggers has been installed.
var requiredCRDNames = []string{
	"pipelineresources.tekton.dev", "pipelineresources.tekton.dev",
	"pipelineruns.tekton.dev", "triggerbindings.tekton.dev", "triggertemplates.tekton.dev",
	"clustertasks.tekton.dev", "taskruns.tekton.dev", "tasks.tekton.dev",
}

// getClientConfig returns client config to be used to create client
func getClientConfig() (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return kubeconfig.ClientConfig()
}

// getCRDInterface returns CustomResourceDefinitionInterface
func getCRDInterface() (v1beta1.CustomResourceDefinitionInterface, error) {
	// obtain client config
	clientConfig, err := getClientConfig()
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

// isCRDFound retuns true if crdName is found.  Otherwise, it returns false.
func isCRDFound(crdInterface v1beta1.CustomResourceDefinitionInterface, crdName string) (bool, error) {
	crd, err := crdInterface.Get(crdName, metav1.GetOptions{})
	if err != nil {
		// return false if CRD is not found
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return crd != nil, nil
}

// areCRDsInstalled returns true if all given CRDs are fouund.   Otherwise, it returns false.
// If crdNames is empty, it returns false.
func areCRDsInstalled(crdNames []string) (bool, error) {
	// If crdNames is empty, it returns false.
	if len(crdNames) == 0 {
		return false, nil
	}

	// get CRD interface
	crdInterface, err := getCRDInterface()
	if err != nil {
		return false, err
	}

	// check each CRD name
	for _, name := range crdNames {
		found, err := isCRDFound(crdInterface, name)
		if err != nil {
			// return false if CRD is not found
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		// return false immmediately if a CRD is not found
		if !found {
			return false, nil
		}
	}

	// CRDs are found
	return true, nil
}

// isTektonPipelinesInstalled returns true if Tekton Pipeline CRD is installed.   Otherwiise, it returns false.
func isTektonPipelinesInstalled() (bool, error) {
	return areCRDsInstalled(requiredCRDNames)
}
