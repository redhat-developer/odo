package release

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	digest "github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"
	"sigs.k8s.io/yaml"

	imagev1 "github.com/openshift/api/image/v1"
	operatorv1alpha1 "github.com/openshift/api/operator/v1alpha1"
	imageclient "github.com/openshift/client-go/image/clientset/versioned"
	"github.com/openshift/library-go/pkg/image/dockerv1client"
	imagereference "github.com/openshift/library-go/pkg/image/reference"
	"github.com/openshift/oc/pkg/cli/image/extract"
	"github.com/openshift/oc/pkg/cli/image/imagesource"
	imagemanifest "github.com/openshift/oc/pkg/cli/image/manifest"
	"github.com/openshift/oc/pkg/cli/image/mirror"
)

// NewMirrorOptions creates the options for mirroring a release.
func NewMirrorOptions(streams genericclioptions.IOStreams) *MirrorOptions {
	return &MirrorOptions{
		IOStreams:       streams,
		ParallelOptions: imagemanifest.ParallelOptions{MaxPerRegistry: 6},
	}
}

// NewMirror creates a command to mirror an existing release.
//
// Example command to mirror a release to a local repository to work offline
//
// $ oc adm release mirror \
//     --from=registry.svc.ci.openshift.org/openshift/v4.0 \
//     --to=mycompany.com/myrepository/repo
//
func NewMirror(f kcmdutil.Factory, parentName string, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewMirrorOptions(streams)
	cmd := &cobra.Command{
		Use:   "mirror",
		Short: "Mirror a release to a different image registry location",
		Long: templates.LongDesc(`
			Mirror an OpenShift release image to another registry

			Copies the images and update payload for a given release from one registry to another.
			By default this command will not alter the payload and will print out the configuration
			that must be applied to a cluster to use the mirror, but you may opt to rewrite the
			update to point to the new location and lose the cryptographic integrity of the update.

			The common use for this command is to mirror a specific OpenShift release version to
			a private registry for use in a disconnected or offline context. The command copies all
			images that are part of a release into the target repository and then prints the
			correct information to give to OpenShift to use that content offline. An alternate mode
			is to specify --to-image-stream, which imports the images directly into an OpenShift
			image stream.

			You may use --to-dir to specify a directory to download release content into, and add
			the file:// prefix to the --to flag. The command will print the 'oc image mirror' command
			that can be used to upload the release to another registry.
		`),
		Example: templates.Examples(`
			# Perform a dry run showing what would be mirrored, including the mirror objects
			%[1]s 4.2.2 --to myregistry.local/openshift/release --dry-run

			# Mirror a release into the current directory
			%[1]s 4.2.2 --to file://openshift/release

			# Mirror a release to another directory in the default location
			%[1]s 4.2.2 --to-dir /tmp/releases

			# Upload a release from the current directory to another server
			%[1]s --from file://openshift/release --to myregistry.com/openshift/release
			`),
		Run: func(cmd *cobra.Command, args []string) {
			kcmdutil.CheckErr(o.Complete(cmd, f, args))
			kcmdutil.CheckErr(o.Run())
		},
	}
	flags := cmd.Flags()
	o.SecurityOptions.Bind(flags)
	o.ParallelOptions.Bind(flags)

	flags.StringVar(&o.From, "from", o.From, "Image containing the release payload.")
	flags.StringVar(&o.To, "to", o.To, "An image repository to push to.")
	flags.StringVar(&o.ToImageStream, "to-image-stream", o.ToImageStream, "An image stream to tag images into.")
	flags.StringVar(&o.FromDir, "from-dir", o.ToDir, "A directory to import images from.")
	flags.StringVar(&o.ToDir, "to-dir", o.ToDir, "A directory to export images to.")
	flags.BoolVar(&o.ToMirror, "to-mirror", o.ToMirror, "Output the mirror mappings instead of mirroring.")
	flags.BoolVar(&o.DryRun, "dry-run", o.DryRun, "Display information about the mirror without actually executing it.")

	flags.BoolVar(&o.SkipRelease, "skip-release-image", o.SkipRelease, "Do not push the release image.")
	flags.StringVar(&o.ToRelease, "to-release-image", o.ToRelease, "Specify an alternate locations for the release image instead as tag 'release' in --to")
	return cmd
}

