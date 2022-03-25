package component

import (
	"errors"
	"fmt"

	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/storage"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

type provider interface {
	GetLabels() map[string]string
	GetAnnotations() map[string]string
	GetName() string
	GetEnvVars() []v12.EnvVar
	GetLinkedSecrets() []SecretMount
}

// PushedComponent is an abstraction over the cluster representation of the component
type PushedComponent interface {
	provider
	GetApplication() string
	GetType() (string, error)
	GetStorage() ([]storage.Storage, error)
}

type defaultPushedComponent struct {
	application   string
	storage       []storage.Storage
	provider      provider
	client        kclient.ClientInterface
	storageClient storage.Client
}

func (d defaultPushedComponent) GetLabels() map[string]string {
	return d.provider.GetLabels()
}

func (d defaultPushedComponent) GetAnnotations() map[string]string {
	return d.provider.GetAnnotations()
}

func (d defaultPushedComponent) GetName() string {
	return d.provider.GetName()
}

func (d defaultPushedComponent) GetType() (string, error) {
	return getType(d.provider)
}

func (d defaultPushedComponent) GetEnvVars() []v12.EnvVar {
	return d.provider.GetEnvVars()
}

func (d defaultPushedComponent) GetLinkedSecrets() []SecretMount {
	return d.provider.GetLinkedSecrets()
}

// GetStorage gets the storage using the storage client of the given pushed component
func (d defaultPushedComponent) GetStorage() ([]storage.Storage, error) {
	if d.storage == nil {
		if _, ok := d.provider.(*devfileComponent); ok {
			storageList, err := d.storageClient.ListFromCluster()
			if err != nil {
				return nil, err
			}
			d.storage = storageList.Items
		}
	}
	return d.storage, nil
}

func (d defaultPushedComponent) GetApplication() string {
	return d.application
}

type devfileComponent struct {
	d v1.Deployment
}

func (d devfileComponent) GetLinkedSecrets() (secretMounts []SecretMount) {
	for _, container := range d.d.Spec.Template.Spec.Containers {
		for _, env := range container.EnvFrom {
			if env.SecretRef != nil {
				secretMounts = append(secretMounts, SecretMount{
					SecretName:  env.SecretRef.Name,
					MountVolume: false,
				})
			}
		}
	}

	for _, volume := range d.d.Spec.Template.Spec.Volumes {
		if volume.Secret != nil {
			mountPath := ""
			for _, container := range d.d.Spec.Template.Spec.Containers {
				for _, mount := range container.VolumeMounts {
					if mount.Name == volume.Name {
						mountPath = mount.MountPath
						break
					}
				}
			}
			secretMounts = append(secretMounts, SecretMount{
				SecretName:  volume.Secret.SecretName,
				MountVolume: true,
				MountPath:   mountPath,
			})
		}
	}

	return secretMounts
}

func (d devfileComponent) GetEnvVars() []v12.EnvVar {
	var envs []v12.EnvVar
	for _, container := range d.d.Spec.Template.Spec.Containers {
		envs = append(envs, container.Env...)
	}
	return envs
}

func (d devfileComponent) GetLabels() map[string]string {
	return d.d.Labels
}
func (d devfileComponent) GetAnnotations() map[string]string {
	return d.d.Annotations
}

func (d devfileComponent) GetName() string {
	return d.d.Labels[componentlabels.KubernetesInstanceLabel]
}

func getType(component provider) (string, error) {

	// For backwards compatibility with previously deployed components that could be non-odo, check the annotation first
	// then check to see if there is a label with the project type
	if componentType, ok := component.GetAnnotations()[componentlabels.OdoProjectTypeAnnotation]; ok {
		return componentType, nil
	} else if componentType, ok = component.GetLabels()[componentlabels.OdoProjectTypeAnnotation]; ok {
		return componentType, nil
	}

	return "", fmt.Errorf("%s component doesn't provide a type annotation; consider pushing the component again", component.GetName())
}

func newPushedComponent(applicationName string, p provider, c kclient.ClientInterface, storageClient storage.Client) PushedComponent {
	return &defaultPushedComponent{
		application:   applicationName,
		provider:      p,
		client:        c,
		storageClient: storageClient,
	}
}

// GetPushedComponent returns an abstraction over the cluster representation of the component
func GetPushedComponent(c kclient.ClientInterface, componentName, applicationName string) (PushedComponent, error) {
	d, err := c.GetOneDeployment(componentName, applicationName)
	if err != nil {
		if isIgnorableError(err) {
			return nil, nil
		}
		return nil, err
	}
	storageClient := storage.NewClient(storage.ClientOptions{
		Client:     c,
		Deployment: d,
	})

	return newPushedComponent(applicationName, &devfileComponent{d: *d}, c, storageClient), nil
}

func isIgnorableError(err error) bool {
	for {
		e := errors.Unwrap(err)
		if e != nil {
			err = e
		} else {
			break
		}
	}

	if _, ok := err.(*kclient.DeploymentNotFoundError); ok {
		return true
	}
	return kerrors.IsNotFound(err) || kerrors.IsForbidden(err) || kerrors.IsUnauthorized(err)

}
