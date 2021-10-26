package application

import (
	"github.com/openshift/odo/v2/pkg/kclient"
	"github.com/pkg/errors"
	"k8s.io/klog"

	applabels "github.com/openshift/odo/v2/pkg/application/labels"
	"github.com/openshift/odo/v2/pkg/component"
	"github.com/openshift/odo/v2/pkg/occlient"
	"github.com/openshift/odo/v2/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	appAPIVersion = "odo.dev/v1alpha1"
	appKind       = "Application"
	appList       = "List"
)

// List all applications names in current project by looking at `app` labels in deployments
func List(client *kclient.Client) ([]string, error) {
	if client == nil {
		return nil, nil
	}

	// Get all Deployments with the "app" label
	deploymentAppNames, err := client.GetDeploymentLabelValues(applabels.ApplicationLabel, applabels.ApplicationLabel)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list applications from deployments")
	}

	// Filter out any names, as there could be multiple components but within the same application
	return util.RemoveDuplicates(deploymentAppNames), nil
}

// Exists checks whether the given app exist or not in the list of applications
func Exists(app string, client *kclient.Client) (bool, error) {

	appList, err := List(client)

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
func Delete(client *kclient.Client, name string) error {
	klog.V(4).Infof("Deleting application %s", name)

	labels := applabels.GetLabels(name, false)

	// delete application from cluster
	err := client.Delete(labels, false)
	if err != nil {
		return errors.Wrapf(err, "unable to delete application %s", name)
	}

	return nil
}

// GetMachineReadableFormat returns resource information in machine readable format
func GetMachineReadableFormat(client *occlient.Client, appName string, projectName string) App {
	componentList, _ := component.GetComponentNames(client, appName)
	appDef := App{
		TypeMeta: metav1.TypeMeta{
			Kind:       appKind,
			APIVersion: appAPIVersion,
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
func GetMachineReadableFormatForList(apps []App) AppList {
	return AppList{
		TypeMeta: metav1.TypeMeta{
			Kind:       appList,
			APIVersion: appAPIVersion,
		},
		ListMeta: metav1.ListMeta{},
		Items:    apps,
	}
}
