package genericclioptions

import (
	"fmt"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
)

// resolveProjectAndNamespace resolve project in Context and namespace in Kubernetes and OpenShift clients
func (o *internalCxt) resolveProjectAndNamespace(cmdline cmdline.Cmdline) error {
	var namespace string
	projectFlag := ""
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
