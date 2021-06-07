/*
	This file contains code for various service backends supported by odo. Different backends have different logics for
	Complete, Validate and Run functions. These are covered in this file.
*/
package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/service"
	svc "github.com/openshift/odo/pkg/service"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// This CompleteServiceCreate contains logic to complete the "odo service create" call for the case of Operator backend
func (b *OperatorBackend) CompleteServiceCreate(o *CreateOptions, cmd *cobra.Command, args []string) (err error) {
	// since interactive mode is not supported for Operators yet, set it to false
	o.interactive = false

	// if user has just used "odo service create", simply return
	if o.fromFile == "" && len(args) == 0 {
		return
	}

	// if user wants to create service from file and use a name given on CLI
	if o.fromFile != "" {
		if len(args) == 1 {
			o.ServiceName = args[0]
		}
		return
	}

	// split the name provided on CLI and populate servicetype & customresource
	o.ServiceType, b.CustomResource, err = svc.SplitServiceKindName(args[0])
	if err != nil {
		return fmt.Errorf("invalid service name, use the format <operator-type>/<crd-name>")
	}

	// if two args are given, first is service type and second one is service name
	if len(args) == 2 {
		o.ServiceName = args[1]
	}

	return nil
}

func (b *OperatorBackend) ValidateServiceCreate(o *CreateOptions) (err error) {
	d := svc.NewDynamicCRD()
	// if the user wants to create service from a file, we check for
	// existence of file and validate if the requested operator and CR
	// exist on the cluster
	if o.fromFile != "" {
		if _, err := os.Stat(o.fromFile); err != nil {
			return errors.Wrap(err, "unable to find specified file")
		}

		// Parse the file to find Operator and CR info
		fileContents, err := ioutil.ReadFile(o.fromFile)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(fileContents, &d.OriginalCRD)
		if err != nil {
			return err
		}

		// Check if the operator and the CR exist on cluster
		var csv olm.ClusterServiceVersion
		b.CustomResource, csv, err = svc.GetCSV(o.KClient, d.OriginalCRD)
		if err != nil {
			return err
		}

		// all is well, let's populate the fields required for creating operator backed service
		b.group, b.version, b.resource, err = svc.GetGVRFromOperator(csv, b.CustomResource)
		if err != nil {
			return err
		}

		err = d.ValidateMetadataInCRD()
		if err != nil {
			return err
		}

		if o.ServiceName != "" && !o.DryRun {
			// First check if service with provided name already exists
			svcFullName := strings.Join([]string{b.CustomResource, o.ServiceName}, "/")
			exists, err := svc.OperatorSvcExists(o.KClient, svcFullName)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("service %q already exists; please provide a different name or delete the existing service first", svcFullName)
			}

			d.SetServiceName(o.ServiceName)
		} else {
			o.ServiceName, err = d.GetServiceNameFromCRD()
			if err != nil {
				return err
			}
		}

		// CRD is valid. We can use it further to create a service from it.
		b.CustomResourceDefinition = d.OriginalCRD

		return nil
	} else if b.CustomResource != "" {
		// make sure that CSV of the specified ServiceType exists
		csv, err := o.KClient.GetClusterServiceVersion(o.ServiceType)
		if err != nil {
			// error only occurs when OperatorHub is not installed.
			// k8s does't have it installed by default but OCP does
			return err
		}
		b.group, b.version, b.resource, err = svc.GetGVRFromOperator(csv, b.CustomResource)
		if err != nil {
			return err
		}

		if len(o.parameters) != 0 {
			var cr *olm.CRDDescription
			hasCR := false
			CRs := o.KClient.GetCustomResourcesFromCSV(&csv)
			for _, custRes := range *CRs {
				c := custRes
				if c.Kind == b.CustomResource {
					cr = &c
					hasCR = true
					break
				}
			}
			if !hasCR {
				return fmt.Errorf("the %q resource doesn't exist in specified %q operator", b.CustomResource, o.ServiceType)
			}

			crBuilder := service.NewCRBuilder(cr)
			var errorStrs []string

			for key, value := range o.ParametersMap {
				err := crBuilder.SetAndValidate(key, value)
				if err != nil {
					errorStrs = append(errorStrs, err.Error())
				}
			}

			if len(errorStrs) > 0 {
				return errors.New(strings.Join(errorStrs, "\n"))
			}

			builtCRD, err := crBuilder.Map()
			if err != nil {
				return err
			}

			d.OriginalCRD = builtCRD
		} else {
			almExample, err := svc.GetAlmExample(csv, b.CustomResource, o.ServiceType)
			if err != nil {
				return err
			}

			d.OriginalCRD = almExample
		}

		if o.ServiceName != "" && !o.DryRun {
			// First check if service with provided name already exists
			svcFullName := strings.Join([]string{b.CustomResource, o.ServiceName}, "/")
			exists, err := svc.OperatorSvcExists(o.KClient, svcFullName)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("service %q already exists; please provide a different name or delete the existing service first", svcFullName)
			}

			d.SetServiceName(o.ServiceName)
		}

		err = d.ValidateMetadataInCRD()
		if err != nil {
			return err
		}

		// CRD is valid. We can use it further to create a service from it.
		b.CustomResourceDefinition = d.OriginalCRD

		if o.ServiceName == "" {
			o.ServiceName, err = d.GetServiceNameFromCRD()
			if err != nil {
				return err
			}
		}

		return nil
	} else {
		// This block is executed only when user has neither provided a
		// file nor a valid `odo service create <operator-name>` to start
		// the service from an Operator. So we raise an error because the
		// correct way is to execute:
		// `odo service create <operator-name>/<crd-name>`

		return fmt.Errorf("please use a valid command to start an Operator backed service; desired format: %q", "odo service create <operator-name>/<crd-name>")
	}
}

