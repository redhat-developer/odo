package clientset

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

func newErrorResource(r schema.GroupVersionResource) errorResourceInterface {
	return errorResourceInterface{resource: r}
}

type errorResourceInterface struct {
	resource schema.GroupVersionResource
}

func (i errorResourceInterface) Namespace(string) dynamic.ResourceInterface {
	return i
}

func (i errorResourceInterface) err() error {
	return fmt.Errorf("resource %+v not supported", i.resource)
}

func (i errorResourceInterface) Create(obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, i.err()
}

func (i errorResourceInterface) Update(obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, i.err()
}

func (i errorResourceInterface) UpdateStatus(obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return nil, i.err()
}

func (i errorResourceInterface) Delete(name string, options *metav1.DeleteOptions, subresources ...string) error {
	return i.err()
}

func (i errorResourceInterface) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return i.err()
}

func (i errorResourceInterface) Get(name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, i.err()
}

func (i errorResourceInterface) List(opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return nil, i.err()
}

func (i errorResourceInterface) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return nil, i.err()
}

func (i errorResourceInterface) Patch(name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, i.err()
}
