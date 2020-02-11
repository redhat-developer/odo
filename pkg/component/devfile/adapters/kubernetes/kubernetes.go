package kubernetes

import (
	"github.com/golang/glog"
	"github.com/openshift/odo/pkg/component/devfile/adapters/common"
	devfileCommon "github.com/openshift/odo/pkg/devfile/versions/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/log"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

// Adapter maps Devfiles to Kubernetes resources and actions
type Adapter struct {
	common.DevfileAdapter
	Client kclient.Client
}

// Start creates Kubernetes resources that correspond to the devfile if they don't already exist
func (k Adapter) Start() error {

	err := k.Devfile.Data.Validate()
	if err != nil {
		return errors.Wrap(err, "The devfile is invalid")
	}

	componentName := k.DevfileComp.Name

	if k.componentExists(componentName) {
		log.Info("The component already exists")
		return nil
	}

	var containers []corev1.Container
	// Only components with aliases are considered because without an alias commands cannot reference them
	for _, comp := range k.Devfile.Data.GetAliasedComponents() {
		if comp.Type == devfileCommon.DevfileComponentTypeDockerimage {
			glog.V(3).Infof("Found component %v with alias %v\n", comp.Type, *comp.Alias)
			envVars := convertEnvs(comp.Env)
			resourceReqs := getResourceReqs(comp)
			container := kclient.GenerateContainer(*comp.Alias, *comp.Image, false, comp.Command, comp.Args, envVars, resourceReqs)
			containers = append(containers, *container)
		}
	}

	labels := map[string]string{
		"component": componentName,
	}

	podTemplateSpec := kclient.GeneratePodTemplateSpec(componentName, k.Client.Namespace, labels, containers)
	deploymentSpec := kclient.GenerateDeploymentSpec(*podTemplateSpec)

	glog.V(3).Infof("Successfully created component %v", componentName)
	glog.V(3).Infof("Creating deployment %v", deploymentSpec.Template.GetName())
	glog.V(3).Infof("The component name is %v", componentName)

	_, err = k.Client.CreateDeployment(componentName, *deploymentSpec)
	if err != nil {
		return err
	}
	log.Infof("Successfully created component %v", componentName)
	return nil
}

func (k Adapter) componentExists(name string) bool {
	_, err := k.Client.GetDeploymentByName(name)
	return err == nil
}
