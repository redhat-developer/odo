package component

import (
	"fmt"
	appsv1 "github.com/openshift/api/apps/v1"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/util"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// PushedComponent is an abstraction over the cluster representation of the component
type PushedComponent interface {
	GetLabels() map[string]string
	GetAnnotations() map[string]string
	GetName() string
	GetType() (string, error)
	GetSource() (string, string, error)
	GetEnvVars() []v12.EnvVar
	GetLinkedSecretNames() []string
}

type s2iComponent struct {
	dc *appsv1.DeploymentConfig
}

func (s s2iComponent) GetLinkedSecretNames() (secretNames []string) {
	for _, env := range s.dc.Spec.Template.Spec.Containers[0].EnvFrom {
		if env.SecretRef != nil {
			secretNames = append(secretNames, env.SecretRef.Name)
		}
	}
	return secretNames
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
	d *v1.Deployment
}

func (d devfileComponent) GetLinkedSecretNames() (secretNames []string) {
	for _, container := range d.d.Spec.Template.Spec.Containers {
		for _, env := range container.EnvFrom {
			if env.SecretRef != nil {
				secretNames = append(secretNames, env.SecretRef.Name)
			}
		}
	}
	return secretNames
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
	return d.d.Name
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

func getSource(component PushedComponent) (string, string, error) {
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

func getType(component PushedComponent) (string, error) {
	if componentType, ok := component.GetLabels()[componentlabels.ComponentTypeLabel]; ok {
		return componentType, nil
	}
	return "", fmt.Errorf("%s component doesn't provide a type label", component.GetName())
}

// GetPushedComponent returns an abstraction over the cluster representation of the component
func GetPushedComponent(c *occlient.Client, componentName, applicationName string) (PushedComponent, error) {
	d, err := c.GetKubeClient().AppsV1().Deployments(c.Namespace).Get(componentName, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			// if it's not found, check if there's a deploymentconfig
			deploymentName, err := util.NamespaceOpenShiftObject(componentName, applicationName)
			if err != nil {
				return nil, err
			}
			dc, err := c.GetDeploymentConfigFromName(deploymentName)
			if err != nil {
				if kerrors.IsNotFound(err) {
					return nil, nil
				}
			} else {
				return &s2iComponent{dc: dc}, nil
			}
		}
		return nil, err
	}
	return &devfileComponent{d: d}, nil
}
