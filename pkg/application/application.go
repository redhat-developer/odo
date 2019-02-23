package application

import (
	"github.com/golang/glog"
	"github.com/pkg/errors"

	applabels "github.com/openshift/odo/pkg/application/labels"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/preference"
	"github.com/openshift/odo/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	appPrefixMaxLen   = 12
	appNameMaxRetries = 3
	appAPIVersion     = "odo.openshift.io/v1alpha1"
	appKind           = "app"
	appList           = "List"
)

// GetDefaultAppName returns randomly generated application name with unique configurable prefix suffixed by a randomly generated string which can be used as a default name in case the user doesn't provide a name.
func GetDefaultAppName() (string, error) {
	var appName string

	// Get the desired app name prefix from odo config
	cfg, err := preference.New()
	if err != nil {
		return "", errors.Wrap(err, "unable to fetch config")
	}

	// If there's no prefix in config file or it is equal to $DIR, use safe default which is the name of current directory
	if cfg.OdoSettings.NamePrefix == nil || *cfg.OdoSettings.NamePrefix == "" {
		prefix, err := component.GetComponentDir("", config.NONE)
		if err != nil {
			return "", errors.Wrap(err, "unable to generate random app name")
		}
		appName, err = util.GetRandomName(prefix, appPrefixMaxLen, []string{}, appNameMaxRetries)
		if err != nil {
			return "", errors.Wrap(err, "unable to generate random app name")
		}
	} else {
		appName, err = util.GetRandomName(*cfg.OdoSettings.NamePrefix, appPrefixMaxLen, []string{}, appNameMaxRetries)
	}
	if err != nil {
		return "", errors.Wrap(err, "unable to generate random app name")
	}
	return util.GetDNS1123Name(appName), nil
}

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
