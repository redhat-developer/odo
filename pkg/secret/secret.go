package secret

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/kclient"

	"strings"

	"github.com/redhat-developer/odo/pkg/occlient"
	corev1 "k8s.io/api/core/v1"

	applabels "github.com/redhat-developer/odo/pkg/application/labels"

	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
)

// DetermineSecretName resolves the name of the secret that corresponds to the supplied component name and port
func DetermineSecretName(client *occlient.Client, componentName, applicationName, port string) (string, error) {
	labelSelector := fmt.Sprintf("%v=%v", applabels.ApplicationLabel, applicationName) +
		fmt.Sprintf(",%v=%v", componentlabels.ComponentLabel, componentName)
	secrets, err := client.GetKubeClient().ListSecrets(labelSelector)
	if err != nil {
		return "", err
	}

	if len(secrets) == 0 {
		return "", fmt.Errorf(`A secret should have been created for component %s. 
Please delete the component and recreate it using 'odo create'`, componentName)
	}

	// when the port is not supplied, then we either select the only one exposed is so is the case,
	// or when there multiple ports exposed, we fail
	if len(port) == 0 {
		if len(secrets) == 1 {
			return secrets[0].Name, nil
		}
		return "", fmt.Errorf("Unable to properly link to component %s. "+
			"Please select one of the following ports: '%s' "+
			"by supplying the --port option and rerun the command", componentName, strings.Join(availablePorts(secrets), ","))
	}

	// search each secret to see which port is corresponds to
	for _, secret := range secrets {
		if secret.Annotations[kclient.ComponentPortAnnotationName] == port {
			return secret.Name, nil
		}
	}
	return "", fmt.Errorf("Unable to properly link to component %s using port %s. "+
		"Please select one of the following ports: '%s' "+
		"by supplying the --port option and rerun the command", componentName, port, strings.Join(availablePorts(secrets), ","))
}

func availablePorts(secrets []corev1.Secret) []string {
	ports := make([]string, 0, len(secrets))
	for _, secret := range secrets {
		ports = append(ports, secret.Annotations[kclient.ComponentPortAnnotationName])
	}
	return ports
}
