package service

import (
	"context"
	e "errors"

	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/binding"
	"github.com/redhat-developer/service-binding-operator/pkg/binding/registry"
	"github.com/redhat-developer/service-binding-operator/pkg/client/kubernetes"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"

	"reflect"

	"github.com/redhat-developer/service-binding-operator/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var _ pipeline.Service = &service{}

type CrdReader func(gvk *schema.GroupVersionResource) (*unstructured.Unstructured, error)

type Builder interface {
	WithClient(client dynamic.Interface) Builder
	WithCrdReader(reader CrdReader) Builder
	Build(content *unstructured.Unstructured, options ...buildOption) (pipeline.Service, error)
	LookOwnedResources(val bool) Builder
}

type buildOption func(s *service)

func Id(v string) buildOption {
	return func(s *service) {
		s.id = &v
	}
}

func CrdReaderOption(reader CrdReader) buildOption {
	return func(s *service) {
		s.CrdReader = reader
	}
}

type builder struct {
	client                dynamic.Interface
	typeLookup            kubernetes.K8STypeLookup
	crdReader             CrdReader
	lookForOwnedResources bool
}

func NewBuilder(typeLookup kubernetes.K8STypeLookup) Builder {
	return &builder{
		typeLookup: typeLookup,
	}
}

func (b *builder) WithClient(client dynamic.Interface) Builder {
	b.client = client
	return b
}

func (b *builder) WithCrdReader(reader CrdReader) Builder {
	b.crdReader = reader
	return b
}

func (b *builder) LookOwnedResources(val bool) Builder {
	b.lookForOwnedResources = val
	return b
}

func (b *builder) Build(content *unstructured.Unstructured, options ...buildOption) (pipeline.Service, error) {
	gvr, err := b.typeLookup.ResourceForKind(content.GroupVersionKind())
	if err != nil {
		return nil, err
	}
	s := &service{
		client:                b.client,
		resource:              content,
		groupVersionResource:  gvr,
		lookForOwnedResources: b.lookForOwnedResources,
		namespace:             content.GetNamespace(),
	}
	for _, o := range options {
		o(s)
	}
	if s.CrdReader == nil {
		if b.crdReader == nil {
			b.crdReader = func(gvr *schema.GroupVersionResource) (*unstructured.Unstructured, error) {
				var err error
				var u *unstructured.Unstructured
				for _, crd := range crdGVRs {
					u, err = b.client.Resource(crd).Get(context.Background(), gvr.GroupResource().String(), metav1.GetOptions{})
					if err == nil {
						return u, nil
					}
				}
				return nil, err
			}
		}
		s.CrdReader = b.crdReader
	}
	return s, nil
}

var crdGVRs = []schema.GroupVersionResource{
	{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	},
	{
		Group:    "apiextensions.k8s.io",
		Version:  "v1beta1",
		Resource: "customresourcedefinitions",
	},
}

var bindableResourceGVRs = []schema.GroupVersionResource{
	{Group: "", Version: "v1", Resource: "configmaps"},
	{Group: "", Version: "v1", Resource: "secrets"},
	{Group: "", Version: "v1", Resource: "services"},
	{Group: "route.openshift.io", Version: "v1", Resource: "routes"},
}

type service struct {
	client               dynamic.Interface
	namespace            string
	resource             *unstructured.Unstructured
	groupVersionResource *schema.GroupVersionResource
	CrdReader
	crd                   pipeline.CRD
	crdLookup             bool
	lookForOwnedResources bool
	bindingDefinitions    []binding.Definition
	id                    *string
}

func (s *service) OwnedResources() ([]*unstructured.Unstructured, error) {
	uid := s.Resource().GetUID()
	var result []*unstructured.Unstructured
	if !s.lookForOwnedResources {
		return result, nil
	}
	for _, gvr := range bindableResourceGVRs {
		list, err := s.client.Resource(gvr).Namespace(s.namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		for i := range list.Items {
			item := list.Items[i]
			for _, ownerRef := range item.GetOwnerReferences() {
				if reflect.DeepEqual(ownerRef.UID, uid) {
					result = append(result, &item)
				}
			}
		}
	}
	return result, nil
}

func (s *service) Id() *string {
	return s.id
}

func (s *service) Resource() *unstructured.Unstructured {
	return s.resource
}

func (s *service) CustomResourceDefinition() (pipeline.CRD, error) {
	if s.crd == nil {
		if s.crdLookup {
			return nil, nil
		}

		u, err := s.CrdReader(s.groupVersionResource)
		if errors.IsNotFound(err) {
			s.crdLookup = true
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		s.crd = &customResourceDefinition{resource: u, client: s.client, ns: s.namespace, serviceGVR: s.groupVersionResource}
		return s.crd, err
	}
	return s.crd, nil
}

func (s *service) AddBindingDef(def binding.Definition) {
	s.bindingDefinitions = append(s.bindingDefinitions, def)
}

func (s *service) BindingDefs() []binding.Definition {
	return s.bindingDefinitions
}

func (s *service) IsBindable() (bool, error) {
	crd, err := s.CustomResourceDefinition()
	if err != nil {
		return false, err
	}
	return crd.IsBindable()
}

type customResourceDefinition struct {
	resource   *unstructured.Unstructured
	serviceGVR *schema.GroupVersionResource
	client     dynamic.Interface
	ns         string
}

func (c *customResourceDefinition) Resource() *unstructured.Unstructured {
	if b, err := c.IsBindable(); b && err == nil {
		return c.resource
	}
	gvk := c.serviceGVR.GroupVersion().WithKind(c.kind())
	if annotations, found := registry.ServiceAnnotations.GetAnnotations(gvk); found {
		c.resource.SetAnnotations(util.MergeMaps(c.resource.GetAnnotations(), annotations))
	}
	return c.resource
}

func (c *customResourceDefinition) kind() string {
	val, found, _ := unstructured.NestedString(c.resource.Object, "spec", "names", "kind")
	if found {
		return val
	}
	return ""
}

func (c *customResourceDefinition) IsBindable() (bool, error) {
	descriptor, err := c.Descriptor()
	if err != nil {
		return false, err
	}
	annotations := make(map[string]string)
	if descriptor != nil {
		util.MergeMaps(annotations, descriptor.BindingAnnotations())
	}
	util.MergeMaps(annotations, c.resource.GetAnnotations())
	if len(annotations) == 0 {
		return false, nil
	}
	val, found := annotations[binding.ProvisionedServiceAnnotationKey]
	if found && val == "true" {
		return true, nil
	}

	for k := range annotations {
		if ok, err := binding.IsServiceBindingAnnotation(k); ok && err == nil {
			return true, nil
		}
	}
	return false, nil
}

func (c *customResourceDefinition) Descriptor() (*pipeline.CRDDescription, error) {
	csvs, err := c.client.Resource(olmv1alpha1.SchemeGroupVersion.WithResource("clusterserviceversions")).Namespace(c.ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(csvs.Items) == 0 {
		return nil, nil
	}
	for _, csv := range csvs.Items {
		ownedPath := []string{"spec", "customresourcedefinitions", "owned"}

		ownedCRDs, exists, err := unstructured.NestedSlice(csv.Object, ownedPath...)
		if err != nil {
			return nil, err
		}
		if !exists {
			continue
		}

		for _, crd := range ownedCRDs {
			crdDesciption := &pipeline.CRDDescription{}
			data, ok := crd.(map[string]interface{})
			if !ok {
				return nil, e.New("cannot cast to map")
			}
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(data, crdDesciption)
			if err != nil {
				return nil, err
			}
			if crdDesciption.Name == c.resource.GetName() && crdDesciption.Kind == c.kind() && crdDesciption.Version == c.serviceGVR.Version {
				return crdDesciption, nil
			}
		}
	}
	return nil, nil
}
