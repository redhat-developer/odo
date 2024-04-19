//
// Copyright Red Hat
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/devfile/api/v2/pkg/attributes"
	devfileCtx "github.com/devfile/library/v2/pkg/devfile/parser/context"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	errPkg "github.com/devfile/library/v2/pkg/devfile/parser/errors"
	"github.com/devfile/library/v2/pkg/util"
	registryLibrary "github.com/devfile/registry-support/registry-library/library"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	parserUtil "github.com/devfile/library/v2/pkg/devfile/parser/util"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	apiOverride "github.com/devfile/api/v2/pkg/utils/overriding"
	"github.com/devfile/api/v2/pkg/validation"
	versionpkg "github.com/hashicorp/go-version"
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
		return d, &errPkg.NonCompliantDevfile{Err: err.Error()}
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
	// Path is a relative or absolute devfile path on disk.
	// It can also be a relative or absolute path to a folder containing one or more devfiles,
	// in which case the library will try to pick an existing one, based on the following priority order:
	// devfile.yaml > .devfile.yaml > devfile.yml > .devfile.yml
	Path string
	// URL is the URL address of the specific devfile.
	URL string
	// Data is the devfile content in []byte format.
	Data []byte
	// FlattenedDevfile defines if the returned devfileObj is flattened content (true) or raw content (false).
	// The value is default to be true.
	FlattenedDevfile *bool
	// ConvertKubernetesContentInUri defines if the kubernetes resources definition from uri will be converted to inlined in devObj(true) or not (false).
	// The value is default to be true.
	ConvertKubernetesContentInUri *bool
	// RegistryURLs is a list of registry hosts which parser should pull parent devfile from.
	// If registryUrl is defined in devfile, this list will be ignored.
	RegistryURLs []string
	// Token is a GitHub, GitLab, or Bitbucket personal access token used with a private git repo URL
	Token string
	// DefaultNamespace is the default namespace to use
	// If namespace is defined under devfile's parent kubernetes object, this namespace will be ignored.
	DefaultNamespace string
	// Context is the context used for making Kubernetes requests
	Context context.Context
	// K8sClient is the Kubernetes client instance used for interacting with a cluster
	K8sClient client.Client
	// ExternalVariables override variables defined in the Devfile
	ExternalVariables map[string]string
	// HTTPTimeout overrides the request and response timeout values for reading a parent devfile reference from the registry.  If a negative value is specified, the default timeout will be used.
	HTTPTimeout *int
	// SetBooleanDefaults sets the boolean properties to their default values after a devfile been parsed.
	// The value is true by default.  Clients can set this to false if they want to set the boolean properties themselves
	SetBooleanDefaults *bool
	// ImageNamesAsSelector sets the information that will be used to handle image names as selectors when parsing the Devfile.
	// Not setting this field or setting it to nil disables the logic of handling image names as selectors.
	ImageNamesAsSelector *ImageSelectorArgs
	// DownloadGitResources downloads the resources from Git repository if true
	DownloadGitResources *bool
	// DevfileUtilsClient exposes the interface for mock implementation.
	DevfileUtilsClient parserUtil.DevfileUtils
}

// ImageSelectorArgs defines the structure to leverage for using image names as selectors after parsing the Devfile.
// The fields defined here will be used together to compute the final image names that will be built and pushed,
// and replaced in all matching Image, Container or Kubernetes/OpenShift components.
//
// For Kubernetes/OpenShift components, replacement is done only in core Kubernetes resources
// (CronJob, DaemonSet, Deployment, Job, Pod, ReplicaSet, ReplicationController, StatefulSet) that are *inlined* in those components.
// Resources referenced via URIs will not be resolved. So you may want to also set ConvertKubernetesContentInUri to true in the parser args.
//
// For example, if Registry is set to "<local-registry>/<user-org>" and Tag is set to "some-dynamic-unique-tag",
// all container and Kubernetes/OpenShift components matching a relative image name (say "my-image-name") of an Image component
// will be replaced in the resulting Devfile by: "<local-registry>/<user-org>/<devfile-name>-my-image-name:some-dynamic-unique-tag".
type ImageSelectorArgs struct {
	// Registry is the registry base path under which images matching selectors will be built and pushed to. Required.
	//
	// Example: <local-registry>/<user-org>
	Registry string
	// Tag represents a tag identifier under which images matching selectors will be built and pushed to.
	// This should ideally be set to a unique identifier for each run of the caller tool.
	Tag string
}

