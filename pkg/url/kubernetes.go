package url

import (
	"fmt"
	"sort"

	"github.com/devfile/library/pkg/devfile/generator"
	routev1 "github.com/openshift/api/route/v1"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/occlient"
	urlLabels "github.com/openshift/odo/pkg/url/labels"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
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
}

// ListFromCluster lists both route and ingress based URLs from the cluster
func (k kubernetesClient) ListFromCluster() (URLList, error) {
	labelSelector := fmt.Sprintf("%v=%v", componentlabels.ComponentLabel, k.componentName)
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
	for _, i := range ingresses {
		clusterURL := getMachineReadableFormatIngress(i)
		clusterURLs = append(clusterURLs, clusterURL)
	}
	for _, r := range routes {
		// ignore the routes created by ingresses
		if r.OwnerReferences != nil && r.OwnerReferences[0].Kind == "Ingress" {
			continue
		}
		clusterURL := getMachineReadableFormat(r)
		clusterURLs = append(clusterURLs, clusterURL)
	}

	return getMachineReadableFormatForList(clusterURLs), nil
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
	if k.localConfig != nil {
		// get the URLs present on the localConfigProvider
		localURLS, err := k.localConfig.ListURLs()
		if err != nil {
			return URLList{}, err
		}
		for _, url := range localURLS {
			if !k.isRouteSupported && url.Kind == localConfigProvider.ROUTE {
				continue
			}
			localURL := ConvertEnvinfoURL(url, k.componentName)
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
	urlList := getMachineReadableFormatForList(urls)
	return urlList, nil
}

// Delete deletes the URL with the given name and kind
func (k kubernetesClient) Delete(name string, kind localConfigProvider.URLKind) error {
	selector := util.ConvertLabelsToSelector(urlLabels.GetLabels(name, k.componentName, k.appName, false))

	switch kind {
	case localConfigProvider.INGRESS:
		ingress, err := k.client.GetKubeClient().GetOneIngressFromSelector(selector)
		if err != nil {
			return err
		}
		return k.client.GetKubeClient().DeleteIngress(ingress.Name)
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
	serviceName := k.componentName
	ingressDomain := fmt.Sprintf("%v.%v", url.Name, url.Spec.Host)

	// generate the owner reference
	deployment, err := k.client.GetKubeClient().GetOneDeployment(k.componentName, k.appName)
	if err != nil {
		return "", err
	}
	ownerReference := generator.GetOwnerReference(deployment)

	if url.Spec.Secure {
		if len(url.Spec.TLSSecret) != 0 {
			// get the user given secret
			_, err := k.client.GetKubeClient().GetSecret(url.Spec.TLSSecret, k.client.Namespace)
			if err != nil {
				return "", errors.Wrap(err, "unable to get the provided secret: "+url.Spec.TLSSecret)
			}
		} else {
			// get the default secret
			defaultTLSSecretName := k.componentName + "-tlssecret"
			_, err := k.client.GetKubeClient().GetSecret(defaultTLSSecretName, k.client.Namespace)

			// create tls secret if it does not exist
			if kerrors.IsNotFound(err) {
				selfSignedCert, err := kclient.GenerateSelfSignedCertificate(url.Spec.Host)
				if err != nil {
					return "", errors.Wrap(err, "unable to generate self-signed certificate for clutser: "+url.Spec.Host)
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
				secret, err := k.client.GetKubeClient().CreateTLSSecret(selfSignedCert.CertPem, selfSignedCert.KeyPem, objectMeta)
				if err != nil {
					return "", errors.Wrap(err, "unable to create tls secret")
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

	ingressName, err := getResourceName(url.Name, k.componentName, k.appName)
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
			ServiceName:   serviceName,
			IngressDomain: ingressDomain,
			PortNumber:    intstr.FromInt(url.Spec.Port),
			TLSSecretName: url.Spec.TLSSecret,
			Path:          url.Spec.Path,
		},
	}
	ingress := generator.GetIngress(ingressParam)
	// Pass in the namespace name, link to the service (componentName) and labels to create a ingress
	i, err := k.client.GetKubeClient().CreateIngress(*ingress)
	if err != nil {
		return "", errors.Wrap(err, "unable to create ingress")
	}
	return GetURLString(GetProtocol(routev1.Route{}, *i), "", ingressDomain, false), nil
}

// createRoute creates a route for the given URL with the given labels
func (k kubernetesClient) createRoute(url URL, labels map[string]string) (string, error) {
	// to avoid error due to duplicate ingress name defined in different devfile components
	// we avoid using the getResourceName() and use the previous method from s2i
	// as the host name, which is automatically created on openshift,
	// can become more than 63 chars, which is invalid
	routeName, err := util.NamespaceOpenShiftObject(url.Name, k.appName)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create namespaced name")
	}
	serviceName := k.componentName

	deployment, err := k.client.GetKubeClient().GetOneDeployment(k.componentName, k.appName)
	if err != nil {
		return "", err
	}
	ownerReference := generator.GetOwnerReference(deployment)

	// Pass in the namespace name, link to the service (componentName) and labels to create a route
	route, err := k.client.CreateRoute(routeName, serviceName, intstr.FromInt(url.Spec.Port), labels, url.Spec.Secure, url.Spec.Path, ownerReference)
	if err != nil {
		if kerrors.IsAlreadyExists(err) {
			return "", fmt.Errorf("url named %q already exists in the same app named %q", url.Name, k.appName)
		}
		return "", errors.Wrap(err, "unable to create route")
	}
	return GetURLString(GetProtocol(*route, iextensionsv1.Ingress{}), route.Spec.Host, "", true), nil
}

// getResourceName gets the route/ingress resource name
func getResourceName(urlName, componentName, appName string) (string, error) {
	resourceName, err := util.NamespaceKubernetesObject(urlName, componentName)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create namespaced name")
	}

	resourceName, err = util.NamespaceKubernetesObject(resourceName, appName)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create namespaced name")
	}
	return resourceName, nil
}
