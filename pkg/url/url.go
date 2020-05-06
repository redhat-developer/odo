package url

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"

	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/log"

	routev1 "github.com/openshift/api/route/v1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/util/experimental"
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

// Get returns URL definition for given URL name
func (urls URLList) Get(urlName string) URL {
	for _, url := range urls.Items {
		if url.Name == urlName {
			return url
		}
	}
	return URL{}

}

// Get returns URL definition for given URL name
func Get(client *occlient.Client, localConfig *config.LocalConfigInfo, urlName string, applicationName string) (URL, error) {
	remoteUrlName, err := util.NamespaceOpenShiftObject(urlName, applicationName)
	if err != nil {
		return URL{}, errors.Wrapf(err, "unable to create namespaced name")
	}

	// Check whether remote already created the route
	remoteExist := true
	route, err := client.GetRoute(remoteUrlName)
	if err != nil {
		remoteExist = false
	}

	localConfigURLs := localConfig.GetURL()
	for _, configURL := range localConfigURLs {
		localURL := ConvertConfigURL(configURL)
		// search local URL, if it exist in local, update state with remote status
		if localURL.Name == urlName {
			if remoteExist {
				localURL.Status.State = StateTypePushed
			} else {
				localURL.Status.State = StateTypeNotPushed
			}
			return localURL, nil
		}
	}

	if err == nil && remoteExist {
		// Remote exist, but not in local, so it's deleted status
		clusterURL := getMachineReadableFormat(*route)
		clusterURL.Status.State = StateTypeLocallyDeleted
		return clusterURL, nil
	}

	// can't find the URL in local and remote
	return URL{}, errors.New(fmt.Sprintf("the url %v does not exist", urlName))
}

// GetIngress returns ingress spec for given URL name
func GetIngress(kClient *kclient.Client, envSpecificInfo *envinfo.EnvSpecificInfo, urlName string) (iextensionsv1.Ingress, error) {

	// Check whether remote already created the ingress
	ingress, err := kClient.GetIngress(urlName)
	if err == nil {
		return *ingress, nil
	}

	ingresses := envSpecificInfo.GetURL()
	for _, envIngress := range ingresses {
		// search local URL check if it exist in local envinfo
		if envIngress.Name == urlName {
			return iextensionsv1.Ingress{}, errors.New(fmt.Sprintf("the url %v is not created, but exists in local envinfo file. Please run 'odo push'.", urlName))
		}
	}

	// can't find the URL in local and remote
	return iextensionsv1.Ingress{}, errors.New(fmt.Sprintf("the url %v does not exist", urlName))
}

