package testingutil

import (
	"github.com/devfile/library/pkg/devfile/parser"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	v12 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/redhat-developer/odo/pkg/libdevfile"
)

func CreateFakeIngressFromDevfile(devfileObj parser.DevfileObj, ingressComponentName string, label map[string]string) *v1.Ingress {
	ing := &v1.Ingress{}
	u, _ := libdevfile.GetK8sComponentAsUnstructured(devfileObj, ingressComponentName, "", devfilefs.DefaultFs{})
	_ = runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), ing)
	ing.SetLabels(label)
	return ing
}

func CreateFakeRouteFromDevfile(devfileObj parser.DevfileObj, routeComponentName string, label map[string]string) *v12.Route {
	route := &v12.Route{}
	u, _ := libdevfile.GetK8sComponentAsUnstructured(devfileObj, routeComponentName, "", devfilefs.DefaultFs{})
	_ = runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), route)
	route.SetLabels(label)
	return route
}
