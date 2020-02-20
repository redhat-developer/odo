package component

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"

	adapterCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/versions/common"
	"github.com/openshift/odo/pkg/kclient"
)

// Adapter is a component adapter implementation for Kubernetes
type Adapter struct {
	Client  kclient.Client
	Adapter adapterCommon.AdapterMetadata
}

// New instantiantes a component adapter
func New(commonAdapter adapterCommon.AdapterMetadata, client kclient.Client) Adapter {
	return Adapter{
		Client:  client,
		Adapter: commonAdapter,
	}
}

// Start updates the component if a matching component exists or creates one if it doesn't exist
func (a Adapter) Start() (err error) {
	componentName := a.Adapter.ComponentName

	var containers []corev1.Container
	// Only components with aliases are considered because without an alias commands cannot reference them
	for _, comp := range a.Adapter.Devfile.Data.GetAliasedComponents() {
		if comp.Type == common.DevfileComponentTypeDockerimage {
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

	if len(containers) == 0 {
		return fmt.Errorf("No valid components found in the devfile")
	}

	objectMeta := kclient.CreateObjectMeta(componentName, a.Client.Namespace, labels, nil)
	podTemplateSpec := kclient.GeneratePodTemplateSpec(objectMeta, a.Client.Namespace, containers)
	deploymentSpec := kclient.GenerateDeploymentSpec(*podTemplateSpec)

	glog.V(3).Infof("Creating deployment %v", deploymentSpec.Template.GetName())
	glog.V(3).Infof("The component name is %v", componentName)

	if componentExists(a.Client, componentName) {
		glog.V(3).Info("The component already exists, attempting to update it")
		_, err = a.Client.UpdateDeployment(componentName, *deploymentSpec)
		if err != nil {
			return err
		}
		glog.V(3).Infof("Successfully updated component %v", componentName)
	} else {
		_, err = a.Client.CreateDeployment(*deploymentSpec)
		if err != nil {
			return err
		}
		glog.V(3).Infof("Successfully created component %v", componentName)
	}

	podSelector := fmt.Sprintf("component=%s", componentName)
	watchOptions := metav1.ListOptions{
		LabelSelector: podSelector,
	}

	_, err = a.Client.WaitAndGetPod(watchOptions, corev1.PodRunning, "Waiting for component to start")
	return err
}

func componentExists(client kclient.Client, name string) bool {
	_, err := client.GetDeploymentByName(name)
	return err == nil
}
