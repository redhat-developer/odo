package fake

import (
	"fmt"

	"github.com/devfile/library/pkg/devfile/generator"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/kclient/unions"
	"github.com/redhat-developer/odo/pkg/url/labels"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/redhat-developer/odo/pkg/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GetKubernetesIngressListWithMultiple(componentName, appName string, networkingV1Supported, extensionV1Supported bool) *unions.KubernetesIngressList {
	kubernetesIngressList := unions.NewEmptyKubernetesIngressList()
	kubernetesIngress1 := unions.NewKubernetesIngressFromParams(generator.IngressParams{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example-0",
			Labels: map[string]string{
				applabels.ApplicationLabel:                       appName,
				componentlabels.ComponentKubernetesInstanceLabel: componentName,
				applabels.ManagedBy:                              "odo",
				applabels.ManagerVersion:                         version.VERSION,
				labels.URLLabel:                                  "example-0",
				applabels.App:                                    appName,
			},
		},
		IngressSpecParams: generator.IngressSpecParams{
			IngressDomain: "example-0.com",
			ServiceName:   "example-0",
			PortNumber:    intstr.FromInt(8080),
		},
	})
	if !networkingV1Supported {
		kubernetesIngress1.NetworkingV1Ingress = nil
	}
	if !extensionV1Supported {
		kubernetesIngress1.ExtensionV1Beta1Ingress = nil
	}
	kubernetesIngressList.Items = append(kubernetesIngressList.Items, kubernetesIngress1)
	kubernetesIngress2 := unions.NewKubernetesIngressFromParams(generator.IngressParams{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example-1",
			Labels: map[string]string{
				applabels.ApplicationLabel:                       "app",
				componentlabels.ComponentKubernetesInstanceLabel: componentName,
				applabels.ManagedBy:                              "odo",
				applabels.ManagerVersion:                         version.VERSION,
				labels.URLLabel:                                  "example-1",
				applabels.App:                                    "app",
			},
		},
		IngressSpecParams: generator.IngressSpecParams{
			IngressDomain: "example-1.com",
			ServiceName:   "example-1",
			PortNumber:    intstr.FromInt(9090),
		},
	})
	if !networkingV1Supported {
		kubernetesIngress2.NetworkingV1Ingress = nil
	}
	if !extensionV1Supported {
		kubernetesIngress2.ExtensionV1Beta1Ingress = nil
	}
	kubernetesIngressList.Items = append(kubernetesIngressList.Items, kubernetesIngress2)
	return kubernetesIngressList
}

func GetSingleKubernetesIngress(urlName, componentName, appName string, networkingv1Supported, extensionv1Supported bool) *unions.KubernetesIngress {
	kubernetesIngress := unions.NewKubernetesIngressFromParams(generator.IngressParams{
		ObjectMeta: metav1.ObjectMeta{
			Name: urlName,
			Labels: map[string]string{
				applabels.ApplicationLabel:                       appName,
				componentlabels.ComponentKubernetesInstanceLabel: componentName,
				applabels.ManagedBy:                              "odo",
				applabels.ManagerVersion:                         version.VERSION,
				labels.URLLabel:                                  urlName,
				applabels.App:                                    appName,
			},
		},
		IngressSpecParams: generator.IngressSpecParams{
			IngressDomain: fmt.Sprintf("%s.com", urlName),
			ServiceName:   urlName,
			PortNumber:    intstr.FromInt(8080),
		},
	})
	if !networkingv1Supported {
		kubernetesIngress.NetworkingV1Ingress = nil
	}
	if !extensionv1Supported {
		kubernetesIngress.ExtensionV1Beta1Ingress = nil
	}
	return kubernetesIngress
}

// GetSingleSecureKubernetesIngress gets a single secure ingress with the given secret name
// if no secret name is provided, the default one is used
func GetSingleSecureKubernetesIngress(urlName, componentName, appName, secretName string, networkingV1Supported, extensionV1Supported bool) *unions.KubernetesIngress {
	if secretName == "" {
		suffix := util.GetAdler32Value(urlName + appName + componentName)

		secretName = urlName + "-" + suffix + "-tls"
	}
	kubernetesIngress := unions.NewKubernetesIngressFromParams(generator.IngressParams{
		ObjectMeta: metav1.ObjectMeta{
			Name: urlName,
			Labels: map[string]string{
				applabels.ApplicationLabel:                       appName,
				componentlabels.ComponentKubernetesInstanceLabel: componentName,
				applabels.ManagedBy:                              "odo",
				applabels.ManagerVersion:                         version.VERSION,
				labels.URLLabel:                                  urlName,
				applabels.App:                                    appName,
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
		kubernetesIngress.NetworkingV1Ingress = nil
	}
	if !extensionV1Supported {
		kubernetesIngress.ExtensionV1Beta1Ingress = nil
	}
	return kubernetesIngress
}
