package fake

import (
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/url/labels"
	"github.com/openshift/odo/pkg/version"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GetIngressListWithMultiple(componentName string) *extensionsv1.IngressList {
	return &extensionsv1.IngressList{
		Items: []extensionsv1.Ingress{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "example-0",
					Labels: map[string]string{
						applabels.ApplicationLabel:     "",
						componentlabels.ComponentLabel: componentName,
						applabels.OdoManagedBy:         "odo",
						applabels.OdoVersion:           version.VERSION,
						labels.URLLabel:                "example-0",
					},
				},
				Spec: *kclient.GenerateIngressSpec(kclient.IngressParameter{ServiceName: "example-0", PortNumber: intstr.FromInt(8080)}),
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "example-1",
					Labels: map[string]string{
						applabels.ApplicationLabel:     "",
						componentlabels.ComponentLabel: componentName,
						applabels.OdoManagedBy:         "odo",
						applabels.OdoVersion:           version.VERSION,
						labels.URLLabel:                "example-1",
					},
				},
				Spec: *kclient.GenerateIngressSpec(kclient.IngressParameter{ServiceName: "example-1", PortNumber: intstr.FromInt(8080)}),
			},
		},
	}
}

func GetSingleIngress(urlName, componentName string) *extensionsv1.Ingress {
	return &extensionsv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: urlName,
			Labels: map[string]string{
				applabels.ApplicationLabel:     "",
				componentlabels.ComponentLabel: componentName,
				applabels.OdoManagedBy:         "odo",
				applabels.OdoVersion:           version.VERSION,
				labels.URLLabel:                urlName,
				applabels.App:                  "",
			},
		},
		Spec: *kclient.GenerateIngressSpec(kclient.IngressParameter{ServiceName: urlName, PortNumber: intstr.FromInt(8080)}),
	}
}
