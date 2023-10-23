package alizer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/apis/recognizer"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/registry"
	"github.com/redhat-developer/odo/pkg/util"
)

type Alizer struct {
	registryClient registry.Client
}

var _ Client = (*Alizer)(nil)

func NewAlizerClient(registryClient registry.Client) *Alizer {
	return &Alizer{
		registryClient: registryClient,
	}
}

// DetectFramework uses the alizer library in order to detect the devfile
// to use depending on the files in the path
func (o *Alizer) DetectFramework(ctx context.Context, path string) (DetectedFramework, error) {
	types := []model.DevfileType{}
	components, err := o.registryClient.ListDevfileStacks(ctx, "", "", "", false, false)
	if err != nil {
		return DetectedFramework{}, err
	}
	for _, component := range components.Items {
		types = append(types, model.DevfileType{
			Name:        component.Name,
			Language:    component.Language,
			ProjectType: component.ProjectType,
			Tags:        component.Tags,
		})
	}
	typ, err := recognizer.SelectDevFileFromTypes(path, types)
	if err != nil {
		return DetectedFramework{}, err
	}
	// Get the default stack version that will be downloaded
	var defaultVersion string
	for _, version := range components.Items[typ].Versions {
		if version.IsDefault {
			defaultVersion = version.Version
		}
	}
	return DetectedFramework{
		Type:           types[typ],
		DefaultVersion: defaultVersion,
		Registry:       components.Items[typ].Registry,
		Architectures:  components.Items[typ].Architectures,
	}, nil
}

// DetectName retrieves the name of the project (if available).
// If source code is detected:
// 1. Detect the name (pom.xml for java, package.json for nodejs, etc.)
// 2. If unable to detect the name, use the directory name
//
// If no source is detected:
// 1. Use the directory name
//
// Last step. Sanitize the name so it's valid for a component name
//
// Use:
// import "github.com/redhat-developer/alizer/pkg/apis/recognizer"
// components, err := recognizer.DetectComponents("./")
//
// In order to detect the name, the name will first try to find out the name based on the program (pom.xml, etc.) but then if not, it will use the dir name.
func (o *Alizer) DetectName(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}

	// Check if the path exists using os.Stat
	dir, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	// Check to see if the path is a directory, and fail if it is not
	if !dir.IsDir() {
		return "", fmt.Errorf("alizer DetectName %q path is not a directory", path)
	}

	// Step 1.
	// Get the name of the directory from the devfile absolute path
	// Use that path with Alizer to get the name of the project,
	// if unable to find the name, we will use the directory name
	components, err := recognizer.DetectComponents(path)
	if err != nil {
		return "", err
	}
	klog.V(4).Infof("Found components: %v", components)

	// Take the first name that is found
	var detectedName string
	if len(components) > 0 {
		detectedName = components[0].Name
	}

	// Step 2. If unable to detect the name, we will use the directory name.
	// Alizer will not correctly default to the directory name when unable to detect it via pom.xml, package.json, etc.
	// So we must do it ourselves
	// The directory name SHOULD be the path (we use a previous check to see if it's "itsdir"
	if detectedName == "" {
		detectedName = filepath.Base(path)
	}

	// Step 3.  Sanitize the name
	// Make sure that detectedName conforms with Kubernetes naming rules
	// If not, we will use the directory name
	name := util.GetDNS1123Name(detectedName)
	klog.V(4).Infof("Path: %s, Detected name: %s, Sanitized name: %s", path, detectedName, name)
	if name == "" {
		return "", fmt.Errorf("unable to sanitize name to DNS1123 format: %q", name)
	}

	return name, nil
}

func (o *Alizer) DetectPorts(path string) ([]int, error) {
	//TODO(rm3l): Find a better way not to call recognizer.DetectComponents multiple times (in DetectFramework, DetectName and DetectPorts)
	components, err := recognizer.DetectComponents(path)
	if err != nil {
		return nil, err
	}

	if len(components) == 0 {
		klog.V(4).Infof("no components found at path %q", path)
		return nil, nil
	}

	return components[0].Ports, nil
}

func NewDetectionResult(typ model.DevfileType, registry api.Registry, appPorts []int, devfileVersion, name string) *api.DetectionResult {
	return &api.DetectionResult{
		Devfile:          typ.Name,
		DevfileRegistry:  registry.Name,
		ApplicationPorts: appPorts,
		DevfileVersion:   devfileVersion,
		Name:             name,
	}
}
