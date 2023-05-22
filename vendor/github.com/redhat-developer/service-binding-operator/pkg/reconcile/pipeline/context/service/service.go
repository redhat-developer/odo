package service

import (
	"context"
	"reflect"
	"strings"

	"github.com/redhat-developer/service-binding-operator/pkg/binding"
	"github.com/redhat-developer/service-binding-operator/pkg/binding/registry"
	"github.com/redhat-developer/service-binding-operator/pkg/client/kubernetes"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"

	"github.com/redhat-developer/service-binding-operator/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/jsonpath"
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

// getValuesByJSONPath returns values from the given map matching the provided JSONPath
// 'path' argument takes JSONPath expressions enclosed by curly braces {}
// see https://kubernetes.io/docs/reference/kubectl/jsonpath/ for more details
// It returns zero or more filtered values back,
// or error if the jsonpath is invalid or it cannot be applied on the given map
func getValuesByJSONPath(obj map[string]interface{}, path string) ([]reflect.Value, error) {
	j := jsonpath.New("")
	err := j.Parse(path)
	if err != nil {
		return nil, err
	}
	result, err := j.FindResults(obj)
	if err != nil {
		return nil, err
	}
	if len(result) > 1 {
		w := strings.Builder{}
		for i := range result {
			if err := j.PrintResults(&w, result[i]); err != nil {
				return nil, err
			}
		}
		return []reflect.Value{reflect.ValueOf(w.String())}, nil
	}
	return result[0], nil
}

func (c *customResourceDefinition) IsBindable() (bool, error) {

	value, err := getValuesByJSONPath(c.resource.Object, "{..schema.openAPIV3Schema.properties.status.properties.binding.properties.name.type}")
	if err == nil && len(value) > 0 && value[0].Interface().(string) == "string" {
		return true, nil
	}

	annotations := make(map[string]string)
	util.MergeMaps(annotations, c.resource.GetAnnotations())
	if len(annotations) == 0 {
		return false, nil
	}
	for k := range annotations {
		if ok, err := binding.IsServiceBindingAnnotation(k); ok && err == nil {
			return true, nil
		}
	}
	return false, nil
}
