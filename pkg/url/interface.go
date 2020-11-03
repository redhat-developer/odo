package url

import (
	"fmt"
	"github.com/openshift/odo/pkg/component"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/kclient"
	urlLabels "github.com/openshift/odo/pkg/url/labels"
	"github.com/pkg/errors"
	iextensionsv1 "k8s.io/api/extensions/v1beta1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog"
	"strings"
)

type url interface {
	Create() error
	Delete() error
	GetName() string
	GetHost() string
	GetProtocol() string
	GetPort() int
	IsSecure() bool
	GetKind() URLKind
	GetTLSSecret() string
	GetPath() string
	GetStatus() URLStatus
}

type URLKind string

const (
	INGRESS URLKind = "ingress"
	ROUTE   URLKind = "route"
)

func Coalescing(a, b interface{}) interface{} {
	if a != nil {
		return a
	} else if b != nil {
		return b
	}
	return nil
}

func NewPushedURL(data interface{}, localState envinfo.EnvInfoURL) url {
	if value, ok := data.(iextensionsv1.Ingress); ok {
		return ingressURL{
			localState: &localState,
			ingress:    &value,
		}
	}
	return nil
}

type ingressURL struct {
	component  component.KubernetesComponent
	localState *envinfo.EnvInfoURL
	ingress    *iextensionsv1.Ingress
}

func (i ingressURL) Delete() error {
	client, err := kclient.GetInstance()
	if err != nil {
		return err
	}
	return client.DeleteIngress(i.localState.Name)
}

func NewNonPushedIngress(localState *envinfo.EnvInfoURL, kubernetesComponent component.KubernetesComponent) url {
	return ingressURL{
		component:  kubernetesComponent,
		localState: localState,
		ingress:    nil,
	}
}

func NewIngress(ingress *iextensionsv1.Ingress, localState *envinfo.EnvInfoURL) url {
	return ingressURL{
		ingress:    ingress,
		localState: localState,
	}
}

func (i ingressURL) Create() error {
	client, err := kclient.GetInstance()
	if err != nil {
		return err
	}

	labels := urlLabels.GetLabels(i.localState.Name, i.component.GetName(), i.component.GetApplication(), true)

	if i.localState.Host == "" {
		return errors.Errorf("the host cannot be empty")
	}
	serviceName := i.component.GetName()
	ingressDomain := fmt.Sprintf("%v.%v", i.localState.Name, i.localState.Host)

	ownerReference := i.component.GetOwnerReference()
	if i.localState.Secure {
		if len(i.localState.TLSSecret) != 0 {
			_, err := client.KubeClient.CoreV1().Secrets(client.Namespace).Get(i.localState.TLSSecret, metav1.GetOptions{})
			if err != nil {
				return errors.Wrap(err, "unable to get the provided secret: "+i.localState.TLSSecret)
			}
		}
		if len(i.localState.TLSSecret) == 0 {
			defaultTLSSecretName := i.component.GetName() + "-tlssecret"
			_, err := client.KubeClient.CoreV1().Secrets(client.Namespace).Get(defaultTLSSecretName, metav1.GetOptions{})
			// create tls secret if it does not exist
			if kerrors.IsNotFound(err) {
				selfsignedcert, err := kclient.GenerateSelfSignedCertificate(i.localState.Host)
				if err != nil {
					return errors.Wrap(err, "unable to generate self-signed certificate for clutser: "+i.localState.Host)
				}
				// create tls secret
				secretlabels := componentlabels.GetLabels(i.component.GetName(), i.component.GetApplication(), true)
				objectMeta := metav1.ObjectMeta{
					Name:   defaultTLSSecretName,
					Labels: secretlabels,
					OwnerReferences: []v1.OwnerReference{
						ownerReference,
					},
				}
				secret, err := client.CreateTLSSecret(selfsignedcert.CertPem, selfsignedcert.KeyPem, objectMeta)
				if err != nil {
					return errors.Wrap(err, "unable to create tls secret")
				}
				i.localState.TLSSecret = secret.Name
			} else if err != nil {
				return err
			} else {
				// tls secret found for this component
				i.localState.TLSSecret = defaultTLSSecretName
			}

		}

	}
	ingressParam := kclient.IngressParameter{ServiceName: serviceName, IngressDomain: ingressDomain, PortNumber: intstr.FromInt(i.localState.Port), TLSSecretName: i.localState.TLSSecret, Path: i.localState.Path}
	ingressSpec := kclient.GenerateIngressSpec(ingressParam)
	objectMeta := kclient.CreateObjectMeta(i.component.GetName(), client.Namespace, labels, nil)
	// to avoid error due to duplicate ingress name defined in different devfile components
	objectMeta.Name = fmt.Sprintf("%s-%s", i.localState.Name, i.component.GetName())
	objectMeta.OwnerReferences = append(objectMeta.OwnerReferences, ownerReference)
	// Pass in the namespace name, link to the service (componentName) and labels to create a ingress
	_, err = client.CreateIngress(objectMeta, *ingressSpec)
	if err != nil {
		return errors.Wrap(err, "unable to create ingress")
	}
	//return GetURLString(GetProtocol(routev1.Route{}, *ingress), "", ingressDomain, false), nil
	return nil
}

