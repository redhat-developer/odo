package describe

import (
	"fmt"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/service"
	svc "github.com/openshift/odo/pkg/service"
	olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type operatorBackend struct {
}

func NewOperatorBackend() *operatorBackend {
	return &operatorBackend{}
}

func (ohb *operatorBackend) CompleteDescribeService(dso *DescribeServiceOptions, args []string) error {
	oprType, CR, err := service.SplitServiceKindName(args[0])
	if err != nil {
		return err
	}
	// we check if the cluster supports ClusterServiceVersion or not.
	isCSVSupported, err := svc.IsCSVSupported()
	if err != nil {
		// if there is an error checking it, we return the error.
		return err
	}
	// if its not supported then we return an error
	if !isCSVSupported {
		return errors.New("it seems the cluster doesn't support Operators. Please install OLM and try again")
	}
	dso.OperatorType = oprType
	dso.CustomResource = CR
	return nil
}

func (ohb *operatorBackend) ValidateDescribeService(dso *DescribeServiceOptions) error {
	var err error
	if dso.OperatorType == "" || dso.CustomResource == "" {
		return fmt.Errorf("invalid service name, use the format <operator-type>/<crd-name>")
	}
	// make sure that CSV of the specified OperatorType exists
	dso.CSV, err = dso.KClient.GetClusterServiceVersion(dso.OperatorType)
	if err != nil {
		// error only occurs when OperatorHub is not installed.
		// k8s does't have it installed by default but OCP does
		return err
	}

	// Get the specific CR that matches "kind"
	crs := dso.KClient.GetCustomResourcesFromCSV(&dso.CSV)

	var cr *olm.CRDDescription
	hasCR := false
	for _, custRes := range *crs {
		c := custRes
		if c.Kind == dso.CustomResource {
			cr = &c
			hasCR = true
			break
		}
	}
	if !hasCR {
		return fmt.Errorf("the %q resource doesn't exist in specified %q operator", dso.CustomResource, dso.OperatorType)
	}

	dso.CR = cr
	return nil

}

func (ohb *operatorBackend) RunDescribeService(dso *DescribeServiceOptions) error {
	if log.IsJSON() {
		machineoutput.OutputSuccess(svc.ConvertCRDToJSONRepr(dso.CR))
		return nil
	}
	output, err := yaml.Marshal(svc.ConvertCRDToRepr(dso.CR))
	if err != nil {
		return err
	}

	fmt.Print(string(output))
	return nil
}
