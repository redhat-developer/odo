package kubernetes

import (
	devfileCommon "github.com/openshift/odo/pkg/devfile/versions/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func convertEnvs(vars []devfileCommon.DockerimageEnv) []corev1.EnvVar {
	kVars := []corev1.EnvVar{}
	for _, env := range vars {
		kVars = append(kVars, corev1.EnvVar{
			Name:  *env.Name,
			Value: *env.Value,
		})
	}
	return kVars
}

func getResourceReqs(comp devfileCommon.DevfileComponent) corev1.ResourceRequirements {
	reqs := corev1.ResourceRequirements{}
	limits := make(corev1.ResourceList)
	if comp.MemoryLimit != nil {
		memoryLimit, err := resource.ParseQuantity(*comp.MemoryLimit)
		if err == nil {
			limits[corev1.ResourceMemory] = memoryLimit
		}
		reqs.Limits = limits
	}
	return reqs
}
