package servicebindingrequest

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

func TestSBRControllerBuildSBRPredicate(t *testing.T) {
	// keep the predicate around
	pred := buildSBRPredicate(log.NewLog("test-log"))

	// the expected behavior is that every create event triggers a reconciliation
	t.Run("create", func(t *testing.T) {
		if got := pred.Create(event.CreateEvent{}); !got {
			t.Errorf("newSBRPredicate() = %v, want %v", got, true)
		}
	})

	// update exercises changes that should or not trigger the reconciliation
	t.Run("update", func(t *testing.T) {
		sbrA := &v1alpha1.ServiceBindingRequest{
			Spec: v1alpha1.ServiceBindingRequestSpec{
				BackingServiceSelector: &v1alpha1.BackingServiceSelector{
					GroupVersionKind: metav1.GroupVersionKind{Group: "test", Version: "v1alpha1", Kind: "TestHost"},
					ResourceRef:      "",
				},
			},
		}
		sbrB := &v1alpha1.ServiceBindingRequest{
			Spec: v1alpha1.ServiceBindingRequestSpec{
				BackingServiceSelector: &v1alpha1.BackingServiceSelector{
					GroupVersionKind: metav1.GroupVersionKind{Group: "test", Version: "v1", Kind: "TestHost"},
					ResourceRef:      "",
				},
			},
		}

		tests := []struct {
			name string
			want bool
			a    runtime.Object
			b    runtime.Object
		}{
			{name: "same-spec", want: false, a: sbrA, b: sbrA},
			{name: "changed-spec", want: true, a: sbrA, b: sbrB},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := pred.Update(event.UpdateEvent{ObjectOld: tt.a, ObjectNew: tt.b}); got != tt.want {
					t.Errorf("newSBRPredicate() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	// delete verifies that SBRs will be reconciled prior to its deletion
	t.Run("delete", func(t *testing.T) {
		tests := []struct {
			name           string
			want           bool
			confirmDeleted bool
		}{
			// FIXME: validate whether this is the behavior we want
			{name: "delete-not-confirmed", confirmDeleted: false, want: true},
			{name: "delete-confirmed", confirmDeleted: true, want: false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := pred.Delete(event.DeleteEvent{DeleteStateUnknown: tt.confirmDeleted}); got != tt.want {
					t.Errorf("newSBRPredicate() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}

func TestSBRControllerBuildGVKPredicate(t *testing.T) {
	pred := buildGVKPredicate(log.NewLog("test-log"))

	// update verifies whether only the accepted manifests trigger the reconciliation process
	t.Run("update", func(t *testing.T) {
		deploymentA := &appsv1.Deployment{}
		configMapA := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 1,
			},
		}
		configMapB := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 2,
			},
		}
		secretA := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 1,
			},
		}
		secretB := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 2,
			},
		}

		tests := []struct {
			name   string
			wanted bool
			a      runtime.Object
			b      runtime.Object
			metaA  metav1.Object
			metaB  metav1.Object
		}{
			{
				name:   "non supported update",
				wanted: false,
				a:      deploymentA,
				b:      deploymentA,
				metaA:  deploymentA.GetObjectMeta(),
				metaB:  deploymentA.GetObjectMeta(),
			},
			{
				name:   "supported update no changes",
				wanted: false,
				a:      configMapA,
				b:      configMapA,
				metaA:  configMapA.GetObjectMeta(),
				metaB:  configMapA.GetObjectMeta(),
			},
			{
				name:   "supported update no changes",
				wanted: false,
				a:      secretA,
				b:      secretA,
				metaA:  secretA.GetObjectMeta(),
				metaB:  secretA.GetObjectMeta(),
			},
			{
				name:   "supported update generation changed",
				wanted: true,
				a:      configMapA,
				b:      configMapB,
				metaA:  configMapA.GetObjectMeta(),
				metaB:  configMapB.GetObjectMeta(),
			},
			{
				name:   "supported update generation changed",
				wanted: true,
				a:      secretA,
				b:      secretB,
				metaA:  secretA.GetObjectMeta(),
				metaB:  secretB.GetObjectMeta(),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				e := event.UpdateEvent{
					MetaOld:   tt.metaA,
					MetaNew:   tt.metaB,
					ObjectOld: tt.a,
					ObjectNew: tt.b,
				}
				if got := pred.Update(e); got != tt.wanted {
					t.Errorf("newGVKPredicate() = %v, want %v", got, tt.wanted)
				}
			})
		}
	})
}
