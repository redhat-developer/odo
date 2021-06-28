package unions

import (
	"fmt"
	"github.com/devfile/library/pkg/devfile/generator"
	"github.com/openshift/odo/pkg/odogenerator"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/api/networking/v1"
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
	ki.ExtensionV1Beta1Ingress = generator.GetIngress(ingressParams)
	return ki
}

//IsGenerated returns true if ths KubernetesIngress was generated using generators
func (ki *KubernetesIngress) IsGenerated() bool {
	return ki.isGenerated
}

func (ki *KubernetesIngress) GetName() string {
	if ki.NetworkingV1Ingress != nil {
		return ki.NetworkingV1Ingress.GetName()
	} else if ki.ExtensionV1Beta1Ingress != nil {
		return ki.ExtensionV1Beta1Ingress.GetName()
	}
	return ""
}

func (ki *KubernetesIngress) GetProtocol() string {
	if (ki.NetworkingV1Ingress != nil && len(ki.NetworkingV1Ingress.Spec.TLS) > 0) || (ki.ExtensionV1Beta1Ingress != nil && len(ki.ExtensionV1Beta1Ingress.Spec.TLS) > 0) {
		return "https"
	}
	return "http"
}

func (ki *KubernetesIngress) GetHost() string {
	if ki.NetworkingV1Ingress != nil {
		return ki.NetworkingV1Ingress.Spec.Rules[0].Host
	} else if ki.ExtensionV1Beta1Ingress != nil {
		return ki.ExtensionV1Beta1Ingress.Spec.Rules[0].Host
	}
	return ""
}

func (ki *KubernetesIngress) GetURLString() string {
	return fmt.Sprintf("%v://%v", ki.GetProtocol(), ki.GetHost())
}

type KubernetesIngressList struct {
	Items []*KubernetesIngress
}

func NewEmptyKubernetesIngressList() *KubernetesIngressList {
	return &KubernetesIngressList{}
}

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
