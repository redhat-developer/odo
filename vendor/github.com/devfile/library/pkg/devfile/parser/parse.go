package parser

import (
	"encoding/json"
	"fmt"
	"github.com/devfile/library/pkg/util"
	"net/url"
	"path"
	"strings"

	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"k8s.io/klog"

	"reflect"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	apiOverride "github.com/devfile/api/v2/pkg/utils/overriding"
	"github.com/devfile/api/v2/pkg/validation"
	"github.com/pkg/errors"
)

// ParseDevfile func validates the devfile integrity.
// Creates devfile context and runtime objects
func parseDevfile(d DevfileObj, flattenedDevfile bool) (DevfileObj, error) {

	// Validate devfile
	err := d.Ctx.Validate()
	if err != nil {
		return d, err
	}

	// Create a new devfile data object
	d.Data, err = data.NewDevfileData(d.Ctx.GetApiVersion())
	if err != nil {
		return d, err
	}

	// Unmarshal devfile content into devfile struct
	err = json.Unmarshal(d.Ctx.GetDevfileContent(), &d.Data)
	if err != nil {
		return d, errors.Wrapf(err, "failed to decode devfile content")
	}

	if flattenedDevfile {
		err = parseParentAndPlugin(d)
		if err != nil {
			return DevfileObj{}, err
		}
	}

	// Successful
	return d, nil
}

// ParserArgs is the struct to pass into parser functions which contains required info for parsing devfile.
// It accepts devfile path, devfile URL or devfile content in []byte format.
type ParserArgs struct {
	// Path is a relative or absolute devfile path.
	Path string
	// URL is the URL address of the specific devfile.
	URL string
	// Data is the devfile content in []byte format.
	Data []byte
	// FlattenedDevfile defines if the returned devfileObj is flattened content (true) or raw content (false).
	// The value is default to be true.
	FlattenedDevfile *bool
	// RegistryURLs is a list of registry hosts which parser should pull parent devfile from.
	// If registryUrl is defined in devfile, this list will be ignored.
	RegistryURLs []string
}

// ParseDevfile func populates the devfile data, parses and validates the devfile integrity.
// Creates devfile context and runtime objects
func ParseDevfile(args ParserArgs) (d DevfileObj, err error) {
	if args.Data != nil {
		d.Ctx = devfileCtx.DevfileCtx{}
		err = d.Ctx.SetDevfileContentFromBytes(args.Data)
		if err != nil {
			return d, errors.Wrap(err, "failed to set devfile content from bytes")
		}
	} else if args.Path != "" {
		d.Ctx = devfileCtx.NewDevfileCtx(args.Path)
	} else if args.URL != "" {
		d.Ctx = devfileCtx.NewURLDevfileCtx(args.URL)
	} else {
		return d, errors.Wrap(err, "the devfile source is not provided")
	}

	if args.RegistryURLs != nil {
		d.Ctx.SetRegistryURLs(args.RegistryURLs)
	}

	flattenedDevfile := true
	if args.FlattenedDevfile != nil {
		flattenedDevfile = *args.FlattenedDevfile
	}

	return populateAndParseDevfile(d, flattenedDevfile)
}

func populateAndParseDevfile(d DevfileObj, flattenedDevfile bool) (DevfileObj, error) {
	var err error

	// Fill the fields of DevfileCtx struct
	if d.Ctx.GetURL() != "" {
		err = d.Ctx.PopulateFromURL()
	} else if d.Ctx.GetDevfileContent() != nil {
		err = d.Ctx.PopulateFromRaw()
	} else {
		err = d.Ctx.Populate()
	}
	if err != nil {
		return d, err
	}

	return parseDevfile(d, flattenedDevfile)
}

// Parse func populates the flattened devfile data, parses and validates the devfile integrity.
// Creates devfile context and runtime objects
// Deprecated, use ParseDevfile() instead
func Parse(path string) (d DevfileObj, err error) {

	// NewDevfileCtx
	d.Ctx = devfileCtx.NewDevfileCtx(path)

	return populateAndParseDevfile(d, true)
}

// ParseRawDevfile populates the raw devfile data without overriding and merging
// Deprecated, use ParseDevfile() instead
func ParseRawDevfile(path string) (d DevfileObj, err error) {
	// NewDevfileCtx
	d.Ctx = devfileCtx.NewDevfileCtx(path)

	return populateAndParseDevfile(d, false)
}

// ParseFromURL func parses and validates the devfile integrity.
// Creates devfile context and runtime objects
// Deprecated, use ParseDevfile() instead
func ParseFromURL(url string) (d DevfileObj, err error) {
	d.Ctx = devfileCtx.NewURLDevfileCtx(url)
	return populateAndParseDevfile(d, true)
}

// ParseFromData func parses and validates the devfile integrity.
// Creates devfile context and runtime objects
// Deprecated, use ParseDevfile() instead
func ParseFromData(data []byte) (d DevfileObj, err error) {
	d.Ctx = devfileCtx.DevfileCtx{}
	err = d.Ctx.SetDevfileContentFromBytes(data)
	if err != nil {
		return d, errors.Wrap(err, "failed to set devfile content from bytes")
	}
	return populateAndParseDevfile(d, true)
}

