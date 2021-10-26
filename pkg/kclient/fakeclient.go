package kclient

import (
	fakeServiceCatalogClientSet "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	odoFake "github.com/openshift/odo/v2/pkg/kclient/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeKubeClientset "k8s.io/client-go/kubernetes/fake"
)

// FakeClientset holds fake ClientSets
// this is returned by FakeNew to access methods of fake client sets
type FakeClientset struct {
	Kubernetes              *fakeKubeClientset.Clientset
	ServiceCatalogClientSet *fakeServiceCatalogClientSet.Clientset
}

// FakeNew creates new fake client for testing
// returns Client that is filled with fake clients and
// FakeClientSet that holds fake Clientsets to access Actions, Reactors etc... in fake client
// fake ingress support is set to default ie only extension v1 beta 1 is supported
func FakeNew() (*Client, *FakeClientset) {
	return FakeNewWithIngressSupports(false, true)
}

// FakeNewWithIngressSupports creates new fake client for testing
// returns Client that is filled with fake clients and
// FakeClientSet that holds fake Clientsets to access Actions, Reactors etc... in fake
func FakeNewWithIngressSupports(networkingv1Supported, extensionV1Supported bool) (*Client, *FakeClientset) {
	var client Client
	var fkclientset FakeClientset

	fkclientset.Kubernetes = fakeKubeClientset.NewSimpleClientset()
	client.KubeClient = fkclientset.Kubernetes

	fkclientset.ServiceCatalogClientSet = fakeServiceCatalogClientSet.NewSimpleClientset()
	client.serviceCatalogClient = fkclientset.ServiceCatalogClientSet.ServicecatalogV1beta1()
	client.appsClient = fkclientset.Kubernetes.AppsV1()
	client.isExtensionV1Beta1IngressSupported = extensionV1Supported
	client.isNetworkingV1IngressSupported = networkingv1Supported
	client.checkIngressSupports = false
	client.SetDiscoveryInterface(NewKubernetesFakedDiscovery(true, true))

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

func NewKubernetesFakedDiscovery(extv1b1supported, nwv1suppored bool) *odoFake.FakeDiscovery {
	fd := odoFake.NewFakeDiscovery()
	extingress := metav1.GroupVersionResource{
		Group:    "extensions",
		Version:  "v1beta1",
		Resource: "ingress",
	}
	netv1ingress := metav1.GroupVersionResource{
		Group:    "networking.k8s.io",
		Version:  "v1",
		Resource: "ingress",
	}
	if extv1b1supported {
		fd.AddResourceList(extingress.String(), &metav1.APIResourceList{
			GroupVersion: "extensions/v1beta1",
			APIResources: []metav1.APIResource{{
				Name:         "ingress",
				SingularName: "ingress",
				Namespaced:   true,
				Kind:         "ingress",
			}},
		})
	}

	if nwv1suppored {
		fd.AddResourceList(netv1ingress.String(), &metav1.APIResourceList{
			GroupVersion: "networking.k8s.io/v1",
			APIResources: []metav1.APIResource{{
				Name:         "ingress",
				SingularName: "ingress",
				Namespaced:   true,
				Kind:         "ingress",
			}},
		})
	}
	return fd
}
