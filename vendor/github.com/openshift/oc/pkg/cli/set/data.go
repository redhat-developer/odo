package set

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/klog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/dynamic"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	dataLong = templates.LongDesc(`
		Add, update, or remove data keys from secrets and config maps

		Secrets and config maps allow users to store keys and values that can be passed into
		pods or loaded by other Kubernetes resources. This command lets you set or remove keys
		from those objects from arguments or files. Use the --from-file flag when you want to
		load the contents of a file or directory, or pass command line arguments that contain
		either a KEY=VALUE pair (to set a value) or KEY- (to remove that key).

		You may also use this command as part of a chain to modify an object before submitting
		to the server with the --local and --dry-run flags. This allows you to update local
		resources to contain additional keys.

		Experimental: This command is under active development and may change without notice.`)

	dataExample = templates.Examples(`
	  # Set the 'password' key of a secret
	  %[1]s data secret/foo password=this_is_secret

	  # Remove the 'password' key from a secret
	  %[1]s data secret/foo password-

	  # Update the 'haproxy.conf' key of a config map from a file on disk
	  %[1]s data configmap/bar --from-file=../haproxy.conf

	  # Update a secret with the contents of a directory, one key per file
	  %[1]s data secret/foo --from-file=secret-dir`)
)

type DataOptions struct {
	PrintFlags *genericclioptions.PrintFlags

	SetData    map[string]string
	RemoveData []string
	// FileSources is supported to be consistent with kubectl create, and is
	// checked against SetData for duplicates.
	FileSources []string
	// LiteralSources is supported to be consistent with kubectl create, and is
	// checked against SetData for duplicates.
	LiteralSources []string

	Selector string
	All      bool
	Local    bool

	Mapper              meta.RESTMapper
	Client              dynamic.Interface
	Printer             printers.ResourcePrinter
	Builder             func() *resource.Builder
	Encoder             runtime.Encoder
	Namespace           string
	ExplicitNamespace   bool
	UpdateDataForObject func(obj runtime.Object, fn func(data map[string][]byte) error) (bool, error)
	Command             []string
	Resources           []string
	DryRun              bool

	FlagSet func(string) bool

	resource.FilenameOptions
	genericclioptions.IOStreams
}

func NewDataOptions(streams genericclioptions.IOStreams) *DataOptions {
	return &DataOptions{
		PrintFlags: genericclioptions.NewPrintFlags("data updated").WithTypeSetter(scheme.Scheme),
		IOStreams:  streams,
	}
}

// NewCmdData implements the set data command
func NewCmdData(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewDataOptions(streams)
	cmd := &cobra.Command{
		Use:     "data RESOURCE/NAME [KEY=VALUE|KEY- ...] [--from-file=file|dir|key=path]",
		Short:   "Update the data within a config map or secret",
		Long:    dataLong,
		Example: fmt.Sprintf(dataExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			kcmdutil.CheckErr(o.Complete(f, cmd, args))
			kcmdutil.CheckErr(o.Validate())
			kcmdutil.CheckErr(o.Run())
		},
	}
	usage := "to use to edit the resource"
	kcmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, usage)
	cmd.Flags().StringVarP(&o.Selector, "selector", "l", o.Selector, "Selector (label query) to filter on")
	cmd.Flags().BoolVar(&o.All, "all", o.All, "If true, select all resources in the namespace of the specified resource types")
	cmd.Flags().BoolVar(&o.Local, "local", o.Local, "If true, set image will NOT contact api-server but run locally.")
	cmd.Flags().StringSliceVar(
		&o.FileSources,
		"from-file",
		[]string{},
		"Specify a file using its file path, in which case the file basename will be used as the key"+
			"or optionally with a key and file path, in which case the given key will be used.  Specifying a "+
			"directory will iterate each named file in the directory whose basename is a valid secret key.")
	cmd.Flags().StringArrayVar(
		&o.LiteralSources,
		"from-literal",
		[]string{},
		"Specify a key and literal value to set (i.e. mykey=somevalue)")

	o.PrintFlags.AddFlags(cmd)
	kcmdutil.AddDryRunFlag(cmd)

	return cmd
}

func (o *DataOptions) Complete(f kcmdutil.Factory, cmd *cobra.Command, args []string) error {
	o.Resources = args
	if i := cmd.ArgsLenAtDash(); i != -1 {
		o.Resources = args[:i]
		o.Command = args[i:]
	}
	if len(o.Filenames) == 0 && len(args) < 1 {
		return kcmdutil.UsageErrorf(cmd, "one or more resources must be specified as <resource> <name> or <resource>/<name>")
	}

	var err error
	o.Namespace, o.ExplicitNamespace, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	o.Mapper, err = f.ToRESTMapper()
	if err != nil {
		return err
	}
	o.Builder = f.NewBuilder
	o.UpdateDataForObject = updateDataForObject

	o.DryRun = kcmdutil.GetDryRunFlag(cmd)
	if o.DryRun {
		o.PrintFlags.Complete("%s (dry run)")
	}
	o.Printer, err = o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}

	clientConfig, err := f.ToRESTConfig()
	if err != nil {
		return err
	}
	o.Client, err = dynamic.NewForConfig(clientConfig)
	if err != nil {
		return err
	}

	args, dataArgs, err := kcmdutil.GetResourcesAndPairs(args, "data")
	if err != nil {
		return err
	}

	o.Resources = args

	o.SetData, o.RemoveData, err = kcmdutil.ParsePairs(dataArgs, "data", true)
	if err != nil {
		return err
	}

	kv, err := keyValuesFromFileSources(o.FileSources)
	if err != nil {
		return err
	}
	for k, v := range kv {
		if _, ok := o.SetData[k]; ok && k != v {
			return fmt.Errorf("cannot set key %q in both argument and flag", k)
		}
		o.SetData[k] = v
	}
	kv, err = keyValuesFromLiteralSources(o.LiteralSources)
	if err != nil {
		return err
	}
	for k, v := range kv {
		if _, ok := o.SetData[k]; ok && k != v {
			return fmt.Errorf("cannot set key %q in both argument and flag", k)
		}
		o.SetData[k] = v
	}

	return nil
}

func (o *DataOptions) Validate() error {
	if len(o.SetData) == 0 && len(o.RemoveData) == 0 {
		return fmt.Errorf("must add, update, or remove at least one key from this object")
	}
	for _, remove := range o.RemoveData {
		if _, ok := o.SetData[remove]; ok {
			return fmt.Errorf("you may not remove and set the key %q in the same invocation", remove)
		}
	}
	return nil
}

func (o *DataOptions) Run() error {
	b := o.Builder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		LocalParam(o.Local).
		ContinueOnError().
		NamespaceParam(o.Namespace).DefaultNamespace().
		FilenameParam(o.ExplicitNamespace, &o.FilenameOptions).
		Flatten()

	if !o.Local {
		b = b.
			LabelSelectorParam(o.Selector).
			ResourceTypeOrNameArgs(o.All, o.Resources...).
			Latest()
	}

	singleItemImplied := false
	infos, err := b.Do().IntoSingleItemImplied(&singleItemImplied).Infos()
	if err != nil {
		return err
	}

	allErrs := []error{}

	patches := CalculatePatchesExternal(infos, func(info *resource.Info) (bool, error) {
		changed := false
		valid, err := o.UpdateDataForObject(info.Object, func(data map[string][]byte) error {
			for k, v := range o.SetData {
				if existing, ok := data[k]; ok && string(existing) == v {
					continue
				}
				data[k] = []byte(v)
				changed = true
			}
			for _, k := range o.RemoveData {
				if _, ok := data[k]; !ok {
					continue
				}
				delete(data, k)
				changed = true
			}
			return nil
		})
		if valid && !changed {
			if o.Local || o.DryRun {
				if err := o.Printer.PrintObj(info.Object, o.Out); err != nil {
					allErrs = append(allErrs, err)
				}
			} else {
				fmt.Fprintf(o.ErrOut, "info: %s was not changed\n", info.Name)
			}
		}
		return valid && changed, err
	})

	for _, patch := range patches {
		info := patch.Info
		name := getObjectName(info)
		if patch.Err != nil {
			allErrs = append(allErrs, fmt.Errorf("error: %s %v\n", name, patch.Err))
			continue
		}

		if string(patch.Patch) == "{}" || len(patch.Patch) == 0 {
			klog.V(1).Infof("info: %s was not changed\n", name)
			continue
		}

		if o.Local || o.DryRun {
			if err := o.Printer.PrintObj(info.Object, o.Out); err != nil {
				allErrs = append(allErrs, err)
			}
			continue
		}

		actual, err := o.Client.Resource(info.Mapping.Resource).Namespace(info.Namespace).Patch(info.Name, types.StrategicMergePatchType, patch.Patch, metav1.PatchOptions{})
		if err != nil {
			allErrs = append(allErrs, err)
			continue
		}

		if err := o.Printer.PrintObj(actual, o.Out); err != nil {
			allErrs = append(allErrs, err)
		}
	}
	return utilerrors.NewAggregate(allErrs)
}

func isBinary(data []byte) bool {
	for _, b := range data {
		if b > 127 || (b < 32 && b != 9 && b != 10 && b != 13) {
			return true
		}
	}
	return false
}

func updateDataForObject(obj runtime.Object, fn func(data map[string][]byte) error) (bool, error) {
	switch t := obj.(type) {
	case *corev1.Secret:
		if t.Data == nil {
			t.Data = make(map[string][]byte)
		}
		return true, fn(t.Data)
		// ReplicationController
	case *corev1.ConfigMap:
		if t.BinaryData == nil {
			t.BinaryData = make(map[string][]byte)
		}
		for k, v := range t.Data {
			t.BinaryData[k] = []byte(v)
		}
		t.Data = nil
		if err := fn(t.BinaryData); err != nil {
			return true, err
		}
		for k, v := range t.BinaryData {
			if isBinary(v) {
				continue
			}
			delete(t.BinaryData, k)
			if t.Data == nil {
				t.Data = make(map[string]string)
			}
			t.Data[k] = string(v)
		}
		return true, nil

		// Deployment
	default:
		return false, fmt.Errorf("the object is not a config map or secret and cannot have data updated: %T", t)
	}
}

func keyValuesFromFileSources(fileSources []string) (map[string]string, error) {
	data := make(map[string]string)
	for _, fileSource := range fileSources {
		keyName, filePath, err := util.ParseFileSource(fileSource)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(filePath)
		if err != nil {
			switch err := err.(type) {
			case *os.PathError:
				return nil, fmt.Errorf("error reading %s: %v", filePath, err.Err)
			default:
				return nil, fmt.Errorf("error reading %s: %v", filePath, err)
			}
		}
		if info.IsDir() {
			if strings.Contains(fileSource, "=") {
				return nil, fmt.Errorf("cannot give a key name for a directory path")
			}
			fileList, err := ioutil.ReadDir(filePath)
			if err != nil {
				return nil, fmt.Errorf("error listing files in %s: %v", filePath, err)
			}
			for _, item := range fileList {
				itemPath := filepath.Join(filePath, item.Name())
				if item.Mode().IsRegular() {
					keyName = item.Name()
					if err = addKeyFromFileToMap(data, keyName, itemPath); err != nil {
						return nil, err
					}
				}
			}
		} else {
			if err := addKeyFromFileToMap(data, keyName, filePath); err != nil {
				return nil, err
			}
		}
	}
	return data, nil
}

func addKeyFromFileToMap(data map[string]string, keyName, filePath string) error {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return addKeyFromLiteralToMap(data, keyName, contents)
}

func addKeyFromLiteralToMap(data map[string]string, keyName string, contents []byte) error {
	if _, entryExists := data[keyName]; entryExists {
		return fmt.Errorf("cannot add key %s, another key by that name already exists", keyName)
	}
	data[keyName] = string(contents)
	return nil
}

func keyValuesFromLiteralSources(sources []string) (map[string]string, error) {
	kvs := make(map[string]string)
	for _, s := range sources {
		k, v, err := parseLiteralSource(s)
		if err != nil {
			return nil, err
		}
		kvs[k] = v
	}
	return kvs, nil
}

// parseLiteralSource parses the source key=val pair into its component pieces.
// This functionality is distinguished from strings.SplitN(source, "=", 2) since
// it returns an error in the case of empty keys, values, or a missing equals sign.
func parseLiteralSource(source string) (keyName, value string, err error) {
	// leading equal is invalid
	if strings.Index(source, "=") == 0 {
		return "", "", fmt.Errorf("invalid literal source %v, expected key=value", source)
	}
	// split after the first equal (so values can have the = character)
	items := strings.SplitN(source, "=", 2)
	if len(items) != 2 {
		return "", "", fmt.Errorf("invalid literal source %v, expected key=value", source)
	}
	return items[0], strings.Trim(items[1], "\"'"), nil
}
