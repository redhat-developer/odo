package pipelines

import (
	"fmt"

	securityv1typedclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/serviceaccount"
)

type scc struct {
	client securityv1typedclient.SecurityContextConstraintsInterface
}

func newSCC() (*scc, error) {

	// obtain client config
	clientConfig, err := getClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get client config due to %w", err)
	}

	securityClient, err := securityv1typedclient.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return &scc{
		client: securityClient.SecurityContextConstraints(),
	}, nil

}

func (s *scc) addSCCToUser(sccName, namespace, saName string) error {

	scc, err := s.client.Get(sccName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	added, newUsers := appendIfMissing(scc.Users, serviceaccount.MakeUsername(namespace, saName))
	if !added {
		return nil
	}

	scc.Users = newUsers
	_, err = s.client.Update(scc)
	if err != nil {
		return err
	}
	return nil
}

func appendIfMissing(slice []string, s string) (bool, []string) {
	for _, ele := range slice {
		if ele == s {
			return false, slice
		}
	}
	return true, append(slice, s)
}
