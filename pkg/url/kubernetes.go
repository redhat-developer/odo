package url

import (
	"fmt"
	"sort"

	"github.com/redhat-developer/odo/pkg/unions"

	"github.com/devfile/library/pkg/devfile/generator"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	"github.com/redhat-developer/odo/pkg/occlient"
	urlLabels "github.com/redhat-developer/odo/pkg/url/labels"
	"github.com/redhat-developer/odo/pkg/util"
	appsV1 "k8s.io/api/apps/v1"
	iextensionsv1 "k8s.io/api/extensions/v1beta1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog"
)

// kubernetesClient contains information required for devfile based URL based operations
type kubernetesClient struct {
	generic
	isRouteSupported bool
	client           occlient.Client

	// if we don't have access to the local config
	// we can use the deployment to call ListFromCluster() and
	// directly list storage from the cluster without the local config
	deployment *appsV1.Deployment
}

// ListFromCluster lists both route and ingress based URLs from the cluster
func (k kubernetesClient) ListFromCluster() (URLList, error) {
	if k.componentName == "" || k.appName == "" {
		return URLList{}, fmt.Errorf("the component name, the app name or both are empty")
	}
	labelSelector := componentlabels.GetSelector(k.componentName, k.appName)
	klog.V(4).Infof("Listing ingresses with label selector: %v", labelSelector)
	ingresses, err := k.client.GetKubeClient().ListIngresses(labelSelector)
	if err != nil {
		return URLList{}, errors.Wrap(err, "unable to list ingress")
	}

	var routes []routev1.Route
	if k.isRouteSupported {
		routes, err = k.client.ListRoutes(labelSelector)
		if err != nil {
			return URLList{}, errors.Wrap(err, "unable to list routes")
		}
	}

	var clusterURLs []URL
	clusterURLs = append(clusterURLs, NewURLsFromKubernetesIngressList(ingresses)...)
	for _, r := range routes {
		// ignore the routes created by ingresses
		if r.OwnerReferences != nil && r.OwnerReferences[0].Kind == "Ingress" {
			continue
		}
		clusterURL := NewURL(r)
		clusterURLs = append(clusterURLs, clusterURL)
	}

	return NewURLList(clusterURLs), nil
}

// List lists both route/ingress based URLs and local URLs with respective states
func (k kubernetesClient) List() (URLList, error) {
	// get the URLs present on the cluster
	clusterURLMap := make(map[string]URL)
	var clusterURLs URLList
	var err error
	if k.client.GetKubeClient() != nil {
		clusterURLs, err = k.ListFromCluster()
		if err != nil {
			return URLList{}, errors.Wrap(err, "unable to list routes")
		}
	}

	for _, url := range clusterURLs.Items {
		clusterURLMap[url.Name] = url
	}

	localMap := make(map[string]URL)
	if k.localConfigProvider != nil {
		// get the URLs present on the localConfigProvider
		localURLS, err := k.localConfigProvider.ListURLs()
		if err != nil {
			return URLList{}, err
		}
		for _, url := range localURLS {
			if !k.isRouteSupported && url.Kind == localConfigProvider.ROUTE {
				continue
			}
			localURL := NewURLFromEnvinfoURL(url, k.componentName)
			if localURL.Spec.Protocol == "" {
				if localURL.Spec.Secure {
					localURL.Spec.Protocol = "https"
				} else {
					localURL.Spec.Protocol = "http"
				}
			}
			localMap[url.Name] = localURL
		}
	}

	// find the URLs which are present on the cluster but not on the localConfigProvider
	// if not found on the localConfigProvider, mark them as 'StateTypeLocallyDeleted'
	// else mark them as 'StateTypePushed'
	var urls sortableURLs
	for URLName, clusterURL := range clusterURLMap {
		_, found := localMap[URLName]
		if found {
			// URL is in both local env file and cluster
			clusterURL.Status.State = StateTypePushed
			urls = append(urls, clusterURL)
		} else {
			// URL is on the cluster but not in local env file
			clusterURL.Status.State = StateTypeLocallyDeleted
			urls = append(urls, clusterURL)
		}
	}

	// find the URLs which are present on the localConfigProvider but not on the cluster
	// if not found on the cluster, mark them as 'StateTypeNotPushed'
	for localName, localURL := range localMap {
		_, remoteURLFound := clusterURLMap[localName]
		if !remoteURLFound {
			// URL is in the local env file but not pushed to cluster
			localURL.Status.State = StateTypeNotPushed
			urls = append(urls, localURL)
		}
	}

	// sort urls by name to get consistent output
	sort.Sort(urls)
	urlList := NewURLList(urls)
	return urlList, nil
}

