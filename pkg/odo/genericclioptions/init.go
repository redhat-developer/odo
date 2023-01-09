package genericclioptions

import (
	"context"
	"path/filepath"

	"github.com/devfile/library/v2/pkg/devfile/parser"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/files"
	"github.com/redhat-developer/odo/pkg/odo/cli/messages"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"
	"github.com/redhat-developer/odo/pkg/version"
)

// runPreInit executes the Init command before running the main command
func runPreInit(ctx context.Context, workingDir string, deps *clientset.Clientset, cmdline cmdline.Cmdline, msg string) error {
	isEmptyDir, err := location.DirIsEmpty(deps.FS, workingDir)
	if err != nil {
		return err
	}
	if isEmptyDir {
		return NewNoDevfileError(workingDir)
	}

	initFlags := deps.InitClient.GetFlags(cmdline.GetFlags())

	err = deps.InitClient.InitDevfile(ctx, initFlags, workingDir,
		func(interactiveMode bool) {
			scontext.SetInteractive(cmdline.Context(), interactiveMode)
			if interactiveMode {
				log.Title(msg, messages.SourceCodeDetected, "odo version: "+version.VERSION)
				log.Info("\n" + messages.InteractiveModeEnabled)
			}
		},
		func(newDevfileObj parser.DevfileObj) error {
			dErr := newDevfileObj.WriteYamlDevfile()
			if dErr != nil {
				return dErr
			}
			dErr = files.ReportLocalFileGeneratedByOdo(deps.FS, workingDir, filepath.Base(newDevfileObj.Ctx.GetAbsPath()))
			if dErr != nil {
				klog.V(4).Infof("error trying to report local file generated: %v", dErr)
			}
			return nil
		})
	if err != nil {
		return err
	}
	return nil
}
