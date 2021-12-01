package component

import (
	"fmt"

	"github.com/pkg/errors"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/url"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
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
	GetURLs() ([]url.URL, error)
	GetApplication() string
	GetType() (string, error)
	GetStorage() ([]storage.Storage, error)
}

type defaultPushedComponent struct {
	application   string
	urls          []url.URL
	storage       []storage.Storage
	provider      provider
	client        *occlient.Client
	storageClient storage.Client
	urlClient     url.Client
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

func (d defaultPushedComponent) GetURLs() ([]url.URL, error) {
	if d.urls == nil {
		urls, err := d.urlClient.ListFromCluster()
		if err != nil && !isIgnorableError(err) {
			return []url.URL{}, err
		}
		d.urls = urls.Items
	}
	return d.urls, nil
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
	return d.d.Labels[componentlabels.ComponentLabel]
}

func (d devfileComponent) GetType() (string, error) {
	return getType(d)
}

func getType(component provider) (string, error) {
	if componentType, ok := component.GetAnnotations()[componentlabels.ComponentTypeAnnotation]; ok {
		return componentType, nil
	} else if _, ok = component.GetLabels()[componentlabels.ComponentTypeLabel]; ok {
		klog.V(1).Info("No annotation assigned; retuning 'Not available' since labels are assigned. Annotations will be assigned when user pushes again.")
		return NotAvailable, nil
	}
	return "", fmt.Errorf("%s component doesn't provide a type annotation; consider pushing the component again", component.GetName())
}

// GetPushedComponents retrieves a map of PushedComponents from the cluster, keyed by their name
func GetPushedComponents(c *occlient.Client, applicationName string) (map[string]PushedComponent, error) {
	applicationSelector := fmt.Sprintf("%s=%s", applabels.ApplicationLabel, applicationName)

	deploymentList, err := c.GetKubeClient().ListDeployments(applicationSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list components")
	}
	res := make(map[string]PushedComponent, len(deploymentList.Items))
	for _, d := range deploymentList.Items {
		deployment := d
		storageClient := storage.NewClient(storage.ClientOptions{
			OCClient:   *c,
			Deployment: &deployment,
		})

		urlClient := url.NewClient(url.ClientOptions{
			OCClient:   *c,
			Deployment: &deployment,
		})
		comp := newPushedComponent(applicationName, &devfileComponent{d: d}, c, storageClient, urlClient)
		res[comp.GetName()] = comp
	}

	return res, nil
}

func newPushedComponent(applicationName string, p provider, c *occlient.Client, storageClient storage.Client, urlClient url.Client) PushedComponent {
	return &defaultPushedComponent{
		application:   applicationName,
		provider:      p,
		client:        c,
		storageClient: storageClient,
		urlClient:     urlClient,
	}
}

// GetPushedComponent returns an abstraction over the cluster representation of the component
func GetPushedComponent(c *occlient.Client, componentName, applicationName string) (PushedComponent, error) {
	d, err := c.GetKubeClient().GetOneDeployment(componentName, applicationName)
	if err != nil {
		if isIgnorableError(err) {
			return nil, nil
		}
		return nil, err
	}
	storageClient := storage.NewClient(storage.ClientOptions{
		OCClient:   *c,
		Deployment: d,
	})

	urlClient := url.NewClient(url.ClientOptions{
		OCClient:   *c,
		Deployment: d,
	})
	return newPushedComponent(applicationName, &devfileComponent{d: *d}, c, storageClient, urlClient), nil
}

func isIgnorableError(err error) bool {
	e := errors.Cause(err)
	if e != nil {
		err = e
	}
	if _, ok := err.(*kclient.DeploymentNotFoundError); ok {
		return true
	}
	return kerrors.IsNotFound(err) || kerrors.IsForbidden(err) || kerrors.IsUnauthorized(err)
}
