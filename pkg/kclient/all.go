package kclient

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

// Code into this file is heavily inspired from https://github.com/ahmetb/kubectl-tree

func (c *Client) GetAllResourcesFromSelector(selector string, ns string) ([]unstructured.Unstructured, error) {
	apis, err := findAPIs(c.cachedDiscoveryClient)
	if err != nil {
		return nil, err
	}
	return getAllResources(c.DynamicClient, apis.list, ns, selector)
}

func getAllResources(client dynamic.Interface, apis []apiResource, ns string, selector string) ([]unstructured.Unstructured, error) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var out []unstructured.Unstructured

	start := time.Now()
	klog.V(2).Infof("starting to query %d APIs in concurrently", len(apis))

	var errResult error
	for _, api := range apis {
		if !api.r.Namespaced {
			klog.V(4).Infof("[query api] api (%s) is non-namespaced, skipping", api.r.Name)
			continue
		}
		wg.Add(1)
		go func(a apiResource) {
			defer wg.Done()
			klog.V(4).Infof("[query api] start: %s", a.GroupVersionResource())
			v, err := queryAPI(client, a, ns, selector)
			if err != nil {
				klog.V(4).Infof("[query api] error querying: %s, error=%v", a.GroupVersionResource(), err)
				errResult = err
				return
			}
			mu.Lock()
			out = append(out, v...)
			mu.Unlock()
			klog.V(4).Infof("[query api]  done: %s, found %d apis", a.GroupVersionResource(), len(v))
		}(api)
	}

	klog.V(2).Infof("fired up all goroutines to query APIs")
	wg.Wait()
	klog.V(2).Infof("all goroutines have returned in %v", time.Since(start))
	klog.V(2).Infof("query result: error=%v, objects=%d", errResult, len(out))
	return out, errResult
}

func queryAPI(client dynamic.Interface, api apiResource, ns string, selector string) ([]unstructured.Unstructured, error) {
	var out []unstructured.Unstructured

	var next string
	for {
		var intf dynamic.ResourceInterface
		nintf := client.Resource(api.GroupVersionResource())
		intf = nintf.Namespace(ns)
		resp, err := intf.List(context.TODO(), metav1.ListOptions{
			Limit:         250,
			Continue:      next,
			LabelSelector: selector,
		})
		if err != nil {
			klog.V(2).Infof("listing resources failed (%s): %v", api.GroupVersionResource(), err)
			return nil, nil
		}
		out = append(out, resp.Items...)

		next = resp.GetContinue()
		if next == "" {
			break
		}
	}
	return out, nil
}

type apiResource struct {
	r  metav1.APIResource
	gv schema.GroupVersion
}

func (a apiResource) GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    a.gv.Group,
		Version:  a.gv.Version,
		Resource: a.r.Name,
	}
}

type resourceNameLookup map[string][]apiResource

type resourceMap struct {
	list []apiResource
	m    resourceNameLookup
}

func findAPIs(client discovery.DiscoveryInterface) (*resourceMap, error) {
	start := time.Now()
	resList, err := client.ServerPreferredResources()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch api groups from kubernetes: %w", err)
	}
	klog.V(2).Infof("queried api discovery in %v", time.Since(start))
	klog.V(3).Infof("found %d items (groups) in server-preferred APIResourceList", len(resList))

	rm := &resourceMap{
		m: make(resourceNameLookup),
	}
	for _, group := range resList {
		klog.V(5).Infof("iterating over group %s/%s (%d apis)", group.GroupVersion, group.APIVersion, len(group.APIResources))
		gv, err := schema.ParseGroupVersion(group.GroupVersion)
		if err != nil {
			return nil, fmt.Errorf("%q cannot be parsed into groupversion: %w", group.GroupVersion, err)
		}

		for _, apiRes := range group.APIResources {
			klog.V(5).Infof("  api=%s namespaced=%v", apiRes.Name, apiRes.Namespaced)
			if !contains(apiRes.Verbs, "list") {
				klog.V(4).Infof("    api (%s) doesn't have required verb, skipping: %v", apiRes.Name, apiRes.Verbs)
				continue
			}
			v := apiResource{
				gv: gv,
				r:  apiRes,
			}
			names := apiNames(apiRes, gv)
			klog.V(6).Infof("names: %s", strings.Join(names, ", "))
			for _, name := range names {
				rm.m[name] = append(rm.m[name], v)
			}
			rm.list = append(rm.list, v)
		}
	}
	klog.V(5).Infof("  found %d apis", len(rm.m))
	return rm, nil
}

func contains(v []string, s string) bool {
	for _, vv := range v {
		if vv == s {
			return true
		}
	}
	return false
}

// return all names that could refer to this APIResource
func apiNames(a metav1.APIResource, gv schema.GroupVersion) []string {
	var out []string
	singularName := a.SingularName
	if singularName == "" {
		// TODO(ahmetb): sometimes SingularName is empty (e.g. Deployment), use lowercase Kind as fallback - investigate why
		singularName = strings.ToLower(a.Kind)
	}
	pluralName := a.Name
	shortNames := a.ShortNames
	names := append([]string{singularName, pluralName}, shortNames...)
	for _, n := range names {
		fmtBare := n                                                                // e.g. deployment
		fmtWithGroup := strings.Join([]string{n, gv.Group}, ".")                    // e.g. deployment.apps
		fmtWithGroupVersion := strings.Join([]string{n, gv.Version, gv.Group}, ".") // e.g. deployment.v1.apps

		out = append(out,
			fmtBare, fmtWithGroup, fmtWithGroupVersion)
	}
	return out
}
