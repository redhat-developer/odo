package testutils

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func BuildTestRESTMapper() meta.RESTMapper {
	restMapper := meta.NewDefaultRESTMapper(
		[]schema.GroupVersion{
			{Version: "v1"},
		},
	)
	restMapper.Add(
		schema.GroupVersionKind{Kind: "Secret", Version: "v1"},
		meta.RESTScopeNamespace,
	)
	restMapper.Add(
		schema.GroupVersionKind{Kind: "ConfigMap", Version: "v1"},
		meta.RESTScopeNamespace,
	)
	restMapper.Add(
		schema.GroupVersionKind{Kind: "Deployment", Version: "v1", Group: "apps"},
		meta.RESTScopeNamespace,
	)
	return restMapper
}
