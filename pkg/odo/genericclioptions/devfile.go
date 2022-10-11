package genericclioptions

import (
	"context"
	"fmt"

	"github.com/devfile/library/pkg/devfile/parser"
	dfutil "github.com/devfile/library/pkg/util"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/devfile/validate"
	fcontext "github.com/redhat-developer/odo/pkg/odo/commonflags/context"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	odoutil "github.com/redhat-developer/odo/pkg/util"
)

func getDevfileInfo(ctx context.Context) (
	devfilePath string,
	devfileObj *parser.DevfileObj,
	componentName string,
	err error,
) {
	workingDir := odocontext.GetWorkingDirectory(ctx)
	devfilePath = location.DevfileLocation(workingDir)
	isDevfile := odoutil.CheckPathExists(devfilePath)
	if isDevfile {
		devfilePath, err = dfutil.GetAbsPath(devfilePath)
		if err != nil {
			return "", nil, "", err
		}
		// Parse devfile and validate
		variables := fcontext.GetVariables(ctx)
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
	}

	componentName, err = component.GatherName(workingDir, devfileObj)
	if err != nil {
		return "", nil, "", err
	}

	return devfilePath, devfileObj, componentName, nil
}
