package genericclioptions

import (
	"fmt"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/util"
)

// resolveProjectAndNamespace resolve project in Context and namespace in Kubernetes and OpenShift clients
func (o *internalCxt) resolveProjectAndNamespace(cmdline cmdline.Cmdline) error {
	var namespace string
	projectFlag := cmdline.FlagValueIfSet(util.ProjectFlagName)
	if len(projectFlag) > 0 {
		// if namespace flag was set, check that the specified namespace exists and use it
		_, err := o.KClient.GetNamespaceNormal(projectFlag)
		if err != nil {
			return err
		}
		namespace = projectFlag
	} else {
		namespace = o.KClient.GetCurrentNamespace()
		if len(namespace) <= 0 {
			errFormat := "Could not get current namespace. Please create or set a namespace\n"
			err := checkProjectCreateOrDeleteOnlyOnInvalidNamespace(cmdline, errFormat)
			if err != nil {
				return err
			}
		}

		// check that the specified namespace exists
		_, err := o.KClient.GetNamespaceNormal(namespace)
		if err != nil {
			var errFormat string
			if kerrors.IsForbidden(err) {
				errFormat = "You are currently not logged in into the cluster. Use `odo login` first to perform any operation on cluster"
			} else {
				errFormat = fmt.Sprintf("You don't have permission to create or set namespace %q or the namespace doesn't exist. Please create or set a different namespace\n\t", namespace)
			}

			// errFormat := fmt.Sprint(e1, "%s project create|set <project_name>")
			err = checkProjectCreateOrDeleteOnlyOnInvalidNamespaceNoFmt(cmdline, errFormat)
			if err != nil {
				return err
			}
		}
	}
	o.KClient.SetNamespace(namespace)
	o.project = namespace
	return nil
}

// resolveApp resolves the app
// If `--app` flag is used, return its value
// Or If app is set in envfile, return its value
// Or if createAppIfNeeded, returns the default app name
func resolveApp(cmdline cmdline.Cmdline, localConfiguration localConfigProvider.LocalConfigProvider, createAppIfNeeded bool) string {
	appFlag := cmdline.FlagValueIfSet(util.ApplicationFlagName)
	if len(appFlag) > 0 {
		return appFlag
	}

	app := localConfiguration.GetApplication()
	if app == "" && createAppIfNeeded {
		app = defaultAppName
	}
	return app
}

// resolveComponent resolves component
// If `--component` flag is used, return its value
// Or Return the value in envfile
func resolveComponent(cmdline cmdline.Cmdline, localConfiguration localConfigProvider.LocalConfigProvider) string {
	cmpFlag := cmdline.FlagValueIfSet(util.ComponentFlagName)
	if len(cmpFlag) > 0 {
		return cmpFlag
	}

	return localConfiguration.GetName()
}

func resolveProject(cmdline cmdline.Cmdline) string {
	return cmdline.FlagValueIfSet(util.ProjectFlagName)
}

// checkComponentExistsOrFail checks if the specified component exists with the given context and returns error if not.
// KClient, component and application should have been set before to call this method
func (o *internalCxt) checkComponentExistsOrFail() error {
	exists, err := component.Exists(o.KClient, o.component, o.application)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("component %v does not exist in application %s", o.component, o.application)
	}
	return nil
}
