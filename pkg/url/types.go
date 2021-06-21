package url

import (
	"fmt"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/unions"
	urlLabels "github.com/openshift/odo/pkg/url/labels"
	"k8s.io/api/extensions/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// URL is
type URL struct {
	v1.TypeMeta   `json:",inline"`
	v1.ObjectMeta `json:"metadata,omitempty"`
	Spec          URLSpec   `json:"spec,omitempty"`
	Status        URLStatus `json:"status,omitempty"`
}

// URLSpec is
type URLSpec struct {
	Host         string                      `json:"host,omitempty"`
	Protocol     string                      `json:"protocol,omitempty"`
	Port         int                         `json:"port,omitempty"`
	Secure       bool                        `json:"secure"`
	Kind         localConfigProvider.URLKind `json:"kind,omitempty"`
	TLSSecret    string                      `json:"tlssecret,omitempty"`
	ExternalPort int                         `json:"externalport,omitempty"`
	Path         string                      `json:"path,omitempty"`
}

// URLList is a list of applications
type URLList struct {
	v1.TypeMeta `json:",inline"`
	v1.ListMeta `json:"metadata,omitempty"`
	Items       []URL `json:"items"`
}

// URLStatus is Status of url
type URLStatus struct {
	// "Pushed" or "Not Pushed" or "Locally Delted"
	State StateType `json:"state"`
}

type StateType string

const (
	// StateTypePushed means that URL is present both locally and on cluster/container
	StateTypePushed = "Pushed"
	// StateTypeNotPushed means that URL is only in local config, but not on the cluster/container
	StateTypeNotPushed = "Not Pushed"
	// StateTypeLocallyDeleted means that URL was deleted from the local config, but it is still present on the cluster/container
	StateTypeLocallyDeleted = "Locally Deleted"
)

func NewURLFromKubernetesIngress(ki *unions.KubernetesIngress) URL {
	if ki.IsGenerated() {
		return URL{}
	}
	u := URL{
		TypeMeta: metav1.TypeMeta{Kind: "url", APIVersion: apiVersion},
	}
	if ki.NetworkingV1Ingress != nil {
		u.ObjectMeta = metav1.ObjectMeta{Name: ki.NetworkingV1Ingress.Labels[urlLabels.URLLabel]}
		u.Spec = URLSpec{
			Host:   ki.NetworkingV1Ingress.Spec.Rules[0].Host,
			Port:   int(ki.NetworkingV1Ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number),
			Secure: ki.NetworkingV1Ingress.Spec.TLS != nil,
			Path:   ki.NetworkingV1Ingress.Spec.Rules[0].HTTP.Paths[0].Path,
			Kind:   localConfigProvider.INGRESS,
		}
		if len(ki.NetworkingV1Ingress.Spec.TLS) > 0 {
			u.Spec.TLSSecret = ki.NetworkingV1Ingress.Spec.TLS[0].SecretName
		}
		if u.Spec.Secure {
			u.Spec.Protocol = "https"
		} else {
			u.Spec.Protocol = "http"
		}
	} else if ki.ExtensionV1Beta1Ingress != nil {
		u.ObjectMeta = metav1.ObjectMeta{Name: ki.ExtensionV1Beta1Ingress.Labels[urlLabels.URLLabel]}
		u.Spec = URLSpec{
			Host:   ki.ExtensionV1Beta1Ingress.Spec.Rules[0].Host,
			Port:   int(ki.ExtensionV1Beta1Ingress.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort.IntVal),
			Secure: ki.ExtensionV1Beta1Ingress.Spec.TLS != nil,
			Path:   ki.ExtensionV1Beta1Ingress.Spec.Rules[0].HTTP.Paths[0].Path,
			Kind:   localConfigProvider.INGRESS,
		}
		if len(ki.ExtensionV1Beta1Ingress.Spec.TLS) > 0 {
			u.Spec.TLSSecret = ki.ExtensionV1Beta1Ingress.Spec.TLS[0].SecretName
		}
		if u.Spec.Secure {
			u.Spec.Protocol = "https"
		} else {
			u.Spec.Protocol = "http"
		}
	}
	return u
}

// Get returns URL definition for given URL name
func (urls URLList) Get(urlName string) URL {
	for _, url := range urls.Items {
		if url.Name == urlName {
			return url
		}
	}
	return URL{}

}

func (u URL) GetKubernetesIngress(serviceName string) *unions.KubernetesIngress {
	port := intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: int32(u.Spec.Port),
	}
	ki := &unions.KubernetesIngress{
		NetworkingV1Ingress: &networkingv1.Ingress{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Ingress",
				APIVersion: "networking.k8s.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: u.Name,
			},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{
						Host: u.Spec.Host,
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										Path: u.Spec.Path,

										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: serviceName,
												Port: networkingv1.ServiceBackendPort{
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
				Name: u.Name,
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: u.Spec.Host,
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: u.Spec.Path,
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
	}

	if len(u.Spec.TLSSecret) > 0 {
		ki.NetworkingV1Ingress.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts: []string{
					u.Spec.Host,
				},
				SecretName: u.Spec.TLSSecret,
			},
		}

		ki.ExtensionV1Beta1Ingress.Spec.TLS = []v1beta1.IngressTLS{
			{
				Hosts: []string{
					u.Spec.Host,
				},
				SecretName: u.Spec.TLSSecret,
			},
		}
	}
	return ki
}
