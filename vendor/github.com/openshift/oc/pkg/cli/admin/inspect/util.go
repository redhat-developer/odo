package inspect

import (
	"fmt"
	"path"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"

	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// resourceContext is used to keep track of previously seen objects
type resourceContext struct {
	visited sets.String
}

func NewResourceContext() *resourceContext {
	return &resourceContext{
		visited: sets.NewString(),
	}
}

func objectReferenceToString(ref *configv1.ObjectReference) string {
	resource := ref.Resource
	group := ref.Group
	name := ref.Name
	if len(name) > 0 {
		name = "/" + name
	}
	if len(group) > 0 {
		group = "." + group
	}
	return resource + group + name
}

func unstructuredToString(obj *unstructured.Unstructured) string {
	resource := obj.GetKind()
	var group string
	if gv, err := schema.ParseGroupVersion(obj.GetAPIVersion()); err != nil {
		group = gv.Group
	}
	name := obj.GetName()
	if len(name) > 0 {
		name = "/" + name
	}
	if len(group) > 0 {
		group = "." + group
	}
	return resource + group + name

}

func objectReferenceToResourceInfos(clientGetter genericclioptions.RESTClientGetter, ref *configv1.ObjectReference) ([]*resource.Info, error) {
	b := resource.NewBuilder(clientGetter).
		Unstructured().
		ResourceTypeOrNameArgs(true, objectReferenceToString(ref)).
		NamespaceParam(ref.Namespace).DefaultNamespace().AllNamespaces(len(ref.Namespace) == 0).
		Flatten().
		Latest()

	infos, err := b.Do().Infos()
	if err != nil {
		return nil, err
	}

	return infos, nil
}

func groupResourceToInfos(clientGetter genericclioptions.RESTClientGetter, ref schema.GroupResource, namespace string) ([]*resource.Info, error) {
	resourceString := ref.Resource
	if len(ref.Group) > 0 {
		resourceString = fmt.Sprintf("%s.%s", resourceString, ref.Group)
	}
	b := resource.NewBuilder(clientGetter).
		Unstructured().
		ResourceTypeOrNameArgs(false, resourceString).
		SelectAllParam(true).
		NamespaceParam(namespace).
		Latest()

	return b.Do().Infos()
}

// infoToContextKey receives a resource.Info and returns a unique string for use in keeping track of objects previously seen
func infoToContextKey(info *resource.Info) string {
	name := info.Name
	if meta.IsListType(info.Object) {
		name = "*"
	}
	return fmt.Sprintf("%s/%s/%s/%s", info.Namespace, info.ResourceMapping().GroupVersionKind.Group, info.ResourceMapping().Resource.Resource, name)
}

// objectRefToContextKey is a variant of infoToContextKey that receives a configv1.ObjectReference and returns a unique string for use in keeping track of object references previously seen
func objectRefToContextKey(objRef *configv1.ObjectReference) string {
	return fmt.Sprintf("%s/%s/%s/%s", objRef.Namespace, objRef.Group, objRef.Resource, objRef.Name)
}

func resourceToContextKey(resource schema.GroupResource, namespace string) string {
	return fmt.Sprintf("%s/%s/%s/%s", namespace, resource.Group, resource.Resource, "*")
}

// dirPathForInfo receives a *resource.Info and returns a relative path
// corresponding to the directory location of that object on disk
func dirPathForInfo(baseDir string, info *resource.Info) string {
	groupName := "core"
	if len(info.Mapping.GroupVersionKind.Group) > 0 {
		groupName = info.Mapping.GroupVersionKind.Group
	}

	groupPath := path.Join(baseDir, namespaceResourcesDirname, info.Namespace, groupName)
	if len(info.Namespace) == 0 {
		groupPath = path.Join(baseDir, clusterScopedResourcesDirname, "/"+groupName)
	}
	if meta.IsListType(info.Object) {
		return groupPath
	}

	objPath := path.Join(groupPath, info.ResourceMapping().Resource.Resource)
	if len(info.Namespace) == 0 {
		objPath = path.Join(groupPath, info.ResourceMapping().Resource.Resource)
	}
	return objPath
}

// filenameForInfo receives a *resource.Info and returns the basename
func filenameForInfo(info *resource.Info) string {
	if meta.IsListType(info.Object) {
		return info.ResourceMapping().Resource.Resource + ".yaml"
	}

	return info.Name + ".yaml"
}
