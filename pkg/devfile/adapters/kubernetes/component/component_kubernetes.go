package component

import (
	"strings"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/pkg/errors"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/service"
)

// componentKubernetes represents a devfile component of type Kubernetes
type componentKubernetes struct {
	client        occlient.Client
	component     devfilev1.Component
	componentName string
	appName       string
}

func newComponentKubernetes(client occlient.Client, component devfilev1.Component, componentName string, appName string) componentKubernetes {
	return componentKubernetes{
		client:        client,
		component:     component,
		componentName: componentName,
		appName:       appName,
	}
}

// Apply a component of type Kubernetes by creating resources into a Kubernetes cluster
func (o componentKubernetes) Apply(devfileObj parser.DevfileObj, devfilePath string) error {
	// validate if the GVRs represented by Kubernetes inlined components are supported by the underlying cluster
	_, err := service.ValidateResourceExist(o.client.GetKubeClient(), o.component, devfilePath)
	if err != nil {
		return err
	}

	labels := componentlabels.GetLabels(o.componentName, o.appName, true)
	u, err := service.GetK8sComponentAsUnstructured(o.component.Kubernetes, devfilePath, devfilefs.DefaultFs{})
	if err != nil {
		return err
	}

	log.Infof("\nDeploying Kubernetes %s: %s", u.GetKind(), u.GetName())
	isOperatorBackedService, err := service.PushKubernetesResource(o.client.GetKubeClient(), u, labels)
	if err != nil {
		return errors.Wrap(err, "failed to create service(s) associated with the component")
	}
	if isOperatorBackedService {
		log.Successf("Kubernetes resource %q on the cluster; refer %q to know how to link it to the component", strings.Join([]string{u.GetKind(), u.GetName()}, "/"), "odo link -h")

	}
	return nil
}
