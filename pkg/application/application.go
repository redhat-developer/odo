package application

import (
	"github.com/golang/glog"
	"github.com/pkg/errors"

	applabels "github.com/openshift/odo/pkg/application/labels"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/occlient"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	appPrefixMaxLen   = 12
	appNameMaxRetries = 3
	appAPIVersion     = "odo.openshift.io/v1alpha1"
	appKind           = "app"
	appList           = "List"
)

// List all applications in current project
func List(client *occlient.Client) ([]string, error) {
	return ListInProject(client)
}

// ListInProject lists all applications in given project by Querying the cluster
func ListInProject(client *occlient.Client) ([]string, error) {

	// Get applications from cluster
	appNames, err := client.GetLabelValues(applabels.ApplicationLabel, applabels.ApplicationLabel)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list applications")
	}

	return appNames, nil
}

// Exists checks whether the given app exist or not
func Exists(app string, client *occlient.Client) (bool, error) {

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

// Delete deletes the given application
func Delete(client *occlient.Client, name string) error {
	glog.V(4).Infof("Deleting application %s", name)

	labels := applabels.GetLabels(name, false)

	// delete application from cluster
	err := client.Delete(labels)
	if err != nil {
		return errors.Wrapf(err, "unable to delete application %s", name)
	}

	return nil
}

// GetMachineReadableFormat returns resource information in machine readable format
func GetMachineReadableFormat(client *occlient.Client, appName string, projectName string) App {
	componentList, _ := component.List(client, appName)
	var compList []string
	for _, comp := range componentList.Items {
		compList = append(compList, comp.Name)
	}
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
			Components: compList,
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
