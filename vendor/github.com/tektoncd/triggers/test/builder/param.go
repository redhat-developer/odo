package builder

import "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"

func Param(name, value string) v1alpha1.Param {
	return v1alpha1.Param{
		Name: name,
		Value: v1alpha1.ArrayOrString{
			Type:      v1alpha1.ParamTypeString,
			StringVal: value,
		},
	}
}
