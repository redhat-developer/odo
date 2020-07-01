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

package cel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	celext "github.com/google/cel-go/ext"
	"github.com/tektoncd/triggers/pkg/interceptors"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"

	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

// Interceptor implements a CEL based interceptor that uses CEL expressions
// against the incoming body and headers to match, if the expression returns
// a true value, then the interception is "successful".
type Interceptor struct {
	KubeClientSet          kubernetes.Interface
	Logger                 *zap.SugaredLogger
	CEL                    *triggersv1.CELInterceptor
	EventListenerNamespace string
}

// NewInterceptor creates a prepopulated Interceptor.
func NewInterceptor(cel *triggersv1.CELInterceptor, k kubernetes.Interface, ns string, l *zap.SugaredLogger) interceptors.Interceptor {
	return &Interceptor{
		Logger:                 l,
		CEL:                    cel,
		KubeClientSet:          k,
		EventListenerNamespace: ns,
	}
}

// ExecuteTrigger is an implementation of the Interceptor interface.
func (w *Interceptor) ExecuteTrigger(request *http.Request) (*http.Response, error) {
	env, err := makeCelEnv(w.EventListenerNamespace, w.KubeClientSet)
	if err != nil {
		return nil, fmt.Errorf("error creating cel environment: %w", err)
	}

	var payload = []byte(`{}`)
	if request.Body != nil {
		defer request.Body.Close()
		payload, err = ioutil.ReadAll(request.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading request body: %w", err)
		}
	}

	evalContext, err := makeEvalContext(payload, request)
	if err != nil {
		return nil, fmt.Errorf("error making the evaluation context: %w", err)
	}

	if w.CEL.Filter != "" {
		out, err := evaluate(w.CEL.Filter, env, evalContext)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate expression '%s': %w", w.CEL.Filter, err)
		}

		if out != types.True {
			return nil, fmt.Errorf("expression %s did not return true", w.CEL.Filter)
		}
	}

	for _, u := range w.CEL.Overlays {
		val, err := evaluate(u.Expression, env, evalContext)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate overlay expression '%s': %w", u.Expression, err)
		}

		var raw interface{}
		var b []byte

		switch val.(type) {
		case types.String:
			raw, err = val.ConvertToNative(reflect.TypeOf(&structpb.Value{}))
			if err == nil {
				b, err = json.Marshal(raw.(*structpb.Value).GetStringValue())
			}
		case types.Double, types.Int:
			raw, err = val.ConvertToNative(reflect.TypeOf(&structpb.Value{}))
			if err == nil {
				b, err = json.Marshal(raw.(*structpb.Value).GetNumberValue())
			}

		default:
			raw, err = val.ConvertToNative(reflect.TypeOf([]byte{}))
			b = raw.([]byte)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to convert overlay result to bytes: %w", err)
		}

		payload, err = sjson.SetRawBytes(payload, u.Key, b)
		if err != nil {
			return nil, fmt.Errorf("failed to sjson for key '%s' to '%s': %w", u.Key, val, err)
		}
	}

	return &http.Response{
		Header: request.Header,
		Body:   ioutil.NopCloser(bytes.NewBuffer(payload)),
	}, nil

}

func evaluate(expr string, env *cel.Env, data map[string]interface{}) (ref.Val, error) {
	parsed, issues := env.Parse(expr)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	checked, issues := env.Check(parsed)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	prg, err := env.Program(checked)
	if err != nil {
		return nil, err
	}

	out, _, err := prg.Eval(data)
	return out, err
}

func makeCelEnv(ns string, k kubernetes.Interface) (*cel.Env, error) {
	mapStrDyn := decls.NewMapType(decls.String, decls.Dyn)
	return cel.NewEnv(
		Triggers(ns, k),
		celext.Strings(),
		cel.Declarations(
			decls.NewIdent("body", mapStrDyn, nil),
			decls.NewIdent("header", mapStrDyn, nil)))
}

func makeEvalContext(body []byte, r *http.Request) (map[string]interface{}, error) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal(body, &jsonMap)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"body": jsonMap, "header": r.Header}, nil
}
