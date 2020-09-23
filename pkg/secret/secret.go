package secret

import (
	"fmt"

	"strings"

	"github.com/openshift/odo/pkg/occlient"
	corev1 "k8s.io/api/core/v1"
)
import (
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
)

// DetermineSecretName resolves the name of the secret that corresponds to the supplied component name and port
func DetermineSecretName(client *occlient.Client, componentName, applicationName, port string) (string, error) {
	labelSelector := fmt.Sprintf("%v=%v", applabels.ApplicationLabel, applicationName) +
		fmt.Sprintf(",%v=%v", componentlabels.ComponentLabel, componentName)
	secrets, err := client.ListSecrets(labelSelector)
	if err != nil {
		return "", err
	}

	if len(secrets) == 0 {
		return "", fmt.Errorf(`a secret should have been created for component %s. 
Please delete the component and recreate it using 'odo create'`, componentName)
	}

	// when the port is not supplied, then we either select the only one exposed is so is the case,
	// or when there multiple ports exposed, we fail
	if len(port) == 0 {
		if len(secrets) == 1 {
			return secrets[0].Name, nil
		}
		return "", fmt.Errorf("unable to properly link to component %s. "+
			"Please select one of the following ports: '%s' "+
			"by supplying the --port option and rerun the command", componentName, strings.Join(availablePorts(secrets), ","))
	}

	// search each secret to see which port is corresponds to
	for _, secret := range secrets {
		if secret.Annotations[occlient.ComponentPortAnnotationName] == port {
			return secret.Name, nil
		}
	}
	return "", fmt.Errorf("unable to properly link to component %s using port %s. "+
		"Please select one of the following ports: '%s' "+
		"by supplying the --port option and rerun the command", componentName, port, strings.Join(availablePorts(secrets), ","))
}

func availablePorts(secrets []corev1.Secret) []string {
	ports := make([]string, 0, len(secrets))
	for _, secret := range secrets {
		ports = append(ports, secret.Annotations[occlient.ComponentPortAnnotationName])
	}
	return ports
}
