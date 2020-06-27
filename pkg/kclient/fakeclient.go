package kclient

import (
	fakeKubeClientset "k8s.io/client-go/kubernetes/fake"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FakeClientset holds fake ClientSets
// this is returned by FakeNew to access methods of fake client sets
type FakeClientset struct {
	Kubernetes *fakeKubeClientset.Clientset
}

// FakeNew creates new fake client for testing
// returns Client that is filled with fake clients and
// FakeClientSet that holds fake Clientsets to access Actions, Reactors etc... in fake client
func FakeNew() (*Client, *FakeClientset) {
	var client Client
	var fkclientset FakeClientset

	fkclientset.Kubernetes = fakeKubeClientset.NewSimpleClientset()
	client.KubeClient = fkclientset.Kubernetes

	return &client, &fkclientset
}

//FakePodStatus returns a pod with the status
func FakePodStatus(status corev1.PodPhase, podName string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   podName,
			Labels: map[string]string{},
		},
		Status: corev1.PodStatus{
			Phase: status,
		},
	}
}
