package url

import (
	"fmt"
	"reflect"

	applabels "github.com/openshift/odo/pkg/application/labels"
	"github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/pkg/errors"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/klog"
)

const apiVersion = "odo.dev/v1alpha1"

// ListPushed lists the URLs in an application that are in cluster. The results can further be narrowed
/// down if a component name is provided, which will only list URLs for the
// given component
func ListPushed(client *occlient.Client, componentName string, applicationName string) (URLList, error) {

	labelSelector := fmt.Sprintf("%v=%v", applabels.ApplicationLabel, applicationName)

	if componentName != "" {
		labelSelector = labelSelector + fmt.Sprintf(",%v=%v", componentlabels.ComponentLabel, componentName)
	}

	klog.V(4).Infof("Listing routes with label selector: %v", labelSelector)
	routes, err := client.ListRoutes(labelSelector)

	if err != nil {
		return URLList{}, errors.Wrap(err, "unable to list route names")
	}

	var urls []URL
	for _, r := range routes {
		if r.OwnerReferences != nil && r.OwnerReferences[0].Kind == "Ingress" {
			continue
		}
		a := NewURL(r)
		urls = append(urls, a)
	}

	urlList := NewURLList(urls)
	return urlList, nil

}

// ListPushedIngress lists the ingress URLs on cluster for the given component
func ListPushedIngress(client kclient.ClientInterface, componentName string) (URLList, error) {
	labelSelector := fmt.Sprintf("%v=%v", componentlabels.ComponentLabel, componentName)
	klog.V(4).Infof("Listing ingresses with label selector: %v", labelSelector)
	ingresses, err := client.ListIngresses(labelSelector)
	if err != nil {
		return URLList{}, fmt.Errorf("unable to list ingress names %w", err)
	}

	var urls []URL
	urls = append(urls, NewURLsFromKubernetesIngressList(ingresses)...)
	urlList := NewURLList(urls)
	return urlList, nil
}

type sortableURLs []URL

func (s sortableURLs) Len() int {
	return len(s)
}

func (s sortableURLs) Less(i, j int) bool {
	return s[i].Name <= s[j].Name
}

// generic contains information required for all the URL clients
type generic struct {
	appName       string
	componentName string
	localConfig   localConfigProvider.LocalConfigProvider
}

type Client interface {
	Create(url URL) (string, error)
	Delete(string, localConfigProvider.URLKind) error
	ListFromCluster() (URLList, error)
	List() (URLList, error)
}

type ClientOptions struct {
	OCClient            occlient.Client
	IsRouteSupported    bool
	LocalConfigProvider localConfigProvider.LocalConfigProvider
	Deployment          *v1.Deployment
}

// NewClient gets the appropriate URL client based on the parameters
func NewClient(options ClientOptions) Client {
	var genericInfo generic

	if options.LocalConfigProvider != nil {
		genericInfo = generic{
			appName:       options.LocalConfigProvider.GetApplication(),
			componentName: options.LocalConfigProvider.GetName(),
			localConfig:   options.LocalConfigProvider,
		}
	}

	if options.Deployment != nil {
		genericInfo.appName = options.Deployment.Labels[applabels.ApplicationLabel]
		genericInfo.componentName = options.Deployment.Labels[labels.ComponentLabel]
	}

	return kubernetesClient{
		generic:          genericInfo,
		isRouteSupported: options.IsRouteSupported,
		client:           options.OCClient,
	}
}

type PushParameters struct {
	LocalConfig      localConfigProvider.LocalConfigProvider
	URLClient        Client
	IsRouteSupported bool
}

// Push creates and deletes the required URLs
func Push(parameters PushParameters) error {
	urlLOCAL := make(map[string]URL)

	localConfigURLs, err := parameters.LocalConfig.ListURLs()
	if err != nil {
		return err
	}

	// get the local URLs
	for _, url := range localConfigURLs {
		if !parameters.IsRouteSupported && url.Kind == localConfigProvider.ROUTE {
			// display warning since Host info is missing
			log.Warningf("Unable to create ingress, missing host information for Endpoint %v, please check instructions on URL creation (refer `odo url create --help`)\n", url.Name)
			continue
		}

		urlLOCAL[url.Name] = NewURLFromLocalURL(url)
	}

	log.Info("\nApplying URL changes")

	urlCLUSTER := make(map[string]URL)

	// get the URLs on the cluster
	urlList, err := parameters.URLClient.ListFromCluster()
	if err != nil {
		return err
	}

	for _, url := range urlList.Items {
		urlCLUSTER[url.Name] = url
	}

	urlChange := false

	// find URLs to delete
	for urlName, urlSpec := range urlCLUSTER {
		val, ok := urlLOCAL[urlName]
		configMismatch := false
		if ok {
			// since the host stored in an ingress
			// is the combination of name and host of the url
			if val.Spec.Kind == localConfigProvider.INGRESS {
				// in case of a secure ingress type URL with no user given tls secret
				// the default secret name is used during creation
				// thus setting it to the local URLs to avoid config mismatch
				if val.Spec.Secure && val.Spec.TLSSecret == "" {
					val.Spec.TLSSecret = getDefaultTLSSecretName(parameters.LocalConfig.GetName(), parameters.LocalConfig.GetApplication())
				}
				val.Spec.Host = fmt.Sprintf("%v.%v", urlName, val.Spec.Host)
			} else if val.Spec.Kind == localConfigProvider.ROUTE {
				// we don't allow the host input for route based URLs
				// removing it for the urls from the cluster to avoid config mismatch
				urlSpec.Spec.Host = ""
			}

			if val.Spec.Protocol == "" {
				if val.Spec.Secure {
					val.Spec.Protocol = "https"
				} else {
					val.Spec.Protocol = "http"
				}
			}

			if !reflect.DeepEqual(val.Spec, urlSpec.Spec) {
				configMismatch = true
				klog.V(4).Infof("config and cluster mismatch for url %s", urlName)
			}
		}

		if !ok || configMismatch {
			// delete the url
			err := parameters.URLClient.Delete(urlName, urlSpec.Spec.Kind)
			if err != nil {
				return err
			}
			log.Successf("URL %s successfully deleted", urlName)
			urlChange = true
			delete(urlCLUSTER, urlName)
			continue
		}
	}

	// find URLs to create
	for urlName, urlInfo := range urlLOCAL {
		_, ok := urlCLUSTER[urlName]
		if !ok {
			host, err := parameters.URLClient.Create(urlInfo)
			if err != nil {
				return err
			}
			log.Successf("URL %s: %s%s created", urlName, host, urlInfo.Spec.Path)
			urlChange = true
		}
	}

	if !urlChange {
		log.Success("URLs are synced with the cluster, no changes are required.")
	}

	return nil
}
