package genericclioptions

import (
	"fmt"
	
	"github.com/devfile/library/v2/pkg/devfile/parser"
	dfutil "github.com/devfile/library/v2/pkg/util"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/devfile/validate"
	odoutil "github.com/redhat-developer/odo/pkg/util"
)

func getDevfileInfo(workingDir string, variables map[string]string) (
	devfilePath string,
	devfileObj *parser.DevfileObj,
	componentName string,
	err error,
) {
	devfilePath = location.DevfileLocation(workingDir)
	isDevfile := odoutil.CheckPathExists(devfilePath)
	if isDevfile {
		devfilePath, err = dfutil.GetAbsPath(devfilePath)
		if err != nil {
			return "", nil, "", err
		}
		// Parse devfile and validate
		var devObj parser.DevfileObj
		devObj, err = devfile.ParseAndValidateFromFileWithVariables(devfilePath, variables)
		if err != nil {
			return "", nil, "", fmt.Errorf("failed to parse the devfile %s: %w", devfilePath, err)
		}
		devfileObj = &devObj
		err = validate.ValidateDevfileData(devfileObj.Data)
		if err != nil {
			return "", nil, "", err
		}

		componentName, err = component.GatherName(workingDir, devfileObj)
		if err != nil {
			return "", nil, "", err
		}
	}

	return devfilePath, devfileObj, componentName, nil
}
