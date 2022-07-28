package component

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	odolabels "github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/storage"
)

type provider interface {
	GetLabels() map[string]string
	GetAnnotations() map[string]string
	GetName() string
	GetEnvVars() []v1.EnvVar
	GetLinkedSecrets() []SecretMount
}

// SecretMount describes a Secret mount (either as environment variables with envFrom or as a volume)
type SecretMount struct {
	ServiceName string
	SecretName  string
	MountVolume bool
	MountPath   string
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
	storageClient storage.Client
}

var _ provider = (*defaultPushedComponent)(nil)
var _ PushedComponent = (*defaultPushedComponent)(nil)

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

func (d defaultPushedComponent) GetEnvVars() []v1.EnvVar {
	return d.provider.GetEnvVars()
}

func (d defaultPushedComponent) GetLinkedSecrets() []SecretMount {
	return d.provider.GetLinkedSecrets()
}

// GetStorage gets the storage using the storage client of the given pushed component
func (d defaultPushedComponent) GetStorage() ([]storage.Storage, error) {
	if d.storage == nil {
		if _, ok := d.provider.(*devfileComponent); ok {
			storageList, err := d.storageClient.List()
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
	d appsv1.Deployment
}

var _ provider = (*devfileComponent)(nil)

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

func (d devfileComponent) GetEnvVars() []v1.EnvVar {
	var envs []v1.EnvVar
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
	return odolabels.GetComponentName(d.d.Labels)
}

func getType(component provider) (string, error) {
	res, err := odolabels.GetProjectType(component.GetLabels(), component.GetAnnotations())
	if err != nil {
		return "", fmt.Errorf("%s component doesn't provide a type annotation; consider pushing the component again", component.GetName())
	}
	return res, nil
}
