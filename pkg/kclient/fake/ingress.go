package fake

import (
	"fmt"

	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/kclient/generator"
	"github.com/openshift/odo/pkg/url/labels"
	"github.com/openshift/odo/pkg/version"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GetIngressListWithMultiple(componentName, appName string) *extensionsv1.IngressList {
	return &extensionsv1.IngressList{
		Items: []extensionsv1.Ingress{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "example-0",
					Labels: map[string]string{
						applabels.ApplicationLabel:     appName,
						componentlabels.ComponentLabel: componentName,
						applabels.OdoManagedBy:         "odo",
						applabels.OdoVersion:           version.VERSION,
						labels.URLLabel:                "example-0",
						applabels.App:                  appName,
					},
				},
				Spec: *generator.GetIngressSpec(generator.IngressParams{IngressDomain: "example-0.com", ServiceName: "example-0", PortNumber: intstr.FromInt(8080)}),
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "example-1",
					Labels: map[string]string{
						applabels.ApplicationLabel:     "app",
						componentlabels.ComponentLabel: componentName,
						applabels.OdoManagedBy:         "odo",
						applabels.OdoVersion:           version.VERSION,
						labels.URLLabel:                "example-1",
						applabels.App:                  "app",
					},
				},
				Spec: *generator.GetIngressSpec(generator.IngressParams{IngressDomain: "example-1.com", ServiceName: "example-1", PortNumber: intstr.FromInt(9090)}),
			},
		},
	}
}

func GetSingleIngress(urlName, componentName, appName string) *extensionsv1.Ingress {
	return &extensionsv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: urlName,
			Labels: map[string]string{
				applabels.ApplicationLabel:     appName,
				componentlabels.ComponentLabel: componentName,
				applabels.OdoManagedBy:         "odo",
				applabels.OdoVersion:           version.VERSION,
				labels.URLLabel:                urlName,
				applabels.App:                  appName,
			},
		},
		Spec: *generator.GetIngressSpec(generator.IngressParams{IngressDomain: fmt.Sprintf("%s.com", urlName), ServiceName: urlName, PortNumber: intstr.FromInt(8080)}),
	}
}
