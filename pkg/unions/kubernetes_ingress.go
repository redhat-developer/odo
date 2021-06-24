package unions

import (
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
