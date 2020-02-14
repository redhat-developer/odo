package v1alpha1_test

import (
	"context"
	"testing"

	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	bldr "github.com/tektoncd/triggers/test/builder"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_EventListenerValidate(t *testing.T) {
	tests := []struct {
		name string
		el   *v1alpha1.EventListener
	}{{
		name: "TriggerTemplate Does Not Exist",
		el: bldr.EventListener("name", "namespace",
			bldr.EventListenerSpec(
				bldr.EventListenerTrigger("tb", "dne", "v1alpha1"))),
	}, {
		name: "Valid EventListener No TriggerBinding",
		el: bldr.EventListener("name", "namespace",
			bldr.EventListenerSpec(
				bldr.EventListenerTrigger("", "tt", "v1alpha1"))),
	}, {
		name: "Valid EventListener No Interceptor",
		el: bldr.EventListener("name", "namespace",
			bldr.EventListenerSpec(
				bldr.EventListenerTrigger("tb", "tt", "v1alpha1"))),
	}, {
		name: "Valid EventListener Interceptor",
		el: bldr.EventListener("name", "namespace",
			bldr.EventListenerSpec(
				bldr.EventListenerTrigger("tb", "tt", "v1alpha1",
					bldr.EventListenerTriggerInterceptor("svc", "v1", "Service", "namespace")))),
	}, {
		name: "Valid EventListener Interceptor With Header",
		el: bldr.EventListener("name", "namespace",
			bldr.EventListenerSpec(
				bldr.EventListenerTrigger("tb", "tt", "v1alpha1",
					bldr.EventListenerTriggerInterceptor("svc", "v1", "Service", "namespace",
						bldr.EventInterceptorParam("Valid-Header-Key", "valid value"))))),
	}, {
		name: "Valid EventListener Interceptor With Headers",
		el: bldr.EventListener("name", "namespace",
			bldr.EventListenerSpec(
				bldr.EventListenerTrigger("tb", "tt", "v1alpha1",
					bldr.EventListenerTriggerInterceptor("svc", "v1", "Service", "namespace",
						bldr.EventInterceptorParam("Valid-Header-Key1", "valid value1"),
						bldr.EventInterceptorParam("Valid-Header-Key1", "valid value2"),
						bldr.EventInterceptorParam("Valid-Header-Key2", "valid value"))))),
	}, {
		name: "Valid EventListener Two Triggers",
		el: bldr.EventListener("name", "namespace",
			bldr.EventListenerSpec(
				bldr.EventListenerTrigger("tb", "tt", "v1alpha1",
					bldr.EventListenerTriggerInterceptor("svc", "v1", "Service", "namespace"),
				),
				bldr.EventListenerTrigger("tb", "tt", "v1alpha1"))),
	}, {
		name: "Valid EventListener with CEL interceptor",
		el: bldr.EventListener("name", "namespace",
			bldr.EventListenerSpec(
				bldr.EventListenerTrigger("tb", "tt", "v1alpha1",
					bldr.EventListenerCELInterceptor("body.value == 'test'")))),
	}, {
		name: "Valid EventListener with no trigger name",
		el: bldr.EventListener("name", "namespace",
			bldr.EventListenerSpec(
				bldr.EventListenerTrigger("tb", "tt", "v1alpha1"))),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.el.Validate(context.Background())
			if err != nil {
				t.Errorf("EventListener.Validate() expected no error, but got one, EventListener: %v, error: %v", test.el, err)
			}
		})
	}
}

