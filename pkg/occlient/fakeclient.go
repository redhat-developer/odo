package occlient

import (
	fakeRouteClientset "github.com/openshift/client-go/route/clientset/versioned/fake"
)

// fkClientSet holds fake ClientSets
// this is returned by FakeNew to access methods of fake client sets
type FakeClientset struct {
	RouteClientset *fakeRouteClientset.Clientset
}

// FakeNew creates new fake client for testing
// returns Client that is filled with fake clients and
// fkClientSet that holds fake Clientsets to access Actions, Reactors etc... in fake client
func FakeNew() (*Client, *FakeClientset) {
	var client Client
	var fkclientset FakeClientset

	fkclientset.RouteClientset = fakeRouteClientset.NewSimpleClientset()
	client.routeClient = fkclientset.RouteClientset.Route()

	return &client, &fkclientset
}