func parseParentAndPlugin(d DevfileObj) (err error) {
	flattenedParent := &v1.DevWorkspaceTemplateSpecContent{}
	parent := d.Data.GetParent()
	if parent != nil {
		if !reflect.DeepEqual(parent, &v1.Parent{}) {

			var parentDevfileObj DevfileObj
			if parent.Uri != "" {
				parentDevfileObj, err = parseFromURI(parent.Uri, d.Ctx)
				if err != nil {
					return err
				}
			} else if parent.Id != "" {
				parentDevfileObj, err = parseFromRegistry(parent.Id, parent.RegistryUrl, d.Ctx)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("parent URI or parent Id undefined, currently only URI and Id are suppported")
			}

			parentWorkspaceContent := parentDevfileObj.Data.GetDevfileWorkspace()
			if !reflect.DeepEqual(parent.ParentOverrides, v1.ParentOverrides{}) {
				flattenedParent, err = apiOverride.OverrideDevWorkspaceTemplateSpec(parentWorkspaceContent, parent.ParentOverrides)
				if err != nil {
					return err
				}
			} else {
				flattenedParent = parentWorkspaceContent
			}

			klog.V(4).Infof("adding data of devfile with URI: %v", parent.Uri)
		}
	}

	flattenedPlugins := []*v1.DevWorkspaceTemplateSpecContent{}
	components, err := d.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}
	for _, component := range components {
		if component.Plugin != nil && !reflect.DeepEqual(component.Plugin, &v1.PluginComponent{}) {
			plugin := component.Plugin
			var pluginDevfileObj DevfileObj
			if plugin.Uri != "" {
				pluginDevfileObj, err = parseFromURI(plugin.Uri, d.Ctx)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("plugin URI undefined, currently only URI is suppported")
			}
			pluginWorkspaceContent := pluginDevfileObj.Data.GetDevfileWorkspace()
			flattenedPlugin := pluginWorkspaceContent
			if !reflect.DeepEqual(plugin.PluginOverrides, v1.PluginOverrides{}) {
				flattenedPlugin, err = apiOverride.OverrideDevWorkspaceTemplateSpec(pluginWorkspaceContent, plugin.PluginOverrides)
				if err != nil {
					return err
				}
			}
			flattenedPlugins = append(flattenedPlugins, flattenedPlugin)
		}
	}

	mergedContent, err := apiOverride.MergeDevWorkspaceTemplateSpec(d.Data.GetDevfileWorkspace(), flattenedParent, flattenedPlugins...)
	if err != nil {
		return err
	}
	d.Data.SetDevfileWorkspace(*mergedContent)
	// remove parent from flatterned devfile
	d.Data.SetParent(nil)

	return nil
}

func parseFromURI(uri string, curDevfileCtx devfileCtx.DevfileCtx) (DevfileObj, error) {
	// validate URI
	err := validation.ValidateURI(uri)
	if err != nil {
		return DevfileObj{}, err
	}
	// NewDevfileCtx
	var d DevfileObj
	absoluteURL := strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://")

	// relative path on disk
	if !absoluteURL && curDevfileCtx.GetAbsPath() != "" {
		d.Ctx = devfileCtx.NewDevfileCtx(path.Join(path.Dir(curDevfileCtx.GetAbsPath()), uri))
	} else if absoluteURL {
		// absolute URL address
		d.Ctx = devfileCtx.NewURLDevfileCtx(uri)
	} else if curDevfileCtx.GetURL() != "" {
		// relative path to a URL
		u, err := url.Parse(curDevfileCtx.GetURL())
		if err != nil {
			return DevfileObj{}, err
		}
		u.Path = path.Join(path.Dir(u.Path), uri)
		d.Ctx = devfileCtx.NewURLDevfileCtx(u.String())
	}
	d.Ctx.SetURIMap(curDevfileCtx.GetURIMap())
	return populateAndParseDevfile(d, true)
}

func parseFromRegistry(parentId, registryURL string, curDevfileCtx devfileCtx.DevfileCtx) (DevfileObj, error) {
	if registryURL != "" {
		devfileContent, err := getDevfileFromRegistry(parentId, registryURL)
		if err != nil {
			return DevfileObj{}, err
		}
		return ParseDevfile(ParserArgs{Data: devfileContent, RegistryURLs: curDevfileCtx.GetRegistryURLs()})
	} else if curDevfileCtx.GetRegistryURLs() != nil {
		for _, registry := range curDevfileCtx.GetRegistryURLs() {
			devfileContent, err := getDevfileFromRegistry(parentId, registry)
			if devfileContent != nil && err == nil {
				return ParseDevfile(ParserArgs{Data: devfileContent, RegistryURLs: curDevfileCtx.GetRegistryURLs()})
			}
		}
	} else {
		return DevfileObj{}, fmt.Errorf("failed to fetch from registry, registry URL is not provided")
	}

	return DevfileObj{}, fmt.Errorf("failed to get parent Id: %s from registry URLs provided", parentId)
}

func getDevfileFromRegistry(parentId, registryURL string) ([]byte, error) {
	if !strings.HasPrefix(registryURL, "http://") && !strings.HasPrefix(registryURL, "https://") {
		return nil, fmt.Errorf("the provided registryURL: %s is not a valid URL", registryURL)
	}
	param := util.HTTPRequestParams{
		URL: fmt.Sprintf("%s/devfiles/%s", registryURL, parentId),
	}
	return util.HTTPGetRequest(param, 0)
}
