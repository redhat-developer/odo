package servicebindingrequest

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

// Secret represents the data collected by this operator, and later handled as a secret.
type Secret struct {
	logger *log.Log          // logger instance
	client dynamic.Interface // Kubernetes API client
	ns     string
	name   string
}

// buildResourceClient creates a resource client to handle corev1/secret resource.
func (s *Secret) buildResourceClient() dynamic.ResourceInterface {
	gvr := corev1.SchemeGroupVersion.WithResource(SecretResource)
	return s.client.Resource(gvr).Namespace(s.ns)
}

// createOrUpdate will take informed payload and either create a new secret or update an existing
// one. It can return error when Kubernetes client does.
func (s *Secret) createOrUpdate(payload map[string][]byte) (*unstructured.Unstructured, error) {
	logger := s.logger.WithValues("Namespace", s.ns, "Name", s.name)
	secretObj := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.ns,
			Name:      s.name,
		},
		Data: payload,
	}

	gvk := corev1.SchemeGroupVersion.WithKind(SecretKind)
	u, err := converter.ToUnstructuredAsGVK(secretObj, gvk)
	if err != nil {
		return nil, err
	}

	resourceClient := s.buildResourceClient()

	logger.Debug("Attempt to create secret...")
	_, err = resourceClient.Create(u, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, err
	}

	logger.Debug("Secret already exists, updating contents instead...")
	_, err = resourceClient.Update(u, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Commit will store informed data as a secret, commit it against the API server. It can forward
// errors from custom environment parser component, or from the API server itself.
func (s *Secret) Commit(payload map[string][]byte) (*unstructured.Unstructured, error) {
	return s.createOrUpdate(payload)
}

// Get an unstructured object from the secret handled by this component. It can return errors in case
// the API server does.
func (s *Secret) Get() (*unstructured.Unstructured, bool, error) {
	resourceClient := s.buildResourceClient()
	u, err := resourceClient.Get(s.name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, false, err
	}
	return u, u != nil, nil
}

// Delete the secret represented by this component. It can return error when the API server does.
func (s *Secret) Delete() error {
	resourceClient := s.buildResourceClient()
	err := resourceClient.Delete(s.name, &metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

// NewSecret instantiate a new Secret.
func NewSecret(
	client dynamic.Interface,
	ns string,
	name string,
) *Secret {
	return &Secret{
		logger: log.NewLog("secret"),
		client: client,

		name: name,
		ns:   ns,
	}
}