// Delete deletes a URL
func Delete(client *occlient.Client, kClient *kclient.Client, urlName string, applicationName string, urlType envinfo.URLKind) error {
	if urlType == envinfo.INGRESS {
		return kClient.DeleteIngress(urlName)
	} else if urlType == envinfo.ROUTE {
		if applicationName != "" {
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
	urlKind         envinfo.URLKind
}

// Create creates a URL and returns url string and error if any
// portNumber is the target port number for the route and is -1 in case no port number is specified in which case it is automatically detected for components which expose only one service port)
func Create(client *occlient.Client, kClient *kclient.Client, parameters CreateParameters, isRouteSupported bool, isExperimental bool) (string, error) {

	if parameters.urlKind != envinfo.INGRESS && parameters.urlKind != envinfo.ROUTE {
		return "", fmt.Errorf("urlKind %s is not supported for URL creation", parameters.urlKind)
	}

	if !parameters.secureURL && parameters.secretName != "" {
		return "", fmt.Errorf("secret name can only be used for secure URLs")
	}

	labels := urlLabels.GetLabels(parameters.urlName, parameters.componentName, parameters.applicationName, true)

	serviceName := ""

	if isExperimental && parameters.urlKind == envinfo.INGRESS && kClient != nil {
		if parameters.host == "" {
			return "", errors.Errorf("the host cannot be empty")
		}
		serviceName := parameters.componentName
		ingressDomain := fmt.Sprintf("%v.%v", parameters.urlName, parameters.host)
		deployment, err := kClient.GetDeploymentByName(parameters.componentName)
		if err != nil {
			return "", err
		}
		ownerReference := kclient.GenerateOwnerReference(deployment)
		if parameters.secureURL {
			if len(parameters.secretName) != 0 {
				_, err := kClient.KubeClient.CoreV1().Secrets(kClient.Namespace).Get(parameters.secretName, metav1.GetOptions{})
				if err != nil {
					return "", errors.Wrap(err, "unable to get the provided secret: "+parameters.secretName)
				}
			}
			if len(parameters.secretName) == 0 {
				defaultTLSSecretName := parameters.componentName + "-tlssecret"
				_, err := kClient.KubeClient.CoreV1().Secrets(kClient.Namespace).Get(defaultTLSSecretName, metav1.GetOptions{})
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

		ingressParam := kclient.IngressParameter{ServiceName: serviceName, IngressDomain: ingressDomain, PortNumber: intstr.FromInt(parameters.portNumber), TLSSecretName: parameters.secretName}
		ingressSpec := kclient.GenerateIngressSpec(ingressParam)
		objectMeta := kclient.CreateObjectMeta(parameters.componentName, kClient.Namespace, labels, nil)
		objectMeta.Name = parameters.urlName
		objectMeta.OwnerReferences = append(objectMeta.OwnerReferences, ownerReference)
		// Pass in the namespace name, link to the service (componentName) and labels to create a ingress
		ingress, err := kClient.CreateIngress(objectMeta, *ingressSpec)
		if err != nil {
			return "", errors.Wrap(err, "unable to create ingress")
		}
		return GetURLString(GetProtocol(routev1.Route{}, *ingress, isExperimental), "", ingressDomain, isExperimental), nil
	} else {
		if !isRouteSupported {
			return "", errors.Errorf("routes are not available on non OpenShift clusters")
		}

		var ownerReference metav1.OwnerReference
		if !isExperimental || kClient == nil {
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
			serviceName = parameters.componentName

			deployment, err := kClient.GetDeploymentByName(parameters.componentName)
			if err != nil {
				return "", err
			}
			ownerReference = kclient.GenerateOwnerReference(deployment)
		}

		// Pass in the namespace name, link to the service (componentName) and labels to create a route
		route, err := client.CreateRoute(parameters.urlName, serviceName, intstr.FromInt(parameters.portNumber), labels, parameters.secureURL, ownerReference)
		if err != nil {
			return "", errors.Wrap(err, "unable to create route")
		}
		return GetURLString(GetProtocol(*route, iextensionsv1.Ingress{}, isExperimental), route.Spec.Host, "", isExperimental), nil
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

// ListPushedIngress lists the ingress URLs for the given component
func ListPushedIngress(client *kclient.Client, componentName string) (iextensionsv1.IngressList, error) {
	labelSelector := fmt.Sprintf("%v=%v", componentlabels.ComponentLabel, componentName)
	klog.V(4).Infof("Listing ingresses with label selector: %v", labelSelector)
	ingresses, err := client.ListIngresses(labelSelector)
	if err != nil {
		return iextensionsv1.IngressList{}, errors.Wrap(err, "unable to list ingress names")
	}

	var urls []iextensionsv1.Ingress
	for _, i := range ingresses {
		a := getMachineReadableFormatIngress(i)
		urls = append(urls, a)
	}

	urlList := getMachineReadableFormatForIngressList(urls)
	return urlList, nil
}

// List returns all URLs for given component.
// If componentName is empty string, it lists all url in a given application.
func List(client *occlient.Client, localConfig *config.LocalConfigInfo, componentName string, applicationName string) (URLList, error) {

	labelSelector := fmt.Sprintf("%v=%v", applabels.ApplicationLabel, applicationName)

	if componentName != "" {
		labelSelector = labelSelector + fmt.Sprintf(",%v=%v", componentlabels.ComponentLabel, componentName)
	}

	routes, err := client.ListRoutes(labelSelector)
	if err != nil {
		return URLList{}, errors.Wrap(err, "unable to list route names")
	}

	localConfigURLs := localConfig.GetURL()

	var urls []URL

	for _, r := range routes {
		clusterURL := getMachineReadableFormat(r)
		var found bool = false
		for _, configURL := range localConfigURLs {
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

	for _, configURL := range localConfigURLs {
		localURL := ConvertConfigURL(configURL)
		var found bool = false
		for _, r := range routes {
			clusterURL := getMachineReadableFormat(r)
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

// GetProtocol returns the protocol string
func GetProtocol(route routev1.Route, ingress iextensionsv1.Ingress, isExperimental bool) string {
	if isExperimental {
		if ingress.Spec.TLS != nil {
			return "https"
		}
	} else {
		if route.Spec.TLS != nil {
			return "https"
		}
	}
	return "http"
}

// ConvertConfigURL converts ConfigURL to URL
func ConvertConfigURL(configURL config.ConfigURL) URL {
	return URL{
		TypeMeta: metav1.TypeMeta{
			Kind:       "url",
			APIVersion: "odo.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: configURL.Name,
		},
		Spec: URLSpec{
			Port: configURL.Port,
		},
	}
}

// GetURLString returns a string representation of given url
func GetURLString(protocol, URL string, ingressDomain string, isExperimentalMode bool) string {
	if isExperimentalMode && URL == "" {
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

// GetURLName returns a url name from the component name and the given port number
func GetURLName(componentName string, componentPort int) string {
	if componentPort == -1 {
		return componentName
	}
	return fmt.Sprintf("%v-%v", componentName, componentPort)
}

// GetValidPortNumber checks if the given port number is a valid component port or not
// if port number is not provided and the component is a single port component, the component port is returned
// port number is -1 if the user does not specify any port
func GetValidPortNumber(componentName string, portNumber int, portList []string) (int, error) {
	var componentPorts []int
	for _, p := range portList {
		port, err := strconv.Atoi(strings.Split(p, "/")[0])
		if err != nil {
			return port, err
		}
		componentPorts = append(componentPorts, port)
	}
	// port number will be -1 if the user doesn't specify any port
	if portNumber == -1 {
		switch {
		case len(componentPorts) > 1:
			return portNumber, errors.Errorf("port for the component %s is required as it exposes %d ports: %s", componentName, len(componentPorts), strings.Trim(strings.Replace(fmt.Sprint(componentPorts), " ", ",", -1), "[]"))
		case len(componentPorts) == 1:
			return componentPorts[0], nil
		default:
			return portNumber, errors.Errorf("no port is exposed by the component %s", componentName)
		}
	} else {
		for _, port := range componentPorts {
			if portNumber == port {
				return portNumber, nil
			}
		}
	}

	return portNumber, fmt.Errorf("given port %d is not exposed on given component, available ports are: %s", portNumber, strings.Trim(strings.Replace(fmt.Sprint(componentPorts), " ", ",", -1), "[]"))
}

// GetValidExposedPortNumber checks if the given exposed port number is a valid port or not
// if exposed port is not provided, a random free port will be generated and returned
func GetValidExposedPortNumber(exposedPort int) (int, error) {
	// exposed port number will be -1 if the user doesn't specify any port
	if exposedPort == -1 {
		freePort, err := util.HttpGetFreePort()
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
		TypeMeta:   metav1.TypeMeta{Kind: "url", APIVersion: "odo.openshift.io/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{Name: r.Labels[urlLabels.URLLabel]},
		Spec:       URLSpec{Host: r.Spec.Host, Port: r.Spec.Port.TargetPort.IntValue(), Protocol: GetProtocol(r, iextensionsv1.Ingress{}, experimental.IsExperimentalModeEnabled()), Secure: r.Spec.TLS != nil},
	}

}

func getMachineReadableFormatForList(urls []URL) URLList {
	return URLList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: "odo.openshift.io/v1alpha1",
		},
		ListMeta: metav1.ListMeta{},
		Items:    urls,
	}
}

func getMachineReadableFormatIngress(i iextensionsv1.Ingress) iextensionsv1.Ingress {
	return iextensionsv1.Ingress{
		TypeMeta:   metav1.TypeMeta{Kind: "Ingress", APIVersion: "extensions/v1beta1"},
		ObjectMeta: metav1.ObjectMeta{Name: i.Labels[urlLabels.URLLabel]},
		Spec:       iextensionsv1.IngressSpec{TLS: i.Spec.TLS, Rules: i.Spec.Rules},
	}

}

func getMachineReadableFormatForIngressList(ingresses []iextensionsv1.Ingress) iextensionsv1.IngressList {
	return iextensionsv1.IngressList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: "udo.udo.io/v1alpha1",
		},
		ListMeta: metav1.ListMeta{},
		Items:    ingresses,
	}
}

type PushParameters struct {
	ComponentName             string
	ApplicationName           string
	ConfigURLs                []config.ConfigURL
	EnvURLS                   []envinfo.EnvInfoURL
	IsRouteSupported          bool
	IsExperimentalModeEnabled bool
}

// Push creates and deletes the required URLs
func Push(client *occlient.Client, kClient *kclient.Client, parameters PushParameters) error {
	urlLOCAL := make(map[string]URL)

	// in case the component is a s2i one
	// kClient will be nil
	if parameters.IsExperimentalModeEnabled && kClient != nil {
		urls := parameters.EnvURLS
		for _, url := range urls {
			if url.Kind != envinfo.DOCKER {
				urlLOCAL[url.Name] = URL{
					Spec: URLSpec{
						Host:      url.Host,
						Port:      url.Port,
						Secure:    url.Secure,
						tLSSecret: url.TLSSecret,
						urlKind:   url.Kind,
					},
				}
			}
		}
	} else {
		urls := parameters.ConfigURLs
		for _, url := range urls {
			urlLOCAL[url.Name] = URL{
				Spec: URLSpec{
					Port:    url.Port,
					Secure:  url.Secure,
					urlKind: envinfo.ROUTE,
				},
			}
		}
	}

	urlCLUSTER := make(map[string]URL)
	if parameters.IsExperimentalModeEnabled && kClient != nil {
		urlList, err := ListPushedIngress(kClient, parameters.ComponentName)
		if err != nil {
			return err
		}
		for _, url := range urlList.Items {
			urlCLUSTER[url.Name] = URL{
				Spec: URLSpec{
					Port:    int(url.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort.IntVal),
					urlKind: envinfo.INGRESS,
				},
			}
		}
	}

	if parameters.IsRouteSupported {
		urlPushedRoutes, err := ListPushed(client, parameters.ComponentName, parameters.ApplicationName)
		if err != nil {
			return err
		}
		for _, urlRoute := range urlPushedRoutes.Items {
			urlCLUSTER[urlRoute.Name] = URL{
				Spec: URLSpec{
					Port:    urlRoute.Spec.Port,
					urlKind: envinfo.ROUTE,
				},
			}
		}
	}

	log.Info("\nApplying URL changes")
	urlChange := false

	// find URLs to delete
	for urlName, urlSpec := range urlCLUSTER {
		val, ok := urlLOCAL[urlName]
		if !ok {
			if urlSpec.Spec.urlKind == envinfo.INGRESS && kClient == nil {
				continue
			}
			// delete the url
			err := Delete(client, kClient, urlName, parameters.ApplicationName, urlSpec.Spec.urlKind)
			if err != nil {
				return err
			}
			log.Successf("URL %s successfully deleted", urlName)
			urlChange = true
			continue
		} else {
			if !reflect.DeepEqual(val.Spec, urlSpec.Spec) {
				return errors.Errorf("config mismatch for URL with the same name %s", val.Name)
			}
		}
	}

	// find URLs to create
	for urlName, urlInfo := range urlLOCAL {
		_, ok := urlCLUSTER[urlName]
		if !ok {
			if urlInfo.Spec.urlKind == envinfo.INGRESS && kClient == nil {
				continue
			}

			createParameters := CreateParameters{
				urlName:         urlName,
				portNumber:      urlInfo.Spec.Port,
				secureURL:       urlInfo.Spec.Secure,
				componentName:   parameters.ComponentName,
				applicationName: parameters.ApplicationName,
				host:            urlInfo.Spec.Host,
				secretName:      urlInfo.Spec.tLSSecret,
				urlKind:         urlInfo.Spec.urlKind,
			}
			host, err := Create(client, kClient, createParameters, parameters.IsRouteSupported, parameters.IsExperimentalModeEnabled)
			if err != nil {
				return err
			}
			log.Successf("URL %s: %s created", urlName, host)
			urlChange = true
		}
	}

	if !urlChange {
		log.Success("URLs are synced with the cluster, no changes are required.")
	}

	return nil
}
