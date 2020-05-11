package occlient

import (
	fakeServiceCatalogClientSet "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	fakeAppsClientset "github.com/openshift/client-go/apps/clientset/versioned/fake"
	fakeBuildClientset "github.com/openshift/client-go/build/clientset/versioned/fake"
	fakeImageClientset "github.com/openshift/client-go/image/clientset/versioned/fake"
	fakeProjClientset "github.com/openshift/client-go/project/clientset/versioned/fake"
	fakeRouteClientset "github.com/openshift/client-go/route/clientset/versioned/fake"
	fakeKubeClientset "k8s.io/client-go/kubernetes/fake"
)

// FakeClientset holds fake ClientSets
// this is returned by FakeNew to access methods of fake client sets
type FakeClientset struct {
	Kubernetes              *fakeKubeClientset.Clientset
	AppsClientset           *fakeAppsClientset.Clientset
	BuildClientset          *fakeBuildClientset.Clientset
	ImageClientset          *fakeImageClientset.Clientset
	RouteClientset          *fakeRouteClientset.Clientset
	ProjClientset           *fakeProjClientset.Clientset
	ServiceCatalogClientSet *fakeServiceCatalogClientSet.Clientset
}

// FakeNew creates new fake client for testing
// returns Client that is filled with fake clients and
// FakeClientSet that holds fake Clientsets to access Actions, Reactors etc... in fake client
func FakeNew() (*Client, *FakeClientset) {
	var client Client
	var fkclientset FakeClientset

	fkclientset.Kubernetes = fakeKubeClientset.NewSimpleClientset()
	client.kubeClient = fkclientset.Kubernetes

	fkclientset.AppsClientset = fakeAppsClientset.NewSimpleClientset()
	client.appsClient = fkclientset.AppsClientset.AppsV1()

	fkclientset.BuildClientset = fakeBuildClientset.NewSimpleClientset()
	client.buildClient = fkclientset.BuildClientset.BuildV1()

	fkclientset.RouteClientset = fakeRouteClientset.NewSimpleClientset()
	client.routeClient = fkclientset.RouteClientset.RouteV1()

	fkclientset.ImageClientset = fakeImageClientset.NewSimpleClientset()
	client.imageClient = fkclientset.ImageClientset.ImageV1()

	fkclientset.ProjClientset = fakeProjClientset.NewSimpleClientset()
	client.projectClient = fkclientset.ProjClientset.ProjectV1()

	fkclientset.BuildClientset = fakeBuildClientset.NewSimpleClientset()
	client.buildClient = fkclientset.BuildClientset.BuildV1()

	fkclientset.ServiceCatalogClientSet = fakeServiceCatalogClientSet.NewSimpleClientset()
	client.serviceCatalogClient = fkclientset.ServiceCatalogClientSet.ServicecatalogV1beta1()

	return &client, &fkclientset
}
