package url

import (
	"fmt"
	"reflect"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/redhat-developer/odo/pkg/kclient/unions"
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	urlLabels "github.com/redhat-developer/odo/pkg/url/labels"
	iextensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const URLKind = "URL"

// URL is an abstraction giving network access to the component from outside the cluster
type URL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              URLSpec   `json:"spec,omitempty"`
	Status            URLStatus `json:"status,omitempty"`
}

// URLSpec contains the specifications of a URL
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

// URLList is a list of urls
type URLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []URL `json:"items"`
}

// URLStatus is the current status of a url
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

// NewURL gives machine readable URL definition
func NewURL(r routev1.Route) URL {
	return URL{
		TypeMeta: metav1.TypeMeta{
			Kind:       URLKind,
			APIVersion: machineoutput.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Labels[urlLabels.URLLabel],
		},
		Spec: URLSpec{
			Host:     r.Spec.Host,
			Port:     r.Spec.Port.TargetPort.IntValue(),
			Protocol: GetProtocol(r, iextensionsv1.Ingress{}),
			Secure:   r.Spec.TLS != nil,
			Path:     r.Spec.Path,
			Kind:     localConfigProvider.ROUTE,
		},
	}

}

func NewURLList(urls []URL) URLList {
	return URLList{
		TypeMeta: metav1.TypeMeta{
			Kind:       machineoutput.ListKind,
			APIVersion: machineoutput.APIVersion,
		},
		ListMeta: metav1.ListMeta{},
		Items:    urls,
	}
}

// NewURLFromConfigURL creates a URL from a ConfigURL
func NewURLFromConfigURL(configURL localConfigProvider.LocalURL) URL {
	return URL{
		TypeMeta: metav1.TypeMeta{
			Kind:       URLKind,
			APIVersion: machineoutput.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: configURL.Name,
		},
		Spec: URLSpec{
			Port:   configURL.Port,
			Secure: configURL.Secure,
			Kind:   localConfigProvider.ROUTE,
			Path:   "/",
		},
	}
}

// NewURLFromEnvinfoURL creates a URL from a EnvinfoURL
func NewURLFromEnvinfoURL(envinfoURL localConfigProvider.LocalURL, serviceName string) URL {
	hostString := fmt.Sprintf("%s.%s", envinfoURL.Name, envinfoURL.Host)
	// default to route kind if none is provided
	kind := envinfoURL.Kind
	if kind == "" {
		kind = localConfigProvider.ROUTE
	}
	url := URL{
		TypeMeta: metav1.TypeMeta{
			Kind:       URLKind,
			APIVersion: machineoutput.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: envinfoURL.Name,
		},
		Spec: URLSpec{
			Host:      envinfoURL.Host,
			Protocol:  envinfoURL.Protocol,
			Port:      envinfoURL.Port,
			Secure:    envinfoURL.Secure,
			Kind:      kind,
			TLSSecret: envinfoURL.TLSSecret,
			Path:      envinfoURL.Path,
		},
	}
	if kind == localConfigProvider.INGRESS {
		url.Spec.Host = hostString
		if envinfoURL.Secure && len(envinfoURL.TLSSecret) > 0 {
			url.Spec.TLSSecret = envinfoURL.TLSSecret
		} else if envinfoURL.Secure {
			url.Spec.TLSSecret = fmt.Sprintf("%s-tlssecret", serviceName)
		}
	}
	return url
}

// NewURLFromLocalURL creates a URL from a localConfigProvider.LocalURL
func NewURLFromLocalURL(localURL localConfigProvider.LocalURL) URL {
	return URL{
		TypeMeta: metav1.TypeMeta{
			Kind:       URLKind,
			APIVersion: machineoutput.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: localURL.Name,
		},
		Spec: URLSpec{
			Host:      localURL.Host,
			Protocol:  localURL.Protocol,
			Port:      localURL.Port,
			Secure:    localURL.Secure,
			Kind:      localURL.Kind,
			TLSSecret: localURL.TLSSecret,
			Path:      localURL.Path,
		},
	}
}

func NewURLsFromKubernetesIngressList(kil *unions.KubernetesIngressList) []URL {
	var urlList []URL
	for _, item := range kil.Items {
		urlItem := NewURLFromKubernetesIngress(item, true)
		if !reflect.DeepEqual(urlItem, URL{}) {
			urlList = append(urlList, urlItem)
		}
	}
	return urlList
}

func NewURLFromKubernetesIngress(ki *unions.KubernetesIngress, skipIfGenerated bool) URL {
	if skipIfGenerated && ki.IsGenerated() {
		return URL{}
	}
	u := URL{
		TypeMeta: metav1.TypeMeta{
			Kind:       URLKind,
			APIVersion: machineoutput.APIVersion,
		},
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

func (urls URLList) AreOutOfSync() bool {
	outOfSync := false
	for _, u := range urls.Items {
		if u.Status.State != StateTypePushed {
			outOfSync = true
			break
		}
	}
	return outOfSync
}