// Delete deletes the URL with the given name and kind
func (k kubernetesClient) Delete(name string, kind localConfigProvider.URLKind) error {
	if k.componentName == "" || k.appName == "" {
		return fmt.Errorf("the component name, the app name or both are empty")
	}
	selector := util.ConvertLabelsToSelector(urlLabels.GetLabels(name, k.componentName, k.appName, false))

	switch kind {
	case localConfigProvider.INGRESS:
		ingress, err := k.client.GetKubeClient().GetOneIngressFromSelector(selector)
		if err != nil {
			return err
		}
		return k.client.GetKubeClient().DeleteIngress(ingress.GetName())
	case localConfigProvider.ROUTE:
		route, err := k.client.GetOneRouteFromSelector(selector)
		if err != nil {
			return err
		}
		return k.client.DeleteRoute(route.Name)
	default:
		return fmt.Errorf("url type is not supported")
	}
}

// Create creates a route or ingress based on the given URL
func (k kubernetesClient) Create(url URL) (string, error) {
	if k.componentName == "" || k.appName == "" {
		return "", fmt.Errorf("the component name, the app name or both are empty")
	}

	if url.Spec.Kind != localConfigProvider.INGRESS && url.Spec.Kind != localConfigProvider.ROUTE {
		return "", fmt.Errorf("urlKind %s is not supported for URL creation", url.Spec.Kind)
	}

	if !url.Spec.Secure && url.Spec.TLSSecret != "" {
		return "", fmt.Errorf("secret name can only be used for secure URLs")
	}

	labels := urlLabels.GetLabels(url.Name, k.componentName, k.appName, true)

	if url.Spec.Kind == localConfigProvider.INGRESS {
		return k.createIngress(url, labels)
	} else {
		if !k.isRouteSupported {
			return "", errors.Errorf("routes are not available on non OpenShift clusters")
		}

		return k.createRoute(url, labels)
	}

}

