package secret

import (
	"fmt"

	"strings"

	"github.com/openshift/odo/pkg/occlient"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	applabels "github.com/openshift/odo/pkg/application/labels"

	componentlabels "github.com/openshift/odo/pkg/component/labels"
)

var secretTypeMeta = TypeMeta("Secret", "v1")

type objectMetaFunc func(om *v1.ObjectMeta)

func TypeMeta(kind, apiVersion string) v1.TypeMeta {
	return v1.TypeMeta{
		Kind:       kind,
		APIVersion: apiVersion,
	}
}

func ObjectMeta(n types.NamespacedName, opts ...objectMetaFunc) v1.ObjectMeta {
	om := v1.ObjectMeta{
		Namespace: n.Namespace,
		Name:      n.Name,
	}
	for _, o := range opts {
		o(&om)
	}
	return om

}

// DetermineSecretName resolves the name of the secret that corresponds to the supplied component name and port
func DetermineSecretName(client *occlient.Client, componentName, applicationName, port string) (string, error) {
	labelSelector := fmt.Sprintf("%v=%v", applabels.ApplicationLabel, applicationName) +
		fmt.Sprintf(",%v=%v", componentlabels.ComponentLabel, componentName)
	secrets, err := client.ListSecrets(labelSelector)
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
		if secret.Annotations[occlient.ComponentPortAnnotationName] == port {
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
		ports = append(ports, secret.Annotations[occlient.ComponentPortAnnotationName])
	}
	return ports
}

func CreateDockerConfigSecret(name types.NamespacedName, configData []byte) (*corev1.Secret, error) {
	return createSecret(name, ".dockerconfigjson", corev1.SecretTypeDockerConfigJson, configData)
}

func createSecret(name types.NamespacedName, key string, st corev1.SecretType, configData []byte) (*corev1.Secret, error) {

	secret := &corev1.Secret{
		TypeMeta:   secretTypeMeta,
		ObjectMeta: ObjectMeta(name),
		Type:       st,
		Data: map[string][]byte{
			key: configData,
		},
	}
	return secret, nil
}
