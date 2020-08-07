package servicebindingrequest

import (
	"errors"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/annotations"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/nested"
)

var (
	// ErrUnspecifiedBackingServiceNamespace is returned when the namespace of a service is
	// unspecified.
	ErrUnspecifiedBackingServiceNamespace = errors.New("backing service namespace is unspecified")
	// EmptyBackingServiceSelectorsErr is returned when no backing service selectors have been
	// informed in the Service Binding Request.
	ErrEmptyBackingServiceSelectors = errors.New("backing service selectors are empty")
)

func findService(
	client dynamic.Interface,
	ns string,
	gvk schema.GroupVersionKind,
	resourceRef string,
) (
	*unstructured.Unstructured,
	error,
) {
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)

	if len(ns) == 0 {
		return nil, ErrUnspecifiedBackingServiceNamespace
	}

	// delegate the search selector's namespaced resource client
	return client.
		Resource(gvr).
		Namespace(ns).
		Get(resourceRef, metav1.GetOptions{})
}

// CRDGVR is the plural GVR for Kubernetes CRDs.
var CRDGVR = schema.GroupVersionResource{
	Group:    "apiextensions.k8s.io",
	Version:  "v1beta1",
	Resource: "customresourcedefinitions",
}

func findServiceCRD(client dynamic.Interface, gvk schema.GroupVersionKind) (*unstructured.Unstructured, error) {
	// gvr is the plural guessed resource for the given GVK
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	// crdName is the string'fied GroupResource, e.g. "deployments.apps"
	crdName := gvr.GroupResource().String()
	// delegate the search to the CustomResourceDefinition resource client
	return client.Resource(CRDGVR).Get(crdName, metav1.GetOptions{})
}

func loadDescriptor(anns map[string]string, path string, descriptor string, root string) {
	if !strings.HasPrefix(descriptor, "binding:") {
		return
	}

	n := annotations.ServiceBindingOperatorAnnotationPrefix + root + "." + path
	v := strings.Split(descriptor, ":")

	if strings.HasPrefix(descriptor, "binding:env:") {
		if len(v) > 4 {
			n = n + "-" + v[4]
			anns[n] = strings.Join(v[0:4], ":")
		}
		if len(v) == 4 {
			anns[n] = strings.Join(v[0:4], ":")
		}

	}

	if strings.HasPrefix(descriptor, "binding:volumemount:") {
		anns[n] = strings.Join(v[0:3], ":")
	}

}

func convertCRDDescriptionToAnnotations(crdDescription *olmv1alpha1.CRDDescription) map[string]string {
	anns := make(map[string]string)
	for _, sd := range crdDescription.StatusDescriptors {
		for _, xd := range sd.XDescriptors {
			loadDescriptor(anns, sd.Path, xd, "status")
		}
	}

	for _, sd := range crdDescription.SpecDescriptors {
		for _, xd := range sd.XDescriptors {
			loadDescriptor(anns, sd.Path, xd, "spec")
		}
	}

	return anns
}

// findCRDDescription attempts to find the CRDDescription resource related CustomResourceDefinition.
func findCRDDescription(
	ns string,
	client dynamic.Interface,
	bssGVK schema.GroupVersionKind,
	crd *unstructured.Unstructured,
) (*olmv1alpha1.CRDDescription, error) {
	return NewOLM(client, ns).SelectCRDByGVK(bssGVK, crd)
}

type bindableResource struct {
	gvk        schema.GroupVersionKind
	gvr        schema.GroupVersionResource
	inputPath  string
	outputPath string
}

var bindableResources = []bindableResource{
	{
		gvk:        schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"},
		gvr:        schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"},
		inputPath:  "data",
		outputPath: "",
	},
	{
		gvk:        schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"},
		gvr:        schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"},
		inputPath:  "data",
		outputPath: "",
	},
	{
		gvk:        schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
		gvr:        schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"},
		inputPath:  "spec.clusterIP",
		outputPath: "clusterIP",
	},
	{
		gvk: schema.GroupVersionKind{
			Group:   "route.openshift.io",
			Version: "v1",
			Kind:    "Route",
		},
		gvr:        schema.GroupVersionResource{Group: "route.openshift.io", Version: "v1", Resource: "routes"},
		inputPath:  "spec.host",
		outputPath: "host",
	},
}

func getOwnedResources(
	client dynamic.Interface,
	ns string,
	gvk schema.GroupVersionKind,
	name string,
	uid types.UID,
) (
	[]*unstructured.Unstructured,
	error,
) {
	var resources []*unstructured.Unstructured
	for _, br := range bindableResources {
		lst, err := client.Resource(br.gvr).Namespace(ns).List(metav1.ListOptions{})
		if err != nil {
			return resources, err
		}
		for _, item := range lst.Items {
			owners := item.GetOwnerReferences()
			for _, owner := range owners {
				if owner.UID == uid {
					resources = append(resources, &item)
				}
			}
		}
	}
	return resources, nil
}

func buildOwnedResourceContext(
	client dynamic.Interface,
	obj *unstructured.Unstructured,
	ownerEnvVarPrefix *string,
	restMapper meta.RESTMapper,
	inputPath string,
	outputPath string,
) (*ServiceContext, error) {
	svcCtx, err := buildServiceContext(
		client, obj.GetNamespace(), obj.GetObjectKind().GroupVersionKind(), obj.GetName(),
		ownerEnvVarPrefix, restMapper)
	if err != nil {
		return nil, err
	}
	svcCtx.EnvVars, _, err = nested.GetValue(obj.Object, inputPath, outputPath)
	return svcCtx, err
}

func buildOwnedResourceContexts(
	client dynamic.Interface,
	objs []*unstructured.Unstructured,
	ownerEnvVarPrefix *string,
	restMapper meta.RESTMapper,
) ([]*ServiceContext, error) {
	ctxs := make(ServiceContextList, 0)

	for _, obj := range objs {
		for _, br := range bindableResources {
			if br.gvk != obj.GetObjectKind().GroupVersionKind() {
				continue
			}
			svcCtx, err := buildOwnedResourceContext(
				client,
				obj,
				ownerEnvVarPrefix,
				restMapper,
				br.inputPath,
				br.outputPath,
			)
			if err != nil {
				return nil, err
			}
			ctxs = append(ctxs, svcCtx)
		}
	}

	return ctxs, nil
}