type MirrorOptions struct {
	genericclioptions.IOStreams

	SecurityOptions imagemanifest.SecurityOptions
	ParallelOptions imagemanifest.ParallelOptions

	From    string
	FromDir string

	To            string
	ToImageStream string

	// modifies the targets
	ToRelease   string
	SkipRelease bool

	ToMirror bool
	ToDir    string

	DryRun                        bool
	PrintImageContentInstructions bool

	ClientFn func() (imageclient.Interface, string, error)

	ImageStream *imagev1.ImageStream
	TargetFn    func(component string) imagereference.DockerImageReference
}

func (o *MirrorOptions) Complete(cmd *cobra.Command, f kcmdutil.Factory, args []string) error {
	switch {
	case len(args) == 0 && len(o.From) == 0:
		return fmt.Errorf("must specify a release image with --from")
	case len(args) == 1 && len(o.From) == 0:
		o.From = args[0]
	case len(args) == 1 && len(o.From) > 0:
		return fmt.Errorf("you may not specify an argument and --from")
	case len(args) > 1:
		return fmt.Errorf("only one argument is accepted")
	}

	args, err := findArgumentsFromCluster(f, []string{o.From})
	if err != nil {
		return err
	}
	if len(args) != 1 {
		return fmt.Errorf("only one release image may be mirrored")
	}
	o.From = args[0]

	o.ClientFn = func() (imageclient.Interface, string, error) {
		cfg, err := f.ToRESTConfig()
		if err != nil {
			return nil, "", err
		}
		client, err := imageclient.NewForConfig(cfg)
		if err != nil {
			return nil, "", err
		}
		ns, _, err := f.ToRawKubeConfigLoader().Namespace()
		if err != nil {
			return nil, "", err
		}
		return client, ns, nil
	}
	o.PrintImageContentInstructions = true
	return nil
}

const replaceComponentMarker = "X-X-X-X-X-X-X"
const replaceVersionMarker = "V-V-V-V-V-V-V"

