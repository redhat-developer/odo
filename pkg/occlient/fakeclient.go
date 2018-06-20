package occlient

import (
	fakeAppClientset "github.com/openshift/client-go/apps/clientset/versioned/fake"
	fakeBuildClientset "github.com/openshift/client-go/build/clientset/versioned/fake"
	fakeRouteClientset "github.com/openshift/client-go/route/clientset/versioned/fake"
	fakeKubeClientset "k8s.io/client-go/kubernetes/fake"
)

// FakeClientset holds fake ClientSets
// this is returned by FakeNew to access methods of fake client sets
type FakeClientset struct {
	Kubernetes     *fakeKubeClientset.Clientset
	RouteClientset *fakeRouteClientset.Clientset
	AppClientset   *fakeAppClientset.Clientset
	BuildClientset *fakeBuildClientset.Clientset
}

// FakeNew creates new fake client for testing
// returns Client that is filled with fake clients and
// FakeClientSet that holds fake Clientsets to access Actions, Reactors etc... in fake client
func FakeNew() (*Client, *FakeClientset) {
	var client Client
	var fkclientset FakeClientset

	fkclientset.RouteClientset = fakeRouteClientset.NewSimpleClientset()
	client.routeClient = fkclientset.RouteClientset.Route()

	fkclientset.Kubernetes = fakeKubeClientset.NewSimpleClientset()
	client.kubeClient = fkclientset.Kubernetes

	fkclientset.AppClientset = fakeAppClientset.NewSimpleClientset()
	client.appsClient = fkclientset.AppClientset.Apps()

	fkclientset.BuildClientset = fakeBuildClientset.NewSimpleClientset()
	client.buildClient = fkclientset.BuildClientset.Build()

	return &client, &fkclientset
}