// ParseDevfile func populates the devfile data, parses and validates the devfile integrity.
// Creates devfile context and runtime objects
func ParseDevfile(args ParserArgs) (d DevfileObj, err error) {
	if args.ImageNamesAsSelector != nil && strings.TrimSpace(args.ImageNamesAsSelector.Registry) == "" {
		return DevfileObj{}, errors.New("registry is mandatory when setting ImageNamesAsSelector in the parser args")
	}

	if args.Data != nil {
		d.Ctx, err = devfileCtx.NewByteContentDevfileCtx(args.Data)
		if err != nil {
			return d, err
		}
	} else if args.Path != "" {
		d.Ctx = devfileCtx.NewDevfileCtx(args.Path)
	} else if args.URL != "" {
		d.Ctx = devfileCtx.NewURLDevfileCtx(args.URL)
	} else {
		return d, fmt.Errorf("the devfile source is not provided")
	}

	if args.Token != "" {
		d.Ctx.SetToken(args.Token)
	}

	if args.DevfileUtilsClient == nil {
		args.DevfileUtilsClient = parserUtil.NewDevfileUtilsClient()
	}

	downloadGitResources := true
	if args.DownloadGitResources != nil {
		downloadGitResources = *args.DownloadGitResources
	}

	tool := resolverTools{
		defaultNamespace:     args.DefaultNamespace,
		registryURLs:         args.RegistryURLs,
		context:              args.Context,
		k8sClient:            args.K8sClient,
		httpTimeout:          args.HTTPTimeout,
		downloadGitResources: downloadGitResources,
		devfileUtilsClient:   args.DevfileUtilsClient,
	}

	flattenedDevfile := true
	if args.FlattenedDevfile != nil {
		flattenedDevfile = *args.FlattenedDevfile
	}

	d, err = populateAndParseDevfile(d, &resolutionContextTree{}, tool, flattenedDevfile)
	if err != nil {
		return d, err
	}

	setBooleanDefaults := true
	if args.SetBooleanDefaults != nil {
		setBooleanDefaults = *args.SetBooleanDefaults
	}
	//set defaults only if parsing succeeded
	if err == nil && setBooleanDefaults {
		err := setDefaults(d)
		if err != nil {
			return d, errors.Wrap(err, "failed to setDefaults")
		}
	}

	convertUriToInlined := true
	if args.ConvertKubernetesContentInUri != nil {
		convertUriToInlined = *args.ConvertKubernetesContentInUri
	}

	if convertUriToInlined {
		d.Ctx.SetConvertUriToInlined(true)
		err = parseKubeResourceFromURI(d, tool.devfileUtilsClient)
		if err != nil {
			return d, err
		}
	}

	return d, err
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
	// httpTimeout is the timeout value in seconds passed in from the client.
	httpTimeout *int
	// downloadGitResources downloads the resources from Git repository if true
	downloadGitResources bool
	// devfileUtilsClient exposes the Git Interface to be able to use mock implementation.
	devfileUtilsClient parserUtil.DevfileUtils
}

