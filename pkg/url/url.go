package url

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strconv"

	"github.com/openshift/odo/pkg/log"

	"github.com/devfile/library/pkg/devfile/generator"
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
	kerrors "k8s.io/apimachinery/pkg/api/errors"

	iextensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog"
)

const apiVersion = "odo.dev/v1alpha1"

// Get returns URL definition for given URL name
func (urls URLList) Get(urlName string) URL {
	for _, url := range urls.Items {
		if url.Name == urlName {
			return url
		}
	}
	return URL{}

}

// Delete deletes a URL
func Delete(client *occlient.Client, kClient *kclient.Client, urlName string, applicationName string, urlType localConfigProvider.URLKind, isS2i bool) error {
	if urlType == localConfigProvider.INGRESS {
		return kClient.DeleteIngress(urlName)
	} else if urlType == localConfigProvider.ROUTE {
		if isS2i {
			// Namespace the URL name
			var err error
			urlName, err = util.NamespaceOpenShiftObject(urlName, applicationName)
			if err != nil {
				return errors.Wrapf(err, "unable to create namespaced name")
			}
		}

		return client.DeleteRoute(urlName)
	}
	return errors.New("url type is not supported")
}

type CreateParameters struct {
	urlName         string
	portNumber      int
	secureURL       bool
	componentName   string
	applicationName string
	host            string
	secretName      string
	urlKind         localConfigProvider.URLKind
	path            string
}

// Create creates a URL and returns url string and error if any
// portNumber is the target port number for the route and is -1 in case no port number is specified in which case it is automatically detected for components which expose only one service port)
func Create(client *occlient.Client, kClient *kclient.Client, parameters CreateParameters, isRouteSupported bool, isS2I bool) (string, error) {

	if parameters.urlKind != localConfigProvider.INGRESS && parameters.urlKind != localConfigProvider.ROUTE {
		return "", fmt.Errorf("urlKind %s is not supported for URL creation", parameters.urlKind)
	}

	if !parameters.secureURL && parameters.secretName != "" {
		return "", fmt.Errorf("secret name can only be used for secure URLs")
	}

	labels := urlLabels.GetLabels(parameters.urlName, parameters.componentName, parameters.applicationName, true)

	serviceName := ""

	if !isS2I && parameters.urlKind == localConfigProvider.INGRESS && kClient != nil {
		if parameters.host == "" {
			return "", errors.Errorf("the host cannot be empty")
		}
		serviceName := parameters.componentName
		ingressDomain := fmt.Sprintf("%v.%v", parameters.urlName, parameters.host)

		deployment, err := kClient.GetOneDeployment(parameters.componentName, parameters.applicationName)
		if err != nil {
			return "", err
		}
		ownerReference := generator.GetOwnerReference(deployment)
		if parameters.secureURL {
			if len(parameters.secretName) != 0 {
				_, err := kClient.KubeClient.CoreV1().Secrets(kClient.Namespace).Get(context.TODO(), parameters.secretName, metav1.GetOptions{})
				if err != nil {
					return "", errors.Wrap(err, "unable to get the provided secret: "+parameters.secretName)
				}
			}
			if len(parameters.secretName) == 0 {
				defaultTLSSecretName := parameters.componentName + "-tlssecret"
				_, err := kClient.KubeClient.CoreV1().Secrets(kClient.Namespace).Get(context.TODO(), defaultTLSSecretName, metav1.GetOptions{})
				// create tls secret if it does not exist
				if kerrors.IsNotFound(err) {
					selfsignedcert, err := kclient.GenerateSelfSignedCertificate(parameters.host)
					if err != nil {
						return "", errors.Wrap(err, "unable to generate self-signed certificate for clutser: "+parameters.host)
					}
					// create tls secret
					secretlabels := componentlabels.GetLabels(parameters.componentName, parameters.applicationName, true)
					objectMeta := metav1.ObjectMeta{
						Name:   defaultTLSSecretName,
						Labels: secretlabels,
						OwnerReferences: []v1.OwnerReference{
							ownerReference,
						},
					}
					secret, err := kClient.CreateTLSSecret(selfsignedcert.CertPem, selfsignedcert.KeyPem, objectMeta)
					if err != nil {
						return "", errors.Wrap(err, "unable to create tls secret")
					}
					parameters.secretName = secret.Name
				} else if err != nil {
					return "", err
				} else {
					// tls secret found for this component
					parameters.secretName = defaultTLSSecretName
				}

			}

		}

		objectMeta := generator.GetObjectMeta(parameters.componentName, kClient.Namespace, labels, nil)
		// to avoid error due to duplicate ingress name defined in different devfile components
		objectMeta.Name = fmt.Sprintf("%s-%s", parameters.urlName, parameters.componentName)
		objectMeta.OwnerReferences = append(objectMeta.OwnerReferences, ownerReference)

		ingressParam := generator.IngressParams{
			ObjectMeta: objectMeta,
			IngressSpecParams: generator.IngressSpecParams{
				ServiceName:   serviceName,
				IngressDomain: ingressDomain,
				PortNumber:    intstr.FromInt(parameters.portNumber),
				TLSSecretName: parameters.secretName,
				Path:          parameters.path,
			},
		}
		ingress := generator.GetIngress(ingressParam)
		// Pass in the namespace name, link to the service (componentName) and labels to create a ingress
		i, err := kClient.CreateIngress(*ingress)
		if err != nil {
			return "", errors.Wrap(err, "unable to create ingress")
		}
		return GetURLString(GetProtocol(routev1.Route{}, *i), "", ingressDomain, false), nil
	} else {
		if !isRouteSupported {
			return "", errors.Errorf("routes are not available on non OpenShift clusters")
		}

		var ownerReference metav1.OwnerReference
		if isS2I || kClient == nil {
			var err error
			parameters.urlName, err = util.NamespaceOpenShiftObject(parameters.urlName, parameters.applicationName)
			if err != nil {
				return "", errors.Wrapf(err, "unable to create namespaced name")
			}
			serviceName, err = util.NamespaceOpenShiftObject(parameters.componentName, parameters.applicationName)
			if err != nil {
				return "", errors.Wrapf(err, "unable to create namespaced name")
			}

			// since the serviceName is same as the DC name, we use that to get the DC
			// to which this route belongs. A better way could be to get service from
			// the name and set it as owner of the route
			dc, err := client.GetDeploymentConfigFromName(serviceName)
			if err != nil {
				return "", errors.Wrapf(err, "unable to get DeploymentConfig %s", serviceName)
			}

			ownerReference = occlient.GenerateOwnerReference(dc)
		} else {
			// to avoid error due to duplicate ingress name defined in different devfile components
			parameters.urlName = fmt.Sprintf("%s-%s", parameters.urlName, parameters.componentName)
			serviceName = parameters.componentName

			deployment, err := kClient.GetOneDeployment(parameters.componentName, parameters.applicationName)
			if err != nil {
				return "", err
			}
			ownerReference = generator.GetOwnerReference(deployment)
		}

		// Pass in the namespace name, link to the service (componentName) and labels to create a route
		route, err := client.CreateRoute(parameters.urlName, serviceName, intstr.FromInt(parameters.portNumber), labels, parameters.secureURL, parameters.path, ownerReference)
		if err != nil {
			return "", errors.Wrap(err, "unable to create route")
		}
		return GetURLString(GetProtocol(*route, iextensionsv1.Ingress{}), route.Spec.Host, "", true), nil
	}

}

