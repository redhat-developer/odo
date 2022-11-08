package genericclioptions

import (
	"context"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/log"
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
			return newDevfileObj.WriteYamlDevfile()
		})
	if err != nil {
		return err
	}
	return nil
}
