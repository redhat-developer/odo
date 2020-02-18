package pipelines

import (
	"fmt"

	securityv1typedclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/serviceaccount"
)

// interface to invoke methods on a SCC (SecurityContextConstraints) object
type sccAccessor interface {
	addSCCToUser(sccName, namespace, saName string) error
}

// scc represents an OpenShift SecurityContextConstraints object with accessor interface that we need for our operations
type scc struct {
	client   securityv1typedclient.SecurityContextConstraintsInterface
	accessor sccAccessor
}

// newSCC creates a SCC object
func newSCC() (*scc, error) {

	// obtain client config
	clientConfig, err := getClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get client config due to %w", err)
	}

	// obtain security client
	securityClient, err := securityv1typedclient.NewForConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain Security Client due to %w", err)
	}

	return &scc{
		client: securityClient.SecurityContextConstraints(),
	}, nil

}

// addSCCToUser adds the a given namespace/saName to SecurityContextContaints named by sscName
func (s *scc) addSCCToUser(sccName, namespace, saName string) error {

	// get scc object by calling APIs
	sccObj, err := s.client.Get(sccName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get SCC '%s' : %w", sccName, err)
	}

	// add use to sccObj if it is not in there already
	added, newUsers := appendIfMissing(sccObj.Users, serviceaccount.MakeUsername(namespace, saName))
	if added {
		// update sccObj
		sccObj.Users = newUsers
		_, err = s.client.Update(sccObj)
		if err != nil {
			return fmt.Errorf("failed to add SA '%s/%s' to '%s' : %w", namespace, saName, sccName, err)
		}
	}

	return nil
}

// appendIfMissing appends "s" to "slice" if "s" is not in "slice."  It returns slice and true
// if "s" has been added.   Otherwise, it returns slice and false
func appendIfMissing(slice []string, s string) (bool, []string) {
	for _, elem := range slice {
		if elem == s {
			return false, slice
		}
	}
	return true, append(slice, s)
}
