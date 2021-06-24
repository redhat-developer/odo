package fake

import (
	"fmt"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery/fake"
	"runtime"
	"strings"
	"sync"
)

type ResourceMapEntry struct {
	list *metav1.APIResourceList
}

type FakeDiscovery struct {
	*fake.FakeDiscovery

	lock        sync.Mutex
	resourceMap map[string]*ResourceMapEntry
}

func NewFakeDiscovery() *FakeDiscovery {
	return &FakeDiscovery{resourceMap: make(map[string]*ResourceMapEntry, 7)}
}

func (c *FakeDiscovery) AddResourceList(key string, are *metav1.APIResourceList) {
	c.resourceMap[key] = &ResourceMapEntry{list: are}
}

func (c *FakeDiscovery) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	found := false
	arl := metav1.APIResourceList{
		GroupVersion: groupVersion,
		APIResources: nil,
	}
	for k, v := range c.resourceMap {
		if strings.Contains(k, groupVersion) {
			found = true
			arl.APIResources = append(arl.APIResources, v.list.APIResources...)
		}
	}
	if found {
		return &arl, nil
	}
	return nil, kerrors.NewNotFound(schema.GroupResource{}, "")
}

func (c *FakeDiscovery) ServerVersion() (*version.Info, error) {
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
