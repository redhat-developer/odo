package url

import (
	"fmt"
	"net"
	"reflect"
	"strconv"

	"github.com/openshift/odo/pkg/urltype"

	"github.com/openshift/odo/pkg/log"

	routev1 "github.com/openshift/api/route/v1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/occlient"
	urlLabels "github.com/openshift/odo/pkg/url/labels"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	iextensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog"
)

const apiVersion = "odo.dev/v1alpha1"

func getURLTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{Kind: "url", APIVersion: apiVersion}
}

// ListPushed lists the URLs in an application that are in cluster. The results can further be narrowed
/// down if a component name is provided, which will only list URLs for the
// given component
func ListPushed(client *occlient.Client, componentName string, applicationName string) (urltype.URLList, error) {

	labelSelector := fmt.Sprintf("%v=%v", applabels.ApplicationLabel, applicationName)

	if componentName != "" {
		labelSelector = labelSelector + fmt.Sprintf(",%v=%v", componentlabels.ComponentLabel, componentName)
	}

	klog.V(4).Infof("Listing routes with label selector: %v", labelSelector)
	routes, err := client.ListRoutes(labelSelector)

	if err != nil {
		return urltype.URLList{}, errors.Wrap(err, "unable to list route names")
	}

	var urls []urltype.URL
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

// ListPushedIngress lists the ingress URLs on cluster for the given component
func ListPushedIngress(client *kclient.Client, componentName string) (urltype.URLList, error) {
	labelSelector := fmt.Sprintf("%v=%v", componentlabels.ComponentLabel, componentName)
	klog.V(4).Infof("Listing ingresses with label selector: %v", labelSelector)
	ingresses, err := client.ListIngresses(labelSelector)
	if err != nil {
		return urltype.URLList{}, fmt.Errorf("unable to list ingress names %w", err)
	}

	var urls []urltype.URL
	for _, i := range ingresses {
		urls = append(urls, i.GetURL())
	}

	urlList := getMachineReadableFormatForList(urls)
	return urlList, nil
}

type sortableURLs []urltype.URL

func (s sortableURLs) Len() int {
	return len(s)
}

func (s sortableURLs) Less(i, j int) bool {
	return s[i].Name <= s[j].Name
}

func (s sortableURLs) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// GetProtocol returns the protocol string
func GetProtocol(route routev1.Route, ingress iextensionsv1.Ingress) string {
	if !reflect.DeepEqual(ingress, iextensionsv1.Ingress{}) && ingress.Spec.TLS != nil {
		return "https"
	} else if !reflect.DeepEqual(route, routev1.Route{}) && route.Spec.TLS != nil {
		return "https"
	}
	return "http"
}

// ConvertConfigURL converts ConfigURL to URL
func ConvertConfigURL(configURL localConfigProvider.LocalURL) urltype.URL {
	return urltype.URL{
		TypeMeta: metav1.TypeMeta{
			Kind:       "url",
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: configURL.Name,
		},
		Spec: urltype.URLSpec{
			Port:   configURL.Port,
			Secure: configURL.Secure,
			Kind:   localConfigProvider.ROUTE,
			Path:   "/",
		},
	}
}

// ConvertEnvinfoURL converts EnvinfoURL to URL
func ConvertEnvinfoURL(envinfoURL localConfigProvider.LocalURL, serviceName string) urltype.URL {
	hostString := fmt.Sprintf("%s.%s", envinfoURL.Name, envinfoURL.Host)
	// default to route kind if none is provided
	kind := envinfoURL.Kind
	if kind == "" {
		kind = localConfigProvider.ROUTE
	}
	url := urltype.URL{
		TypeMeta: metav1.TypeMeta{
			Kind:       "url",
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: envinfoURL.Name,
		},
		Spec: urltype.URLSpec{
			Host:      envinfoURL.Host,
			Protocol:  envinfoURL.Protocol,
			Port:      envinfoURL.Port,
			Secure:    envinfoURL.Secure,
			Kind:      kind,
			TLSSecret: envinfoURL.TLSSecret,
			Path:      envinfoURL.Path,
		},
	}
	if kind == localConfigProvider.INGRESS {
		url.Spec.Host = hostString
		if envinfoURL.Secure && len(envinfoURL.TLSSecret) > 0 {
			url.Spec.TLSSecret = envinfoURL.TLSSecret
		} else if envinfoURL.Secure {
			url.Spec.TLSSecret = fmt.Sprintf("%s-tlssecret", serviceName)
		}
	}
	return url
}

// ConvertLocalURL converts localConfigProvider.LocalURL to URL
func ConvertLocalURL(localURL localConfigProvider.LocalURL) urltype.URL {
	return urltype.URL{
		TypeMeta: metav1.TypeMeta{
			Kind:       "url",
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: localURL.Name,
		},
		Spec: urltype.URLSpec{
			Host:      localURL.Host,
			Protocol:  localURL.Protocol,
			Port:      localURL.Port,
			Secure:    localURL.Secure,
			Kind:      localURL.Kind,
			TLSSecret: localURL.TLSSecret,
			Path:      localURL.Path,
		},
	}
}

// GetURLString returns a string representation of given url
func GetURLString(protocol, URL, ingressDomain string, isS2I bool) string {
	if protocol == "" && URL == "" && ingressDomain == "" {
		return ""
	}
	if !isS2I && URL == "" {
		return protocol + "://" + ingressDomain
	}
	return protocol + "://" + URL
}

// Exists checks if the url exists in the component or not
// urlName is the name of the url for checking
// componentName is the name of the component to which the url's existence is checked
// applicationName is the name of the application to which the url's existence is checked
func Exists(client *occlient.Client, urlName string, componentName string, applicationName string) (bool, error) {
	urls, err := ListPushed(client, componentName, applicationName)
	if err != nil {
		return false, errors.Wrap(err, "unable to list the urls")
	}

	for _, url := range urls.Items {
		if url.Name == urlName {
			return true, nil
		}
	}
	return false, nil
}

// GetValidExposedPortNumber checks if the given exposed port number is a valid port or not
// if exposed port is not provided, a random free port will be generated and returned
func GetValidExposedPortNumber(exposedPort int) (int, error) {
	// exposed port number will be -1 if the user doesn't specify any port
	if exposedPort == -1 {
		freePort, err := util.HTTPGetFreePort()
		if err != nil {
			return -1, err
		}
		return freePort, nil
	} else {
		// check if the given port is available
		listener, err := net.Listen("tcp", ":"+strconv.Itoa(exposedPort))
		if err != nil {
			return -1, errors.Wrapf(err, "given port %d is not available, please choose another port", exposedPort)
		}
		defer listener.Close()
		return exposedPort, nil
	}
}

// getMachineReadableFormat gives machine readable URL definition
func getMachineReadableFormat(r routev1.Route) urltype.URL {
	return urltype.URL{
		TypeMeta:   metav1.TypeMeta{Kind: "url", APIVersion: apiVersion},
		ObjectMeta: metav1.ObjectMeta{Name: r.Labels[urlLabels.URLLabel]},
		Spec:       urltype.URLSpec{Host: r.Spec.Host, Port: r.Spec.Port.TargetPort.IntValue(), Protocol: GetProtocol(r, iextensionsv1.Ingress{}), Secure: r.Spec.TLS != nil, Path: r.Spec.Path, Kind: localConfigProvider.ROUTE},
	}

}

func getMachineReadableFormatForList(urls []urltype.URL) urltype.URLList {
	return urltype.URLList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: apiVersion,
		},
		ListMeta: metav1.ListMeta{},
		Items:    urls,
	}
}

