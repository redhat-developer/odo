package odogenerator

import (
	"github.com/devfile/library/pkg/devfile/generator"
	"k8s.io/api/networking/v1"
)

// TODO: These functions are replicated from devfile library generators and it makes more sense that they reside there
// getNetworkingV1IngressSpec gets an networking v1 ingress spec
func getNetworkingV1IngressSpec(ingressSpecParams generator.IngressSpecParams) *v1.IngressSpec {
	path := "/"
	pathType := v1.PathTypeImplementationSpecific
	if ingressSpecParams.Path != "" {
		path = ingressSpecParams.Path
	}
	ingressSpec := &v1.IngressSpec{
		Rules: []v1.IngressRule{
			{
				Host: ingressSpecParams.IngressDomain,
				IngressRuleValue: v1.IngressRuleValue{
					HTTP: &v1.HTTPIngressRuleValue{
						Paths: []v1.HTTPIngressPath{
							{
								Path: path,
								Backend: v1.IngressBackend{
									Service: &v1.IngressServiceBackend{
										Name: ingressSpecParams.ServiceName,
										Port: v1.ServiceBackendPort{
											//Looks like we can either set name or number. going with number as it is more important
											Number: ingressSpecParams.PortNumber.IntVal,
										},
									},
									Resource: nil,
								},
								//this field is compulsory
								PathType: &pathType,
							},
						},
					},
				},
			},
		},
	}
	secretNameLength := len(ingressSpecParams.TLSSecretName)
	if secretNameLength != 0 {
		ingressSpec.TLS = []v1.IngressTLS{
			{
				Hosts: []string{
					ingressSpecParams.IngressDomain,
				},
				SecretName: ingressSpecParams.TLSSecretName,
			},
		}
	}

	return ingressSpec
}

func GetNetworkingV1Ingress(ingressParams generator.IngressParams) *v1.Ingress {
	var ip *v1.Ingress
	ingressSpec := getNetworkingV1IngressSpec(ingressParams.IngressSpecParams)
	ip = &v1.Ingress{
		TypeMeta:   ingressParams.TypeMeta,
		ObjectMeta: ingressParams.ObjectMeta,
		Spec:       *ingressSpec,
	}
	return ip
}
