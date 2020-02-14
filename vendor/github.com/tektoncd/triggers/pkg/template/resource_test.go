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

package template

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	"github.com/tektoncd/triggers/test"
	bldr "github.com/tektoncd/triggers/test/builder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_MergeInDefaultParams(t *testing.T) {
	var (
		oneParam = pipelinev1.Param{
			Name:  "oneid",
			Value: pipelinev1.ArrayOrString{StringVal: "onevalue"},
		}
		oneParamSpec = pipelinev1.ParamSpec{
			Name:    "oneid",
			Default: &pipelinev1.ArrayOrString{StringVal: "onedefault"},
		}
		wantDefaultOneParam = pipelinev1.Param{
			Name:  "oneid",
			Value: pipelinev1.ArrayOrString{StringVal: "onedefault"},
		}
		twoParamSpec = pipelinev1.ParamSpec{
			Name:    "twoid",
			Default: &pipelinev1.ArrayOrString{StringVal: "twodefault"},
		}
		wantDefaultTwoParam = pipelinev1.Param{
			Name:  "twoid",
			Value: pipelinev1.ArrayOrString{StringVal: "twodefault"},
		}
		threeParamSpec = pipelinev1.ParamSpec{
			Name:    "threeid",
			Default: &pipelinev1.ArrayOrString{StringVal: "threedefault"},
		}
		wantDefaultThreeParam = pipelinev1.Param{
			Name:  "threeid",
			Value: pipelinev1.ArrayOrString{StringVal: "threedefault"},
		}
		noDefaultParamSpec = pipelinev1.ParamSpec{
			Name: "nodefault",
		}
	)
	type args struct {
		params     []pipelinev1.Param
		paramSpecs []pipelinev1.ParamSpec
	}
	tests := []struct {
		name string
		args args
		want []pipelinev1.Param
	}{
		{
			name: "add one default param",
			args: args{
				params:     []pipelinev1.Param{},
				paramSpecs: []pipelinev1.ParamSpec{oneParamSpec},
			},
			want: []pipelinev1.Param{wantDefaultOneParam},
		},
		{
			name: "add multiple default params",
			args: args{
				params:     []pipelinev1.Param{},
				paramSpecs: []pipelinev1.ParamSpec{oneParamSpec, twoParamSpec, threeParamSpec},
			},
			want: []pipelinev1.Param{wantDefaultOneParam, wantDefaultTwoParam, wantDefaultThreeParam},
		},
		{
			name: "do not override existing value",
			args: args{
				params:     []pipelinev1.Param{oneParam},
				paramSpecs: []pipelinev1.ParamSpec{oneParamSpec},
			},
			want: []pipelinev1.Param{oneParam},
		},
		{
			name: "add no default params",
			args: args{
				params:     []pipelinev1.Param{},
				paramSpecs: []pipelinev1.ParamSpec{noDefaultParamSpec},
			},
			want: []pipelinev1.Param{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeInDefaultParams(tt.args.params, tt.args.paramSpecs)
			if diff := cmp.Diff(tt.want, got, cmpopts.SortSlices(test.CompareParams)); diff != "" {
				t.Errorf("MergeInDefaultParams(): -want +got: %s", diff)
			}
		})
	}
}