// ListPushed lists the URLs in an application that are in cluster. The results can further be narrowed
// down if a component name is provided, which will only list URLs for the
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
		a := getMachineReadableFormat(r)
		urls = append(urls, a)
	}

	urlList := getMachineReadableFormatForList(urls)
	return urlList, nil

}

// ListPushedIngress lists the ingress URLs on cluster for the given component
func ListPushedIngress(client *kclient.Client, componentName string) (URLList, error) {
	labelSelector := fmt.Sprintf("%v=%v", componentlabels.ComponentLabel, componentName)
	klog.V(4).Infof("Listing ingresses with label selector: %v", labelSelector)
	ingresses, err := client.ListIngresses(labelSelector)
	if err != nil {
		return URLList{}, errors.Wrap(err, "unable to list ingress names")
	}

	var urls []URL
	for _, i := range ingresses {
		a := getMachineReadableFormatIngress(i)
		urls = append(urls, a)
	}

	urlList := getMachineReadableFormatForList(urls)
	return urlList, nil
}

type sortableURLs []URL

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
func ConvertConfigURL(configURL localConfigProvider.LocalURL) URL {
	return URL{
		TypeMeta: metav1.TypeMeta{
			Kind:       "url",
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: configURL.Name,
		},
		Spec: URLSpec{
			Port:   configURL.Port,
			Secure: configURL.Secure,
			Kind:   localConfigProvider.ROUTE,
			Path:   "/",
		},
	}
}

