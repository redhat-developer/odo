package fake

import (
	"fmt"
	"github.com/openshift/odo/pkg/unions"

	"github.com/devfile/library/pkg/devfile/generator"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/url/labels"
	"github.com/openshift/odo/pkg/version"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GetKubernetesIngressListWithMultiple(componentName, appName string, networkingV1Supported, extensionV1Supported bool) *unions.KubernetesIngressList {
	kil := unions.NewEmptyKubernetesIngressList()
	kil.Items = append(kil.Items, unions.NewKubernetesIngressFromParams(generator.IngressParams{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example-0",
			Labels: map[string]string{
				applabels.ApplicationLabel:     appName,
				componentlabels.ComponentLabel: componentName,
				applabels.ManagedBy:            "odo",
				applabels.ManagerVersion:       version.VERSION,
				labels.URLLabel:                "example-0",
				applabels.App:                  appName,
			},
		},
	}))
	kil.Items = append(kil.Items, unions.NewKubernetesIngressFromParams(generator.IngressParams{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example-1",
			Labels: map[string]string{
				applabels.ApplicationLabel:     "app",
				componentlabels.ComponentLabel: componentName,
				applabels.ManagedBy:            "odo",
				applabels.ManagerVersion:       version.VERSION,
				labels.URLLabel:                "example-1",
				applabels.App:                  "app",
			},
		},
		IngressSpecParams: generator.IngressSpecParams{
			IngressDomain: "example-1.com",
			ServiceName:   "example-1",
			PortNumber:    intstr.FromInt(9090),
		},
	}))
	return kil
}

func GetExtensionV1IngressListWithMultiple(componentName, appName string) *extensionsv1.IngressList {
	return &extensionsv1.IngressList{
		Items: []extensionsv1.Ingress{
			*generator.GetIngress(generator.IngressParams{
				ObjectMeta: metav1.ObjectMeta{
					Name: "example-0",
					Labels: map[string]string{
						applabels.ApplicationLabel:     appName,
						componentlabels.ComponentLabel: componentName,
						applabels.ManagedBy:            "odo",
						applabels.ManagerVersion:       version.VERSION,
						labels.URLLabel:                "example-0",
						applabels.App:                  appName,
					},
				},
				IngressSpecParams: generator.IngressSpecParams{
					IngressDomain: "example-0.com",
					ServiceName:   "example-0",
					PortNumber:    intstr.FromInt(8080),
				},
			}),
			*generator.GetIngress(generator.IngressParams{
				ObjectMeta: metav1.ObjectMeta{
					Name: "example-1",
					Labels: map[string]string{
						applabels.ApplicationLabel:     "app",
						componentlabels.ComponentLabel: componentName,
						applabels.ManagedBy:            "odo",
						applabels.ManagerVersion:       version.VERSION,
						labels.URLLabel:                "example-1",
						applabels.App:                  "app",
					},
				},
				IngressSpecParams: generator.IngressSpecParams{
					IngressDomain: "example-1.com",
					ServiceName:   "example-1",
					PortNumber:    intstr.FromInt(9090),
				},
			}),
		},
	}
}

func GetSingleKubernetesIngress(urlName, componentName, appName string, networkingv1Supported, extensionv1Supported bool) *unions.KubernetesIngress {
	ki := unions.NewKubernetesIngressFromParams(generator.IngressParams{
		ObjectMeta: metav1.ObjectMeta{
			Name: urlName,
			Labels: map[string]string{
				applabels.ApplicationLabel:     appName,
				componentlabels.ComponentLabel: componentName,
				applabels.ManagedBy:            "odo",
				applabels.ManagerVersion:       version.VERSION,
				labels.URLLabel:                urlName,
				applabels.App:                  appName,
			},
		},
		IngressSpecParams: generator.IngressSpecParams{
			IngressDomain: fmt.Sprintf("%s.com", urlName),
			ServiceName:   urlName,
			PortNumber:    intstr.FromInt(8080),
		},
	})
	if !networkingv1Supported {
		ki.NetworkingV1Ingress = nil
	}
	if !extensionv1Supported {
		ki.ExtensionV1Beta1Ingress = nil
	}
	return ki
}

func GetSingleExtensionV1Ingress(urlName, componentName, appName string) *extensionsv1.Ingress {

	return generator.GetIngress(generator.IngressParams{
		ObjectMeta: metav1.ObjectMeta{
			Name: urlName,
			Labels: map[string]string{
				applabels.ApplicationLabel:     appName,
				componentlabels.ComponentLabel: componentName,
				applabels.ManagedBy:            "odo",
				applabels.ManagerVersion:       version.VERSION,
				labels.URLLabel:                urlName,
				applabels.App:                  appName,
			},
		},
		IngressSpecParams: generator.IngressSpecParams{
			IngressDomain: fmt.Sprintf("%s.com", urlName),
			ServiceName:   urlName,
			PortNumber:    intstr.FromInt(8080),
		},
	})
}

func GetSingleSecureKubernetesIngress(urlName, componentName, appName, secretName string, networkingV1Supported, extensionV1Supported bool) *unions.KubernetesIngress {
	if secretName == "" {
		secretName = componentName + "-" + appName + "-tlssecret"
	}
	ki := unions.NewKubernetesIngressFromParams(generator.IngressParams{
		ObjectMeta: metav1.ObjectMeta{
			Name: urlName,
			Labels: map[string]string{
				applabels.ApplicationLabel:     appName,
				componentlabels.ComponentLabel: componentName,
				applabels.ManagedBy:            "odo",
				applabels.ManagerVersion:       version.VERSION,
				labels.URLLabel:                urlName,
				applabels.App:                  appName,
			},
		},
		IngressSpecParams: generator.IngressSpecParams{
			TLSSecretName: secretName,
			IngressDomain: fmt.Sprintf("%s.com", urlName),
			ServiceName:   urlName,
			PortNumber:    intstr.FromInt(8080),
		},
	})
	if !networkingV1Supported {
		ki.NetworkingV1Ingress = nil
	}
	if !extensionV1Supported {
		ki.ExtensionV1Beta1Ingress = nil
	}
	return ki
}

// GetSingleSecureIngress gets a single secure ingress with the given secret name
// if no secret name is provided, the default one is used
func GetSingleSecureIngress(urlName, componentName, appName, secretName string) *extensionsv1.Ingress {

	if secretName == "" {
		secretName = componentName + "-" + appName + "-tlssecret"
	}
	return generator.GetIngress(generator.IngressParams{
		ObjectMeta: metav1.ObjectMeta{
			Name: urlName,
			Labels: map[string]string{
				applabels.ApplicationLabel:     appName,
				componentlabels.ComponentLabel: componentName,
				applabels.ManagedBy:            "odo",
				applabels.ManagerVersion:       version.VERSION,
				labels.URLLabel:                urlName,
				applabels.App:                  appName,
			},
		},
		IngressSpecParams: generator.IngressSpecParams{
			TLSSecretName: secretName,
			IngressDomain: fmt.Sprintf("%s.com", urlName),
			ServiceName:   urlName,
			PortNumber:    intstr.FromInt(8080),
		},
	})
}
