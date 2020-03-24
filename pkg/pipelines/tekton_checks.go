package pipelines

import (
	"errors"
	"fmt"

	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	errs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// These CRDs names are checked to confirm that Tekton Pipelines/Triggers has been installed.
var requiredCRDNames = []string{
	"pipelines.tekton.dev", "pipelineresources.tekton.dev", "pipelineruns.tekton.dev",
	"triggerbindings.tekton.dev", "triggertemplates.tekton.dev",
	"clustertasks.tekton.dev", "taskruns.tekton.dev", "tasks.tekton.dev",
}

// Strategy to check Tekton install which is checking the existence of CRDs
type checkStrategy struct {
	client       v1beta1.CustomResourceDefinitionInterface
	requiredCRDs []string
}

// tektonChecker object that knows how to perform Tekton installation checks
type tektonChecker struct {
	strategy *checkStrategy
}

// newTektonChecker constructs a tektonChecker that is backed by a client configured with user's kubeconfig
func newTektonChecker() (*tektonChecker, error) {
	// obtain client config
	clientConfig, err := getClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get client config due to %w", err)
	}

	// get client set from client config
	cs, err := clientset.NewForConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get APIs client due to %w", err)
	}

	return &tektonChecker{
		strategy: &checkStrategy{
			requiredCRDs: requiredCRDNames,
			client:       cs.ApiextensionsV1beta1().CustomResourceDefinitions(),
		},
	}, nil
}

// getClientConfig returns client config to be used to create client
func getClientConfig() (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return kubeconfig.ClientConfig()
}

// isCRDFound retuns true if crdName is found.  Otherwise, it returns false.
func (s *checkStrategy) isCRDFound(crdName string) (bool, error) {
	crd, err := s.client.Get(crdName, metav1.GetOptions{})
	if err != nil {
		// return false if CRD is not found
		if errs.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get CRD for '%s' due to %w", crdName, err)
	}

	return crd != nil, nil
}

// checkInstall returns true if all given CRDs are fouund.   Otherwise, it returns false.
func (s *checkStrategy) checkInstall() (bool, error) {
	// If s.requiredCRDs is empty, we throw an error.
	if len(s.requiredCRDs) == 0 {
		return false, errors.New("no CRD names provided")
	}

	// check each CRD name
	for _, name := range s.requiredCRDs {
		found, err := s.isCRDFound(name)
		if err != nil {
			// return false if CRD is not found
			if errs.IsNotFound(err) {
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

// checkInstall returns true if Tekton Pipeline CRD is installed.   Otherwise, it returns false.
func (t *tektonChecker) checkInstall() (bool, error) {
	return t.strategy.checkInstall()
}
