package url

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"

	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/log"

	types "github.com/docker/docker/api/types"
	routev1 "github.com/openshift/api/route/v1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/config"
	dockercomponent "github.com/openshift/odo/pkg/devfile/adapters/docker/component"
	dockerutils "github.com/openshift/odo/pkg/devfile/adapters/docker/utils"
	parsercommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/lclient"
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
				clusterURL := getMachineReadableFormat(*route)
				clusterURL.Status.State = StateTypePushed
				return clusterURL, nil
			} else {
				localURL.Status.State = StateTypeNotPushed
				return localURL, nil
			}
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

// GetIngressOrRoute returns ingress/route spec for given URL name
func GetIngressOrRoute(client *occlient.Client, kClient *kclient.Client, envSpecificInfo *envinfo.EnvSpecificInfo, urlName string, componentName string, routeSupported bool) (URL, error) {
	remoteExist := true
	var ingress *iextensionsv1.Ingress
	var route *routev1.Route
	var getRouteErr error
	// Check whether remote already created the ingress
	ingress, getIngressErr := kClient.GetIngress(urlName)
	if kerrors.IsNotFound(getIngressErr) && routeSupported {
		// Check whether remote already created the route
		route, getRouteErr = client.GetRoute(urlName)
	}
	if kerrors.IsNotFound(getIngressErr) && (!routeSupported || kerrors.IsNotFound(getRouteErr)) {
		remoteExist = false
	} else if (getIngressErr != nil && !kerrors.IsNotFound(getIngressErr)) || (getRouteErr != nil && !kerrors.IsNotFound(getRouteErr)) {
		if getIngressErr != nil {
			return URL{}, errors.Wrap(getIngressErr, "unable to get ingress")
		}
		return URL{}, errors.Wrap(getRouteErr, "unable to get route")
	}

	envinfoURLs := envSpecificInfo.GetURL()
	for _, url := range envinfoURLs {
		// ignore Docker URLs
		if url.Kind == envinfo.DOCKER {
			continue
		}
		if !routeSupported && url.Kind == envinfo.ROUTE {
			continue
		}
		localURL := ConvertEnvinfoURL(url, componentName)
		// search local URL, if it exist in local, update state with remote status
		if localURL.Name == urlName {
			if remoteExist {
				if ingress != nil && ingress.Spec.Rules != nil {
					// Remote exist, but not in local, so it's deleted status
					clusterURL := getMachineReadableFormatIngress(*ingress)
					clusterURL.Status.State = StateTypePushed
					return clusterURL, nil
				} else if route != nil {
					clusterURL := getMachineReadableFormat(*route)
					clusterURL.Status.State = StateTypePushed
					return clusterURL, nil
				}
			} else {
				localURL.Status.State = StateTypeNotPushed
			}
			return localURL, nil
		}
	}

	if remoteExist {
		if ingress != nil && ingress.Spec.Rules != nil {
			// Remote exist, but not in local, so it's deleted status
			clusterURL := getMachineReadableFormatIngress(*ingress)
			clusterURL.Status.State = StateTypeLocallyDeleted
			return clusterURL, nil
		} else if route != nil {
			clusterURL := getMachineReadableFormat(*route)
			clusterURL.Status.State = StateTypeLocallyDeleted
			return clusterURL, nil
		}
	}

	// can't find the URL in local and remote
	return URL{}, errors.New(fmt.Sprintf("the url %v does not exist", urlName))
}

