package component

import (
	"fmt"

	appsv1 "github.com/openshift/api/apps/v1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/storage"
	"github.com/openshift/odo/pkg/url"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
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
	GetSource() (string, string, error)
	GetStorage() ([]storage.Storage, error)
}

type defaultPushedComponent struct {
	application   string
	urls          []url.URL
	storage       []storage.Storage
	provider      provider
	client        *occlient.Client
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

func (d defaultPushedComponent) GetSource() (string, string, error) {
	return getSource(d.provider)
}

func (d defaultPushedComponent) GetEnvVars() []v12.EnvVar {
	return d.provider.GetEnvVars()
}

func (d defaultPushedComponent) GetLinkedSecrets() []SecretMount {
	return d.provider.GetLinkedSecrets()
}

func (d defaultPushedComponent) GetURLs() ([]url.URL, error) {
	if d.urls == nil {
		name := d.GetName()
		var routes url.URLList
		if routeAvailable, err := d.client.IsRouteSupported(); routeAvailable && err == nil {
			routes, err = url.ListPushed(d.client, name, d.GetApplication())
			if err != nil && !isIgnorableError(err) {
				return []url.URL{}, err
			}
		}
		ingresses, err := url.ListPushedIngress(d.client.GetKubeClient(), name)
		if err != nil && !isIgnorableError(err) {
			return []url.URL{}, err
		}
		d.urls = append(routes.Items, ingresses.Items...)
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
			d.storage = append(d.storage, storageList.Items...)
		}
	}
	return d.storage, nil
}

func (d defaultPushedComponent) GetApplication() string {
	return d.application
}

type s2iComponent struct {
	dc appsv1.DeploymentConfig
}

func (s s2iComponent) GetLinkedSecrets() (secretMounts []SecretMount) {
	for _, env := range s.dc.Spec.Template.Spec.Containers[0].EnvFrom {
		if env.SecretRef != nil {
			secretMounts = append(secretMounts, SecretMount{
				SecretName:  env.SecretRef.Name,
				MountVolume: false,
			})
		}
	}
	return secretMounts
}

func (s s2iComponent) GetEnvVars() []v12.EnvVar {
	return s.dc.Spec.Template.Spec.Containers[0].Env
}

func (s s2iComponent) GetLabels() map[string]string {
	return s.dc.Labels
}

func (s s2iComponent) GetAnnotations() map[string]string {
	return s.dc.Annotations
}

func (s s2iComponent) GetName() string {
	return s.dc.Labels[componentlabels.ComponentLabel]
}

func (s s2iComponent) GetType() (string, error) {
	return getType(s)
}

func (s s2iComponent) GetSource() (string, string, error) {
	return getSource(s)
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

func (d devfileComponent) GetSource() (string, string, error) {
	return getSource(d)
}

type noSourceError struct {
	msg string
}

func (n noSourceError) Error() string {
	return n.msg
}

func getSource(component provider) (string, string, error) {
	annotations := component.GetAnnotations()
	if sourceType, ok := annotations[ComponentSourceTypeAnnotation]; ok {
		if !validateSourceType(sourceType) {
			return "", "", fmt.Errorf("unsupported component source type %s", sourceType)
		}
		var sourcePath string
		if sourceType == string(config.GIT) {
			sourcePath = annotations[componentSourceURLAnnotation]
		}

		klog.V(4).Infof("Source for component %s is %s (%s)", component.GetName(), sourcePath, sourceType)
		return sourceType, sourcePath, nil
	}
	return "", "", noSourceError{msg: fmt.Sprintf("%s component doesn't provide a source type annotation", component.GetName())}
}

func getType(component provider) (string, error) {
	if componentType, ok := component.GetAnnotations()[componentlabels.ComponentTypeAnnotation]; ok {
		return componentType, nil
	} else if _, ok = component.GetLabels()[componentlabels.ComponentTypeLabel]; ok {
		klog.V(1).Info("No annotation assigned; retuning 'Not available' since labels are assigned. Annotations will be assigned when user pushes again.")
		return NOTAVAILABLE, nil
	}
	return "", fmt.Errorf("%s component doesn't provide a type annotation; consider pushing the component again", component.GetName())
}

// GetPushedComponents retrieves a map of PushedComponents from the cluster, keyed by their name
func GetPushedComponents(c *occlient.Client, applicationName string) (map[string]PushedComponent, error) {
	applicationSelector := fmt.Sprintf("%s=%s", applabels.ApplicationLabel, applicationName)

	dcList, err := c.ListDeploymentConfigs(applicationSelector)
	if err != nil {
		if !isIgnorableError(err) {
			return nil, err
		}
	}
	res := make(map[string]PushedComponent, len(dcList))
	for _, dc := range dcList {
		comp := newPushedComponent(applicationName, &s2iComponent{dc: dc}, c, nil)
		res[comp.GetName()] = comp
	}

	deploymentList, err := c.GetKubeClient().ListDeployments(applicationSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list components")
	}

	for _, d := range deploymentList.Items {
		deployment := d
		storageClient := storage.NewClient(storage.ClientOptions{
			OCClient:   *c,
			Deployment: &deployment,
		})
		comp := newPushedComponent(applicationName, &devfileComponent{d: d}, c, storageClient)
		res[comp.GetName()] = comp
	}

	return res, nil
}

func newPushedComponent(applicationName string, p provider, c *occlient.Client, storageClient storage.Client) PushedComponent {
	return &defaultPushedComponent{
		application:   applicationName,
		provider:      p,
		client:        c,
		storageClient: storageClient,
	}
}

// GetPushedComponent returns an abstraction over the cluster representation of the component
func GetPushedComponent(c *occlient.Client, componentName, applicationName string) (PushedComponent, error) {
	d, err := c.GetKubeClient().GetOneDeployment(componentName, applicationName)
	if err != nil {
		if isIgnorableError(err) {
			// if it's not found, check if there's a deploymentconfig
			deploymentName, err := util.NamespaceOpenShiftObject(componentName, applicationName)
			if err != nil {
				return nil, err
			}
			dc, err := c.GetDeploymentConfigFromName(deploymentName)
			if err != nil {
				if kerrors.IsNotFound(err) {
					// in case where odo's standard naming practices are not followed, it makes sense to do a double check with component name
					// this is useful when dealing with components that are not managed/created by odo
					dc, err = c.GetDeploymentConfigFromName(componentName)
					if err != nil {
						return nil, nil
					} else {
						return newPushedComponent(applicationName, &s2iComponent{dc: *dc}, c, nil), nil
					}
				}
			} else {
				return newPushedComponent(applicationName, &s2iComponent{dc: *dc}, c, nil), nil
			}
		}
		return nil, err
	}
	storageClient := storage.NewClient(storage.ClientOptions{
		OCClient:   *c,
		Deployment: d,
	})
	return newPushedComponent(applicationName, &devfileComponent{d: *d}, c, storageClient), nil
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
