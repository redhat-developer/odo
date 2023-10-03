package backend

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/redhat-developer/odo/pkg/log"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/alizer"
	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/init/asker"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type AlizerBackend struct {
	askerClient  asker.Asker
	alizerClient alizer.Client
}

var _ InitBackend = (*AlizerBackend)(nil)

func NewAlizerBackend(askerClient asker.Asker, alizerClient alizer.Client) *AlizerBackend {
	return &AlizerBackend{
		askerClient:  askerClient,
		alizerClient: alizerClient,
	}
}

func (o *AlizerBackend) Validate(flags map[string]string, fs filesystem.Filesystem, dir string) error {
	return nil
}

func archList(archs []string) string {
	if len(archs) == 0 {
		return "all"
	}
	return strings.Join(archs, ", ")
}

// SelectDevfile calls the Alizer to detect the devfile and asks for confirmation to the user
func (o *AlizerBackend) SelectDevfile(ctx context.Context, flags map[string]string, fs filesystem.Filesystem, dir string) (*api.DetectionResult, error) {
	type result struct {
		location *api.DetectionResult
		err      error
	}
	var (
		resultChan = make(chan result)
	)
	go func() {
		location, err := func() (location *api.DetectionResult, err error) {
			spinner := log.Spinnerf("Determining a Devfile for the current directory")
			defer spinner.End(err == nil)
			detected, err := o.alizerClient.DetectFramework(ctx, dir)
			if err != nil {
				return nil, err
			}

			msg := fmt.Sprintf("Based on the files in the current directory odo detected\nSupported architectures: %s\nLanguage: %s\nProject type: %s",
				archList(detected.Architectures), detected.Type.Language, detected.Type.ProjectType)

			appPorts, err := o.alizerClient.DetectPorts(dir)
			if err != nil {
				return nil, err
			}
			spinner.End(true)
			appPortsAsString := make([]string, 0, len(appPorts))
			for _, p := range appPorts {
				appPortsAsString = append(appPortsAsString, strconv.Itoa(p))
			}
			if len(appPorts) > 0 {
				msg += fmt.Sprintf("\nApplication ports: %s", strings.Join(appPortsAsString, ", "))
			}

			fmt.Println(msg)
			fmt.Printf("The devfile \"%s:%s\" from the registry %q will be downloaded.\n", detected.Type.Name, detected.DefaultVersion, detected.Registry.Name)
			confirm, err := o.askerClient.AskCorrect()
			if err != nil {
				return nil, err
			}
			if !confirm {
				return nil, nil
			}
			return alizer.NewDetectionResult(detected.Type, detected.Registry, appPorts, detected.DefaultVersion, ""), nil
		}()
		resultChan <- result{
			location: location,
			err:      err,
		}
	}()

	select {
	case res := <-resultChan:
		return res.location, res.err
	case <-ctx.Done():
		return nil, fmt.Errorf("interrupted: %w", ctx.Err())
	}
}

func (o *AlizerBackend) SelectStarterProject(devfile parser.DevfileObj, flags map[string]string) (starter *v1alpha2.StarterProject, err error) {
	return nil, nil
}

func (o *AlizerBackend) PersonalizeName(devfile parser.DevfileObj, flags map[string]string) (string, error) {
	// Get the absolute path to the directory from the Devfile context
	path := devfile.Ctx.GetAbsPath()
	if path == "" {
		return "", fmt.Errorf("cannot determine the absolute path of the directory")
	}
	return o.alizerClient.DetectName(path)
}

func (o *AlizerBackend) PersonalizeDevfileConfig(devfile parser.DevfileObj) (parser.DevfileObj, error) {
	return devfile, nil
}

func (o *AlizerBackend) HandleApplicationPorts(devfileobj parser.DevfileObj, ports []int, flags map[string]string) (parser.DevfileObj, error) {
	return devfileobj, nil
}
