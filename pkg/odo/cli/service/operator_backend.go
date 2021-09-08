/*
	This file contains code for various service backends supported by odo. Different backends have different logics for
	Complete, Validate and Run functions. These are covered in this file.
*/
package service

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ghodss/yaml"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	svc "github.com/openshift/odo/pkg/service"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

		// if the service name is blank then we set it to custom resource name
		if o.ServiceName == "" {
			o.ServiceName = strings.ToLower(b.CustomResource)
		}

		if len(o.parameters) != 0 {
			builtCRD, err := b.buildCRDfromParams(o, csv)
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
		}

		d.SetServiceName(o.ServiceName)

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

		if log.IsJSON() {
			svcFullName := strings.Join([]string{b.CustomResource, o.ServiceName}, "/")
			svc := NewServiceItem(svcFullName)
			svc.Manifest = b.CustomResourceDefinition
			svc.InDevfile = true
			machineoutput.OutputSuccess(svc)
		}
	}
	s.End(true)
	return
}

// ServiceDefined returns true if the service is defined in the devfile
func (b *OperatorBackend) ServiceDefined(ctx *genericclioptions.Context, name string) (bool, error) {
	_, instanceName, err := svc.SplitServiceKindName(name)
	if err != nil {
		return false, err
	}
	return svc.IsDefined(instanceName, ctx.EnvSpecificInfo.GetDevfileObj())
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

func (b *OperatorBackend) buildCRDfromParams(o *CreateOptions, csv olm.ClusterServiceVersion) (map[string]interface{}, error) {
	hasCR, cr := o.KClient.CheckCustomResourceInCSV(b.CustomResource, &csv)
	if !hasCR {
		return nil, fmt.Errorf("the %q resource doesn't exist in specified %q operator", b.CustomResource, o.ServiceType)
	}

	return svc.BuildCRDFromParams(cr, o.ParametersMap)
}

func (b *OperatorBackend) DescribeService(o *DescribeOptions, serviceName, app string) error {

	clusterList, _, err := svc.ListOperatorServices(o.KClient)
	if err != nil {
		return err
	}
	var clusterFound *unstructured.Unstructured
	for i, clusterInstance := range clusterList {
		fullName := strings.Join([]string{clusterInstance.GetKind(), clusterInstance.GetName()}, "/")
		if fullName == serviceName {
			clusterFound = &clusterList[i]
			break
		}
	}

	item := NewServiceItem(serviceName)
	item.Deployed = clusterFound != nil
	if item.Deployed {
		item.Manifest = clusterFound.Object
	} else {
		devfileList, err := svc.ListDevfileServices(o.EnvSpecificInfo.GetDevfileObj())
		if err != nil {
			return err
		}
		devfileService, inDevfile := devfileList[serviceName]

		item.InDevfile = inDevfile
		if item.InDevfile {
			item.Manifest = devfileService.Object
		}
	}

	if log.IsJSON() {
		machineoutput.OutputSuccess(item)
		return nil
	}

	return HumanReadableOutput(os.Stdout, item)
}

// HumanReadableOutput outputs the list of projects in a human readable format
func HumanReadableOutput(w io.Writer, item *serviceItem) error {
	fmt.Fprintf(w, "Version: %s\n", item.Manifest["apiVersion"])
	fmt.Fprintf(w, "Kind: %s\n", item.Manifest["kind"])
	metadata, ok := item.Manifest["metadata"].(map[string]interface{})
	if !ok {
		return errors.New("unable to get name from manifest")
	}
	fmt.Fprintf(w, "Name: %s\n", metadata["name"])
	spec, ok := item.Manifest["spec"].(map[string]interface{})
	if !ok {
		return errors.New("unable to get specifications from manifest")
	}

	fmt.Fprintln(w, "Parameters:")

	wr := tabwriter.NewWriter(w, 5, 2, 3, ' ', tabwriter.TabIndent)
	fmt.Fprint(wr, "NAME", "\t", "VALUE", "\n")
	displayParameters(wr, spec, "")
	wr.Flush()
	return nil
}

func displayParameters(wr *tabwriter.Writer, spec map[string]interface{}, prefix string) {
	keys := make([]string, len(spec))
	i := 0
	for key := range spec {
		keys[i] = key
		i++
	}

	for _, k := range keys {
		v := spec[k]
		switch val := v.(type) {
		case map[string]interface{}:
			displayParameters(wr, val, prefix+k+".")
		default:
			fmt.Fprintf(wr, "%s%s\t%v\n", prefix, k, val)
		}
	}
}
