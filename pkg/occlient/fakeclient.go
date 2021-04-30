package occlient

import (
	"fmt"
	"os"
	"runtime"
	"sync"

	fakeServiceCatalogClientSet "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	fakeAppsClientset "github.com/openshift/client-go/apps/clientset/versioned/fake"
	fakeBuildClientset "github.com/openshift/client-go/build/clientset/versioned/fake"
	fakeImageClientset "github.com/openshift/client-go/image/clientset/versioned/fake"
	fakeProjClientset "github.com/openshift/client-go/project/clientset/versioned/fake"
	fakeRouteClientset "github.com/openshift/client-go/route/clientset/versioned/fake"
	"github.com/openshift/odo/pkg/kclient"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery/fake"
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
	DiscoveryClientSet      *fake.FakeDiscovery
}

// FakeNew creates new fake client for testing
// returns Client that is filled with fake clients and
// FakeClientSet that holds fake Clientsets to access Actions, Reactors etc... in fake client
func FakeNew() (*Client, *FakeClientset) {
	client := &Client{}
	fkclientset := &FakeClientset{}

	kc, fkcs := kclient.FakeNew()
	client.kubeClient = kc
	fkclientset.Kubernetes = fkcs.Kubernetes
	fkclientset.ServiceCatalogClientSet = fkcs.ServiceCatalogClientSet

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

	if os.Getenv("KUBERNETES") != "true" {
		client.GetKubeClient().SetDiscoveryInterface(fakeDiscoveryWithRoute)
		client.GetKubeClient().SetDiscoveryInterface(fakeDiscoveryWithDeploymentConfig)
	} else {
		client.GetKubeClient().SetDiscoveryInterface(&fakeDiscovery{
			resourceMap: map[string]*resourceMapEntry{},
		})
	}

	return client, fkclientset
}

type resourceMapEntry struct {
	list *metav1.APIResourceList
	err  error
}

type fakeDiscovery struct {
	*fake.FakeDiscovery

	lock        sync.Mutex
	resourceMap map[string]*resourceMapEntry
}

var fakeDiscoveryWithRoute = &fakeDiscovery{
	resourceMap: map[string]*resourceMapEntry{
		"project.openshift.io/v1": {
			list: &metav1.APIResourceList{
				GroupVersion: "project.openshift.io/v1",
				APIResources: []metav1.APIResource{{
					Name:         "routes",
					SingularName: "route",
					Namespaced:   true,
					Kind:         "route",
				}},
			},
		},
	},
}

var fakeDiscoveryWithDeploymentConfig = &fakeDiscovery{
	resourceMap: map[string]*resourceMapEntry{
		"apps.openshift.io/v1": {
			list: &metav1.APIResourceList{
				GroupVersion: "apps.openshift.io/v1",
				APIResources: []metav1.APIResource{{
					Name:         "deploymentconfigs",
					SingularName: "deploymentconfigs",
					Namespaced:   true,
					Kind:         "deploymentconfigs",
				}},
			},
		},
	},
}

func (c *fakeDiscovery) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if rl, ok := c.resourceMap[groupVersion]; ok {
		return rl.list, rl.err
	}
	return nil, kerrors.NewNotFound(schema.GroupResource{}, "")
}

func (c *fakeDiscovery) ServerVersion() (*version.Info, error) {
	versionInfo := version.Info{
		Major:        "1",
		Minor:        "16",
		GitVersion:   "v1.16.0+0000000",
		GitCommit:    "",
		GitTreeState: "",
		BuildDate:    "",
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
	return &versionInfo, nil
}
