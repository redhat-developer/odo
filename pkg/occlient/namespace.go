package occlient

import (
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"
)

// CreateNamespace creates new namespace
func (c *Client) CreateNamespace(name string) (*corev1.Namespace, error) {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	newNamespace, err := c.kubeClient.CoreV1().Namespaces().Create(namespace)
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
		watcher, err = c.kubeClient.CoreV1().Namespaces().Watch(metav1.ListOptions{
			FieldSelector: fields.Set{"metadata.name": name}.AsSelector().String(),
		})
		if err != nil {
			return errors.Wrapf(err, "unable to watch namespace")
		}
		defer watcher.Stop()
	}

	err = c.kubeClient.CoreV1().Namespaces().Delete(name, &metav1.DeleteOptions{})
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
				klog.V(4).Infof("Watch event.Type '%s'.", val.Type)

				if namespaceStatus, ok := val.Object.(*corev1.Namespace); ok {
					klog.V(4).Infof("Status of delete of namespace %s is '%s'.", name, namespaceStatus.Status.Phase)
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
			klog.V(4).Infof("Namespace %s deleted", val.Name)
			return nil
		case err := <-watchErrorChannel:
			return err
		case <-time.After(waitForProjectDeletionTimeOut):
			return errors.Errorf("waited %s but couldn't delete namespace %s in time", waitForProjectDeletionTimeOut, name)
		}

	}
	return nil
}
