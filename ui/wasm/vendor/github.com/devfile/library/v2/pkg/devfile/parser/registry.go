//go:build !js

package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfileCtx "github.com/devfile/library/v2/pkg/devfile/parser/context"
	"github.com/devfile/library/v2/pkg/util"
	registryLibrary "github.com/devfile/registry-support/registry-library/library"

	"github.com/pkg/errors"
)

func parseFromRegistry(importReference v1.ImportReference, resolveCtx *resolutionContextTree, tool resolverTools) (d DevfileObj, err error) {
	id := importReference.Id
	registryURL := importReference.RegistryUrl
	destDir := path.Dir(d.Ctx.GetAbsPath())

	if registryURL != "" {
		devfileContent, err := getDevfileFromRegistry(id, registryURL, importReference.Version, tool.httpTimeout)
		if err != nil {
			return DevfileObj{}, err
		}
		d.Ctx, err = devfileCtx.NewByteContentDevfileCtx(devfileContent)
		if err != nil {
			return d, errors.Wrap(err, "failed to set devfile content from bytes")
		}
		newResolveCtx := resolveCtx.appendNode(importReference)

		err = getResourcesFromRegistry(id, registryURL, destDir)
		if err != nil {
			return DevfileObj{}, err
		}

		return populateAndParseDevfile(d, newResolveCtx, tool, true)

	} else if tool.registryURLs != nil {
		for _, registryURL := range tool.registryURLs {
			devfileContent, err := getDevfileFromRegistry(id, registryURL, importReference.Version, tool.httpTimeout)
			if devfileContent != nil && err == nil {
				d.Ctx, err = devfileCtx.NewByteContentDevfileCtx(devfileContent)
				if err != nil {
					return d, errors.Wrap(err, "failed to set devfile content from bytes")
				}
				importReference.RegistryUrl = registryURL
				newResolveCtx := resolveCtx.appendNode(importReference)

				err := getResourcesFromRegistry(id, registryURL, destDir)
				if err != nil {
					return DevfileObj{}, err
				}

				return populateAndParseDevfile(d, newResolveCtx, tool, true)
			}
		}
	} else {
		return DevfileObj{}, fmt.Errorf("failed to fetch from registry, registry URL is not provided")
	}

	return DevfileObj{}, fmt.Errorf("failed to get id: %s from registry URLs provided", id)
}

func getDevfileFromRegistry(id, registryURL, version string, httpTimeout *int) ([]byte, error) {
	if !strings.HasPrefix(registryURL, "http://") && !strings.HasPrefix(registryURL, "https://") {
		return nil, fmt.Errorf("the provided registryURL: %s is not a valid URL", registryURL)
	}
	param := util.HTTPRequestParams{
		URL: fmt.Sprintf("%s/devfiles/%s/%s", registryURL, id, version),
	}

	param.Timeout = httpTimeout
	//suppress telemetry for parent uri references
	param.TelemetryClientName = util.TelemetryIndirectDevfileCall
	return util.HTTPGetRequest(param, 0)
}

func getResourcesFromRegistry(id, registryURL, destDir string) error {
	stackDir, err := ioutil.TempDir(os.TempDir(), fmt.Sprintf("registry-resources-%s", id))
	if err != nil {
		return fmt.Errorf("failed to create dir: %s, error: %v", stackDir, err)
	}
	defer os.RemoveAll(stackDir)
	//suppress telemetry for downloading resources from parent reference
	err = registryLibrary.PullStackFromRegistry(registryURL, id, stackDir, registryLibrary.RegistryOptions{Telemetry: registryLibrary.TelemetryData{Client: util.TelemetryIndirectDevfileCall}})
	if err != nil {
		return fmt.Errorf("failed to pull stack from registry %s", registryURL)
	}

	err = util.CopyAllDirFiles(stackDir, destDir)
	if err != nil {
		return err
	}

	return nil
}