func (b *OperatorBackend) RunServiceCreate(o *CreateOptions) (err error) {
	s := &log.Status{}

	// if cluster has resources of type CSV and o.CustomResource is not
	// empty, we're expected to create an Operator backed service
	if o.DryRun {
		// if it's dry run, only print the alm-example (o.CustomResourceDefinition) and exit
		jsonCR, err := json.MarshalIndent(b.CustomResourceDefinition, "", "  ")
		if err != nil {
			return err
		}

		// convert json to yaml
		yamlCR, err := yaml.JSONToYAML(jsonCR)
		if err != nil {
			return err
		}

		log.Info(string(yamlCR))

		return nil
	} else {
		crdYaml, err := yaml.Marshal(b.CustomResourceDefinition)
		if err != nil {
			return err
		}

		err = svc.AddKubernetesComponentToDevfile(string(crdYaml), o.ServiceName, o.EnvSpecificInfo.GetDevfileObj())
		if err != nil {
			return err
		}
	}
	s.End(true)

	return
}

// ServiceDefined returns true if the service is defined in the devfile
func (b *OperatorBackend) ServiceDefined(o *DeleteOptions) (bool, error) {
	_, instanceName, err := svc.SplitServiceKindName(o.serviceName)
	if err != nil {
		return false, err
	}

	return svc.IsDefined(instanceName, o.EnvSpecificInfo.GetDevfileObj())
}

func (b *OperatorBackend) ServiceExists(o *DeleteOptions) (bool, error) {
	return svc.OperatorSvcExists(o.KClient, o.serviceName)
}

func (b *OperatorBackend) DeleteService(o *DeleteOptions, name string, application string) error {
	// "name" is of the form CR-Name/Instance-Name so we split it
	_, instanceName, err := svc.SplitServiceKindName(name)
	if err != nil {
		return err
	}

	err = svc.DeleteKubernetesComponentFromDevfile(instanceName, o.EnvSpecificInfo.GetDevfileObj())
	if err != nil {
		return errors.Wrap(err, "failed to delete service from the devfile")
	}

	return nil
}
