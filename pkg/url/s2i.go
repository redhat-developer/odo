package url

import (
	"fmt"

	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/occlient"
	urlLabels "github.com/openshift/odo/pkg/url/labels"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	iextensionsv1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog"
)

// s2iClient contains information required for s2i based URL based operations
type s2iClient struct {
	generic
	client occlient.Client
}

// ListFromCluster lists route based URLs from the cluster
func (s s2iClient) ListFromCluster() (URLList, error) {
	labelSelector := fmt.Sprintf("%v=%v", applabels.ApplicationLabel, s.localConfig.GetApplication())

	if s.localConfig.GetName() != "" {
		labelSelector = labelSelector + fmt.Sprintf(",%v=%v", componentlabels.ComponentLabel, s.localConfig.GetName())
	}

	klog.V(4).Infof("Listing routes with label selector: %v", labelSelector)
	routes, err := s.client.ListRoutes(labelSelector)

	if err != nil {
		return URLList{}, errors.Wrap(err, "unable to list route names")
	}

	var urls []URL
	for _, r := range routes {
		a := NewURL(r)
		urls = append(urls, a)
	}

	urlList := NewURLList(urls)
	return urlList, nil
}

// List lists both route based URLs and local URLs with respective states
func (s s2iClient) List() (URLList, error) {
	var urls []URL

	clusterUrls, err := s.ListFromCluster()
	if err != nil {
		return URLList{}, errors.Wrap(err, "unable to list route names")
	}

	localConfigURLs, err := s.localConfig.ListURLs()
	if err != nil {
		return URLList{}, err
	}

	for _, clusterURL := range clusterUrls.Items {
		var found = false
		for _, configURL := range localConfigURLs {
			localURL := NewURLFromConfigURL(configURL)
			if localURL.Name == clusterURL.Name {
				// URL is in both local config and cluster
				clusterURL.Status.State = StateTypePushed
				urls = append(urls, clusterURL)
				found = true
			}
		}

		if !found {
			// URL is on the cluster but not in local config
			clusterURL.Status.State = StateTypeLocallyDeleted
			urls = append(urls, clusterURL)
		}
	}

	for _, configURL := range localConfigURLs {
		localURL := NewURLFromConfigURL(configURL)
		var found = false
		for _, clusterURL := range clusterUrls.Items {
			if localURL.Name == clusterURL.Name {
				found = true
			}
		}
		if !found {
			// URL is in the local config but not on the cluster
			localURL.Status.State = StateTypeNotPushed
			urls = append(urls, localURL)
		}
	}

	urlList := NewURLList(urls)
	return urlList, nil
}

// Delete deletes the URL with the given name and kind
func (s s2iClient) Delete(name string, kind localConfigProvider.URLKind) error {
	// Namespace the URL name
	routeName, err := util.NamespaceOpenShiftObject(name, s.appName)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespaced name")
	}

	return s.client.DeleteRoute(routeName)
}

// Create creates a route based on the given URL
func (s s2iClient) Create(url URL) (string, error) {
	routeName, err := util.NamespaceOpenShiftObject(url.Name, s.appName)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create namespaced name")
	}
	serviceName, err := util.NamespaceOpenShiftObject(s.componentName, s.appName)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create namespaced name")
	}

	labels := urlLabels.GetLabels(url.Name, s.componentName, s.appName, true)

	// since the serviceName is same as the DC name, we use that to get the DC
	// to which this route belongs. A better way could be to get service from
	// the name and set it as owner of the route
	dc, err := s.client.GetDeploymentConfigFromName(serviceName)
	if err != nil {
		return "", errors.Wrapf(err, "unable to get DeploymentConfig %s", serviceName)
	}

	ownerReference := occlient.GenerateOwnerReference(dc)

	// Pass in the namespace name, link to the service (componentName) and labels to create a route
	route, err := s.client.CreateRoute(routeName, serviceName, intstr.FromInt(url.Spec.Port), labels, url.Spec.Secure, url.Spec.Path, ownerReference)
	if err != nil {
		return "", errors.Wrap(err, "unable to create route")
	}
	return GetURLString(GetProtocol(*route, iextensionsv1.Ingress{}), route.Spec.Host, "", true), nil
}