func TestEventListenerValidate_error(t *testing.T) {
	tests := []struct {
		name string
		el   *v1alpha1.EventListener
	}{{
		name: "no triggers",
		el: &v1alpha1.EventListener{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "n",
				Namespace: "namespace",
			},
			Spec: v1alpha1.EventListenerSpec{
				Triggers: nil,
			},
		},
	}, {
		name: "Binding missing name",
		el: &v1alpha1.EventListener{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: v1alpha1.EventListenerSpec{
				Triggers: []v1alpha1.EventListenerTrigger{{
					Bindings: []*v1alpha1.EventListenerBinding{{Name: ""}},
					Template: v1alpha1.EventListenerTemplate{Name: "tt"},
				}},
			},
		},
	}, {
		name: "Both Binding and Bindings Present",
		el: &v1alpha1.EventListener{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: v1alpha1.EventListenerSpec{
				Triggers: []v1alpha1.EventListenerTrigger{{
					Bindings:          []*v1alpha1.EventListenerBinding{{Name: "tb"}},
					DeprecatedBinding: &v1alpha1.EventListenerBinding{Name: "bar"},
					Template:          v1alpha1.EventListenerTemplate{Name: "tt"},
				}},
			},
		},
	}, {
		name: "Template with wrong apiVersion",
		el: &v1alpha1.EventListener{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: v1alpha1.EventListenerSpec{
				Triggers: []v1alpha1.EventListenerTrigger{{
					Bindings: []*v1alpha1.EventListenerBinding{{Name: "tb"}},
					Template: v1alpha1.EventListenerTemplate{Name: "tt", APIVersion: "invalid"},
				}},
			},
		},
	}, {
		name: "Template with missing name",
		el: &v1alpha1.EventListener{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: v1alpha1.EventListenerSpec{
				Triggers: []v1alpha1.EventListenerTrigger{{
					Bindings: []*v1alpha1.EventListenerBinding{{Name: "tb"}},
					Template: v1alpha1.EventListenerTemplate{Name: ""},
				}},
			},
		},
	}, {
		name: "Interceptor Name only",
		el: bldr.EventListener("name", "namespace",
			bldr.EventListenerSpec(
				bldr.EventListenerTrigger("tb", "tt", "v1alpha1",
					bldr.EventListenerTriggerInterceptor("svc", "", "", "")))),
	}, {
		name: "Interceptor Missing ObjectRef",
		el: &v1alpha1.EventListener{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: v1alpha1.EventListenerSpec{
				Triggers: []v1alpha1.EventListenerTrigger{{
					Bindings:     []*v1alpha1.EventListenerBinding{{Name: "tb"}},
					Template:     v1alpha1.EventListenerTemplate{Name: "tt"},
					Interceptors: []*v1alpha1.EventInterceptor{{}},
				}},
			},
		},
	}, {
		name: "Interceptor Empty ObjectRef",
		el: &v1alpha1.EventListener{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: v1alpha1.EventListenerSpec{
				Triggers: []v1alpha1.EventListenerTrigger{{
					Bindings: []*v1alpha1.EventListenerBinding{{Name: "tb"}},
					Template: v1alpha1.EventListenerTemplate{Name: "tt"},
					Interceptors: []*v1alpha1.EventInterceptor{{
						Webhook: &v1alpha1.WebhookInterceptor{
							ObjectRef: &corev1.ObjectReference{
								Name: "",
							},
						},
					}},
				}},
			},
		},
	}, {
		name: "Interceptor Wrong APIVersion",
		el: bldr.EventListener("name", "namespace",
			bldr.EventListenerSpec(
				bldr.EventListenerTrigger("tb", "tt", "v1alpha1",
					bldr.EventListenerTriggerInterceptor("foo", "v3", "Service", "")))),
	}, {
		name: "Interceptor Wrong Kind",
		el: bldr.EventListener("name", "namespace",
			bldr.EventListenerSpec(
				bldr.EventListenerTrigger("tb", "tt", "v1alpha1",
					bldr.EventListenerTriggerInterceptor("foo", "v1", "Deployment", "")))),
	}, {
		name: "Interceptor Non-Canonical Header",
		el: bldr.EventListener("name", "namespace",
			bldr.EventListenerSpec(
				bldr.EventListenerTrigger("tb", "tt", "v1alpha1",
					bldr.EventListenerTriggerInterceptor("foo", "v1", "Deployment", "",
						bldr.EventInterceptorParam("non-canonical-header-key", "valid value"))))),
	}, {
		name: "Interceptor Empty Header Name",
		el: bldr.EventListener("name", "namespace",
			bldr.EventListenerSpec(
				bldr.EventListenerTrigger("tb", "tt", "v1alpha1",
					bldr.EventListenerTriggerInterceptor("foo", "v1", "Deployment", "",
						bldr.EventInterceptorParam("", "valid value"))))),
	}, {
		name: "Interceptor Empty Header Value",
		el: bldr.EventListener("name", "namespace",
			bldr.EventListenerSpec(
				bldr.EventListenerTrigger("tb", "tt", "v1alpha1",
					bldr.EventListenerTriggerInterceptor("foo", "v1", "Deployment", "",
						bldr.EventInterceptorParam("Valid-Header-Key", ""))))),
	}, {
		name: "Multiple interceptors set",
		el: &v1alpha1.EventListener{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: v1alpha1.EventListenerSpec{
				Triggers: []v1alpha1.EventListenerTrigger{{
					Bindings: []*v1alpha1.EventListenerBinding{{Name: "tb"}},
					Template: v1alpha1.EventListenerTemplate{Name: "tt"},
					Interceptors: []*v1alpha1.EventInterceptor{{
						GitHub: &v1alpha1.GitHubInterceptor{},
						GitLab: &v1alpha1.GitLabInterceptor{},
					}},
				}},
			},
		},
	}, {
		name: "Both Interceptor and Interceptors Present",
		el: &v1alpha1.EventListener{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: v1alpha1.EventListenerSpec{
				Triggers: []v1alpha1.EventListenerTrigger{{
					Bindings:              []*v1alpha1.EventListenerBinding{{Name: "tb"}},
					Template:              v1alpha1.EventListenerTemplate{Name: "tt"},
					DeprecatedInterceptor: &v1alpha1.EventInterceptor{},
					Interceptors:          []*v1alpha1.EventInterceptor{{}, {}},
				}},
			},
		},
	}, {
		name: "CEL interceptor with no filter",
		el: &v1alpha1.EventListener{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: v1alpha1.EventListenerSpec{
				Triggers: []v1alpha1.EventListenerTrigger{{
					Bindings: []*v1alpha1.EventListenerBinding{{Name: "tb"}},
					Template: v1alpha1.EventListenerTemplate{Name: "tt"},
					Interceptors: []*v1alpha1.EventInterceptor{{
						CEL: &v1alpha1.CELInterceptor{},
					}},
				}},
			},
		},
	}, {
		name: "Triggers name has invalid label characters",
		el: bldr.EventListener("name", "namespace",
			bldr.EventListenerSpec(
				bldr.EventListenerTrigger("tb", "tt", "v1alpha1",
					bldr.EventListenerTriggerName("github.com/tektoncd/triggers")))),
	}, {
		name: "Triggers name is longer than the allowable label value (63 characters)",
		el: bldr.EventListener("name", "namespace",
			bldr.EventListenerSpec(
				bldr.EventListenerTrigger("tb", "tt", "v1alpha1",
					bldr.EventListenerTriggerName("1234567890123456789012345678901234567890123456789012345678901234")))),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.el.Validate(context.Background())
			if err == nil {
				t.Errorf("EventListener.Validate() expected error, but get none, EventListener: %v", test.el)
			}
		})
	}
}
