package catalog

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"
	"sigs.k8s.io/yaml"

	operatorv1alpha1 "github.com/openshift/api/operator/v1alpha1"

	"github.com/operator-framework/operator-registry/pkg/mirror"

	imgextract "github.com/openshift/oc/pkg/cli/image/extract"
	"github.com/openshift/oc/pkg/cli/image/imagesource"
	imagemanifest "github.com/openshift/oc/pkg/cli/image/manifest"
	imgmirror "github.com/openshift/oc/pkg/cli/image/mirror"
)

var (
	mirrorLong = templates.LongDesc(`
			Mirrors the contents of a catalog into a registry.

			This command will pull down an image containing a catalog database, extract it to disk, query it to find
      all of the images used in the manifests, and then mirror them to a target registry.

			By default, the database is extracted to a temporary directory, but can be saved locally via flags.

			An ImageContentSourcePolicy is written to a file that can be adedd to a cluster with access to the target 
			registry. This will configure the cluster to pull from the mirrors instead of the locations referenced in
			the operator manifests.

			A mapping.txt file is also created that is compatible with "oc image mirror". This may be used to further
			customize the mirroring configuration, but should not be needed in normal circumstances.
		`)
	mirrorExample = templates.Examples(`
# Mirror an operator-registry image and its contents to a registry
%[1]s quay.io/my/image:latest myregistry.com

# Configure a cluster to use a mirrored registry
oc apply -f manifests/imageContentSourcePolicy.yaml

# Edit the mirroring mappings and mirror with "oc image mirror" manually
%[1]s --manifests-only quay.io/my/image:latest myregistry.com
oc image mirror -f manifests/mapping.txt
`)
)

type MirrorCatalogOptions struct {
	*mirror.IndexImageMirrorerOptions
	genericclioptions.IOStreams

	DryRun       bool
	ManifestOnly bool
	DatabasePath string

	FromFileDir string
	FileDir     string

	SecurityOptions imagemanifest.SecurityOptions
	FilterOptions   imagemanifest.FilterOptions
	ParallelOptions imagemanifest.ParallelOptions

	SourceRef imagesource.TypedImageReference
	Dest      string
}

func NewMirrorCatalogOptions(streams genericclioptions.IOStreams) *MirrorCatalogOptions {
	return &MirrorCatalogOptions{
		IOStreams:                 streams,
		IndexImageMirrorerOptions: mirror.DefaultImageIndexMirrorerOptions(),
		ParallelOptions:           imagemanifest.ParallelOptions{MaxPerRegistry: 4},
	}
}

func NewMirrorCatalog(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewMirrorCatalogOptions(streams)

	cmd := &cobra.Command{
		Use:     "mirror SRC DEST",
		Short:   "mirror an operator-registry catalog",
		Long:    mirrorLong,
		Example: fmt.Sprintf(mirrorExample, "oc adm catalog mirror"),
		Run: func(cmd *cobra.Command, args []string) {
			kcmdutil.CheckErr(o.Complete(cmd, args))
			kcmdutil.CheckErr(o.Validate())
			kcmdutil.CheckErr(o.Run())
		},
	}
	flags := cmd.Flags()

	o.SecurityOptions.Bind(flags)
	o.FilterOptions.Bind(flags)
	o.ParallelOptions.Bind(flags)

	flags.StringVar(&o.ManifestDir, "to-manifests", "", "Local path to store manifests.")
	flags.StringVar(&o.DatabasePath, "path", "", "Specify an in-container to local path mapping for the database.")
	flags.BoolVar(&o.DryRun, "dry-run", o.DryRun, "Print the actions that would be taken and exit without writing to the destinations.")
	flags.BoolVar(&o.ManifestOnly, "manifests-only", o.ManifestOnly, "Calculate the manifests required for mirroring, but do not actually mirror image content.")
	flags.StringVar(&o.FileDir, "dir", o.FileDir, "The directory on disk that file:// images will be copied under.")
	flags.StringVar(&o.FromFileDir, "from-dir", o.FromFileDir, "The directory on disk that file:// images will be read from. Overrides --dir")
	return cmd
}

