package url

import (
	"fmt"

	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

type s2iClient struct {
	generic
	client occlient.Client
}

func (s s2iClient) ListCluster() (URLList, error) {
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
		if r.OwnerReferences != nil && r.OwnerReferences[0].Kind == "Ingress" {
			continue
		}
		a := getMachineReadableFormat(r)
		urls = append(urls, a)
	}

	urlList := getMachineReadableFormatForList(urls)
	return urlList, nil
}

func (s s2iClient) List() (URLList, error) {
	var urls []URL

	clusterUrls, err := s.ListCluster()
	if err != nil {
		return URLList{}, errors.Wrap(err, "unable to list route names")
	}

	for _, clusterURL := range clusterUrls.Items {
		var found bool = false
		for _, configURL := range s.localConfig.ListURLs() {
			localURL := ConvertConfigURL(configURL)
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

	for _, configURL := range s.localConfig.ListURLs() {
		localURL := ConvertConfigURL(configURL)
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

	urlList := getMachineReadableFormatForList(urls)
	return urlList, nil
}
