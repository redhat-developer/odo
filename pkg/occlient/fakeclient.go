package occlient

import (
	fkappsclientset "github.com/openshift/client-go/apps/clientset/versioned/fake"
	fkbuildclientset "github.com/openshift/client-go/build/clientset/versioned/fake"
	fkimageclientset "github.com/openshift/client-go/image/clientset/versioned/fake"
	fkprojectclientset "github.com/openshift/client-go/project/clientset/versioned/fake"
	fkrouteClientset "github.com/openshift/client-go/route/clientset/versioned/fake"
	fkkubernetes "k8s.io/client-go/kubernetes/fake"
)

// FkClientSet : hold fake ClientSets
type FkClientSet struct {
	kubeClientset    *fkkubernetes.Clientset
	imageClientset   *fkimageclientset.Clientset
	appsClientset    *fkappsclientset.Clientset
	buildClientset   *fkbuildclientset.Clientset
	projectClientset *fkprojectclientset.Clientset
	routeClientset   *fkrouteClientset.Clientset
}

// FakeNew : create new fake client
func FakeNew() (*Client, *FkClientSet) {
	var client Client
	var fkclientset FkClientSet

	routeClient := fkrouteClientset.NewSimpleClientset()
	fkclientset.routeClientset = routeClient
	client.routeClient = fkclientset.routeClientset.Route()

	return &client, &fkclientset
}
