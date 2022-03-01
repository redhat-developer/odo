package delete

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Client interface {
	ListKubernetesComponents(devfileObj parser.DevfileObj, path string) ([]unstructured.Unstructured, error)
	UnDeploy(devfileObj parser.DevfileObj, path string) error
	DeleteComponent(devfileObj parser.DevfileObj, componentName string) error
}