// GetContainer returns Docker URL definition for given URL name
func GetContainerURL(client *lclient.Client, envSpecificInfo *envinfo.EnvSpecificInfo, urlName string, componentName string) (URL, error) {
	localURLs := envSpecificInfo.GetURL()
	containers, err := dockerutils.GetComponentContainers(*client, componentName)
	if err != nil {
		return URL{}, errors.Wrap(err, "unable to get component containers")
	}
	var remoteExist = false
	var dockerURL URL
	// iterating through each container's HostConfig, generate and cache the dockerURL if found a match urlName
	for _, c := range containers {
		containerJSON, err := client.Client.ContainerInspect(client.Context, c.ID)
		if err != nil {
			return URL{}, err
		}
		for internalPort, portbinding := range containerJSON.HostConfig.PortBindings {
			remoteURLName := containerJSON.Config.Labels[internalPort.Port()]
			if remoteURLName != urlName {
				continue
			}
			// found urlName in Docker container's config
			remoteExist = true
			externalport, err := strconv.Atoi(portbinding[0].HostPort)
			if err != nil {
				return URL{}, err
			}
			dockerURL = getMachineReadableFormatDocker(internalPort.Int(), externalport, portbinding[0].HostIP, remoteURLName)
		}
	}

	// iterating through URLs in local env.yaml
	for _, localurl := range localURLs {
		if localurl.Kind != envinfo.DOCKER || localurl.Name != urlName {
			continue
		}
		localURL := getMachineReadableFormatDocker(localurl.Port, localurl.ExposedPort, dockercomponent.LocalhostIP, localurl.Name)
		// found urlName in local env file
		if remoteExist {
			// URL is in both env file and Docker HostConfig
			localURL.Status.State = StateTypePushed
			return localURL, nil
		} else {
			// URL only exists in local env file
			localURL.Status.State = StateTypeNotPushed
			return localURL, nil
		}
	}
	// URL only exists in pushed Docker container
	if remoteExist {
		dockerURL.Status.State = StateTypeLocallyDeleted
		return dockerURL, nil
	}
	// can't find the URL in local env.yaml or Docker containers
	return URL{}, errors.New(fmt.Sprintf("the url %v does not exist", urlName))
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
	path            string
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

		ingressParam := kclient.IngressParameter{ServiceName: serviceName, IngressDomain: ingressDomain, PortNumber: intstr.FromInt(parameters.portNumber), TLSSecretName: parameters.secretName, Path: parameters.path}
		ingressSpec := kclient.GenerateIngressSpec(ingressParam)
		objectMeta := kclient.CreateObjectMeta(parameters.componentName, kClient.Namespace, labels, nil)
		objectMeta.Name = parameters.urlName
		objectMeta.OwnerReferences = append(objectMeta.OwnerReferences, ownerReference)
		// Pass in the namespace name, link to the service (componentName) and labels to create a ingress
		ingress, err := kClient.CreateIngress(objectMeta, *ingressSpec)
		if err != nil {
			return "", errors.Wrap(err, "unable to create ingress")
		}
		return GetURLString(GetProtocol(routev1.Route{}, *ingress), "", ingressDomain, isExperimental), nil
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
		route, err := client.CreateRoute(parameters.urlName, serviceName, intstr.FromInt(parameters.portNumber), labels, parameters.secureURL, parameters.path, ownerReference)
		if err != nil {
			return "", errors.Wrap(err, "unable to create route")
		}
		return GetURLString(GetProtocol(*route, iextensionsv1.Ingress{}), route.Spec.Host, "", isExperimental), nil
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
		var found = false
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

// ListIngressAndRoute returns all Ingress and Route for given component.
func ListIngressAndRoute(oclient *occlient.Client, client *kclient.Client, envSpecificInfo *envinfo.EnvSpecificInfo, componentName string, routeSupported bool) (URLList, error) {
	labelSelector := fmt.Sprintf("%v=%v", componentlabels.ComponentLabel, componentName)
	klog.V(4).Infof("Listing ingresses with label selector: %v", labelSelector)
	ingresses, err := client.ListIngresses(labelSelector)
	if err != nil {
		return URLList{}, errors.Wrap(err, "unable to list ingress")
	}
	routes := []routev1.Route{}
	if routeSupported {
		routes, err = oclient.ListRoutes(labelSelector)
		if err != nil {
			return URLList{}, errors.Wrap(err, "unable to list routes")
		}
	}
	localEnvinfoURLs := envSpecificInfo.GetURL()

	var urls []URL

	clusterURLMap := make(map[string]URL)
	localMap := make(map[string]URL)
	for _, i := range ingresses {
		clusterURL := getMachineReadableFormatIngress(i)
		clusterURLMap[clusterURL.Name] = clusterURL
	}
	for _, r := range routes {
		if r.OwnerReferences != nil && r.OwnerReferences[0].Kind == "Ingress" {
			continue
		}
		clusterURL := getMachineReadableFormat(r)
		clusterURLMap[clusterURL.Name] = clusterURL
	}
	for _, envinfoURL := range localEnvinfoURLs {
		// only checks for Ingress and Route URLs
		if envinfoURL.Kind == envinfo.DOCKER {
			continue
		}
		if !routeSupported && envinfoURL.Kind == envinfo.ROUTE {
			continue
		}
		localURL := ConvertEnvinfoURL(envinfoURL, componentName)
		localMap[localURL.Name] = localURL
	}

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

	for localName, localURL := range localMap {
		_, remoteURLFound := clusterURLMap[localName]
		if !remoteURLFound {
			// URL is in the local env file but not pushed to cluster
			localURL.Status.State = StateTypeNotPushed
			urls = append(urls, localURL)
		}
	}
	urlList := getMachineReadableFormatForList(urls)
	return urlList, nil
}

// ListDockerURL returns all Docker URLs for given component.
func ListDockerURL(client *lclient.Client, componentName string, envSpecificInfo *envinfo.EnvSpecificInfo) (URLList, error) {
	containers, err := dockerutils.GetComponentContainers(*client, componentName)
	if err != nil {
		return URLList{}, errors.Wrap(err, "unable to list component container")
	}

	localURLs := envSpecificInfo.GetURL()

	var urls []URL

	containerJSONMap := make(map[string]types.ContainerJSON)

	// iterating through each container's HostConfig
	// find out if there is a match config in local env.yaml
	// if found a match, then the URL is <Pushed>
	// if the config does not exist in local env.yaml file, then the URL is <Locally Deleted>
	for _, c := range containers {
		var found = false
		containerJSON, err := client.Client.ContainerInspect(client.Context, c.ID)
		if err != nil {
			return URLList{}, err
		}
		containerJSONMap[c.ID] = containerJSON
		for internalPort, portbinding := range containerJSON.HostConfig.PortBindings {
			externalport, err := strconv.Atoi(portbinding[0].HostPort)
			if err != nil {
				return URLList{}, err
			}
			dockerURL := getMachineReadableFormatDocker(internalPort.Int(), externalport, portbinding[0].HostIP, containerJSON.Config.Labels[internalPort.Port()])
			for _, localurl := range localURLs {
				// only checks for Docker URLs
				if localurl.Kind != envinfo.DOCKER {
					continue
				}
				if localurl.Port == dockerURL.Spec.Port && localurl.ExposedPort == dockerURL.Spec.ExternalPort {
					// URL is in both env file and Docker HostConfig
					dockerURL.Status.State = StateTypePushed
					urls = append(urls, dockerURL)
					found = true
					break
				}
			}
			if !found {
				// URL is in Docker HostConfig but not in env file
				dockerURL.Status.State = StateTypeLocallyDeleted
				urls = append(urls, dockerURL)
			}
		}
	}

	// iterating through URLs in local env.yaml
	// find out if there is a match config in Docker container
	// if the config does not exist in Docker container, then the URL is <Not Pushed>
	for _, localurl := range localURLs {
		// only checks for Docker URLs
		if localurl.Kind != envinfo.DOCKER {
			continue
		}
		var found = false
		localURL := getMachineReadableFormatDocker(localurl.Port, localurl.ExposedPort, dockercomponent.LocalhostIP, localurl.Name)
		for _, c := range containers {
			containerJSON := containerJSONMap[c.ID]
			for internalPort, portbinding := range containerJSON.HostConfig.PortBindings {
				externalport, err := strconv.Atoi(portbinding[0].HostPort)
				if err != nil {
					return URLList{}, err
				}
				if localURL.Spec.Port == internalPort.Int() && localURL.Spec.ExternalPort == externalport {
					found = true
					break
				}
			}
		}
		if !found {
			// URL is in the env file but not pushed to Docker container
			localURL.Status.State = StateTypeNotPushed
			urls = append(urls, localURL)
		}
	}
	urlList := getMachineReadableFormatForList(urls)
	return urlList, nil
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
func ConvertConfigURL(configURL config.ConfigURL) URL {
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
			Kind:   envinfo.ROUTE,
			Path:   "/",
		},
	}
}