func (o *MirrorOptions) Run() error {
	if len(o.From) == 0 && o.ImageStream == nil {
		return fmt.Errorf("must specify a release image with --from")
	}

	outputs := 0
	if len(o.To) > 0 {
		outputs++
	}
	if len(o.ToImageStream) > 0 {
		outputs++
	}
	if len(o.ToDir) > 0 {
		if outputs == 0 {
			outputs++
		}
	}
	if o.ToMirror {
		if outputs == 0 {
			outputs++
		}
	}
	if outputs != 1 {
		return fmt.Errorf("must specify an image repository or image stream to mirror the release to")
	}

	if o.SkipRelease && len(o.ToRelease) > 0 {
		return fmt.Errorf("--skip-release-image and --to-release-image may not both be specified")
	}

	var recreateRequired bool
	var hasPrefix bool
	var targetFn func(name string) imagesource.TypedImageReference
	var dst string
	if len(o.ToImageStream) > 0 {
		dst = imagereference.DockerImageReference{
			Registry:  "example.com",
			Namespace: "somenamespace",
			Name:      "mirror",
		}.Exact()
	} else {
		dst = o.To
	}
	if len(dst) == 0 {
		if len(o.ToDir) > 0 {
			dst = "file://openshift/release"
		} else {
			dst = "openshift/release"
		}
	}

	var toDisk bool
	var version string
	if strings.Contains(dst, "${component}") {
		format := strings.Replace(dst, "${component}", replaceComponentMarker, -1)
		format = strings.Replace(format, "${version}", replaceVersionMarker, -1)
		dstRef, err := imagesource.ParseReference(format)
		if err != nil {
			return fmt.Errorf("--to must be a valid image reference: %v", err)
		}
		toDisk = dstRef.Type == imagesource.DestinationFile
		targetFn = func(name string) imagesource.TypedImageReference {
			if len(name) == 0 {
				name = "release"
			}
			value := strings.Replace(dst, "${component}", name, -1)
			value = strings.Replace(value, "${version}", version, -1)
			ref, err := imagesource.ParseReference(value)
			if err != nil {
				klog.Fatalf("requested component %q could not be injected into %s: %v", name, dst, err)
			}
			return ref
		}
		replaceCount := strings.Count(dst, "${component}")
		recreateRequired = replaceCount > 1 || (replaceCount == 1 && !strings.Contains(dstRef.Ref.Tag, replaceComponentMarker))

	} else {
		ref, err := imagesource.ParseReference(dst)
		if err != nil {
			return fmt.Errorf("--to must be a valid image repository: %v", err)
		}
		toDisk = ref.Type == imagesource.DestinationFile
		if len(ref.Ref.ID) > 0 || len(ref.Ref.Tag) > 0 {
			return fmt.Errorf("--to must be to an image repository and may not contain a tag or digest")
		}
		targetFn = func(name string) imagesource.TypedImageReference {
			copied := ref
			if len(name) > 0 {
				copied.Ref.Tag = fmt.Sprintf("%s-%s", version, name)
			} else {
				copied.Ref.Tag = version
			}
			return copied
		}
		hasPrefix = true
	}

	o.TargetFn = func(name string) imagereference.DockerImageReference {
		ref := targetFn(name)
		return ref.Ref
	}

	if recreateRequired {
		return fmt.Errorf("when mirroring to multiple repositories, use the new release command with --from-release and --mirror")
	}

	verifier := imagemanifest.NewVerifier()
	is := o.ImageStream
	if is == nil {
		o.ImageStream = &imagev1.ImageStream{}
		is = o.ImageStream
		// load image references
		buf := &bytes.Buffer{}
		extractOpts := NewExtractOptions(genericclioptions.IOStreams{Out: buf, ErrOut: o.ErrOut})
		extractOpts.SecurityOptions = o.SecurityOptions
		extractOpts.ImageMetadataCallback = func(m *extract.Mapping, dgst, contentDigest digest.Digest, config *dockerv1client.DockerImageConfig) {
			verifier.Verify(dgst, contentDigest)
		}
		extractOpts.From = o.From
		extractOpts.File = "image-references"
		if err := extractOpts.Run(); err != nil {
			return fmt.Errorf("unable to retrieve release image info: %v", err)
		}
		if err := json.Unmarshal(buf.Bytes(), &is); err != nil {
			return fmt.Errorf("unable to load image-references from release payload: %v", err)
		}
		if is.Kind != "ImageStream" || is.APIVersion != "image.openshift.io/v1" {
			return fmt.Errorf("unrecognized image-references in release payload")
		}
		if !verifier.Verified() {
			err := fmt.Errorf("the release image failed content verification and may have been tampered with")
			if !o.SecurityOptions.SkipVerification {
				return err
			}
			fmt.Fprintf(o.ErrOut, "warning: %v\n", err)
		}
	}
	version = is.Name

	// sourceFn is given a chance to rewrite source mappings
	sourceFn := func(ref imagesource.TypedImageReference) imagesource.TypedImageReference {
		return ref
	}
	var mappings []mirror.Mapping
	if len(o.From) > 0 {
		src := o.From
		srcRef, err := imagesource.ParseReference(src)
		if err != nil {
			return fmt.Errorf("invalid --from: %v", err)
		}

		// if the source ref is a file type, provide a function that checks the local file store for a given manifest
		// before continuing, to allow mirroring an entire release to disk in a single file://REPO.
		if srcRef.Type == imagesource.DestinationFile {
			if repo, err := (&imagesource.Options{FileDir: o.FromDir}).Repository(context.TODO(), srcRef); err == nil {
				sourceFn = func(ref imagesource.TypedImageReference) imagesource.TypedImageReference {
					if ref.Type == imagesource.DestinationFile || len(ref.Ref.ID) == 0 {
						return ref
					}
					manifests, err := repo.Manifests(context.TODO())
					if err != nil {
						klog.V(2).Infof("Unable to get local manifest service: %v", err)
						return ref
					}
					ok, err := manifests.Exists(context.TODO(), digest.Digest(ref.Ref.ID))
					if err != nil {
						klog.V(2).Infof("Unable to get check for local manifest: %v", err)
						return ref
					}
					if !ok {
						return ref
					}
					updated := srcRef
					updated.Ref.Tag = ""
					updated.Ref.ID = ref.Ref.ID
					klog.V(2).Infof("Rewrote %s to %s", ref, updated)
					return updated
				}
			} else {
				klog.V(2).Infof("Unable to build local file lookup: %v", err)
			}
		}

		if len(o.ToRelease) > 0 {
			dstRef, err := imagesource.ParseReference(o.ToRelease)
			if err != nil {
				return fmt.Errorf("invalid --to-release-image: %v", err)
			}
			mappings = append(mappings, mirror.Mapping{
				Source:      srcRef,
				Destination: dstRef,
				Name:        o.ToRelease,
			})
		} else if !o.SkipRelease {
			dstRef := targetFn("")
			mappings = append(mappings, mirror.Mapping{
				Source:      srcRef,
				Destination: dstRef,
				Name:        "release",
			})
		}
	}

	repositories := make(map[string]struct{})

	// build the mapping list for mirroring and rewrite if necessary
	for i := range is.Spec.Tags {
		tag := &is.Spec.Tags[i]
		if tag.From == nil || tag.From.Kind != "DockerImage" {
			continue
		}
		from, err := imagereference.Parse(tag.From.Name)
		if err != nil {
			return fmt.Errorf("release tag %q is not valid: %v", tag.Name, err)
		}
		if len(from.Tag) > 0 || len(from.ID) == 0 {
			return fmt.Errorf("image-references should only contain pointers to images by digest: %s", tag.From.Name)
		}

		// Allow mirror refs to be sourced locally
		srcMirrorRef := imagesource.TypedImageReference{Ref: from, Type: imagesource.DestinationRegistry}
		srcMirrorRef = sourceFn(srcMirrorRef)

		// Create a unique map of repos as keys
		currentRepo := from.AsRepository().String()
		repositories[currentRepo] = struct{}{}

		dstMirrorRef := targetFn(tag.Name)
		mappings = append(mappings, mirror.Mapping{
			Source:      srcMirrorRef,
			Destination: dstMirrorRef,
			Name:        tag.Name,
		})
		klog.V(2).Infof("Mapping %#v", mappings[len(mappings)-1])

		dstRef := targetFn(tag.Name)
		dstRef.Ref.Tag = ""
		dstRef.Ref.ID = from.ID
		tag.From.Name = dstRef.Ref.Exact()
	}

	if len(mappings) == 0 {
		fmt.Fprintf(o.ErrOut, "warning: Release image contains no image references - is this a valid release?\n")
	}

	if o.ToMirror {
		for _, mapping := range mappings {
			fmt.Fprintf(o.Out, "%s %s\n", mapping.Source.String(), mapping.Destination.String())
		}
		return nil
	}

	if len(o.ToImageStream) > 0 {
		remaining := make(map[string]mirror.Mapping)
		for _, mapping := range mappings {
			remaining[mapping.Name] = mapping
		}
		client, ns, err := o.ClientFn()
		if err != nil {
			return err
		}
		hasErrors := make(map[string]error)
		maxPerIteration := 12

		for retries := 4; (len(remaining) > 0 || len(hasErrors) > 0) && retries > 0; {
			if len(remaining) == 0 {
				for _, mapping := range mappings {
					if _, ok := hasErrors[mapping.Name]; ok {
						remaining[mapping.Name] = mapping
						delete(hasErrors, mapping.Name)
					}
				}
				retries--
			}
			err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				isi := &imagev1.ImageStreamImport{
					ObjectMeta: metav1.ObjectMeta{
						Name: o.ToImageStream,
					},
					Spec: imagev1.ImageStreamImportSpec{
						Import: !o.DryRun,
					},
				}
				for _, mapping := range remaining {
					if mapping.Source.Type != imagesource.DestinationRegistry {
						return fmt.Errorf("source mapping %s must point to a registry", mapping.Source)
					}
					isi.Spec.Images = append(isi.Spec.Images, imagev1.ImageImportSpec{
						From: corev1.ObjectReference{
							Kind: "DockerImage",
							Name: mapping.Source.Ref.Exact(),
						},
						To: &corev1.LocalObjectReference{
							Name: mapping.Name,
						},
					})
					if len(isi.Spec.Images) > maxPerIteration {
						break
					}
				}

				// use RESTClient directly here to be able to extend request timeout
				result := &imagev1.ImageStreamImport{}
				if err := client.ImageV1().RESTClient().Post().
					Namespace(ns).
					Resource(imagev1.Resource("imagestreamimports").Resource).
					Body(isi).
					// this instructs the api server to allow our request to take up to an hour - chosen as a high boundary
					Timeout(3 * time.Minute).
					Do().
					Into(result); err != nil {
					return err
				}

				for i, image := range result.Status.Images {
					name := result.Spec.Images[i].To.Name
					klog.V(4).Infof("Import result for %s: %#v", name, image.Status)
					if image.Status.Status == metav1.StatusSuccess {
						delete(remaining, name)
						delete(hasErrors, name)
					} else {
						delete(remaining, name)
						err := errors.FromObject(&image.Status)
						hasErrors[name] = err
						klog.V(2).Infof("Failed to import %s as tag %s: %v", remaining[name].Source, name, err)
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		if len(hasErrors) > 0 {
			var messages []string
			for k, v := range hasErrors {
				messages = append(messages, fmt.Sprintf("%s: %v", k, v))
			}
			sort.Strings(messages)
			if len(messages) == 1 {
				return fmt.Errorf("unable to import a release image: %s", messages[0])
			}
			return fmt.Errorf("unable to import some release images:\n* %s", strings.Join(messages, "\n* "))
		}

		fmt.Fprintf(os.Stderr, "Mirrored %d images to %s/%s\n", len(mappings), ns, o.ToImageStream)
		return nil
	}

	fmt.Fprintf(os.Stderr, "info: Mirroring %d images to %s ...\n", len(mappings), dst)
	var lock sync.Mutex
	opts := mirror.NewMirrorImageOptions(genericclioptions.IOStreams{Out: o.Out, ErrOut: o.ErrOut})
	opts.SecurityOptions = o.SecurityOptions
	opts.ParallelOptions = o.ParallelOptions
	opts.Mappings = mappings
	opts.FileDir = o.ToDir
	opts.DryRun = o.DryRun
	opts.ManifestUpdateCallback = func(registry string, manifests map[digest.Digest]digest.Digest) error {
		lock.Lock()
		defer lock.Unlock()

		// when uploading to a schema1 registry, manifest ids change and we must remap them
		for i := range is.Spec.Tags {
			tag := &is.Spec.Tags[i]
			if tag.From == nil || tag.From.Kind != "DockerImage" {
				continue
			}
			ref, err := imagereference.Parse(tag.From.Name)
			if err != nil {
				return fmt.Errorf("unable to parse image reference %s (%s): %v", tag.Name, tag.From.Name, err)
			}
			if ref.Registry != registry {
				continue
			}
			if changed, ok := manifests[digest.Digest(ref.ID)]; ok {
				ref.ID = changed.String()
				klog.V(4).Infof("During mirroring, image %s was updated to digest %s", tag.From.Name, changed)
				tag.From.Name = ref.Exact()
			}
		}
		return nil
	}
	if err := opts.Run(); err != nil {
		return err
	}

	to := o.ToRelease
	if len(to) == 0 {
		to = targetFn("").Ref.Exact()
	}
	fmt.Fprintf(o.Out, "\nSuccess\nUpdate image:  %s\n", to)
	if len(o.To) > 0 {
		if hasPrefix {
			fmt.Fprintf(o.Out, "Mirror prefix: %s\n", o.To)
		} else {
			fmt.Fprintf(o.Out, "Mirrored to: %s\n", o.To)
		}
	}
	if toDisk {
		if len(o.ToDir) > 0 {
			fmt.Fprintf(o.Out, "\nTo upload local images to a registry, run:\n\n    oc image mirror --from-dir=%s 'file://%s*' REGISTRY/REPOSITORY\n\n", o.ToDir, to)
		} else {
			fmt.Fprintf(o.Out, "\nTo upload local images to a registry, run:\n\n    oc image mirror 'file://%s*' REGISTRY/REPOSITORY\n\n", to)
		}
	} else if len(o.To) > 0 {
		if o.PrintImageContentInstructions {
			if err := printImageContentInstructions(o.Out, o.From, o.To, repositories); err != nil {
				return fmt.Errorf("Error creating mirror usage instructions: %v", err)
			}
		}
	}
	return nil
}

// printImageContentInstructions provides examples to the user for using the new repository mirror
// https://github.com/openshift/installer/blob/master/docs/dev/alternative_release_image_sources.md
func printImageContentInstructions(out io.Writer, from, to string, repositories map[string]struct{}) error {
	type installConfigSubsection struct {
		ImageContentSources []operatorv1alpha1.RepositoryDigestMirrors `json:"imageContentSources"`
	}

	var sources []operatorv1alpha1.RepositoryDigestMirrors

	mirrorRef, err := imagesource.ParseReference(to)
	if err != nil {
		return fmt.Errorf("Unable to parse image reference '%s': %v", to, err)
	}
	if mirrorRef.Type != imagesource.DestinationRegistry {
		return nil
	}
	mirrorRepo := mirrorRef.Ref.AsRepository().String()

	if len(from) != 0 {
		sourceRef, err := imagesource.ParseReference(from)
		if err != nil {
			return fmt.Errorf("Unable to parse image reference '%s': %v", from, err)
		}
		if sourceRef.Type != imagesource.DestinationRegistry {
			return nil
		}
		sourceRepo := sourceRef.Ref.AsRepository().String()
		repositories[sourceRepo] = struct{}{}
	}

	if len(repositories) == 0 {
		return nil
	}

	for repository := range repositories {
		sources = append(sources, operatorv1alpha1.RepositoryDigestMirrors{
			Source:  repository,
			Mirrors: []string{mirrorRepo},
		})
	}
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Source < sources[j].Source
	})

	// Create and display install-config.yaml example
	imageContentSources := installConfigSubsection{
		ImageContentSources: sources}
	installConfigExample, err := yaml.Marshal(imageContentSources)
	if err != nil {
		return fmt.Errorf("Unable to marshal install-config.yaml example yaml: %v", err)
	}
	fmt.Fprintf(out, "\nTo use the new mirrored repository to install, add the following section to the install-config.yaml:\n\n")
	fmt.Fprintf(out, string(installConfigExample))

	// Create and display ImageContentSourcePolicy example
	icsp := operatorv1alpha1.ImageContentSourcePolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: operatorv1alpha1.GroupVersion.String(),
			Kind:       "ImageContentSourcePolicy"},
		ObjectMeta: metav1.ObjectMeta{
			Name: "example",
		},
		Spec: operatorv1alpha1.ImageContentSourcePolicySpec{
			RepositoryDigestMirrors: sources,
		},
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
	fmt.Fprintf(out, "\n\nTo use the new mirrored repository for upgrades, use the following to create an ImageContentSourcePolicy:\n\n")
	fmt.Fprintf(out, string(icspExample))

	return nil
}
