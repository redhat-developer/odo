package url

import (
	"fmt"
	"reflect"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/openshift/odo/pkg/localConfigProvider"
	iextensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

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

// ConvertEnvInfoURL converts EnvInfoURL to URL
func ConvertEnvInfoURL(envInfoURL localConfigProvider.LocalURL, serviceName string) URL {
	hostString := fmt.Sprintf("%s.%s", envInfoURL.Name, envInfoURL.Host)
	// default to route kind if none is provided
	kind := envInfoURL.Kind
	if kind == "" {
		kind = localConfigProvider.ROUTE
	}
	url := URL{
		TypeMeta: metav1.TypeMeta{
			Kind:       "url",
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: envInfoURL.Name,
		},
		Spec: URLSpec{
			Host:      envInfoURL.Host,
			Protocol:  envInfoURL.Protocol,
			Port:      envInfoURL.Port,
			Secure:    envInfoURL.Secure,
			Kind:      kind,
			TLSSecret: envInfoURL.TLSSecret,
			Path:      envInfoURL.Path,
		},
	}
	if kind == localConfigProvider.INGRESS {
		url.Spec.Host = hostString
		if envInfoURL.Secure && len(envInfoURL.TLSSecret) > 0 {
			url.Spec.TLSSecret = envInfoURL.TLSSecret
		} else if envInfoURL.Secure {
			url.Spec.TLSSecret = fmt.Sprintf("%s-tlssecret", serviceName)
		}
	}
	return url
}

// GetURLString returns a string representation of given url
func GetURLString(protocol, URL, ingressDomain string) string {
	if protocol == "" && URL == "" && ingressDomain == "" {
		return ""
	}
	if URL == "" {
		return protocol + "://" + ingressDomain
	}
	return protocol + "://" + URL
}

// getDefaultTLSSecretName returns the name of the default tls secret name
func getDefaultTLSSecretName(componentName, appName string) string {
	return componentName + "-" + appName + "-tlssecret"
}

// ConvertExtensionV1IngressURLToIngress converts IngressURL to Ingress
func ConvertExtensionV1IngressURLToIngress(ingressURL URL, serviceName string) iextensionsv1.Ingress {
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
