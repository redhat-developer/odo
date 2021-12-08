package unions

import (
	"fmt"

	v1alpha2 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/generator"
	"github.com/redhat-developer/odo/pkg/odogenerator"
	"k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/api/networking/v1"
)

type KubernetesIngress struct {
	NetworkingV1Ingress     *v1.Ingress
	ExtensionV1Beta1Ingress *v1beta1.Ingress
	isGenerated             bool
}

//NewNonGeneratedKubernetesIngress returns a new empty KubernetesIngress to be populated by caller. It is not genrated
func NewNonGeneratedKubernetesIngress() *KubernetesIngress {
	return &KubernetesIngress{
		NetworkingV1Ingress:     nil,
		ExtensionV1Beta1Ingress: nil,
		isGenerated:             false,
	}
}

//NewGeneratedKubernetesIngress returns a generated KubernetesIngress to be populated by caller
func NewGeneratedKubernetesIngress() *KubernetesIngress {
	return &KubernetesIngress{
		NetworkingV1Ingress:     nil,
		ExtensionV1Beta1Ingress: nil,
		isGenerated:             true,
	}
}

//NewKubernetesIngressFromParams generates a new KubernetesIngress from the ingress params
func NewKubernetesIngressFromParams(ingressParams generator.IngressParams) *KubernetesIngress {
	ki := NewGeneratedKubernetesIngress()
	ki.NetworkingV1Ingress = odogenerator.GetNetworkingV1Ingress(ingressParams)
	ki.ExtensionV1Beta1Ingress = generator.GetIngress(v1alpha2.Endpoint{}, ingressParams)
	return ki
}

//IsGenerated returns true if ths KubernetesIngress was generated using generators
func (ki *KubernetesIngress) IsGenerated() bool {
	return ki.isGenerated
}

//GetName returns the name of underlying networking v1 or extensions v1 ingress
func (ki *KubernetesIngress) GetName() string {
	if ki.NetworkingV1Ingress != nil {
		return ki.NetworkingV1Ingress.GetName()
	} else if ki.ExtensionV1Beta1Ingress != nil {
		return ki.ExtensionV1Beta1Ingress.GetName()
	}
	return ""
}

//GetProtocol returns `https` if tls is configured on either networking v1 or extensions v1 ingress, else `http`
func (ki *KubernetesIngress) GetProtocol() string {
	if (ki.NetworkingV1Ingress != nil && len(ki.NetworkingV1Ingress.Spec.TLS) > 0) || (ki.ExtensionV1Beta1Ingress != nil && len(ki.ExtensionV1Beta1Ingress.Spec.TLS) > 0) {
		return "https"
	}
	return "http"
}

//GetHost returns the host of underlying networking v1 or extensions v1 ingress
func (ki *KubernetesIngress) GetHost() string {
	if ki.NetworkingV1Ingress != nil {
		return ki.NetworkingV1Ingress.Spec.Rules[0].Host
	} else if ki.ExtensionV1Beta1Ingress != nil {
		return ki.ExtensionV1Beta1Ingress.Spec.Rules[0].Host
	}
	return ""
}

//GetURLString returns the fully formed url of the form `GetProtocol()://GetHost()`
func (ki *KubernetesIngress) GetURLString() string {
	return fmt.Sprintf("%v://%v", ki.GetProtocol(), ki.GetHost())
}

type KubernetesIngressList struct {
	Items []*KubernetesIngress
}

func NewEmptyKubernetesIngressList() *KubernetesIngressList {
	return &KubernetesIngressList{}
}

//GetNetworkingV1IngressList returns a v1.IngressList populated by networking v1 ingresses
//if skipIfExtensionV1Set it true, then if both networking v1 and extension v1 are set for
//specific KubernetesIngress, then it will be skipped form the returned list
func (kil *KubernetesIngressList) GetNetworkingV1IngressList(skipIfExtensionV1Set bool) *v1.IngressList {
	il := v1.IngressList{}
	for _, it := range kil.Items {
		if !skipIfExtensionV1Set || (skipIfExtensionV1Set && it.ExtensionV1Beta1Ingress == nil) {
			if it.NetworkingV1Ingress != nil {
				il.Items = append(il.Items, *it.NetworkingV1Ingress)
			}
		}
	}
	return &il
}

//GetExtensionV1Beta1IngresList returns a v1beta1.IngressList populated by extensions v1 beta1 ingresses
//if skipIfNetworkingV1Set it true, then if both networking v1 and extension v1 are set for
//specific KubernetesIngress, then it will be skipped form the returned list
func (kil *KubernetesIngressList) GetExtensionV1Beta1IngresList(skipIfNetworkingV1Set bool) *v1beta1.IngressList {
	il := v1beta1.IngressList{}
	for _, it := range kil.Items {
		if !skipIfNetworkingV1Set || (skipIfNetworkingV1Set && it.NetworkingV1Ingress == nil) {
			if it.ExtensionV1Beta1Ingress != nil {
				il.Items = append(il.Items, *it.ExtensionV1Beta1Ingress)
			}
		}
	}
	return &il
}