func (i ingressURL) GetName() string {
	return Coalescing(i.ingress.Name, i.localState.Name).(string)
}

func (i ingressURL) GetHost() string {
	return strings.Replace(i.ingress.Spec.Rules[0].Host, i.GetName()+".", "", 1)
}

func (i ingressURL) GetProtocol() string {
	if i.ingress.Spec.TLS != nil {
		return "https"
	}
	return "https"
}

func (i ingressURL) GetPort() int {
	return i.ingress.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort.IntValue()
}

func (i ingressURL) IsSecure() bool {
	return i.ingress.Spec.TLS != nil
}

func (i ingressURL) GetKind() URLKind {
	return INGRESS
}

func (i ingressURL) GetTLSSecret() string {
	return i.ingress.Spec.TLS[0].SecretName
}

func (i ingressURL) GetPath() string {
	return i.ingress.Spec.Rules[0].HTTP.Paths[0].Path
}

func (i ingressURL) GetStatus() URLStatus {
	var urlStatus URLStatus
	if i.localState == nil {
		urlStatus = URLStatus{
			State: StateTypeNotPushed,
		}
	}
	if i.ingress != nil {
		if urlStatus.State == StateTypeLocallyDeleted {
			urlStatus = URLStatus{
				State: StateTypeLocallyDeleted,
			}
		} else {
			urlStatus = URLStatus{
				State: StateTypePushed,
			}
		}
	} else {
		urlStatus = URLStatus{
			State: StateTypeUnknown,
		}
	}
	return urlStatus
}

///////////////////////

type urlClient interface {
	Create(url envinfo.EnvInfoURL, component component.OdoComponent) error
	Delete(url url) error
	List() ([]url, error)
	ListPushed() ([]url, error)
}

type kubernetesClient struct {
	kClient   kclient.Client
	env       envinfo.LocalConfigProvider
	component component.OdoComponent
}

func NewClient() urlClient {
	// TODO use context to determine type of client
	return kubernetesClient{
		// get the kClient from the context or create one
	}
}

func (k kubernetesClient) ListPushed() ([]url, error) {
	labelSelector := fmt.Sprintf("%v=%v", componentlabels.ComponentLabel, k.component.GetName())
	klog.V(4).Infof("Listing ingresses with label selector: %v", labelSelector)
	ingresses, err := k.kClient.ListIngresses(labelSelector)
	if err != nil {
		return []url{}, errors.Wrap(err, "unable to list ingress names")
	}

	var urls []url
	for _, ingress := range ingresses {
		for _, localURL := range k.env.GetURL() {
			if localURL.Name == ingress.Name {
				urls = append(urls, NewIngress(&ingress, &localURL))
			}
		}
	}
	return urls, nil
}

func (k kubernetesClient) List() ([]url, error) {
	remoteUrls, _ := k.ListPushed()

	if _, ok := k.component.(component.KubernetesComponent); !ok {
		return []url{}, nil
	}

	var urls []url
	for _, localURL := range k.env.GetURL() {
		found := false
		for _, remoteUrls := range remoteUrls {
			found = true
			if remoteUrls.GetName() == localURL.Name {
				if remoteUrls.GetStatus().State == StateTypePushed {
					found = true
				}
			}
		}
		if !found {
			url := NewNonPushedIngress(&localURL, k.component.(component.KubernetesComponent))
			urls = append(urls, url)
		}
	}
	return urls, nil
}

func (k kubernetesClient) Delete(url url) error {
	if url.GetKind() == INGRESS {
		return url.Delete()
	}
	return nil
}

func (k kubernetesClient) Create(url envinfo.EnvInfoURL, cmp component.OdoComponent) error {
	if url.Kind == envinfo.INGRESS {
		if kubeCmp, ok := cmp.(component.KubernetesComponent); ok {
			url := NewNonPushedIngress(&url, kubeCmp)
			return url.Create()
		}
	}
	return nil
}

func URLPush(client urlClient) error {
	urls, _ := client.List()

	for _, url := range urls {
		if url.GetStatus().State == StateTypeLocallyDeleted {
			url.Delete()
		} else if url.GetStatus().State == StateTypeNotPushed {
			url.Create()
		}
	}
	return nil
}
