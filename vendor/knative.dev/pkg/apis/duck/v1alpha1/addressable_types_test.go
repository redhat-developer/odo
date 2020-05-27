/*
Copyright 2019 The Knative Authors

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

package v1alpha1

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"knative.dev/pkg/apis"
	v1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/apis/duck/v1beta1"
)

func TestGetURL(t *testing.T) {
	tests := []struct {
		name string
		addr Addressable
		want *apis.URL
	}{{
		name: "just hostname",
		addr: Addressable{
			Hostname: "foo.com",
		},
		want: apis.HTTP("foo.com"),
	}, {
		name: "just url",
		addr: Addressable{
			Addressable: v1beta1.Addressable{
				URL: apis.HTTP("bar.com"),
			},
		},
		want: apis.HTTP("bar.com"),
	}, {
		name: "both fields",
		addr: Addressable{
			Hostname: "foo.bar.svc.cluster.local",
			Addressable: v1beta1.Addressable{
				URL: apis.HTTPS("baz.com"),
			},
		},
		want: apis.HTTPS("baz.com"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.addr.GetURL()
			if got.String() != test.want.String() {
				t.Errorf("GetURL() = %v, wanted %v", got, test.want)
			}
		})
	}
}

func TestConversion(t *testing.T) {
	tests := []struct {
		name        string
		addr        *Addressable
		conv        apis.Convertible
		want        string
		wantErrUp   bool
		wantErrDown bool
	}{{
		name: "v1",
		addr: &Addressable{
			Addressable: v1beta1.Addressable{
				URL: apis.HTTP("bar.com"),
			},
			Hostname: "bar.com",
		},
		conv:        &v1.Addressable{},
		wantErrUp:   false,
		wantErrDown: false,
	}, {
		name: "v1beta1",
		addr: &Addressable{
			Addressable: v1beta1.Addressable{
				URL: apis.HTTP("bar.com"),
			},
			Hostname: "bar.com",
		},
		conv:        &v1beta1.Addressable{},
		wantErrUp:   false,
		wantErrDown: false,
	}, {
		name: "v1alpha1",
		addr: &Addressable{
			Addressable: v1beta1.Addressable{
				URL: apis.HTTPS("bar.com"),
			},
		},
		conv:        &Addressable{},
		wantErrUp:   true,
		wantErrDown: true,
	}, {
		name:        "v1alpha1 - empty",
		addr:        &Addressable{},
		conv:        &Addressable{},
		wantErrUp:   true,
		wantErrDown: true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conv := test.conv
			if err := test.addr.ConvertTo(context.Background(), conv); err != nil {
				if !test.wantErrUp {
					t.Errorf("ConvertTo() = %v", err)
				}
			} else if test.wantErrUp {
				t.Errorf("ConvertTo() = %#v, wanted error", conv)
			}
			got := &Addressable{}
			if err := got.ConvertFrom(context.Background(), conv); err != nil {
				if !test.wantErrDown {
					t.Errorf("ConvertFrom() = %v", err)
				}
				return
			} else if test.wantErrDown {
				t.Errorf("ConvertFrom() = %#v, wanted error", conv)
				return
			}

			if diff := cmp.Diff(test.addr, got); diff != "" {
				t.Errorf("roundtrip (-want, +got) = %v", diff)
			}
		})
	}
}

func TestConvertTo(t *testing.T) {
	tests := []struct {
		name        string
		addr        *Addressable
		conv        apis.Convertible
		want        apis.Convertible
		wantErrUp   bool
		wantErrDown bool
	}{{
		name:        "empty to v1beta1",
		addr:        &Addressable{},
		conv:        &v1beta1.Addressable{},
		want:        &v1beta1.Addressable{},
		wantErrUp:   false,
		wantErrDown: false,
	}, {
		name: "to v1beta1",
		addr: &Addressable{
			Hostname: "bar.com",
		},
		conv: &v1beta1.Addressable{},
		want: &v1beta1.Addressable{
			URL: apis.HTTP("bar.com"),
		},
		wantErrUp:   false,
		wantErrDown: false,
	}, {
		name:        "empty to v1",
		addr:        &Addressable{},
		conv:        &v1.Addressable{},
		want:        &v1.Addressable{},
		wantErrUp:   false,
		wantErrDown: false,
	}, {
		name: "to v1",
		addr: &Addressable{
			Hostname: "bar.com",
		},
		conv: &v1.Addressable{},
		want: &v1.Addressable{
			URL: apis.HTTP("bar.com"),
		},
		wantErrUp:   false,
		wantErrDown: false,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.conv
			if err := test.addr.ConvertTo(context.Background(), got); err != nil {
				if !test.wantErrUp {
					t.Errorf("ConvertTo() = %v", err)
				}
			} else if test.wantErrUp {
				t.Errorf("ConvertTo() = %#v, wanted error", got)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("roundtrip (-want, +got) = %v", diff)
			}
		})
	}
}

func TestConvertFrom(t *testing.T) {
	tests := []struct {
		name string
		in   apis.Convertible
		want *Addressable
	}{{
		name: "v1beta1",
		in:   &v1beta1.Addressable{URL: apis.HTTP("foo.example.com")},
		want: &Addressable{Addressable: v1beta1.Addressable{URL: apis.HTTP("foo.example.com")}, Hostname: "foo.example.com"},
	}, {
		name: "v1",
		in:   &v1.Addressable{URL: apis.HTTP("bar.example.com")},
		want: &Addressable{Addressable: v1beta1.Addressable{URL: apis.HTTP("bar.example.com")}, Hostname: "bar.example.com"},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := &Addressable{}
			if err := got.ConvertFrom(context.Background(), test.in); err != nil {
				t.Errorf("ConvertFrom() = %v", err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("roundtrip (-want, +got) = %v", diff)
			}
		})
	}
}
