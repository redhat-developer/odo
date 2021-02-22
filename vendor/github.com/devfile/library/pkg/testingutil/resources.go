package testingutil

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// FakeResourceRequirements creates a fake resource requirements from cpu and memory
func FakeResourceRequirements(cpu, memory string) (corev1.ResourceRequirements, error) {
	var resReq corev1.ResourceRequirements

	limits := make(corev1.ResourceList)
	var err error
	limits[corev1.ResourceCPU], err = resource.ParseQuantity(cpu)
	if err != nil {
		return resReq, err
	}
	limits[corev1.ResourceMemory], err = resource.ParseQuantity(memory)
	if err != nil {
		return resReq, err
	}
	resReq.Limits = limits

	requests := make(corev1.ResourceList)
	requests[corev1.ResourceCPU], err = resource.ParseQuantity(cpu)
	if err != nil {
		return resReq, err
	}
	requests[corev1.ResourceMemory], err = resource.ParseQuantity(memory)
	if err != nil {
		return resReq, err
	}

	resReq.Requests = requests

	return resReq, nil
}