func populateAndParseDevfile(d DevfileObj, resolveCtx *resolutionContextTree, tool resolverTools, flattenedDevfile bool) (DevfileObj, error) {
	var err error
	if err = resolveCtx.hasCycle(); err != nil {
		return DevfileObj{}, &errPkg.NonCompliantDevfile{Err: err.Error()}
	}
	// Fill the fields of DevfileCtx struct
	if d.Ctx.GetURL() != "" {
		err = d.Ctx.PopulateFromURL(tool.devfileUtilsClient)
	} else if d.Ctx.GetDevfileContent() != nil {
		err = d.Ctx.PopulateFromRaw()
	} else {
		err = d.Ctx.Populate(tool.devfileUtilsClient)
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
	var mainDevfileVersion, parentDevfileVerson, pluginDevfileVerson *versionpkg.Version
	var devfileVersion string
	if devfileVersion = d.Ctx.GetApiVersion(); devfileVersion == "" {
		devfileVersion = d.Data.GetSchemaVersion()
	}

	if devfileVersion != "" {
		mainDevfileVersion, err = versionpkg.NewVersion(devfileVersion)
		if err != nil {
			return fmt.Errorf("fail to parse version of the main devfile")
		}
	}
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
				err = &errPkg.NonCompliantDevfile{Err: "devfile parent does not define any resources"}
			}
			if err != nil {
				return err
			}
			var devfileVersion string
			if devfileVersion = parentDevfileObj.Ctx.GetApiVersion(); devfileVersion == "" {
				devfileVersion = parentDevfileObj.Data.GetSchemaVersion()
			}

			if devfileVersion != "" {
				parentDevfileVerson, err = versionpkg.NewVersion(devfileVersion)
				if err != nil {
					return fmt.Errorf("fail to parse version of parent devfile from: %v", resolveImportReference(parent.ImportReference))
				}

				if parentDevfileVerson.GreaterThan(mainDevfileVersion) {
					return &errPkg.NonCompliantDevfile{Err: fmt.Sprintf("the parent devfile version from %v is greater than the child devfile version from %v", resolveImportReference(parent.ImportReference), resolveImportReference(resolveCtx.importReference))}
				}
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
				err = &errPkg.NonCompliantDevfile{Err: fmt.Sprintf("plugin %s does not define any resources", component.Name)}
			}
			if err != nil {
				return err
			}
			var devfileVersion string
			if devfileVersion = pluginDevfileObj.Ctx.GetApiVersion(); devfileVersion == "" {
				devfileVersion = pluginDevfileObj.Data.GetSchemaVersion()
			}

			if devfileVersion != "" {
				pluginDevfileVerson, err = versionpkg.NewVersion(devfileVersion)
				if err != nil {
					return fmt.Errorf("fail to parse version of plugin devfile from: %v", resolveImportReference(component.Plugin.ImportReference))
				}
				if pluginDevfileVerson.GreaterThan(mainDevfileVersion) {
					return &errPkg.NonCompliantDevfile{Err: fmt.Sprintf("the plugin devfile version from %v is greater than the child devfile version from %v", resolveImportReference(component.Plugin.ImportReference), resolveImportReference(resolveCtx.importReference))}
				}
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
		if util.ValidateFile(newUri) != nil {
			return DevfileObj{}, &errPkg.NonCompliantDevfile{Err: fmt.Sprintf("the provided path is not a valid filepath %s", newUri)}
		}
		srcDir := path.Dir(newUri)
		destDir := path.Dir(curDevfileCtx.GetAbsPath())
		if srcDir != destDir {
			err := util.CopyAllDirFiles(srcDir, destDir)
			if err != nil {
				return DevfileObj{}, err
			}
		}
	} else {
		if absoluteURL {
			// absolute URL address
			newUri = uri
		} else if curDevfileCtx.GetURL() != "" {
			// relative path to a URL
			u, err := url.Parse(curDevfileCtx.GetURL())
			if err != nil {
				return DevfileObj{}, err
			}
			u.Path = path.Join(u.Path, uri)
			newUri = u.String()
		} else {
			return DevfileObj{}, fmt.Errorf("failed to resolve parent uri, devfile context is missing absolute url and path to devfile. %s", resolveImportReference(importReference))
		}

		token := curDevfileCtx.GetToken()
		d.Ctx = devfileCtx.NewURLDevfileCtx(newUri)
		if token != "" {
			d.Ctx.SetToken(token)
		}

		if tool.downloadGitResources {
			destDir := path.Dir(curDevfileCtx.GetAbsPath())
			err = tool.devfileUtilsClient.DownloadGitRepoResources(newUri, destDir, token)
			if err != nil {
				return DevfileObj{}, err
			}
		}
	}
	importReference.Uri = newUri
	newResolveCtx := resolveCtx.appendNode(importReference)

	return populateAndParseDevfile(d, newResolveCtx, tool, true)
}

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
			return d, err
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
		return DevfileObj{}, &errPkg.NonCompliantDevfile{Err: "failed to fetch from registry, registry URL is not provided"}
	}

	return DevfileObj{}, fmt.Errorf("failed to get id: %s from registry URLs provided", id)
}