func getMachineReadableFormatExtensionV1Ingress(i iextensionsv1.Ingress) urltype.URL {
	url := urltype.URL{
		TypeMeta:   getURLTypeMeta(),
		ObjectMeta: metav1.ObjectMeta{Name: i.Labels[urlLabels.URLLabel]},
		Spec:       urltype.URLSpec{Host: i.Spec.Rules[0].Host, Port: int(i.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort.IntVal), Secure: i.Spec.TLS != nil, Path: i.Spec.Rules[0].HTTP.Paths[0].Path, Kind: localConfigProvider.INGRESS},
	}
	if i.Spec.TLS != nil {
		url.Spec.TLSSecret = i.Spec.TLS[0].SecretName
		url.Spec.Protocol = "https"
	} else {
		url.Spec.Protocol = "http"
	}
	return url

}

// getDefaultTLSSecretName returns the name of the default tls secret name
func getDefaultTLSSecretName(componentName string) string {
	return componentName + "-tlssecret"
}

// ConvertExtensionV1IngressURLToIngress converts IngressURL to Ingress
func ConvertExtensionV1IngressURLToIngress(ingressURL urltype.URL, serviceName string) iextensionsv1.Ingress {
	port := intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: int32(ingressURL.Spec.Port),
	}
	ingress := iextensionsv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: ingressURL.Name,
		},
		Spec: iextensionsv1.IngressSpec{
			Rules: []iextensionsv1.IngressRule{
				{
					Host: ingressURL.Spec.Host,
					IngressRuleValue: iextensionsv1.IngressRuleValue{
						HTTP: &iextensionsv1.HTTPIngressRuleValue{
							Paths: []iextensionsv1.HTTPIngressPath{
								{
									Path: ingressURL.Spec.Path,
									Backend: iextensionsv1.IngressBackend{
										ServiceName: serviceName,
										ServicePort: port,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	if len(ingressURL.Spec.TLSSecret) > 0 {
		ingress.Spec.TLS = []iextensionsv1.IngressTLS{
			{
				Hosts: []string{
					ingressURL.Spec.Host,
				},
				SecretName: ingressURL.Spec.TLSSecret,
			},
		}
	}
	return ingress
}

type PushParameters struct {
	LocalConfig      localConfigProvider.LocalConfigProvider
	URLClient        Client
	IsRouteSupported bool
}

// Push creates and deletes the required URLs
func Push(parameters PushParameters) error {
	urlLOCAL := make(map[string]urltype.URL)

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

		urlLOCAL[url.Name] = ConvertLocalURL(url)
	}

	log.Info("\nApplying URL changes")

	urlCLUSTER := make(map[string]urltype.URL)

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
					val.Spec.TLSSecret = getDefaultTLSSecretName(parameters.LocalConfig.GetName())
				}
				val.Spec.Host = fmt.Sprintf("%v.%v", urlName, val.Spec.Host)
			} else if val.Spec.Kind == localConfigProvider.ROUTE {
				// we don't allow the host input for route based URLs
				// removing it for the urls from the cluster to avoid config mismatch
				urlSpec.Spec.Host = ""

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

type ClientOptions struct {
	OCClient            occlient.Client
	IsRouteSupported    bool
	LocalConfigProvider localConfigProvider.LocalConfigProvider
}

type Client interface {
	Create(url urltype.URL) (string, error)
	Delete(string, localConfigProvider.URLKind) error
	ListFromCluster() (urltype.URLList, error)
	List() (urltype.URLList, error)
}

// NewClient gets the appropriate URL client based on the parameters
func NewClient(options ClientOptions) Client {
	genericInfo := generic{
		appName:       options.LocalConfigProvider.GetApplication(),
		componentName: options.LocalConfigProvider.GetName(),
		localConfig:   options.LocalConfigProvider,
	}

	if _, ok := options.LocalConfigProvider.(*config.LocalConfigInfo); ok {
		return s2iClient{
			generic: genericInfo,
			client:  options.OCClient,
		}
	} else {
		return kubernetesClient{
			generic:          genericInfo,
			isRouteSupported: options.IsRouteSupported,
			client:           options.OCClient,
		}
	}
}

// generic contains information required for all the URL clients
type generic struct {
	appName       string
	componentName string
	localConfig   localConfigProvider.LocalConfigProvider
}
