package kclient

import (
	"context"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

const (
	// timeout for waiting for namespace deletion
	waitForNamespaceDeletionTimeOut = 3 * time.Minute

	// timeout for getting the default service account
	getDefaultServiceAccTimeout = 1 * time.Minute
)

// GetNamespaces return list of existing namespaces that user has access to.
func (c *Client) GetNamespaces() ([]string, error) {
	namespaces, err := c.KubeClient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list namespaces")
	}

	var names []string
	for _, p := range namespaces.Items {
		names = append(names, p.Name)
	}
	return names, nil
}

// GetNamespace returns Namespace based on its name
// Errors related to project not being found or forbidden are translated to nil project for compatibility
func (c *Client) GetNamespace(name string) (*corev1.Namespace, error) {
	ns, err := c.KubeClient.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		istatus, ok := err.(kerrors.APIStatus)
		if ok {
			status := istatus.Status()
			if status.Reason == metav1.StatusReasonNotFound || status.Reason == metav1.StatusReasonForbidden {
				return nil, nil
			}
		} else {
			return nil, err
		}

	}
	return ns, err

}

// CreateNamespace creates new namespace
func (c *Client) CreateNamespace(name string) (*corev1.Namespace, error) {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	newNamespace, err := c.KubeClient.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{FieldManager: FieldManager})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create Namespace %s", namespace.ObjectMeta.Name)
	}
	return newNamespace, nil
}

// DeleteNamespace deletes namespace
// if wait=true , it will wait for deletion
func (c *Client) DeleteNamespace(name string, wait bool) error {
	var watcher watch.Interface
	var err error
	if wait {
		watcher, err = c.KubeClient.CoreV1().Namespaces().Watch(context.TODO(), metav1.ListOptions{
			FieldSelector: fields.Set{"metadata.name": name}.AsSelector().String(),
		})
		if err != nil {
			return errors.Wrapf(err, "unable to watch namespace")
		}
		defer watcher.Stop()
	}

	err = c.KubeClient.CoreV1().Namespaces().Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "unable to delete Namespace %s", name)
	}

	if watcher != nil {
		namespaceChannel := make(chan *corev1.Namespace)
		watchErrorChannel := make(chan error)

		go func() {
			for {
				val, ok := <-watcher.ResultChan()
				if !ok {
					watchErrorChannel <- errors.Errorf("watch channel was closed unexpectedly: %+v", val)
					break
				}
				klog.V(3).Infof("Watch event.Type '%s'.", val.Type)

				if namespaceStatus, ok := val.Object.(*corev1.Namespace); ok {
					klog.V(3).Infof("Status of delete of namespace %s is '%s'.", name, namespaceStatus.Status.Phase)
					if val.Type == watch.Deleted {
						namespaceChannel <- namespaceStatus
						break
					}
					if val.Type == watch.Error {
						watchErrorChannel <- errors.Errorf("failed watching the deletion of namespace %s", name)
						break
					}

				}

			}
			close(namespaceChannel)
			close(watchErrorChannel)
		}()

		select {
		case val := <-namespaceChannel:
			klog.V(3).Infof("Namespace %s deleted", val.Name)
			return nil
		case err := <-watchErrorChannel:
			return err
		case <-time.After(waitForNamespaceDeletionTimeOut):
			return errors.Errorf("waited %s but couldn't delete namespace %s in time", waitForNamespaceDeletionTimeOut, name)
		}

	}
	return nil
}

// SetCurrentNamespace change current namespace in kubeconfig
func (c *Client) SetCurrentNamespace(namespace string) error {
	rawConfig, err := c.KubeConfig.RawConfig()
	if err != nil {
		return errors.Wrapf(err, "unable to switch to %s project", namespace)
	}

	rawConfig.Contexts[rawConfig.CurrentContext].Namespace = namespace

	err = clientcmd.ModifyConfig(clientcmd.NewDefaultClientConfigLoadingRules(), rawConfig, true)
	if err != nil {
		return errors.Wrapf(err, "unable to switch to %s project", namespace)
	}

	c.Namespace = namespace
	return nil
}

func (c *Client) GetCurrentNamespace() string {
	return c.Namespace
}

func (c *Client) SetNamespace(ns string) {
	c.Namespace = ns
}

// WaitForServiceAccountInNamespace waits for the given service account to be ready
func (c *Client) WaitForServiceAccountInNamespace(namespace, serviceAccountName string) error {
	if namespace == "" || serviceAccountName == "" {
		return errors.New("namespace and serviceAccountName cannot be empty")
	}
	watcher, err := c.KubeClient.CoreV1().ServiceAccounts(namespace).Watch(context.TODO(), metav1.SingleObject(metav1.ObjectMeta{Name: serviceAccountName}))
	if err != nil {
		return err
	}

	timeout := time.After(getDefaultServiceAccTimeout)
	if watcher != nil {
		defer watcher.Stop()
		for {
			select {
			case val, ok := <-watcher.ResultChan():
				if !ok {
					break
				}
				if serviceAccount, ok := val.Object.(*corev1.ServiceAccount); ok {
					if serviceAccount.Name == serviceAccountName {
						klog.V(3).Infof("Status of creation of service account %s is ready", serviceAccount)
						return nil
					}
				}
			case <-timeout:
				return errors.New("Timed out waiting for service to be ready")
			}
		}
	}
	return nil
}
