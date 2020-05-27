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

package apis

import (
	"context"
	"errors"
	"testing"
)

func TestConvertToViaProxy(t *testing.T) {
	sink := &testResource{}
	proxy := &testResource{}
	source := &testResource{proxy: proxy}

	err := ConvertToViaProxy(context.Background(), source, proxy, sink)

	if err != nil {
		t.Errorf("ConvertToViaProxy returned unexpected err: %s", err)
	}

	if source.to != proxy {
		t.Errorf("expected source to be converted to the proxy")
	}

	if proxy.to != sink {
		t.Errorf("expected proxy to be converted to the sink")
	}
}

func TestConvertToViaProxyError(t *testing.T) {
	tests := []struct {
		name          string
		source, proxy testResource
	}{{
		name: "converting source to proxy fails",
		source: testResource{
			err: errors.New("converting up failed"),
		},
		proxy: testResource{},
	}, {
		name:   "converting proxy to sink fails",
		source: testResource{},
		proxy: testResource{
			err: errors.New("converting up failed"),
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ConvertToViaProxy(context.Background(),
				&test.source,
				&test.proxy,
				nil, /* sink */
			)

			if err == nil {
				t.Errorf("expected error to have occurred")
			}
		})
	}
}

func TestConvertFromViaProxy(t *testing.T) {
	proxy := &testResource{}
	sink := &testResource{}
	source := &testResource{}

	err := ConvertFromViaProxy(context.Background(), source, proxy, sink)

	if err != nil {
		t.Errorf("ConvertFromViaProxy returned unexpected err: %s", err)
	}

	if proxy.from != source {
		t.Errorf("expected proxy to be converted from the source")
	}

	if sink.from != proxy {
		t.Errorf("expected sink to be converted from the proxy")
	}
}

func TestConvertFromViaProxyError(t *testing.T) {
	tests := []struct {
		name        string
		sink, proxy testResource
	}{{
		name: "converting proxy from source fails",
		sink: testResource{},
		proxy: testResource{
			err: errors.New("converting down failed"),
		},
	}, {
		name: "converting sink from proxy fails",
		sink: testResource{
			err: errors.New("converting down failed"),
		},
		proxy: testResource{},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			err := ConvertFromViaProxy(context.Background(),
				nil, /* source */
				&test.proxy,
				&test.sink,
			)
			if err == nil {
				t.Errorf("expected error to have occurred")
			}
		})
	}
}

type testResource struct {
	proxy, to, from Convertible
	err             error
}

var _ Convertible = (*testResource)(nil)

func (r *testResource) ConvertTo(ctx context.Context, to Convertible) error {
	r.to = to
	return r.err
}

func (r *testResource) ConvertFrom(ctx context.Context, from Convertible) error {
	r.from = from
	return r.err
}
