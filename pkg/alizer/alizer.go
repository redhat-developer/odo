package alizer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
	"github.com/redhat-developer/alizer/go/pkg/apis/recognizer"
	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/registry"
	"github.com/redhat-developer/odo/pkg/util"
	"k8s.io/klog"
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
func (o *Alizer) DetectFramework(ctx context.Context, path string) (model.DevFileType, api.Registry, error) {
	types := []model.DevFileType{}
	components, err := o.registryClient.ListDevfileStacks(ctx, "", "", "", false)
	if err != nil {
		return model.DevFileType{}, api.Registry{}, err
	}
	for _, component := range components.Items {
		types = append(types, model.DevFileType{
			Name:        component.Name,
			Language:    component.Language,
			ProjectType: component.ProjectType,
			Tags:        component.Tags,
		})
	}
	typ, err := recognizer.SelectDevFileFromTypes(path, types)
	if err != nil {
		return model.DevFileType{}, api.Registry{}, err
	}
	return types[typ], components.Items[typ].Registry, nil
}

// DetectName retrieves the name of the project (if available)
// If source code is detected:
// 1. Detect the name (pom.xml for java, package.json for nodejs, etc.)
// 2. If unable to detect the name, use the directory name
//
// If no source is detected:
// 1. Use the directory name
//
// Last step. Sanitize the name so it's valid for a component name

// Use:
// import "github.com/redhat-developer/alizer/pkg/apis/recognizer"
// components, err := recognizer.DetectComponents("./")

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

func GetDevfileLocationFromDetection(typ model.DevFileType, registry api.Registry) *api.DevfileLocation {
	return &api.DevfileLocation{
		Devfile:         typ.Name,
		DevfileRegistry: registry.Name,
	}
}