func getDevfileFromRegistry(id, registryURL, version string, httpTimeout *int) ([]byte, error) {
	if !strings.HasPrefix(registryURL, "http://") && !strings.HasPrefix(registryURL, "https://") {
		return nil, &errPkg.NonCompliantDevfile{Err: fmt.Sprintf("the provided registryURL: %s is not a valid URL", registryURL)}
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
	stackDir, err := os.MkdirTemp(os.TempDir(), fmt.Sprintf("registry-resources-%s", id))
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

func parseFromKubeCRD(importReference v1.ImportReference, resolveCtx *resolutionContextTree, tool resolverTools) (d DevfileObj, err error) {

	if tool.k8sClient == nil || tool.context == nil {
		return DevfileObj{}, fmt.Errorf("kubernetes client and context are required to parse from Kubernetes CRD")
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

// setDefaults sets the default values for nil boolean properties after the merging of devWorkspaceTemplateSpec is complete
func setDefaults(d DevfileObj) (err error) {

	var devfileVersion string
	if devfileVersion = d.Ctx.GetApiVersion(); devfileVersion == "" {
		devfileVersion = d.Data.GetSchemaVersion()
	}

	commands, err := d.Data.GetCommands(common.DevfileOptions{})

	if err != nil {
		return err
	}

	//set defaults on the commands
	var cmdGroup *v1.CommandGroup
	for i := range commands {
		command := commands[i]
		cmdGroup = nil

		if command.Exec != nil {
			exec := command.Exec
			val := exec.GetHotReloadCapable()
			exec.HotReloadCapable = &val
			cmdGroup = exec.Group

		} else if command.Composite != nil {
			composite := command.Composite
			val := composite.GetParallel()
			composite.Parallel = &val
			cmdGroup = composite.Group

		} else if command.Apply != nil {
			cmdGroup = command.Apply.Group
		}

		if cmdGroup != nil {
			setIsDefault(cmdGroup)
		}

	}

	//set defaults on the components

	components, err := d.Data.GetComponents(common.DevfileOptions{})

	if err != nil {
		return err
	}

	var endpoints []v1.Endpoint
	for i := range components {
		component := components[i]
		endpoints = nil

		if component.Container != nil {
			container := component.Container
			val := container.GetDedicatedPod()
			container.DedicatedPod = &val

			msVal := container.GetMountSources()
			container.MountSources = &msVal

			endpoints = container.Endpoints

		} else if component.Kubernetes != nil {
			endpoints = component.Kubernetes.Endpoints
			if devfileVersion != string(data.APISchemaVersion200) && devfileVersion != string(data.APISchemaVersion210) {
				val := component.Kubernetes.GetDeployByDefault()
				component.Kubernetes.DeployByDefault = &val
			}
		} else if component.Openshift != nil {
			endpoints = component.Openshift.Endpoints
			if devfileVersion != string(data.APISchemaVersion200) && devfileVersion != string(data.APISchemaVersion210) {
				val := component.Openshift.GetDeployByDefault()
				component.Openshift.DeployByDefault = &val
			}

		} else if component.Volume != nil && devfileVersion != string(data.APISchemaVersion200) {
			volume := component.Volume
			val := volume.GetEphemeral()
			volume.Ephemeral = &val

		} else if component.Image != nil { //we don't need to do a schema version check since Image in v2.2.0.  If used in older specs, a parser error would occur
			dockerImage := component.Image.Dockerfile
			if dockerImage != nil {
				val := dockerImage.GetRootRequired()
				dockerImage.RootRequired = &val
			}
			val := component.Image.GetAutoBuild()
			component.Image.AutoBuild = &val
		}

		if endpoints != nil {
			setEndpoints(endpoints)
		}
	}

	return nil
}

// setIsDefault sets the default value of CommandGroup.IsDefault if nil
func setIsDefault(cmdGroup *v1.CommandGroup) {
	val := cmdGroup.GetIsDefault()
	cmdGroup.IsDefault = &val
}

// setEndpoints sets the default value of Endpoint.Secure if nil
func setEndpoints(endpoints []v1.Endpoint) {
	for i := range endpoints {
		val := endpoints[i].GetSecure()
		endpoints[i].Secure = &val
	}
}

// parseKubeResourceFromURI iterate through all kubernetes & openshift components, and parse from uri and update the content to inlined field in devfileObj
func parseKubeResourceFromURI(devObj DevfileObj, devfileUtilsClient parserUtil.DevfileUtils) error {
	getKubeCompOptions := common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1.KubernetesComponentType,
		},
	}
	getOpenshiftCompOptions := common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1.OpenshiftComponentType,
		},
	}
	kubeComponents, err := devObj.Data.GetComponents(getKubeCompOptions)
	if err != nil {
		return err
	}
	openshiftComponents, err := devObj.Data.GetComponents(getOpenshiftCompOptions)
	if err != nil {
		return err
	}
	for _, kubeComp := range kubeComponents {
		if kubeComp.Kubernetes != nil && kubeComp.Kubernetes.Uri != "" {
			/* #nosec G601 -- not an issue, kubeComp is de-referenced in sequence*/
			err := convertK8sLikeCompUriToInlined(&kubeComp, devObj.Ctx, devfileUtilsClient)
			if err != nil {
				return errors.Wrapf(err, "failed to convert kubernetes uri to inlined for component '%s'", kubeComp.Name)
			}
			err = devObj.Data.UpdateComponent(kubeComp)
			if err != nil {
				return err
			}
		}
	}
	for _, openshiftComp := range openshiftComponents {
		if openshiftComp.Openshift != nil && openshiftComp.Openshift.Uri != "" {
			/* #nosec G601 -- not an issue, openshiftComp is de-referenced in sequence*/
			err := convertK8sLikeCompUriToInlined(&openshiftComp, devObj.Ctx, devfileUtilsClient)
			if err != nil {
				return errors.Wrapf(err, "failed to convert openshift uri to inlined for component '%s'", openshiftComp.Name)
			}
			err = devObj.Data.UpdateComponent(openshiftComp)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// convertK8sLikeCompUriToInlined read in kubernetes resources definition from uri and converts to kubernetest inlined field
func convertK8sLikeCompUriToInlined(component *v1.Component, d devfileCtx.DevfileCtx, devfileUtilsClient parserUtil.DevfileUtils) error {
	var uri string
	if component.Kubernetes != nil {
		uri = component.Kubernetes.Uri
	} else if component.Openshift != nil {
		uri = component.Openshift.Uri
	}
	data, err := getKubernetesDefinitionFromUri(uri, d, devfileUtilsClient)
	if err != nil {
		return err
	}
	if component.Kubernetes != nil {
		component.Kubernetes.Inlined = string(data)
		component.Kubernetes.Uri = ""
	} else if component.Openshift != nil {
		component.Openshift.Inlined = string(data)
		component.Openshift.Uri = ""
	}
	if component.Attributes == nil {
		component.Attributes = attributes.Attributes{}
	}
	component.Attributes.PutString(K8sLikeComponentOriginalURIKey, uri)

	return nil
}

// getKubernetesDefinitionFromUri read in kubernetes resources definition from uri and returns the raw content
func getKubernetesDefinitionFromUri(uri string, d devfileCtx.DevfileCtx, devfileUtilsClient parserUtil.DevfileUtils) ([]byte, error) {
	// validate URI
	err := validation.ValidateURI(uri)
	if err != nil {
		return nil, err
	}

	absoluteURL := strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://")
	var newUri string
	var data []byte
	// relative path on disk
	if !absoluteURL && d.GetAbsPath() != "" {
		newUri = path.Join(path.Dir(d.GetAbsPath()), uri)
		fs := d.GetFs()
		data, err = fs.ReadFile(newUri)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read kubernetes resources definition from path '%s'", newUri)
		}
	} else if absoluteURL || d.GetURL() != "" {
		if d.GetURL() != "" {
			// relative path to a URL
			u, err := url.Parse(d.GetURL())
			if err != nil {
				return nil, err
			}
			u.Path = path.Join(path.Dir(u.Path), uri)
			newUri = u.String()
		} else {
			// absolute URL address
			newUri = uri
		}
		params := util.HTTPRequestParams{URL: newUri}
		if d.GetToken() != "" {
			params.Token = d.GetToken()
		}
		data, err = devfileUtilsClient.DownloadInMemory(params)
		if err != nil {
			return nil, errors.Wrapf(err, "error getting kubernetes resources definition information")
		}
	} else {
		return nil, fmt.Errorf("error getting kubernetes resources definition information, unable to resolve the file uri: %v", uri)
	}
	return data, nil
}