func Test_applyParamToResourceTemplate(t *testing.T) {
	var (
		oneParam = pipelinev1.Param{
			Name:  "oneid",
			Value: pipelinev1.ArrayOrString{StringVal: "onevalue"},
		}
		rtNoParamVars             = json.RawMessage(`{"foo": "bar"}`)
		wantRtNoParamVars         = json.RawMessage(`{"foo": "bar"}`)
		rtNoMatchingParamVars     = json.RawMessage(`{"foo": "$(params.no.matching.path)"}`)
		wantRtNoMatchingParamVars = json.RawMessage(`{"foo": "$(params.no.matching.path)"}`)
		rtOneParamVar             = json.RawMessage(`{"foo": "bar-$(params.oneid)-bar"}`)
		wantRtOneParamVar         = json.RawMessage(`{"foo": "bar-onevalue-bar"}`)
		rtMultipleParamVars       = json.RawMessage(`{"$(params.oneid)": "bar-$(params.oneid)-$(params.oneid)$(params.oneid)$(params.oneid)-$(params.oneid)-bar"}`)
		wantRtMultipleParamVars   = json.RawMessage(`{"onevalue": "bar-onevalue-onevalueonevalueonevalue-onevalue-bar"}`)
	)
	type args struct {
		param pipelinev1.Param
		rt    json.RawMessage
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "replace no param vars",
			args: args{
				param: oneParam,
				rt:    rtNoParamVars,
			},
			want: wantRtNoParamVars,
		},
		{
			name: "replace no param vars with non match present",
			args: args{
				param: oneParam,
				rt:    rtNoMatchingParamVars,
			},
			want: wantRtNoMatchingParamVars,
		},
		{
			name: "replace one param var",
			args: args{
				param: oneParam,
				rt:    rtOneParamVar,
			},
			want: wantRtOneParamVar,
		},
		{
			name: "replace multiple param vars",
			args: args{
				param: oneParam,
				rt:    rtMultipleParamVars,
			},
			want: wantRtMultipleParamVars,
		}, {
			name: "espcae quotes in param val",
			args: args{
				param: pipelinev1.Param{
					Name: "p1",
					Value: pipelinev1.ArrayOrString{
						StringVal: `{"a":"b"}`,
					},
				},
				rt: json.RawMessage(`{"foo": "$(params.p1)"}`),
			},
			want: json.RawMessage(`{"foo": "{\"a\":\"b\"}"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyParamToResourceTemplate(tt.args.param, tt.args.rt)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("applyParamToResourceTemplate(): -want +got: %s", diff)
			}
		})
	}
}

func Test_ApplyParamsToResourceTemplate(t *testing.T) {
	rt := json.RawMessage(`{"oneparam": "$(params.oneid)", "twoparam": "$(params.twoid)", "threeparam": "$(params.threeid)"`)
	type args struct {
		params []pipelinev1.Param
		rt     json.RawMessage
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "no params",
			args: args{
				params: []pipelinev1.Param{},
				rt:     rt,
			},
			want: rt,
		},
		{
			name: "one param",
			args: args{
				params: []pipelinev1.Param{
					{Name: "oneid", Value: pipelinev1.ArrayOrString{StringVal: "onevalue"}},
				},
				rt: rt,
			},
			want: json.RawMessage(`{"oneparam": "onevalue", "twoparam": "$(params.twoid)", "threeparam": "$(params.threeid)"`),
		},
		{
			name: "multiple params",
			args: args{
				params: []pipelinev1.Param{
					{Name: "oneid", Value: pipelinev1.ArrayOrString{StringVal: "onevalue"}},
					{Name: "twoid", Value: pipelinev1.ArrayOrString{StringVal: "twovalue"}},
					{Name: "threeid", Value: pipelinev1.ArrayOrString{StringVal: "threevalue"}},
				},
				rt: rt,
			},
			want: json.RawMessage(`{"oneparam": "onevalue", "twoparam": "twovalue", "threeparam": "threevalue"`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyParamsToResourceTemplate(tt.args.params, tt.args.rt)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ApplyParamsToResourceTemplate(): -want +got: %s", diff)
			}
		})
	}
}

var (
	triggerBindings = map[string]*triggersv1.TriggerBinding{
		"my-triggerbinding": {
			ObjectMeta: metav1.ObjectMeta{Name: "my-triggerbinding"},
		},
		"tb-params": {
			ObjectMeta: metav1.ObjectMeta{Name: "tb-params"},
			Spec: triggersv1.TriggerBindingSpec{
				Params: []pipelinev1.Param{{
					Name: "foo",
					Value: pipelinev1.ArrayOrString{
						Type:      pipelinev1.ParamTypeString,
						StringVal: "bar",
					},
				}},
			},
		},
	}
	tb = triggerBindings["my-triggerbinding"]
	tt = triggersv1.TriggerTemplate{
		ObjectMeta: metav1.ObjectMeta{Name: "my-triggertemplate"},
	}
	getTB = func(name string, options metav1.GetOptions) (*triggersv1.TriggerBinding, error) {
		if v, ok := triggerBindings[name]; ok {
			return v, nil
		}
		return nil, fmt.Errorf("error invalid name: %s", name)
	}
	getTT = func(name string, options metav1.GetOptions) (*triggersv1.TriggerTemplate, error) {
		if name == "my-triggertemplate" {
			return &tt, nil
		}
		return nil, fmt.Errorf("error invalid name: %s", name)
	}
)

func Test_ResolveTrigger(t *testing.T) {
	tests := []struct {
		name    string
		trigger triggersv1.EventListenerTrigger
		want    ResolvedTrigger
	}{
		{
			name:    "1 binding",
			trigger: bldr.Trigger("my-triggerbinding", "my-triggertemplate", "v1alpha1"),
			want: ResolvedTrigger{
				TriggerBindings: []*triggersv1.TriggerBinding{tb},
				TriggerTemplate: &tt,
			},
		},
		{
			name: "no binding",
			trigger: triggersv1.EventListenerTrigger{
				Template: triggersv1.EventListenerTemplate{
					Name:       "my-triggertemplate",
					APIVersion: "v1alpha1",
				},
			},
			want: ResolvedTrigger{TriggerBindings: []*triggersv1.TriggerBinding{}, TriggerTemplate: &tt},
		},
		{
			name: "multiple bindings",
			trigger: triggersv1.EventListenerTrigger{
				Bindings: []*triggersv1.EventListenerBinding{
					{
						Name:       "my-triggerbinding",
						APIVersion: "v1alpha1",
					},
					{
						Name:       "tb-params",
						APIVersion: "v1alpha1",
					},
				},
				Template: triggersv1.EventListenerTemplate{
					Name:       "my-triggertemplate",
					APIVersion: "v1alpha1",
				},
			},
			want: ResolvedTrigger{
				TriggerBindings: []*triggersv1.TriggerBinding{
					tb,
					triggerBindings["tb-params"],
				},
				TriggerTemplate: &tt,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolveTrigger(tc.trigger, getTB, getTT)
			if err != nil {
				t.Errorf("ResolveTrigger() returned unexpected error: %s", err)
			} else if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ResolveTrigger(): -want +got: %s", diff)
			}
		})
	}
}

func Test_ResolveTrigger_error(t *testing.T) {
	tests := []struct {
		name    string
		trigger triggersv1.EventListenerTrigger
		getTB   getTriggerBinding
		getTT   getTriggerTemplate
	}{
		{
			name:    "error triggerbinding",
			trigger: bldr.Trigger("invalid-tb-name", "my-triggertemplate", "v1alpha1"),
			getTB:   getTB,
			getTT:   getTT,
		},
		{
			name:    "error triggertemplate",
			trigger: bldr.Trigger("my-triggerbinding", "invalid-tt-name", "v1alpha1"),
			getTB:   getTB,
			getTT:   getTT,
		},
		{
			name:    "error triggerbinding and triggertemplate",
			trigger: bldr.Trigger("invalid-tb-name", "invalid-tt-name", "v1alpha1"),
			getTB:   getTB,
			getTT:   getTT,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResolveTrigger(tt.trigger, tt.getTB, tt.getTT)
			if err == nil {
				t.Error("ResolveTrigger() did not return error when expected")
			}
		})
	}
}

func Test_ApplyUIDToResourceTemplate(t *testing.T) {
	tests := []struct {
		name       string
		rt         json.RawMessage
		expectedRt json.RawMessage
	}{
		{
			name:       "No uid",
			rt:         json.RawMessage(`{"rt": "nothing to see here"}`),
			expectedRt: json.RawMessage(`{"rt": "nothing to see here"}`),
		},
		{
			name:       "One uid",
			rt:         json.RawMessage(`{"rt": "uid is $(uid)"}`),
			expectedRt: json.RawMessage(`{"rt": "uid is abcde"}`),
		},
		{
			name:       "Three uid",
			rt:         json.RawMessage(`{"rt1": "uid is $(uid)", "rt2": "nothing", "rt3": "$(uid)-center-$(uid)"}`),
			expectedRt: json.RawMessage(`{"rt1": "uid is abcde", "rt2": "nothing", "rt3": "abcde-center-abcde"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This seeds UID() to return 'abcde'
			UID = func() string { return "abcde" }
			actualRt := ApplyUIDToResourceTemplate(tt.rt, UID())
			if diff := cmp.Diff(string(tt.expectedRt), string(actualRt)); diff != "" {
				t.Errorf("ApplyUIDToResourceTemplate(): -want +got: %s", diff)
			}
		})
	}
}

func TestMergeBindingParams(t *testing.T) {
	tests := []struct {
		name     string
		bindings []*triggersv1.TriggerBinding
		want     []pipelinev1.Param
		wantErr  bool
	}{{
		name: "empty bindings",
		bindings: []*triggersv1.TriggerBinding{
			bldr.TriggerBinding("", "", bldr.TriggerBindingSpec()),
			bldr.TriggerBinding("", "", bldr.TriggerBindingSpec()),
		},
		want: []pipelinev1.Param{},
	}, {
		name: "single binding with multiple params",
		bindings: []*triggersv1.TriggerBinding{
			bldr.TriggerBinding("", "", bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("param1", "value1"),
				bldr.TriggerBindingParam("param2", "value2"),
			)),
		},
		want: []pipelinev1.Param{{
			Name:  "param1",
			Value: pipelinev1.ArrayOrString{StringVal: "value1", Type: pipelinev1.ParamTypeString},
		}, {
			Name:  "param2",
			Value: pipelinev1.ArrayOrString{StringVal: "value2", Type: pipelinev1.ParamTypeString},
		}},
	}, {
		name: "multiple bindings each with multiple params",
		bindings: []*triggersv1.TriggerBinding{
			bldr.TriggerBinding("", "", bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("param1", "value1"),
				bldr.TriggerBindingParam("param2", "value2"),
			)),
			bldr.TriggerBinding("", "", bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("param3", "value3"),
				bldr.TriggerBindingParam("param4", "value4"),
			)),
		},
		want: []pipelinev1.Param{{
			Name:  "param1",
			Value: pipelinev1.ArrayOrString{StringVal: "value1", Type: pipelinev1.ParamTypeString},
		}, {
			Name:  "param2",
			Value: pipelinev1.ArrayOrString{StringVal: "value2", Type: pipelinev1.ParamTypeString},
		}, {
			Name:  "param3",
			Value: pipelinev1.ArrayOrString{StringVal: "value3", Type: pipelinev1.ParamTypeString},
		}, {
			Name:  "param4",
			Value: pipelinev1.ArrayOrString{StringVal: "value4", Type: pipelinev1.ParamTypeString},
		}},
	}, {
		name: "multiple bindings with duplicate params",
		bindings: []*triggersv1.TriggerBinding{
			bldr.TriggerBinding("", "", bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("param1", "value1"),
			)),
			bldr.TriggerBinding("", "", bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("param1", "value3"),
				bldr.TriggerBindingParam("param4", "value4"),
			)),
		},
		wantErr: true,
	}, {
		name: "single binding with duplicate params",
		bindings: []*triggersv1.TriggerBinding{
			bldr.TriggerBinding("", "", bldr.TriggerBindingSpec(
				bldr.TriggerBindingParam("param1", "value1"),
				bldr.TriggerBindingParam("param1", "value3"),
			)),
		},
		wantErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MergeBindingParams(tt.bindings)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unexpected error : %q", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Unexpected output(-want +got): %s", diff)
			}
		})
	}
}
