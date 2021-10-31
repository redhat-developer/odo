package genericclioptions

import (
	"context"
	"fmt"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResolveAppFlag resolves the app from the flag
func ResolveAppFlag(command *cobra.Command) string {
	appFlag := FlagValueIfSet(command, ApplicationFlagName)
	if len(appFlag) > 0 {
		return appFlag
	}
	return defaultAppName
}

// resolveNamespace resolves namespace for devfile component
func (o *internalCxt) resolveNamespace(command *cobra.Command, configProvider localConfigProvider.LocalConfigProvider) error {
	var namespace string
	projectFlag := FlagValueIfSet(command, ProjectFlagName)
	if len(projectFlag) > 0 {
		// if namespace flag was set, check that the specified namespace exists and use it
		_, err := o.KClient.GetClient().CoreV1().Namespaces().Get(context.TODO(), projectFlag, metav1.GetOptions{})
		// do not error out when its odo delete -a, so that we let users delete the local config on missing namespace
		if command.HasParent() && command.Parent().Name() != "project" && !(command.Name() == "delete" && command.Flags().Changed("all")) {
			if err != nil {
				return err
			}
		}
		namespace = projectFlag
	} else {
		namespace = configProvider.GetNamespace()
		if namespace == "" {
			namespace = o.KClient.GetCurrentNamespace()
			if len(namespace) <= 0 {
				errFormat := "Could not get current namespace. Please create or set a namespace\n"
				err := checkProjectCreateOrDeleteOnlyOnInvalidNamespace(command, errFormat)
				if err != nil {
					return err
				}
			}
		}

		// check that the specified namespace exists
		_, err := o.KClient.GetClient().CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
		if err != nil {
			var errFormat string
			if kerrors.IsForbidden(err) {
				errFormat = "You are currently not logged in into the cluster. Use `odo login` first to perform any operation on cluster"
			} else {
				errFormat = fmt.Sprintf("You don't have permission to create or set namespace %q or the namespace doesn't exist. Please create or set a different namespace\n\t", namespace)
			}

			// errFormat := fmt.Sprint(e1, "%s project create|set <project_name>")
			err = checkProjectCreateOrDeleteOnlyOnInvalidNamespaceNoFmt(command, errFormat)
			if err != nil {
				return err
			}
		}
	}
	o.Client.Namespace = namespace
	o.Client.GetKubeClient().Namespace = namespace
	o.KClient.SetNamespace(namespace)
	o.project = namespace
	return nil
}

// resolveApp resolves the app
// If `--app` flag is used, return its value
// Or If app is set in envfile, return its value
// Or if createAppIfNeeded, returns the default app name
func resolveApp(command *cobra.Command, localConfiguration localConfigProvider.LocalConfigProvider, createAppIfNeeded bool) string {
	appFlag := FlagValueIfSet(command, ApplicationFlagName)
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
func resolveComponent(command *cobra.Command, localConfiguration localConfigProvider.LocalConfigProvider) string {
	cmpFlag := FlagValueIfSet(command, ComponentFlagName)
	if len(cmpFlag) > 0 {
		return cmpFlag
	}

	return localConfiguration.GetName()
}

func resolveProject(command *cobra.Command, localConfiguration localConfigProvider.LocalConfigProvider) string {
	projectFlag := FlagValueIfSet(command, ProjectFlagName)
	if projectFlag != "" {
		return projectFlag
	}
	return localConfiguration.GetNamespace()
}

// checkComponentExistsOrFail checks if the specified component exists with the given context and returns error if not.
// KClient, component and application should have been set before to call this method
func (o *internalCxt) checkComponentExistsOrFail() error {
	exists, err := component.Exists(o.KClient, o.component, o.application)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Component %v does not exist in application %s", o.component, o.application)
	}
	return nil
}
