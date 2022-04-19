package api

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/redhat-developer/odo/pkg/libdevfile"
)

func GetDevfileData(devfileObj parser.DevfileObj) DevfileData {
	return DevfileData{
		Devfile:              devfileObj.Data,
		SupportedOdoFeatures: getSupportedOdoFeatures(devfileObj.Data),
	}
}

func getSupportedOdoFeatures(devfileData data.DevfileData) SupportedOdoFeatures {
	return SupportedOdoFeatures{
		Dev:    libdevfile.HasRunCommand(devfileData),
		Deploy: libdevfile.HasDeployCommand(devfileData),
		Debug:  libdevfile.HasDebugCommand(devfileData),
	}
}
