/*
Copyright 2020 The Knative Authors

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

package conversion

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	// injection
	_ "knative.dev/pkg/client/injection/apiextensions/informers/apiextensions/v1/customresourcedefinition/fake"
	_ "knative.dev/pkg/injection/clients/namespacedkube/informers/core/v1/secret/fake"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/resourcesemantics/conversion/internal"

	. "knative.dev/pkg/reconciler/testing"
)

var (
	webhookPath = "/convert"
	testGK      = schema.GroupKind{
		Group: internal.Group,
		Kind:  internal.Kind,
	}

	zygotes = map[string]ConvertibleObject{
		"v1":    &internal.V1Resource{},
		"v2":    &internal.V2Resource{},
		"v3":    &internal.V3Resource{},
		"error": &internal.ErrorResource{},
	}

	kinds = map[schema.GroupKind]GroupKindConversion{
		testGK: {
			DefinitionName: "resource.webhook.pkg.knative.dev",
			HubVersion:     "v1",
			Zygotes:        zygotes,
		},
	}

	rawOpt = cmp.Transformer("raw", func(res []runtime.RawExtension) []string {
		result := make([]string, 0, len(res))
		for _, re := range res {
			result = append(result, string(re.Raw))
		}
		return result
	})

	cmpOpts = []cmp.Option{
		rawOpt,
	}
)

func testAPIVersion(version string) string {
	return testGK.WithVersion(version).GroupVersion().String()
}

func TestWebhookPath(t *testing.T) {
	ctx, _ := SetupFakeContext(t)
	ctx = webhook.WithOptions(ctx, webhook.Options{
		SecretName: "webhook-secret",
	})

	controller := NewConversionController(ctx, "/some-path", nil, nil)
	conversion := controller.Reconciler.(webhook.ConversionController)

	if got, want := conversion.Path(), "/some-path"; got != want {
		t.Errorf("expected controller to return provided path got: %q, want: %q", got, want)
	}
}

func TestConversionToHub(t *testing.T) {
	ctx, conversion := newConversion(t)

	req := &apixv1.ConversionRequest{
		UID:               "some-uid",
		DesiredAPIVersion: testAPIVersion("v1"),
		Objects: []runtime.RawExtension{
			toRaw(t, internal.NewV2("bing")),
			toRaw(t, internal.NewV3("bang")),
		},
	}

	want := &apixv1.ConversionResponse{
		UID:    "some-uid",
		Result: metav1.Status{Status: metav1.StatusSuccess},
		ConvertedObjects: []runtime.RawExtension{
			toRaw(t, internal.NewV1("bing")),
			toRaw(t, internal.NewV1("bang")),
		},
	}

	got := conversion.Convert(ctx, req)
	if diff := cmp.Diff(want, got, cmpOpts...); diff != "" {
		t.Errorf("unexpected response: %s", diff)
	}
}

func TestConversionFromHub(t *testing.T) {
	tests := []struct {
		version string
		in      runtime.Object
		out     runtime.Object
	}{{
		version: "v2",
		in:      internal.NewV1("bing"),
		out:     internal.NewV2("bing"),
	}, {
		version: "v3",
		in:      internal.NewV1("bing"),
		out:     internal.NewV3("bing"),
	}}

	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			ctx, conversion := newConversion(t)
			req := &apixv1.ConversionRequest{
				UID:               "some-uid",
				DesiredAPIVersion: testAPIVersion(test.version),
				Objects: []runtime.RawExtension{
					toRaw(t, test.in),
				},
			}

			want := &apixv1.ConversionResponse{
				UID:    "some-uid",
				Result: metav1.Status{Status: metav1.StatusSuccess},
				ConvertedObjects: []runtime.RawExtension{
					toRaw(t, test.out),
				},
			}

			got := conversion.Convert(ctx, req)
			if diff := cmp.Diff(want, got, cmpOpts...); diff != "" {
				t.Errorf("unexpected response: %s", diff)
			}
		})
	}
}

func TestConversionThroughHub(t *testing.T) {
	tests := []struct {
		name    string
		version string
		in      runtime.Object
		out     runtime.Object
	}{{
		name:    "v3 to v2",
		version: "v2",
		in:      internal.NewV3("bing"),
		out:     internal.NewV2("bing"),
	}, {
		name:    "v2 to v3",
		version: "v3",
		in:      internal.NewV2("bang"),
		out:     internal.NewV3("bang"),
	}}

	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			ctx, conversion := newConversion(t)

			req := &apixv1.ConversionRequest{
				UID:               "some-uid",
				DesiredAPIVersion: testAPIVersion(test.version),
				Objects: []runtime.RawExtension{
					toRaw(t, test.in),
				},
			}

			want := &apixv1.ConversionResponse{
				UID:    "some-uid",
				Result: metav1.Status{Status: metav1.StatusSuccess},
				ConvertedObjects: []runtime.RawExtension{
					toRaw(t, test.out),
				},
			}

			got := conversion.Convert(ctx, req)
			if diff := cmp.Diff(want, got, cmpOpts...); diff != "" {
				t.Errorf("unexpected response: %s", diff)
			}
		})
	}
}

func TestConversionErrorBadGVK(t *testing.T) {
	tests := []struct {
		name string
		gvk  schema.GroupVersionKind
	}{{
		name: "empty group",
		gvk: schema.GroupVersionKind{
			Version: "v1",
			Kind:    "Resource",
		},
	}, {
		name: "empty version",
		gvk: schema.GroupVersionKind{
			Group: "webhook.pkg.knative.dev",
			Kind:  "Resource",
		},
	}, {
		name: "empty kind",
		gvk: schema.GroupVersionKind{
			Group:   "webhook.pkg.knative.dev",
			Version: "v1",
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			obj := internal.NewV2("bing")
			obj.SetGroupVersionKind(test.gvk)

			ctx, conversion := newConversion(t)

			req := &apixv1.ConversionRequest{
				UID:               "some-uid",
				DesiredAPIVersion: testAPIVersion("v1"),
				Objects: []runtime.RawExtension{
					toRaw(t, obj),
				},
			}

			want := &apixv1.ConversionResponse{
				UID: "some-uid",
				Result: metav1.Status{
					Status: metav1.StatusFailure,
				},
			}

			cmpOpts := []cmp.Option{
				cmpopts.IgnoreFields(metav1.Status{}, "Message"),
				cmpopts.EquateEmpty(),
			}

			got := conversion.Convert(ctx, req)
			if diff := cmp.Diff(want, got, cmpOpts...); diff != "" {
				t.Errorf("unexpected response: %s", diff)
			}

			if !strings.HasPrefix(got.Result.Message, "invalid GroupVersionKind") {
				t.Errorf("expected message to start with 'invalid GroupVersionKind' got %q", got.Result.Message)
			}
		})
	}
}

func TestConversionUnknownInputGVK(t *testing.T) {
	unknownObj := &unstructured.Unstructured{}
	unknownObj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "some.api.group.dev",
		Version: "v1",
		Kind:    "Resource",
	})

	ctx, conversion := newConversion(t)

	req := &apixv1.ConversionRequest{
		UID:               "some-uid",
		DesiredAPIVersion: testAPIVersion("v3"),
		Objects: []runtime.RawExtension{
			toRaw(t, unknownObj),
		},
	}

	want := &apixv1.ConversionResponse{
		UID: "some-uid",
		Result: metav1.Status{
			Message: "no conversion support for type [kind=Resource group=some.api.group.dev]",
			Status:  metav1.StatusFailure,
		},
	}

	got := conversion.Convert(ctx, req)
	if diff := cmp.Diff(want, got, cmpOpts...); diff != "" {
		t.Errorf("unexpected response: %s", diff)
	}
}

func TestConversionInvalidTypeMeta(t *testing.T) {
	ctx, conversion := newConversionWithKinds(t, nil)

	req := &apixv1.ConversionRequest{
		UID:               "some-uid",
		DesiredAPIVersion: "some-version",
		Objects: []runtime.RawExtension{
			{Raw: []byte("}")},
		},
	}

	want := &apixv1.ConversionResponse{
		UID: "some-uid",
		Result: metav1.Status{
			Status: metav1.StatusFailure,
		},
	}

	cmpOpts := []cmp.Option{
		cmpopts.IgnoreFields(metav1.Status{}, "Message"),
		cmpopts.EquateEmpty(),
	}

	got := conversion.Convert(ctx, req)
	if diff := cmp.Diff(want, got, cmpOpts...); diff != "" {
		t.Errorf("unexpected response: %s", diff)
	}

	if !strings.HasPrefix(got.Result.Message, "error parsing type meta") {
		t.Errorf("expected message to start with 'error parsing type meta' got %q", got.Result.Message)
	}
}

func TestConversionFailureToUnmarshalInput(t *testing.T) {
	ctx, conversion := newConversion(t)

	req := &apixv1.ConversionRequest{
		UID:               "some-uid",
		DesiredAPIVersion: testAPIVersion("v1"),
		Objects: []runtime.RawExtension{
			toRaw(t, internal.NewErrorResource(internal.ErrorUnmarshal)),
		},
	}

	want := &apixv1.ConversionResponse{
		UID: "some-uid",
		Result: metav1.Status{
			Status: metav1.StatusFailure,
		},
	}

	cmpOpts := []cmp.Option{
		cmpopts.IgnoreFields(metav1.Status{}, "Message"),
		cmpopts.EquateEmpty(),
	}

	got := conversion.Convert(ctx, req)
	if diff := cmp.Diff(want, got, cmpOpts...); diff != "" {
		t.Errorf("unexpected response: %s", diff)
	}

	if !strings.HasPrefix(got.Result.Message, "unable to unmarshal input") {
		t.Errorf("expected message to start with 'unable to unmarshal input' got %q", got.Result.Message)
	}
}

func TestConversionFailureToMarshalOutput(t *testing.T) {
	ctx, conversion := newConversion(t)

	req := &apixv1.ConversionRequest{
		UID:               "some-uid",
		DesiredAPIVersion: testAPIVersion("error"),
		Objects: []runtime.RawExtension{
			// This property should make the Marshal on the
			// ErrorResource to fail
			toRaw(t, internal.NewV1(internal.ErrorMarshal)),
		},
	}

	want := &apixv1.ConversionResponse{
		UID: "some-uid",
		Result: metav1.Status{
			Status: metav1.StatusFailure,
		},
	}

	cmpOpts := []cmp.Option{
		cmpopts.IgnoreFields(metav1.Status{}, "Message"),
		cmpopts.EquateEmpty(),
	}

	got := conversion.Convert(ctx, req)
	if diff := cmp.Diff(want, got, cmpOpts...); diff != "" {
		t.Errorf("unexpected response: %s", diff)
	}

	if !strings.HasPrefix(got.Result.Message, "unable to marshal output") {
		t.Errorf("expected message to start with 'unable to marshal output' got %q", got.Result.Message)
	}
}

func TestConversionFailureToConvert(t *testing.T) {
	// v1 => error resource => v3
	kinds := map[schema.GroupKind]GroupKindConversion{
		testGK: {
			DefinitionName: "resource.webhook.pkg.knative.dev",
			HubVersion:     "error",
			Zygotes:        zygotes,
		},
	}

	tests := []struct {
		name    string
		errorOn string
	}{{
		name:    "error converting from",
		errorOn: internal.ErrorConvertFrom,
	}, {
		name:    "error converting to",
		errorOn: internal.ErrorConvertTo,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, conversion := newConversionWithKinds(t, kinds)
			req := &apixv1.ConversionRequest{
				UID:               "some-uid",
				DesiredAPIVersion: testAPIVersion("v3"),
				Objects: []runtime.RawExtension{
					// Insert failure here
					toRaw(t, internal.NewV1(test.errorOn)),
				},
			}

			want := &apixv1.ConversionResponse{
				UID: "some-uid",
				Result: metav1.Status{
					Status: metav1.StatusFailure,
				},
			}

			cmpOpts := []cmp.Option{
				cmpopts.IgnoreFields(metav1.Status{}, "Message"),
				rawOpt,
			}

			got := conversion.Convert(ctx, req)
			if diff := cmp.Diff(want, got, cmpOpts...); diff != "" {
				t.Errorf("unexpected response: %s", diff)
			}

			if !strings.HasPrefix(got.Result.Message, "conversion failed") {
				t.Errorf("expected message to start with 'conversion failed' got %q", got.Result.Message)
			}
		})
	}

}

func TestConversionFailureInvalidDesiredAPIVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{{
		name:    "multiple path segments",
		version: "bad-api-version/v1/v2",
	}, {
		name:    "empty",
		version: "",
	}, {
		name:    "path segment",
		version: "/",
	}, {
		name:    "no version",
		version: "some.api.group",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, conversion := newConversion(t)

			req := &apixv1.ConversionRequest{
				UID:               "some-uid",
				DesiredAPIVersion: test.version,
				Objects: []runtime.RawExtension{
					toRaw(t, internal.NewV1("bing")),
				},
			}

			want := &apixv1.ConversionResponse{
				UID: "some-uid",
				Result: metav1.Status{
					Message: fmt.Sprintf("desired API version %q is not valid", test.version),
					Status:  metav1.StatusFailure,
				},
			}

			got := conversion.Convert(ctx, req)
			if diff := cmp.Diff(want, got, cmpOpts...); diff != "" {
				t.Errorf("unexpected response: %s", diff)
			}
		})
	}
}

func TestConversionMissingZygotes(t *testing.T) {
	// Assume we're converting from
	// v2 (input)  => v1 (hub) => v3 (output)
	tests := []struct {
		name    string
		zygotes map[string]ConvertibleObject
	}{{
		name: "missing input",
		zygotes: map[string]ConvertibleObject{
			"v1": &internal.V1Resource{},
			"v3": &internal.V3Resource{},
		},
	}, {
		name: "missing output",
		zygotes: map[string]ConvertibleObject{
			"v1": &internal.V1Resource{},
			"v2": &internal.V2Resource{},
		},
	}, {
		name: "missing hub",
		zygotes: map[string]ConvertibleObject{
			"v2": &internal.V2Resource{},
			"v3": &internal.V3Resource{},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			kinds = map[schema.GroupKind]GroupKindConversion{
				testGK: {
					DefinitionName: "resource.webhook.pkg.knative.dev",
					HubVersion:     "v1",
					Zygotes:        test.zygotes,
				},
			}

			ctx, conversion := newConversionWithKinds(t, kinds)

			req := &apixv1.ConversionRequest{
				UID:               "some-uid",
				DesiredAPIVersion: testAPIVersion("v3"),
				Objects: []runtime.RawExtension{
					toRaw(t, internal.NewV2("bing")),
				},
			}

			want := &apixv1.ConversionResponse{
				UID: "some-uid",
				Result: metav1.Status{
					Status: metav1.StatusFailure,
				},
			}

			cmpOpts := []cmp.Option{
				cmpopts.IgnoreFields(metav1.Status{}, "Message"),
				cmpopts.EquateEmpty(),
			}

			got := conversion.Convert(ctx, req)
			if diff := cmp.Diff(want, got, cmpOpts...); diff != "" {
				t.Errorf("unexpected response: %s", diff)
			}

			if !strings.HasPrefix(got.Result.Message, "conversion not supported") {
				t.Errorf("expected message to start with 'conversion not supported' got %q", got.Result.Message)
			}
		})
	}
}

func TestContextDecoration(t *testing.T) {
	ctx, _ := SetupFakeContext(t)
	ctx = webhook.WithOptions(ctx, webhook.Options{
		SecretName: "webhook-secret",
	})

	decoratorCalled := false
	decorator := func(ctx context.Context) context.Context {
		decoratorCalled = true
		return ctx
	}

	controller := NewConversionController(ctx, webhookPath, kinds, decorator)
	r := controller.Reconciler.(*reconciler)
	r.Convert(ctx, &apixv1.ConversionRequest{})

	if !decoratorCalled {
		t.Errorf("context decorator was not invoked")
	}
}

func toRaw(t *testing.T, obj runtime.Object) runtime.RawExtension {
	t.Helper()

	raw, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("unable to marshal resource: %s", err)
	}

	return runtime.RawExtension{Raw: raw}
}

func newConversion(t *testing.T) (context.Context, webhook.ConversionController) {
	return newConversionWithKinds(t, kinds)
}

func newConversionWithKinds(
	t *testing.T,
	kinds map[schema.GroupKind]GroupKindConversion,
) (
	context.Context,
	webhook.ConversionController,
) {

	ctx, _ := SetupFakeContext(t)
	ctx = webhook.WithOptions(ctx, webhook.Options{
		SecretName: "webhook-secret",
	})

	controller := NewConversionController(ctx, webhookPath, kinds, nil)
	return ctx, controller.Reconciler.(*reconciler)
}
