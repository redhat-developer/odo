package deployment

import (
	"strconv"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

// Revision was copied from k8s.io/kubernetes/pkg/controller/deployment/util/deployment_util.go release-1.17
const (
	// RevisionAnnotation is the revision annotation of a deployment's replica sets which records its rollout sequence
	RevisionAnnotation = "deployment.kubernetes.io/revision"
)

// Revision returns the revision number of the input object.
func Revision(obj runtime.Object) (int64, error) {
	acc, err := meta.Accessor(obj)
	if err != nil {
		return 0, err
	}
	v, ok := acc.GetAnnotations()[RevisionAnnotation]
	if !ok {
		return 0, nil
	}
	return strconv.ParseInt(v, 10, 64)
}
