package url

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/odo/pkg/envinfo"

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

	"github.com/golang/glog"
	iextensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Get returns URL defination for given URL name
func (urls URLList) Get(urlName string) URL {
	for _, url := range urls.Items {
		if url.Name == urlName {
			return url
		}
	}
	return URL{}

}

// Delete deletes a URL
func Delete(client *occlient.Client, kClient *kclient.Client, urlName string, applicationName string) error {

	// Namespace the URL name
	namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(urlName, applicationName)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespaced name")
	}
	if experimental.IsExperimentalModeEnabled() {
		return kClient.DeleteIngress(namespacedOpenShiftObject)
	} else {
		return client.DeleteRoute(namespacedOpenShiftObject)
	}
}

// Create creates a URL and returns url string and error if any
// portNumber is the target port number for the route and is -1 in case no port number is specified in which case it is automatically detected for components which expose only one service port)
func Create(client *occlient.Client, kClient *kclient.Client, urlName string, portNumber int, secureURL bool, componentName, applicationName string, clusterHost string, secretName string) (string, error) {
	labels := urlLabels.GetLabels(urlName, componentName, applicationName, true)

	var serviceName string
	if experimental.IsExperimentalModeEnabled() {
		// TODO: Need deployment & service to get the service name
		// serviceName := fmt.Sprintf("%v-%v", componentName, portNumber)
		serviceName := "spring-springtest1-gldv-app"
		// if err != nil {
		// 	return "", errors.Wrapf(err, "unable to create namespaced name")
		// }
		secretName := ""
		ingressDomain := fmt.Sprintf("%v.%v", urlName, clusterHost)
		if secureURL == true {
			// generate SSl certificate
			fmt.Printf("Https is true, creating SSL certificate.")
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				fmt.Printf("unale to generate rsa key ")
				fmt.Println(errors.Cause(err))
				return "", errors.Wrap(err, "unable to generate rsa key")
			}
			template := x509.Certificate{
				SerialNumber: big.NewInt(2),
				Subject: pkix.Name{
					CommonName:   "Udo self-signed certificate",
					Organization: []string{"Udo"},
				},
				NotBefore:             time.Now(),
				NotAfter:              time.Now().Add(time.Hour * 24 * 365),
				KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
				BasicConstraintsValid: true,
				DNSNames:              []string{"*." + clusterHost},
			}

			certificateDerEncoding, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
			if err != nil {
				fmt.Printf("unable to create certificate ")
				fmt.Println(errors.Cause(err))
				return "", errors.Wrap(err, "unable to create certificate")
			}
			out := &bytes.Buffer{}
			pem.Encode(out, &pem.Block{Type: "CERTIFICATE", Bytes: certificateDerEncoding})
			certPemEncode := out.String()
			certPemByteArr := []byte(certPemEncode)

			tlsPrivKeyEncoding := x509.MarshalPKCS1PrivateKey(privateKey)
			pem.Encode(out, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: tlsPrivKeyEncoding})
			keyPemEncode := out.String()
			keyPemByteArr := []byte(keyPemEncode)

			// create tls secret
			secret, err := kClient.CreateTLSSecret(certPemByteArr, keyPemByteArr, componentName, applicationName, portNumber)
			if err != nil {
				fmt.Printf("unable to create tls secret ")
				fmt.Println(errors.Cause(err))
				return "", errors.Wrap(err, "unable to create tls secret: "+secret.Name)
			}
			secretName = secret.Name

		}
		ingressParam := kclient.IngressParamater{Name: urlName, ServiceName: serviceName, IngressDomain: ingressDomain, PortNumber: intstr.FromInt(portNumber), TLSSecretName: secretName}
		// Pass in the namespace name, link to the service (componentName) and labels to create a ingress
		ingress, err := kClient.CreateIngress(ingressParam, labels)
		if err != nil {
			return "", errors.Wrap(err, "unable to create ingress")
		}
		return GetURLString(GetProtocol(routev1.Route{}, *ingress), "", ingressDomain), nil
	} else {
		urlName, err := util.NamespaceOpenShiftObject(urlName, applicationName)
		if err != nil {
			return "", errors.Wrapf(err, "unable to create namespaced name")
		}
		serviceName, err = util.NamespaceOpenShiftObject(componentName, applicationName)
		if err != nil {
			return "", errors.Wrapf(err, "unable to create namespaced name")
		}
		// Pass in the namespace name, link to the service (componentName) and labels to create a route
		route, err := client.CreateRoute(urlName, serviceName, intstr.FromInt(portNumber), labels, secureURL)
		if err != nil {
			return "", errors.Wrap(err, "unable to create route")
		}
		return GetURLString(GetProtocol(*route, iextensionsv1.Ingress{}), route.Spec.Host, ""), nil
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

	glog.V(4).Infof("Listing routes with label selector: %v", labelSelector)
	routes, err := client.ListRoutes(labelSelector)
	if err != nil {
		return URLList{}, errors.Wrap(err, "unable to list route names")
	}

	var urls []URL
	for _, r := range routes {
		a := getMachineReadableFormat(r)
		urls = append(urls, a)
	}

	urlList := getMachineReadableFormatForList(urls)
	return urlList, nil

}