// createIngress creates a ingress for the given URL with the given labels
func (k kubernetesClient) createIngress(url URL, labels map[string]string) (string, error) {
	if url.Spec.Host == "" {
		return "", errors.Errorf("the host cannot be empty")
	}

	service, err := k.client.GetKubeClient().GetOneService(k.componentName, k.appName)
	if err != nil {
		return "", err
	}

	ingressDomain := fmt.Sprintf("%v.%v", url.Name, url.Spec.Host)

	// generate the owner reference
	if k.deployment == nil {
		k.deployment, err = k.client.GetKubeClient().GetOneDeployment(k.componentName, k.appName)
		if err != nil {
			return "", err
		}
	}
	ownerReference := generator.GetOwnerReference(k.deployment)

	if url.Spec.Secure {
		if len(url.Spec.TLSSecret) != 0 {
			// get the user given secret
			_, err = k.client.GetKubeClient().GetSecret(url.Spec.TLSSecret, k.client.Namespace)
			if err != nil {
				return "", errors.Wrap(err, "unable to get the provided secret: "+url.Spec.TLSSecret)
			}
		} else {
			// get the default secret
			defaultTLSSecretName := getDefaultTLSSecretName(url.Name, k.componentName, k.appName)
			_, err = k.client.GetKubeClient().GetSecret(defaultTLSSecretName, k.client.Namespace)

			// create tls secret if it does not exist
			if kerrors.IsNotFound(err) {
				selfSignedCert, e := kclient.GenerateSelfSignedCertificate(url.Spec.Host)
				if e != nil {
					return "", errors.Wrap(e, "unable to generate self-signed certificate for clutser: "+url.Spec.Host)
				}
				// create tls secret
				secretLabels := componentlabels.GetLabels(k.componentName, k.appName, true)
				objectMeta := metav1.ObjectMeta{
					Name:   defaultTLSSecretName,
					Labels: secretLabels,
					OwnerReferences: []v1.OwnerReference{
						ownerReference,
					},
				}
				secret, e := k.client.GetKubeClient().CreateTLSSecret(selfSignedCert.CertPem, selfSignedCert.KeyPem, objectMeta)
				if e != nil {
					return "", errors.Wrap(e, "unable to create tls secret")
				}
				url.Spec.TLSSecret = secret.Name
			} else if err != nil {
				return "", err
			} else {
				// tls secret found for this component
				url.Spec.TLSSecret = defaultTLSSecretName
			}

		}

	}

	suffix := util.GetAdler32Value(url.Name + k.appName + k.componentName)
	ingressName, err := util.NamespaceOpenShiftObject(url.Name, suffix)
	if err != nil {
		return "", err
	}
	objectMeta := generator.GetObjectMeta(k.componentName, k.client.Namespace, labels, nil)
	// to avoid error due to duplicate ingress name defined in different devfile components
	objectMeta.Name = ingressName
	objectMeta.OwnerReferences = append(objectMeta.OwnerReferences, ownerReference)

	ingressParam := generator.IngressParams{
		ObjectMeta: objectMeta,
		IngressSpecParams: generator.IngressSpecParams{
			ServiceName:   service.Name,
			IngressDomain: ingressDomain,
			PortNumber:    intstr.FromInt(url.Spec.Port),
			TLSSecretName: url.Spec.TLSSecret,
			Path:          url.Spec.Path,
		},
	}
	ingress := unions.NewKubernetesIngressFromParams(ingressParam)
	// Pass in the namespace name, link to the service (componentName) and labels to create a ingress
	i, err := k.client.GetKubeClient().CreateIngress(*ingress)
	if err != nil {
		return "", fmt.Errorf("unable to create ingress %w", err)
	}
	return i.GetURLString(), nil
}

// createRoute creates a route for the given URL with the given labels
func (k kubernetesClient) createRoute(url URL, labels map[string]string) (string, error) {
	// to avoid error due to duplicate ingress name defined in different devfile components
	// we avoid using the getResourceName() and use the previous method from s2i
	// as the host name, which is automatically created on openshift,
	// can become more than 63 chars, which is invalid
	suffix := util.GetAdler32Value(url.Name + k.appName + k.componentName)
	routeName, err := util.NamespaceOpenShiftObject(url.Name, suffix)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create namespaced name")
	}

	if k.deployment == nil {
		k.deployment, err = k.client.GetKubeClient().GetOneDeployment(k.componentName, k.appName)
		if err != nil {
			return "", err
		}
	}
	ownerReference := generator.GetOwnerReference(k.deployment)

	service, err := k.client.GetKubeClient().GetOneService(k.componentName, k.appName)
	if err != nil {
		return "", err
	}

	// Pass in the namespace name, link to the service (componentName) and labels to create a route
	route, err := k.client.CreateRoute(routeName, service.Name, intstr.FromInt(url.Spec.Port), labels, url.Spec.Secure, url.Spec.Path, ownerReference)
	if err != nil {
		if kerrors.IsAlreadyExists(err) {
			return "", fmt.Errorf("url named %q already exists in the same app named %q", url.Name, k.appName)
		}
		return "", errors.Wrap(err, "unable to create route")
	}
	return GetURLString(GetProtocol(*route, iextensionsv1.Ingress{}), route.Spec.Host, ""), nil
}
