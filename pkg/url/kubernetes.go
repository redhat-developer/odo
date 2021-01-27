package url

import (
	"fmt"
	"sort"

	routev1 "github.com/openshift/api/route/v1"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

// kubernetesClient contains information required for devfile based URL based operations
type kubernetesClient struct {
	generic
	isRouteSupported bool
	client           occlient.Client
}

// ListCluster lists both route and ingress based URLs from the cluster
func (k kubernetesClient) ListFromCluster() (URLList, error) {
	labelSelector := fmt.Sprintf("%v=%v", componentlabels.ComponentLabel, k.componentName)
	klog.V(4).Infof("Listing ingresses with label selector: %v", labelSelector)
	ingresses, err := k.client.GetKubeClient().ListIngresses(labelSelector)
	if err != nil {
		return URLList{}, errors.Wrap(err, "unable to list ingress")
	}

	var routes []routev1.Route
	if k.isRouteSupported {
		routes, err = k.client.ListRoutes(labelSelector)
		if err != nil {
			return URLList{}, errors.Wrap(err, "unable to list routes")
		}
	}

	var clusterURLs []URL
	for _, i := range ingresses {
		clusterURL := getMachineReadableFormatIngress(i)
		clusterURLs = append(clusterURLs, clusterURL)
	}
	for _, r := range routes {
		// ignore the routes created by ingresses
		if r.OwnerReferences != nil && r.OwnerReferences[0].Kind == "Ingress" {
			continue
		}
		clusterURL := getMachineReadableFormat(r)
		clusterURLs = append(clusterURLs, clusterURL)
	}

	return getMachineReadableFormatForList(clusterURLs), nil
}

// List lists both route/ingress based URLs and local URLs with respective states
func (k kubernetesClient) List() (URLList, error) {
	// get the URLs present on the cluster
	clusterURLMap := make(map[string]URL)
	clusterURLs, err := k.ListFromCluster()
	if err != nil {
		return URLList{}, errors.Wrap(err, "unable to list routes")
	}

	for _, url := range clusterURLs.Items {
		clusterURLMap[getValidURLName(url.Name)] = url
	}

	localMap := make(map[string]URL)
	if k.localConfig != nil {
		// get the URLs present on the localConfigProvider
		localURLS := k.localConfig.ListURLs()
		for _, url := range localURLS {
			if !k.isRouteSupported && url.Kind == localConfigProvider.ROUTE {
				continue
			}
			localURL := ConvertEnvinfoURL(url, k.componentName)
			// use the trimmed URL Name as the key since remote URLs' names are trimmed
			trimmedURLName := getValidURLName(url.Name)
			localMap[trimmedURLName] = localURL
		}
	}

	// find the URLs which are present on the cluster but not on the localConfigProvider
	// if not found on the localConfigProvider, mark them as 'StateTypeLocallyDeleted'
	// else mark them as 'StateTypePushed'
	var urls sortableURLs
	for URLName, clusterURL := range clusterURLMap {
		_, found := localMap[URLName]
		if found {
			// URL is in both local env file and cluster
			clusterURL.Status.State = StateTypePushed
			urls = append(urls, clusterURL)
		} else {
			// URL is on the cluster but not in local env file
			clusterURL.Status.State = StateTypeLocallyDeleted
			urls = append(urls, clusterURL)
		}
	}

	// find the URLs which are present on the localConfigProvider but not on the cluster
	// if not found on the cluster, mark them as 'StateTypeNotPushed'
	for localName, localURL := range localMap {
		_, remoteURLFound := clusterURLMap[localName]
		if !remoteURLFound {
			// URL is in the local env file but not pushed to cluster
			localURL.Status.State = StateTypeNotPushed
			urls = append(urls, localURL)
		}
	}

	// sort urls by name to get consistent output
	sort.Sort(urls)
	urlList := getMachineReadableFormatForList(urls)
	return urlList, nil
}