// ListIngress lists the ingress URLs in an application. The results can further be narrowed
// down if a component name is provided, which will only list URLs for the
// given component
func ListPushedIngress(client *kclient.Client, componentName string) (iextensionsv1.IngressList, error) {
	labelSelector := fmt.Sprintf("%v=%v", componentlabels.ComponentLabel, componentName)
	// TODO: select url name
	//labelSelector := fmt.Sprintf("%v=%v", componentlabels.ComponentLabel, componentName)
	glog.V(4).Infof("Listing ingresses with label selector: %v", labelSelector)
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

// List returns all Ingress URLs for given component.
// If componentName is empty string, it lists all url in a given application.
func ListIngress(client *kclient.Client, envSpecificInfo *envinfo.EnvSpecificInfo, componentName string) (iextensionsv1.IngressList, error) {

	labelSelector := fmt.Sprintf("%v=%v", componentlabels.ComponentLabel, componentName)
	// TODO: select url name
	ingresses, err := client.ListIngresses(labelSelector)
	if err != nil {
		return iextensionsv1.IngressList{}, errors.Wrap(err, "unable to list ingress names")
	}

	// envinfoURLs := envSpecificInfo.GetURL()

	var urls []iextensionsv1.Ingress

	for _, i := range ingresses {
		clusterURL := getMachineReadableFormatIngress(i)
		urls = append(urls, clusterURL)
		// var found bool = false
		// for _, envinfoURL := range envinfoURLs {
		// 	if envinfoURL.Name == clusterURL.Name {
		// 		// URL is in both local config and cluster
		// 		urls = append(urls, clusterURL)
		// 		found = true
		// 	}
		// }
	}

	urlList := getMachineReadableFormatForIngressList(urls)
	return urlList, nil
}

func GetProtocol(route routev1.Route, ingress iextensionsv1.Ingress) string {
	if experimental.IsExperimentalModeEnabled() {
		if ingress.Spec.TLS != nil {
			return "https"
		}
		return "http"
	} else {
		if route.Spec.TLS != nil {
			return "https"
		}
		return "http"
	}

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
func GetURLString(protocol, URL string, ingressDomain string) string {
	if experimental.IsExperimentalModeEnabled() {
		return protocol + "://" + ingressDomain
	}
	return protocol + "://" + URL
}

// Exists checks if the url exists in the component or not
// urlName is the name of the url for checking
// componentName is the name of the component to which the url's existence is checked
// applicationName is the name of the application to which the url's existence is checked
func Exists(client *occlient.Client, kClient *kclient.Client, urlName string, componentName string, applicationName string) (bool, error) {
	if experimental.IsExperimentalModeEnabled() {
		urls, err := ListPushedIngress(kClient, componentName)
		if err != nil {
			return false, errors.Wrap(err, "unable to list the urls")
		}

		for _, url := range urls.Items {
			if url.Name == urlName {
				return true, nil
			}
		}
	} else {
		urls, err := ListPushed(client, componentName, applicationName)
		if err != nil {
			return false, errors.Wrap(err, "unable to list the urls")
		}

		for _, url := range urls.Items {
			if url.Name == urlName {
				return true, nil
			}
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

// GetComponentServicePortNumbers returns the port numbers exposed by the service of the component
// componentName is the name of the component
// applicationName is the name of the application
func GetComponentServicePortNumbers(client *kclient.Client, componentName string, applicationName string) ([]int, error) {
	componentLabels := componentlabels.GetLabels(componentName, applicationName, false)
	componentSelector := util.ConvertLabelsToSelector(componentLabels)

	services, err := client.GetServicesFromSelector(componentSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get the service")
	}

	var ports []int

	for _, service := range services {
		for _, port := range service.Spec.Ports {
			ports = append(ports, int(port.Port))
		}
	}

	return ports, nil
}

// getMachineReadableFormat gives machine readable URL definition
func getMachineReadableFormat(r routev1.Route) URL {
	return URL{
		TypeMeta:   metav1.TypeMeta{Kind: "url", APIVersion: "odo.openshift.io/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{Name: r.Labels[urlLabels.URLLabel]},
		Spec:       URLSpec{Host: r.Spec.Host, Port: r.Spec.Port.TargetPort.IntValue(), Protocol: GetProtocol(r, iextensionsv1.Ingress{})},
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
