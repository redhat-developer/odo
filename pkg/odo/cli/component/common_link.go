package component

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/secret"
	svc "github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
)

type commonLinkOptions struct {
	port             string
	secretName       string
	isTargetAService bool

	suppliedName  string
	operation     func(secretName, componentName, applicationName string) error
	operationName string

	*genericclioptions.Context
}

func newCommonLinkOptions() *commonLinkOptions {
	return &commonLinkOptions{}
}

// Complete completes LinkOptions after they've been created
func (o *commonLinkOptions) complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.operationName = name

	suppliedName := args[0]
	o.suppliedName = suppliedName
	o.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)

	svcExists, err := svc.SvcExists(o.Client, suppliedName, o.Application)
	if err != nil {
		// we consider this error to be non-terminal since it's entirely possible to use odo without the service catalog
		glog.V(4).Infof("Unable to determine if %s is a service. This most likely means the service catalog is not installed. Proceesing to only use components", suppliedName)
		svcExists = false
	}

	cmpExists, err := component.Exists(o.Client, suppliedName, o.Application)
	if err != nil {
		return fmt.Errorf("Unable to determine if component exists:\n%v", err)
	}

	if !cmpExists && !svcExists {
		return fmt.Errorf("Neither a service nor a component named %s could be located. Please create one of the two before attempting to use 'odo %s'", suppliedName, o.operationName)
	}

	o.isTargetAService = svcExists

	if svcExists {
		if cmpExists {
			glog.V(4).Infof("Both a service and component with name %s - assuming a(n) %s to the service is required", suppliedName, o.operationName)
		}

		o.secretName = suppliedName
	} else {
		secretName, err := secret.DetermineSecretName(o.Client, suppliedName, o.Application, o.port)
		if err != nil {
			return err
		}
		o.secretName = secretName
	}

	return nil
}

func (o *commonLinkOptions) validate(wait bool) (err error) {
	if o.isTargetAService {
		// if there is a ServiceBinding, then that means there is already a secret (or there will be soon)
		// which we can link to
		_, err = o.Client.GetServiceBinding(o.secretName, o.Project)
		if err != nil {
			return fmt.Errorf("The service was not created via Odo. Please delete the service and recreate it using 'odo service create %s'", o.secretName)
		}

		if wait {
			// we wait until the secret has been created on the OpenShift
			// this is done because the secret is only created after the Pod that runs the
			// service is in running state.
			// This can take a long time to occur if the image of the service has yet to be downloaded
			log.Progressf("Waiting for secret of service %s to come up", o.secretName)
			_, err = o.Client.WaitAndGetSecret(o.secretName, o.Project)
		} else {
			// we also need to check whether there is a secret with the same name as the service
			// the secret should have been created along with the secret
			_, err = o.Client.GetSecret(o.secretName, o.Project)
			if err != nil {
				return fmt.Errorf("The service %s created by 'odo service create' is being provisioned. You may have to wait a few seconds until OpenShift fully provisions it before executing 'odo %s'.", o.secretName, o.operationName)
			}
		}
	}

	return
}

func (o *commonLinkOptions) run() (err error) {
	linkType := "Component"
	if o.isTargetAService {
		linkType = "Service"
	}

	err = o.operation(o.secretName, o.Component(), o.Application)
	if err != nil {
		return err
	}

	log.Successf("%s %s has been successfully %sed from the component %s", linkType, o.suppliedName, o.operationName, o.Component())
	return
}
