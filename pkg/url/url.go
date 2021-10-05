package url

import (
	"fmt"
	"reflect"

	applabels "github.com/openshift/odo/pkg/application/labels"
	"github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/klog"
)

const apiVersion = "odo.dev/v1alpha1"

// generic contains information required for all the URL clients
type generic struct {
	appName             string
	componentName       string
	localConfigProvider localConfigProvider.LocalConfigProvider
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
			appName:             options.LocalConfigProvider.GetApplication(),
			componentName:       options.LocalConfigProvider.GetName(),
			localConfigProvider: options.LocalConfigProvider,
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
	LocalConfigProvider localConfigProvider.LocalConfigProvider
	URLClient           Client
	IsRouteSupported    bool
}

// Push creates and deletes the required URLs
func Push(parameters PushParameters) error {
	urlLOCAL := make(map[string]URL)

	localConfigProviderURLs, err := parameters.LocalConfigProvider.ListURLs()
	if err != nil {
		return err
	}

	// get the local URLs
	for _, url := range localConfigProviderURLs {
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
					val.Spec.TLSSecret = getDefaultTLSSecretName(parameters.LocalConfigProvider.GetName(), parameters.LocalConfigProvider.GetApplication())
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
