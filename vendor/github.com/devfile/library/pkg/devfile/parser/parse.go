package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/devfile/library/pkg/util"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"net/url"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
func parseDevfile(d DevfileObj, resolveCtx *resolutionContextTree, tool resolverTools, flattenedDevfile bool) (DevfileObj, error) {

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
		err = parseParentAndPlugin(d, resolveCtx, tool)
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
	// DefaultNamespace is the default namespace to use
	// If namespace is defined under devfile's parent kubernetes object, this namespace will be ignored.
	DefaultNamespace string
	// Context is the context used for making Kubernetes requests
	Context context.Context
	// K8sClient is the Kubernetes client instance used for interacting with a cluster
	K8sClient client.Client
}

// ParseDevfile func populates the devfile data, parses and validates the devfile integrity.
// Creates devfile context and runtime objects
func ParseDevfile(args ParserArgs) (d DevfileObj, err error) {
	if args.Data != nil {
		d.Ctx, err = devfileCtx.NewByteContentDevfileCtx(args.Data)
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

	tool := resolverTools{
		defaultNamespace: args.DefaultNamespace,
		registryURLs:     args.RegistryURLs,
		context:          args.Context,
		k8sClient:        args.K8sClient,
	}

	flattenedDevfile := true
	if args.FlattenedDevfile != nil {
		flattenedDevfile = *args.FlattenedDevfile
	}

	return populateAndParseDevfile(d, &resolutionContextTree{}, tool, flattenedDevfile)
}

// resolverTools contains required structs and data for resolving remote components of a devfile (plugins and parents)
type resolverTools struct {
	// DefaultNamespace is the default namespace to use for resolving Kubernetes ImportReferences that do not include one
	defaultNamespace string
	// RegistryURLs is a list of registry hosts which parser should pull parent devfile from.
	// If registryUrl is defined in devfile, this list will be ignored.
	registryURLs []string
	// Context is the context used for making Kubernetes or HTTP requests
	context context.Context
	// K8sClient is the Kubernetes client instance used for interacting with a cluster
	k8sClient client.Client
}

func populateAndParseDevfile(d DevfileObj, resolveCtx *resolutionContextTree, tool resolverTools, flattenedDevfile bool) (DevfileObj, error) {
	var err error
	if err = resolveCtx.hasCycle(); err != nil {
		return DevfileObj{}, err
	}
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

	return parseDevfile(d, resolveCtx, tool, flattenedDevfile)
}

// Parse func populates the flattened devfile data, parses and validates the devfile integrity.
// Creates devfile context and runtime objects
// Deprecated, use ParseDevfile() instead
func Parse(path string) (d DevfileObj, err error) {

	// NewDevfileCtx
	d.Ctx = devfileCtx.NewDevfileCtx(path)
	return populateAndParseDevfile(d, &resolutionContextTree{}, resolverTools{}, true)
}

// ParseRawDevfile populates the raw devfile data without overriding and merging
// Deprecated, use ParseDevfile() instead
func ParseRawDevfile(path string) (d DevfileObj, err error) {
	// NewDevfileCtx
	d.Ctx = devfileCtx.NewDevfileCtx(path)
	return populateAndParseDevfile(d, &resolutionContextTree{}, resolverTools{}, false)
}

// ParseFromURL func parses and validates the devfile integrity.
// Creates devfile context and runtime objects
// Deprecated, use ParseDevfile() instead
func ParseFromURL(url string) (d DevfileObj, err error) {
	d.Ctx = devfileCtx.NewURLDevfileCtx(url)
	return populateAndParseDevfile(d, &resolutionContextTree{}, resolverTools{}, true)
}

// ParseFromData func parses and validates the devfile integrity.
// Creates devfile context and runtime objects
// Deprecated, use ParseDevfile() instead
func ParseFromData(data []byte) (d DevfileObj, err error) {
	d.Ctx, err = devfileCtx.NewByteContentDevfileCtx(data)
	if err != nil {
		return d, errors.Wrap(err, "failed to set devfile content from bytes")
	}
	return populateAndParseDevfile(d, &resolutionContextTree{}, resolverTools{}, true)
}

func parseParentAndPlugin(d DevfileObj, resolveCtx *resolutionContextTree, tool resolverTools) (err error) {
	flattenedParent := &v1.DevWorkspaceTemplateSpecContent{}
	parent := d.Data.GetParent()
	if parent != nil {
		if !reflect.DeepEqual(parent, &v1.Parent{}) {

			var parentDevfileObj DevfileObj
			switch {
			case parent.Uri != "":
				parentDevfileObj, err = parseFromURI(parent.ImportReference, d.Ctx, resolveCtx, tool)
			case parent.Id != "":
				parentDevfileObj, err = parseFromRegistry(parent.ImportReference, resolveCtx, tool)
			case parent.Kubernetes != nil:
				parentDevfileObj, err = parseFromKubeCRD(parent.ImportReference, resolveCtx, tool)
			default:
				return fmt.Errorf("devfile parent does not define any resources")
			}
			if err != nil {
				return err
			}

			parentWorkspaceContent := parentDevfileObj.Data.GetDevfileWorkspaceSpecContent()
			// add attribute to parent elements
			err = addSourceAttributesForOverrideAndMerge(parent.ImportReference, parentWorkspaceContent)
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(parent.ParentOverrides, v1.ParentOverrides{}) {
				// add attribute to parentOverrides elements
				curNodeImportReference := resolveCtx.importReference
				err = addSourceAttributesForOverrideAndMerge(curNodeImportReference, &parent.ParentOverrides)
				if err != nil {
					return err
				}
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
			switch {
			case plugin.Uri != "":
				pluginDevfileObj, err = parseFromURI(plugin.ImportReference, d.Ctx, resolveCtx, tool)
			case plugin.Id != "":
				pluginDevfileObj, err = parseFromRegistry(plugin.ImportReference, resolveCtx, tool)
			case plugin.Kubernetes != nil:
				pluginDevfileObj, err = parseFromKubeCRD(plugin.ImportReference, resolveCtx, tool)
			default:
				return fmt.Errorf("plugin %s does not define any resources", component.Name)
			}
			if err != nil {
				return err
			}
			pluginWorkspaceContent := pluginDevfileObj.Data.GetDevfileWorkspaceSpecContent()
			// add attribute to plugin elements
			err = addSourceAttributesForOverrideAndMerge(plugin.ImportReference, pluginWorkspaceContent)
			if err != nil {
				return err
			}
			flattenedPlugin := pluginWorkspaceContent
			if !reflect.DeepEqual(plugin.PluginOverrides, v1.PluginOverrides{}) {
				// add attribute to pluginOverrides elements
				curNodeImportReference := resolveCtx.importReference
				err = addSourceAttributesForOverrideAndMerge(curNodeImportReference, &plugin.PluginOverrides)
				if err != nil {
					return err
				}
				flattenedPlugin, err = apiOverride.OverrideDevWorkspaceTemplateSpec(pluginWorkspaceContent, plugin.PluginOverrides)
				if err != nil {
					return err
				}
			}
			flattenedPlugins = append(flattenedPlugins, flattenedPlugin)
		}
	}

	mergedContent, err := apiOverride.MergeDevWorkspaceTemplateSpec(d.Data.GetDevfileWorkspaceSpecContent(), flattenedParent, flattenedPlugins...)
	if err != nil {
		return err
	}
	d.Data.SetDevfileWorkspaceSpecContent(*mergedContent)
	// remove parent from flatterned devfile
	d.Data.SetParent(nil)

	return nil
}

func parseFromURI(importReference v1.ImportReference, curDevfileCtx devfileCtx.DevfileCtx, resolveCtx *resolutionContextTree, tool resolverTools) (DevfileObj, error) {
	uri := importReference.Uri
	// validate URI
	err := validation.ValidateURI(uri)
	if err != nil {
		return DevfileObj{}, err
	}
	// NewDevfileCtx
	var d DevfileObj
	absoluteURL := strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://")
	var newUri string

	// relative path on disk
	if !absoluteURL && curDevfileCtx.GetAbsPath() != "" {
		newUri = path.Join(path.Dir(curDevfileCtx.GetAbsPath()), uri)
		d.Ctx = devfileCtx.NewDevfileCtx(newUri)
	} else if absoluteURL {
		// absolute URL address
		newUri = uri
		d.Ctx = devfileCtx.NewURLDevfileCtx(newUri)
	} else if curDevfileCtx.GetURL() != "" {
		// relative path to a URL
		u, err := url.Parse(curDevfileCtx.GetURL())
		if err != nil {
			return DevfileObj{}, err
		}
		u.Path = path.Join(path.Dir(u.Path), uri)
		newUri = u.String()
		d.Ctx = devfileCtx.NewURLDevfileCtx(newUri)
	}
	importReference.Uri = newUri
	newResolveCtx := resolveCtx.appendNode(importReference)

	return populateAndParseDevfile(d, newResolveCtx, tool, true)
}

func parseFromRegistry(importReference v1.ImportReference, resolveCtx *resolutionContextTree, tool resolverTools) (d DevfileObj, err error) {
	id := importReference.Id
	registryURL := importReference.RegistryUrl
	if registryURL != "" {
		devfileContent, err := getDevfileFromRegistry(id, registryURL)
		if err != nil {
			return DevfileObj{}, err
		}
		d.Ctx, err = devfileCtx.NewByteContentDevfileCtx(devfileContent)
		if err != nil {
			return d, errors.Wrap(err, "failed to set devfile content from bytes")
		}
		newResolveCtx := resolveCtx.appendNode(importReference)

		return populateAndParseDevfile(d, newResolveCtx, tool, true)

	} else if tool.registryURLs != nil {
		for _, registryURL := range tool.registryURLs {
			devfileContent, err := getDevfileFromRegistry(id, registryURL)
			if devfileContent != nil && err == nil {
				d.Ctx, err = devfileCtx.NewByteContentDevfileCtx(devfileContent)
				if err != nil {
					return d, errors.Wrap(err, "failed to set devfile content from bytes")
				}
				importReference.RegistryUrl = registryURL
				newResolveCtx := resolveCtx.appendNode(importReference)

				return populateAndParseDevfile(d, newResolveCtx, tool, true)
			}
		}
	} else {
		return DevfileObj{}, fmt.Errorf("failed to fetch from registry, registry URL is not provided")
	}

	return DevfileObj{}, fmt.Errorf("failed to get id: %s from registry URLs provided", id)
}

func getDevfileFromRegistry(id, registryURL string) ([]byte, error) {
	if !strings.HasPrefix(registryURL, "http://") && !strings.HasPrefix(registryURL, "https://") {
		return nil, fmt.Errorf("the provided registryURL: %s is not a valid URL", registryURL)
	}
	param := util.HTTPRequestParams{
		URL: fmt.Sprintf("%s/devfiles/%s", registryURL, id),
	}
	return util.HTTPGetRequest(param, 0)
}

func parseFromKubeCRD(importReference v1.ImportReference, resolveCtx *resolutionContextTree, tool resolverTools) (d DevfileObj, err error) {

	if tool.k8sClient == nil || tool.context == nil {
		return DevfileObj{}, fmt.Errorf("Kubernetes client and context are required to parse from Kubernetes CRD")
	}
	namespace := importReference.Kubernetes.Namespace

	if namespace == "" {
		// if namespace is not set in devfile, use default namespace provided in by consumer
		if tool.defaultNamespace != "" {
			namespace = tool.defaultNamespace
		} else {
			// use current namespace if namespace is not set in devfile and not provided by consumer
			loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
			configOverrides := &clientcmd.ConfigOverrides{}
			config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
			namespace, _, err = config.Namespace()
			if err != nil {
				return DevfileObj{}, fmt.Errorf("kubernetes namespace is not provided, and cannot get current running cluster's namespace: %v", err)
			}
		}
	}

	var dwTemplate v1.DevWorkspaceTemplate
	namespacedName := types.NamespacedName{
		Name:      importReference.Kubernetes.Name,
		Namespace: namespace,
	}
	err = tool.k8sClient.Get(tool.context, namespacedName, &dwTemplate)
	if err != nil {
		return DevfileObj{}, err
	}

	d, err = convertDevWorskapceTemplateToDevObj(dwTemplate)
	if err != nil {
		return DevfileObj{}, err
	}

	importReference.Kubernetes.Namespace = namespace
	newResolveCtx := resolveCtx.appendNode(importReference)

	err = parseParentAndPlugin(d, newResolveCtx, tool)
	return d, err

}

func convertDevWorskapceTemplateToDevObj(dwTemplate v1.DevWorkspaceTemplate) (d DevfileObj, err error) {
	// APIVersion: group/version
	// for example: APIVersion: "workspace.devfile.io/v1alpha2" uses api version v1alpha2, and match to v2 schemas
	tempList := strings.Split(dwTemplate.APIVersion, "/")
	apiversion := tempList[len(tempList)-1]
	d.Data, err = data.NewDevfileData(apiversion)
	if err != nil {
		return DevfileObj{}, err
	}
	d.Data.SetDevfileWorkspaceSpec(dwTemplate.Spec)

	return d, nil

}
