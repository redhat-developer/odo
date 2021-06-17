package unions

import (
	"fmt"
	"github.com/devfile/library/pkg/devfile/generator"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/odogenerator"
	urlLabels "github.com/openshift/odo/pkg/url/labels"
	"github.com/openshift/odo/pkg/urltype"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const apiVersion = "odo.dev/v1alpha1"

type KubernetesIngress struct {
	NetworkingV1Ingress     *v1.Ingress
	ExtensionV1Beta1Ingress *v1beta1.Ingress
	isGenerated             bool
}

//NewEmptyKubernetesIngress returns a new empty KubernetesIngress to be populated by caller
func NewEmptyKubernetesIngress() *KubernetesIngress {
	return &KubernetesIngress{
		NetworkingV1Ingress:     nil,
		ExtensionV1Beta1Ingress: nil,
		isGenerated:             false,
	}
}

//NewKubernetesIngressFromParams generates a new KubernetesIngress from the ingress params
func NewKubernetesIngressFromParams(ingressParams generator.IngressParams) *KubernetesIngress {
	ki := &KubernetesIngress{
		NetworkingV1Ingress:     odogenerator.GetNetworkingV1Ingress(ingressParams),
		ExtensionV1Beta1Ingress: generator.GetIngress(ingressParams),
		isGenerated:             true,
	}
	return ki
}

//NewKubernetesIngressFromURL returns a new KubernetesIngress from specified URL
func NewKubernetesIngressFromURL(ingressURL urltype.URL, serviceName string) *KubernetesIngress {

	port := intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: int32(ingressURL.Spec.Port),
	}
	ki := &KubernetesIngress{
		NetworkingV1Ingress: &v1.Ingress{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Ingress",
				APIVersion: "networking.k8s.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: ingressURL.Name,
			},
			Spec: v1.IngressSpec{
				Rules: []v1.IngressRule{
					{
						Host: ingressURL.Spec.Host,
						IngressRuleValue: v1.IngressRuleValue{
							HTTP: &v1.HTTPIngressRuleValue{
								Paths: []v1.HTTPIngressPath{
									{
										Path: ingressURL.Spec.Path,

										Backend: v1.IngressBackend{
											Service: &v1.IngressServiceBackend{
												Name: serviceName,
												Port: v1.ServiceBackendPort{
													Name:   fmt.Sprintf("%s%d", serviceName, port.IntVal),
													Number: port.IntVal,
												},
											},
											Resource: nil,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		ExtensionV1Beta1Ingress: &v1beta1.Ingress{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Ingress",
				APIVersion: "extensions/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: ingressURL.Name,
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: ingressURL.Spec.Host,
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: ingressURL.Spec.Path,
										Backend: v1beta1.IngressBackend{
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
		},
		isGenerated: true,
	}

	if len(ingressURL.Spec.TLSSecret) > 0 {
		ki.NetworkingV1Ingress.Spec.TLS = []v1.IngressTLS{
			{
				Hosts: []string{
					ingressURL.Spec.Host,
				},
				SecretName: ingressURL.Spec.TLSSecret,
			},
		}

		ki.ExtensionV1Beta1Ingress.Spec.TLS = []v1beta1.IngressTLS{
			{
				Hosts: []string{
					ingressURL.Spec.Host,
				},
				SecretName: ingressURL.Spec.TLSSecret,
			},
		}
	}
	return ki
}

//IsGenerated returns true if ths KubernetesIngress was generated using generators
func (ki *KubernetesIngress) IsGenerated() bool {
	return ki.isGenerated
}

func (ki *KubernetesIngress) GetURL() urltype.URL {
	if ki.IsGenerated() {
		return urltype.URL{}
	}
	u := urltype.URL{
		TypeMeta: metav1.TypeMeta{Kind: "url", APIVersion: apiVersion},
	}
	if ki.NetworkingV1Ingress != nil {
		u.ObjectMeta = metav1.ObjectMeta{Name: ki.NetworkingV1Ingress.Labels[urlLabels.URLLabel]}
		u.Spec = urltype.URLSpec{
			Host:   ki.NetworkingV1Ingress.Spec.Rules[0].Host,
			Port:   int(ki.NetworkingV1Ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number),
			Secure: ki.NetworkingV1Ingress.Spec.TLS != nil,
			Path:   ki.NetworkingV1Ingress.Spec.Rules[0].HTTP.Paths[0].Path,
			Kind:   localConfigProvider.INGRESS,
		}
		if u.Spec.Secure {
			u.Spec.Protocol = "https"
		} else {
			u.Spec.Protocol = "http"
		}
	} else if ki.ExtensionV1Beta1Ingress != nil {
		u.ObjectMeta = metav1.ObjectMeta{Name: ki.ExtensionV1Beta1Ingress.Labels[urlLabels.URLLabel]}
		u.Spec = urltype.URLSpec{
			Host:   ki.ExtensionV1Beta1Ingress.Spec.Rules[0].Host,
			Port:   int(ki.ExtensionV1Beta1Ingress.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort.IntVal),
			Secure: ki.ExtensionV1Beta1Ingress.Spec.TLS != nil,
			Path:   ki.ExtensionV1Beta1Ingress.Spec.Rules[0].HTTP.Paths[0].Path,
			Kind:   localConfigProvider.INGRESS,
		}
		if u.Spec.Secure {
			u.Spec.Protocol = "https"
		} else {
			u.Spec.Protocol = "http"
		}
	}
	return u
}
