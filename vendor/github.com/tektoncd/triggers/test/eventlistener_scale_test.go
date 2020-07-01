// +build e2e

/*
Copyright 2019 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package test

import (
	"fmt"
	"testing"

	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	bldr "github.com/tektoncd/triggers/test/builder"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knativetest "knative.dev/pkg/test"
)

/*
 * Test creating an EventListener with a large number of Triggers.
 * This is a regression test for issue #356.
 */
func TestEventListenerScale(t *testing.T) {
	c, namespace := setup(t)

	defer tearDown(t, c, namespace)
	knativetest.CleanupOnInterrupt(func() { tearDown(t, c, namespace) }, t.Logf)

	t.Log("Start EventListener Scale e2e test")

	saName := "scale-sa"
	createServiceAccount(t, c, namespace, saName)
	// Create an EventListener with 1000 Triggers
	var err error
	el := bldr.EventListener("my-eventlistener", namespace, bldr.EventListenerSpec(
		bldr.EventListenerServiceAccount(saName)),
	)

	for i := 0; i < 1000; i++ {
		trigger := bldr.Trigger("my-triggertemplate", "v1alpha1",
			bldr.EventListenerTriggerBinding("my-triggerbinding", "", "my-triggerbinding", "v1alpha1"),
		)
		trigger.Name = fmt.Sprintf("%d", i)
		el.Spec.Triggers = append(el.Spec.Triggers, trigger)
	}
	el, err = c.TriggersClient.TriggersV1alpha1().EventListeners(namespace).Create(el)
	if err != nil {
		t.Fatalf("Error creating EventListener: %s", err)
	}

	// Verify that the EventListener was created properly
	if err := WaitFor(eventListenerReady(t, c, namespace, el.Name)); err != nil {
		t.Fatalf("EventListener is not ready: %s", err)
	}
	t.Log("EventListener is ready")
}

func createServiceAccount(t *testing.T, c *clients, namespace, name string) {
	t.Helper()
	sa, err := c.KubeClient.CoreV1().ServiceAccounts(namespace).Create(
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: name},
		},
	)
	if err != nil {
		t.Fatalf("Error creating ServiceAccount: %s", err)
	}
	_, err = c.KubeClient.RbacV1().Roles(namespace).Create(
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{Name: "sa-role"},
			Rules: []rbacv1.PolicyRule{{
				APIGroups: []string{triggersv1.GroupName},
				Resources: []string{"eventlisteners", "triggerbindings", "triggertemplates"},
				Verbs:     []string{"get"},
			}, {
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"get", "list", "watch"},
			},
			},
		},
	)
	if err != nil {
		t.Fatalf("Error creating Role: %s", err)
	}
	_, err = c.KubeClient.RbacV1().RoleBindings(namespace).Create(
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: "sa-rolebinding"},
			Subjects: []rbacv1.Subject{{
				Kind:      "ServiceAccount",
				Name:      sa.Name,
				Namespace: namespace,
			}},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     "sa-role",
			},
		},
	)
	if err != nil {
		t.Fatalf("Error creating RoleBinding: %s", err)
	}

}