func (o *MirrorCatalogOptions) Complete(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("must specify source and dest")
	}
	src := args[0]
	dest := args[1]

	srcRef, err := imagesource.ParseReference(src)
	if err != nil {
		return err
	}
	o.SourceRef = srcRef
	o.Dest = dest

	if o.ManifestDir == "" {
		o.ManifestDir = o.SourceRef.Ref.Name + "-manifests"
	}

	if err := os.MkdirAll(o.ManifestDir, os.ModePerm); err != nil {
		return err
	}

	if o.DatabasePath == "" {
		tmpdir, err := ioutil.TempDir("", "")
		if err != nil {
			return err
		}
		o.DatabasePath = "/:" + tmpdir
	} else {
		dir := strings.Split(o.DatabasePath, ":")
		if len(dir) < 2 {
			return fmt.Errorf("invalid path")
		}
		if err := os.MkdirAll(path.Dir(dir[1]), os.ModePerm); err != nil {
			return err
		}
	}

	var mirrorer mirror.ImageMirrorerFunc
	mirrorer = func(mapping map[string]string) error {
		for from, to := range mapping {
			fromRef, err := imagesource.ParseSourceReference(from, nil)
			if err != nil {
				klog.Warningf("couldn't parse %s, skipping mirror: %v", from, err)
				continue
			}

			// remove destination digest if present
			toRef, err := imagesource.ParseReference(to)
			if err != nil {
				klog.Warningf("couldn't parse %s, skipping mirror: %v", to, err)
				continue
			}
			if toRef.Type == imagesource.DestinationRegistry && len(toRef.Ref.ID) != 0 {
				to = toRef.Ref.AsRepository().String()
			}

			toRef, err = imagesource.ParseDestinationReference(to)
			if err != nil {
				klog.Warningf("couldn't parse %s, skipping mirror: %v", to, err)
				continue
			}

			a := imgmirror.NewMirrorImageOptions(o.IOStreams)
			a.SkipMissing = true
			a.DryRun = o.DryRun
			a.SecurityOptions = o.SecurityOptions
			a.FilterOptions = o.FilterOptions
			a.ParallelOptions = o.ParallelOptions
			a.KeepManifestList = true
			a.Mappings = []imgmirror.Mapping{{
				Source:      fromRef[0],
				Destination: toRef,
			}}
			if err := a.Validate(); err != nil {
				klog.Warningf("error configuring image mirroring: %v", err)
			}
			if err := a.Run(); err != nil {
				klog.Warningf("error mirroring image: %v", err)
			}
		}
		return nil
	}

	if o.ManifestOnly {
		mirrorer = func(mapping map[string]string) error {
			return nil
		}
	}

	o.ImageMirrorer = mirrorer

	var extractor mirror.DatabaseExtractorFunc = func(from string) (string, error) {
		e := imgextract.NewOptions(o.IOStreams)
		e.SecurityOptions = o.SecurityOptions
		e.FilterOptions = o.FilterOptions
		e.ParallelOptions = o.ParallelOptions
		e.FileDir = o.FileDir
		if len(o.FromFileDir) > 0 {
			e.FileDir = o.FromFileDir
		}
		e.Paths = []string{o.DatabasePath}
		e.Confirm = true
		if err := e.Complete(cmd, []string{o.SourceRef.String()}); err != nil {
			return "", err
		}
		if err := e.Validate(); err != nil {
			return "", err
		}
		if err := e.Run(); err != nil {
			return "", err
		}
		if len(e.Mappings) < 1 {
			return "", fmt.Errorf("couldn't extract database")
		}
		klog.Infof("wrote database to %s", path.Join(e.Mappings[0].To, "bundles.db"))
		return path.Join(e.Mappings[0].To, "bundles.db"), nil
	}
	o.DatabaseExtractor = extractor
	return nil
}

func (o *MirrorCatalogOptions) Validate() error {
	if o.DatabasePath == "" {
		return fmt.Errorf("must specify path for database")
	}
	if o.ManifestDir == "" {
		return fmt.Errorf("must specify path for manifests")
	}
	return nil
}

func (o *MirrorCatalogOptions) Run() error {
	indexMirrorer, err := mirror.NewIndexImageMirror(o.IndexImageMirrorerOptions.ToOption(),
		mirror.WithSource(o.SourceRef.String()),
		mirror.WithDest(o.Dest),
	)
	if err != nil {
		return err
	}
	mapping, err := indexMirrorer.Mirror()
	if err != nil {
		klog.Warningf("errors during mirroring. the full contents of the catalog may not have been mirrored: %s", err.Error())
	}

	return WriteManifests(o.SourceRef.Ref.Name, o.ManifestDir, mapping)
}

func WriteManifests(name, dir string, mapping map[string]string) error {
	f, err := os.Create(path.Join(dir, "mapping.txt"))
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			klog.Warningf("error closing file")
		}
	}()

	icsp := operatorv1alpha1.ImageContentSourcePolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: operatorv1alpha1.GroupVersion.String(),
			Kind:       "ImageContentSourcePolicy"},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: operatorv1alpha1.ImageContentSourcePolicySpec{
			RepositoryDigestMirrors: []operatorv1alpha1.RepositoryDigestMirrors{},
		},
	}

	for k, v := range mapping {
		fromRef, err := imagesource.ParseReference(k)
		if err != nil {
			klog.Warningf("error parsing source reference for %s", k)
			continue
		}
		toRef, err := imagesource.ParseReference(v)
		if err != nil {
			klog.Warningf("error parsing target reference for %s", v)
			continue
		}
		icsp.Spec.RepositoryDigestMirrors = append(icsp.Spec.RepositoryDigestMirrors, operatorv1alpha1.RepositoryDigestMirrors{
			Source:  fromRef.Ref.AsRepository().String(),
			Mirrors: []string{toRef.Ref.AsRepository().String()},
		})

		// omit digest from target if digest exists
		to := v
		if len(toRef.Ref.ID) > 0 {
			to = toRef.Ref.AsRepository().String()
		}
		if _, err := f.WriteString(fmt.Sprintf("%s=%s\n", k, to)); err != nil {
			return err
		}
	}

	// Create an unstructured object for removing creationTimestamp
	unstructuredObj := unstructured.Unstructured{}
	unstructuredObj.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&icsp)
	if err != nil {
		return fmt.Errorf("ToUnstructured error: %v", err)
	}
	delete(unstructuredObj.Object["metadata"].(map[string]interface{}), "creationTimestamp")

	icspExample, err := yaml.Marshal(unstructuredObj.Object)
	if err != nil {
		return fmt.Errorf("Unable to marshal ImageContentSourcePolicy example yaml: %v", err)
	}

	if err := ioutil.WriteFile(path.Join(dir, "imageContentSourcePolicy.yaml"), icspExample, os.ModePerm); err != nil {
		return fmt.Errorf("error writing ImageContentSourcePolicy")
	}
	klog.Infof("wrote mirroring manifests to %s", dir)
	return nil
}
