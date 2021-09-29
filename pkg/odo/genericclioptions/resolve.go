package genericclioptions

import (
	"context"
	"fmt"

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
	return DefaultAppName
}

// resolveProject resolves project
func (o *internalCxt) resolveProject(localConfiguration localConfigProvider.LocalConfigProvider) error {
	var namespace string
	command := o.command
	projectFlag := FlagValueIfSet(command, ProjectFlagName)
	if len(projectFlag) > 0 {
		// if project flag was set, check that the specified project exists and use it
		project, err := o.Client.GetProject(projectFlag)
		if err != nil || project == nil {
			return err
		}
		namespace = projectFlag
	} else {
		namespace = localConfiguration.GetNamespace()
		if namespace == "" {
			namespace = o.Client.Namespace
			if len(namespace) <= 0 {
				errFormat := "Could not get current project. Please create or set a project\n\t%s project create|set <project_name>"
				err := checkProjectCreateOrDeleteOnlyOnInvalidNamespace(command, errFormat)
				if err != nil {
					return err
				}
			}
		}

		// check that the specified project exists
		_, err := o.Client.GetProject(namespace)
		if err != nil {
			e1 := fmt.Sprintf("You don't have permission to create or set project '%s' or the project doesn't exist. Please create or set a different project\n\t", namespace)
			errFormat := fmt.Sprint(e1, "%s project create|set <project_name>")
			err = checkProjectCreateOrDeleteOnlyOnInvalidNamespace(command, errFormat)
			if err != nil {
				return err
			}
		}
	}
	o.Client.GetKubeClient().Namespace = namespace
	o.Client.Namespace = namespace
	o.Project = namespace
	if o.KClient != nil {
		o.KClient.SetNamespace(namespace)
	}
	return nil
}

// resolveNamespace resolves namespace for devfile component
func (o *internalCxt) resolveNamespace(configProvider localConfigProvider.LocalConfigProvider) error {
	var namespace string
	command := o.command
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
			errFormat := fmt.Sprintf("You don't have permission to create or set namespace '%s' or the namespace doesn't exist. Please create or set a different namespace\n\t", namespace)
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
	o.Project = namespace

	return nil
}

// resolveApp resolves the app
func (o *internalCxt) resolveApp(createAppIfNeeded bool, localConfiguration localConfigProvider.LocalConfigProvider) {
	var app string
	command := o.command
	appFlag := FlagValueIfSet(command, ApplicationFlagName)
	if len(appFlag) > 0 {
		app = appFlag
	} else {
		app = localConfiguration.GetApplication()
		if app == "" {
			if createAppIfNeeded {
				app = DefaultAppName
			}
		}
	}
	o.Application = app
}

// resolveComponent resolves component
func (o *internalCxt) resolveAndSetComponent(command *cobra.Command, localConfiguration localConfigProvider.LocalConfigProvider) (string, error) {
	var cmp string
	cmpFlag := FlagValueIfSet(command, ComponentFlagName)
	if len(cmpFlag) == 0 {
		// retrieve the current component if it exists if we didn't set the component flag
		cmp = localConfiguration.GetName()
	} else {
		// if flag is set, check that the specified component exists
		err := o.checkComponentExistsOrFail(cmpFlag)
		if err != nil {
			return "", err
		}
		cmp = cmpFlag
	}
	o.cmp = cmp
	return cmp, nil
}
