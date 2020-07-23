package mocks

import (
	"github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func etcdClusterMock(ns, name string) *v1beta2.EtcdCluster {
	return &v1beta2.EtcdCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "etcd.database.coreos.com/v1beta2",
			APIVersion: "EtcdCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			UID:       "1234567",
		},
		Spec: v1beta2.ClusterSpec{
			Version: "3.2.13",
			Size:    5,
		},
	}
}

func etcdClusterServiceMock(ns, name string) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "172.30.0.129",
			Ports: []corev1.ServicePort{
				{
					Name:       "tcp-1",
					Protocol:   "TCP",
					Port:       33411,
					TargetPort: intstr.IntOrString{},
				},
			},
		},
	}
}

// CreateEtcdClusterMock returns all the resources required to setup an etcd cluster
// using etcd-operator.
// It creates following resources.
// 1. EtcdCluster resource.
// 2. Service(this gets created in etcd reconcile loop).
func CreateEtcdClusterMock(ns, name string) (*v1beta2.EtcdCluster, *corev1.Service) {
	ec := etcdClusterMock(ns, name)
	sv := etcdClusterServiceMock(ns, name)
	return ec, sv
}