// ConvertEnvinfoURL converts EnvinfoURL to URL
func ConvertEnvinfoURL(envinfoURL envinfo.EnvInfoURL, serviceName string) URL {
	hostString := fmt.Sprintf("%s.%s", envinfoURL.Name, envinfoURL.Host)
	url := URL{
		TypeMeta: metav1.TypeMeta{
			Kind:       "url",
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: envinfoURL.Name,
		},
		Spec: URLSpec{
			Port:   envinfoURL.Port,
			Secure: envinfoURL.Secure,
			Kind:   envinfoURL.Kind,
		},
	}
	if envinfoURL.Kind == envinfo.INGRESS {
		url.Spec.Host = hostString
		if envinfoURL.Secure && len(envinfoURL.TLSSecret) > 0 {
			url.Spec.TLSSecret = envinfoURL.TLSSecret
		} else if envinfoURL.Secure && envinfoURL.Kind == envinfo.INGRESS {
			url.Spec.TLSSecret = fmt.Sprintf("%s-tlssecret", serviceName)
		}
	}
	return url
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
		Spec:       URLSpec{Host: r.Spec.Host, Port: r.Spec.Port.TargetPort.IntValue(), Protocol: GetProtocol(r, iextensionsv1.Ingress{}), Secure: r.Spec.TLS != nil, Path: r.Spec.Path, Kind: envinfo.ROUTE},
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
		Spec:       URLSpec{Host: i.Spec.Rules[0].Host, Port: int(i.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort.IntVal), Secure: i.Spec.TLS != nil, Path: i.Spec.Rules[0].HTTP.Paths[0].Path, Kind: envinfo.INGRESS},
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

func getMachineReadableFormatDocker(internalPort int, externalPort int, hostIP string, urlName string) URL {
	return URL{
		TypeMeta:   metav1.TypeMeta{Kind: "url", APIVersion: apiVersion},
		ObjectMeta: metav1.ObjectMeta{Name: urlName},
		Spec:       URLSpec{Host: hostIP, Port: internalPort, ExternalPort: externalPort},
	}
}

type PushParameters struct {
	ComponentName             string
	ApplicationName           string
	ConfigURLs                []config.ConfigURL
	EnvURLS                   []envinfo.EnvInfoURL
	IsRouteSupported          bool
	IsExperimentalModeEnabled bool
	EndpointMap               map[int32]parsercommon.Endpoint
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
				if parameters.EndpointMap == nil {
					klog.V(4).Infof("No Endpoint entry defined in devfile.")
					return nil
				}
				endpoint, exist := parameters.EndpointMap[int32(url.Port)]
				if !exist || endpoint.Exposure == "none" || endpoint.Exposure == "internal" {
					return fmt.Errorf("port %v defined in env.yaml file for URL %v is not exposed in devfile Endpoint entry", url.Port, url.Name)
				}
				secure := false
				if endpoint.Secure || endpoint.Protocol == "https" || endpoint.Protocol == "wss" {
					secure = true
				}
				path := "/"
				if endpoint.Path != "" {
					path = endpoint.Path
				}
				urlLOCAL[url.Name] = URL{
					Spec: URLSpec{
						Host:      url.Host,
						Port:      url.Port,
						Secure:    secure,
						TLSSecret: url.TLSSecret,
						Kind:      url.Kind,
						Path:      path,
					},
				}
			}
		}
	} else {
		urls := parameters.ConfigURLs
		for _, url := range urls {
			urlLOCAL[url.Name] = URL{
				Spec: URLSpec{
					Port:   url.Port,
					Secure: url.Secure,
					Kind:   envinfo.ROUTE,
					Path:   "/",
				},
			}
		}
	}

	// iterate through endpoints defined in devfile
	// add the url defination into urlLOCAL if it's not defined in env.yaml
	if parameters.IsExperimentalModeEnabled && parameters.EndpointMap != nil {
		for port, endpoint := range parameters.EndpointMap {
			// should not create URL if Exposure is none or internal
			if endpoint.Exposure == "none" || endpoint.Exposure == "internal" {
				continue
			}
			exist := false
			for _, envURL := range urlLOCAL {
				if envURL.Spec.Port == int(port) {
					exist = true
					break
				}
			}
			if !exist {
				// create route against Openshift
				if parameters.IsRouteSupported {
					secure := false
					if endpoint.Secure || endpoint.Protocol == "https" || endpoint.Protocol == "wss" {
						secure = true
					}
					path := "/"
					if endpoint.Path != "" {
						path = endpoint.Path
					}
					name := strings.TrimSpace(util.GetDNS1123Name(strings.ToLower(endpoint.Name)))
					urlLOCAL[name] = URL{
						Spec: URLSpec{
							Port:   int(port),
							Secure: secure,
							Kind:   envinfo.ROUTE,
							Path:   path,
						},
					}
				} else {
					// display warning since Host info is missing
					log.Warningf("Unable to create ingress, missing host information for Endpoint %v, please check instructions on URL creation (refer `odo url create --help`)\n", endpoint.Name)
				}
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
					Host:   url.Spec.Host,
					Port:   url.Spec.Port,
					Kind:   envinfo.INGRESS,
					Secure: url.Spec.Secure,
					Path:   url.Spec.Path,
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
					Port:   urlRoute.Spec.Port,
					Kind:   envinfo.ROUTE,
					Secure: urlRoute.Spec.Secure,
					Path:   urlRoute.Spec.Path,
				},
			}
		}
	}

	log.Info("\nApplying URL changes")
	urlChange := false

	// find URLs to delete
	for urlName, urlSpec := range urlCLUSTER {
		val, ok := urlLOCAL[urlName]

		configMismatch := false
		if ok {
			// since the host stored in an ingress
			// is the combination of name and host of the url
			if val.Spec.Kind == envinfo.INGRESS {
				val.Spec.Host = fmt.Sprintf("%v.%v", urlName, val.Spec.Host)
			}
			if !reflect.DeepEqual(val.Spec, urlSpec.Spec) {
				configMismatch = true
				klog.V(4).Infof("config and cluster mismatch for url %s", urlName)
			}
		}

		if !ok || configMismatch {
			if urlSpec.Spec.Kind == envinfo.INGRESS && kClient == nil {
				continue
			}
			// delete the url
			err := Delete(client, kClient, urlName, parameters.ApplicationName, urlSpec.Spec.Kind)
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
			if urlInfo.Spec.Kind == envinfo.INGRESS && kClient == nil {
				continue
			}

			createParameters := CreateParameters{
				urlName:         urlName,
				portNumber:      urlInfo.Spec.Port,
				secureURL:       urlInfo.Spec.Secure,
				componentName:   parameters.ComponentName,
				applicationName: parameters.ApplicationName,
				host:            urlInfo.Spec.Host,
				secretName:      urlInfo.Spec.TLSSecret,
				urlKind:         urlInfo.Spec.Kind,
				path:            urlInfo.Spec.Path,
			}
			host, err := Create(client, kClient, createParameters, parameters.IsRouteSupported, parameters.IsExperimentalModeEnabled)
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