// ConvertEnvinfoURL converts EnvinfoURL to URL
func ConvertEnvinfoURL(envinfoURL localConfigProvider.LocalURL, serviceName string) URL {
	hostString := fmt.Sprintf("%s.%s", envinfoURL.Name, envinfoURL.Host)
	// default to route kind if none is provided
	kind := envinfoURL.Kind
	if kind == "" {
		kind = localConfigProvider.ROUTE
	}
	url := URL{
		TypeMeta: metav1.TypeMeta{
			Kind:       "url",
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: envinfoURL.Name,
		},
		Spec: URLSpec{
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
func ConvertLocalURL(localURL localConfigProvider.LocalURL) URL {
	return URL{
		TypeMeta: metav1.TypeMeta{
			Kind:       "url",
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: localURL.Name,
		},
		Spec: URLSpec{
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
func getMachineReadableFormat(r routev1.Route) URL {
	return URL{
		TypeMeta:   metav1.TypeMeta{Kind: "url", APIVersion: apiVersion},
		ObjectMeta: metav1.ObjectMeta{Name: r.Labels[urlLabels.URLLabel]},
		Spec:       URLSpec{Host: r.Spec.Host, Port: r.Spec.Port.TargetPort.IntValue(), Protocol: GetProtocol(r, iextensionsv1.Ingress{}), Secure: r.Spec.TLS != nil, Path: r.Spec.Path, Kind: localConfigProvider.ROUTE},
	}

}

func getMachineReadableFormatForList(urls []URL) URLList {
	return URLList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: apiVersion,
		},
		ListMeta: metav1.ListMeta{},
		Items:    urls,
	}
}

func getMachineReadableFormatIngress(i iextensionsv1.Ingress) URL {
	url := URL{
		TypeMeta:   metav1.TypeMeta{Kind: "url", APIVersion: apiVersion},
		ObjectMeta: metav1.ObjectMeta{Name: i.Labels[urlLabels.URLLabel]},
		Spec:       URLSpec{Host: i.Spec.Rules[0].Host, Port: int(i.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort.IntVal), Secure: i.Spec.TLS != nil, Path: i.Spec.Rules[0].HTTP.Paths[0].Path, Kind: localConfigProvider.INGRESS},
	}
	if i.Spec.TLS != nil {
		url.Spec.TLSSecret = i.Spec.TLS[0].SecretName
	}
	return url

}

// ConvertIngressURLToIngress converts IngressURL to Ingress
func ConvertIngressURLToIngress(ingressURL URL, serviceName string) iextensionsv1.Ingress {
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
	IsS2I            bool
}

// Push creates and deletes the required URLs
func Push(client *occlient.Client, parameters PushParameters) error {
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

		urlLOCAL[url.Name] = ConvertLocalURL(url)
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
				val.Spec.Host = fmt.Sprintf("%v.%v", urlName, val.Spec.Host)
			} else if val.Spec.Kind == localConfigProvider.ROUTE {
				// we don't allow the host input for route based URLs
				// based removing it for the urls from the cluster to avoid config mismatch
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
			if urlSpec.Spec.Kind == localConfigProvider.INGRESS && client.GetKubeClient() == nil {
				continue
			}
			// delete the url
			deleteURLName := urlName
			if !parameters.IsS2I && client.GetKubeClient() != nil {
				// route/ingress name is defined as <urlName>-<componentName>
				// to avoid error due to duplicate ingress name defined in different devfile components
				deleteURLName = fmt.Sprintf("%s-%s", urlName, parameters.LocalConfig.GetName())
			}
			err := Delete(client, client.GetKubeClient(), deleteURLName, parameters.LocalConfig.GetApplication(), urlSpec.Spec.Kind, parameters.IsS2I)
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
			if urlInfo.Spec.Kind == localConfigProvider.INGRESS && client.GetKubeClient() == nil {
				continue
			}
			createParameters := CreateParameters{
				urlName:         urlName,
				portNumber:      urlInfo.Spec.Port,
				secureURL:       urlInfo.Spec.Secure,
				componentName:   parameters.LocalConfig.GetName(),
				applicationName: parameters.LocalConfig.GetApplication(),
				host:            urlInfo.Spec.Host,
				secretName:      urlInfo.Spec.TLSSecret,
				urlKind:         urlInfo.Spec.Kind,
				path:            urlInfo.Spec.Path,
			}
			host, err := Create(client, client.GetKubeClient(), createParameters, parameters.IsRouteSupported, parameters.IsS2I)
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
	ListFromCluster() (URLList, error)
	List() (URLList, error)
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
