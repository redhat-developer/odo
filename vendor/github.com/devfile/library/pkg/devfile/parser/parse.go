package parser

import (
	"encoding/json"
	"fmt"
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
	for uri := range devfileCtx.URIMap {
		delete(devfileCtx.URIMap, uri)
	}
	// Successful
	return d, nil
}

// Parse func populates the flattened devfile data, parses and validates the devfile integrity.
// Creates devfile context and runtime objects
func Parse(path string) (d DevfileObj, err error) {

	// NewDevfileCtx
	d.Ctx = devfileCtx.NewDevfileCtx(path)

	// Fill the fields of DevfileCtx struct
	err = d.Ctx.Populate()
	if err != nil {
		return d, err
	}
	return parseDevfile(d, true)
}

// ParseRawDevfile populates the raw devfile data without overriding and merging
func ParseRawDevfile(path string) (d DevfileObj, err error) {
	// NewDevfileCtx
	d.Ctx = devfileCtx.NewDevfileCtx(path)

	// Fill the fields of DevfileCtx struct
	err = d.Ctx.Populate()
	if err != nil {
		return d, err
	}
	return parseDevfile(d, false)
}

// ParseFromURL func parses and validates the devfile integrity.
// Creates devfile context and runtime objects
func ParseFromURL(url string) (d DevfileObj, err error) {
	d.Ctx = devfileCtx.NewURLDevfileCtx(url)
	// Fill the fields of DevfileCtx struct
	err = d.Ctx.PopulateFromURL()
	if err != nil {
		return d, err
	}
	return parseDevfile(d, true)
}

// ParseFromData func parses and validates the devfile integrity.
// Creates devfile context and runtime objects
func ParseFromData(data []byte) (d DevfileObj, err error) {
	d.Ctx = devfileCtx.DevfileCtx{}
	err = d.Ctx.SetDevfileContentFromBytes(data)
	if err != nil {
		return d, errors.Wrap(err, "failed to set devfile content from bytes")
	}
	err = d.Ctx.PopulateFromRaw()
	if err != nil {
		return d, err
	}

	return parseDevfile(d, true)
}

func parseParentAndPlugin(d DevfileObj) (err error) {
	flattenedParent := &v1.DevWorkspaceTemplateSpecContent{}
	if d.Data.GetParent() != nil {
		if !reflect.DeepEqual(d.Data.GetParent(), &v1.Parent{}) {

			parent := d.Data.GetParent()
			var parentDevfileObj DevfileObj
			if d.Data.GetParent().Uri != "" {
				parentDevfileObj, err = parseFromURI(parent.Uri, d.Ctx)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("parent URI undefined, currently only URI is suppported")
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

	// absolute URL address
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		return ParseFromURL(uri)
	}

	// relative path on disk
	if curDevfileCtx.GetAbsPath() != "" {
		return Parse(path.Join(path.Dir(curDevfileCtx.GetAbsPath()), uri))
	}

	if curDevfileCtx.GetURL() != "" {
		u, err := url.Parse(curDevfileCtx.GetURL())
		if err != nil {
			return DevfileObj{}, err
		}

		u.Path = path.Join(path.Dir(u.Path), uri)
		// u.String() is the joint absolute URL path
		return ParseFromURL(u.String())
	}

	return DevfileObj{}, fmt.Errorf("fail to parse from uri: %s", uri)
}
