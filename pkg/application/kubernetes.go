package application

import (
	"github.com/pkg/errors"

	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

type kubernetesClient struct {
	client kclient.ClientInterface
}

func NewClient(client kclient.ClientInterface) Client {
	return kubernetesClient{
		client: client,
	}
}

// List all applications names in current project by looking at `app` labels in deployments
func (o kubernetesClient) List() ([]string, error) {
	if o.client == nil {
		return nil, nil
	}

	// Get all Deployments with the "app" label
	deploymentAppNames, err := o.client.GetDeploymentLabelValues(applabels.ApplicationLabel, applabels.ApplicationLabel)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list applications from deployments")
	}

	// Filter out any names, as there could be multiple components but within the same application
	return util.RemoveDuplicates(deploymentAppNames), nil
}

// Exists checks whether the given app exist or not in the list of applications
func (o kubernetesClient) Exists(app string) (bool, error) {

	appList, err := o.List()

	if err != nil {
		return false, err
	}
	for _, appName := range appList {
		if appName == app {
			return true, nil
		}
	}
	return false, nil
}

// Delete the given application by deleting deployments and services instances belonging to this application
func (o kubernetesClient) Delete(name string) error {
	klog.V(4).Infof("Deleting application %q", name)

	labels := applabels.GetLabels(name, false)

	// delete application from cluster
	err := o.client.Delete(labels, false)
	if err != nil {
		return errors.Wrapf(err, "unable to delete application %s", name)
	}

	return nil
}

// ComponentList returns the list of components for an application
func (o kubernetesClient) ComponentList(name string) ([]component.Component, error) {
	selector := applabels.GetSelector(name)
	compClient := component.NewClient(o.client)
	componentList, err := compClient.List(selector)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get Component list")
	}
	return componentList.Items, nil
}

// GetMachineReadableFormat returns resource information in machine readable format
func (o kubernetesClient) GetMachineReadableFormat(appName string, projectName string) App {
	compClient := component.NewClient(o.client)
	componentList, _ := compClient.GetComponentNames(appName)
	appDef := App{
		TypeMeta: metav1.TypeMeta{
			Kind:       appKind,
			APIVersion: machineoutput.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: projectName,
		},
		Spec: AppSpec{
			Components: componentList,
		},
	}
	return appDef
}

// GetMachineReadableFormatForList returns application list in machine readable format
func (o kubernetesClient) GetMachineReadableFormatForList(apps []App) AppList {
	return AppList{
		TypeMeta: metav1.TypeMeta{
			Kind:       appList,
			APIVersion: machineoutput.APIVersion,
		},
		ListMeta: metav1.ListMeta{},
		Items:    apps,
	}
}
