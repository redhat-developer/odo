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

package kvstore

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	clientgotesting "k8s.io/client-go/testing"
)

const (
	namespace = "mynamespace"
	name      = "mycm"
)

type testStruct struct {
	LastThingProcessed string
	Stuff              []string
}

type testClient struct {
	created   *corev1.ConfigMap
	updated   *corev1.ConfigMap
	clientset v1.CoreV1Interface
}

func NewTestClient(objects ...runtime.Object) *testClient {
	tc := testClient{}
	cs := fake.NewSimpleClientset(objects...)
	cs.PrependReactor("create", "*", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
		createAction := action.(clientgotesting.CreateAction)
		tc.created = createAction.GetObject().(*corev1.ConfigMap)
		return true, tc.created, nil
	})
	cs.PrependReactor("update", "*", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
		updateAction := action.(clientgotesting.UpdateAction)
		tc.updated = updateAction.GetObject().(*corev1.ConfigMap)
		return true, tc.updated, nil
	})
	tc.clientset = cs.CoreV1()
	return &tc
}

func TestInitCreates(t *testing.T) {
	tc := NewTestClient()
	cs := NewConfigMapKVStore(context.Background(), name, namespace, tc.clientset)
	err := cs.Init(context.Background())
	if err != nil {
		t.Errorf("Failed to Init ConfigStore: %v", err)
	}
	if tc.created == nil {
		t.Errorf("ConfigMap not created")
	}
	if len(tc.created.Data) != 0 {
		t.Errorf("ConfigMap data is not empty")
	}
	if tc.updated != nil {
		t.Errorf("ConfigMap updated")
	}
}

func TestLoadNonexisting(t *testing.T) {
	tc := NewTestClient()
	if NewConfigMapKVStore(context.Background(), name, namespace, tc.clientset).Load(context.Background()) == nil {
		t.Error("non-existent store load didn't fail")
	}
}

func TestInitLoads(t *testing.T) {
	tc := NewTestClient([]runtime.Object{configMap(map[string]string{"foo": marshal(t, "bar")})}...)
	cs := NewConfigMapKVStore(context.Background(), name, namespace, tc.clientset)
	err := cs.Init(context.Background())
	if err != nil {
		t.Errorf("Failed to Init ConfigStore: %v", err)
	}
	if tc.created != nil {
		t.Errorf("ConfigMap created")
	}
	if tc.updated != nil {
		t.Errorf("ConfigMap updated")
	}
	var ret string
	err = cs.Get(context.Background(), "foo", &ret)
	if err != nil {
		t.Errorf("failed to return string: %v", err)
	}
	if ret != "bar" {
		t.Errorf("got back unexpected value, wanted %q got %q", "bar", ret)
	}
	if cs.Get(context.Background(), "not there", &ret) == nil {
		t.Error("non-existent key didn't error")
	}
}

func TestLoadSaveUpdate(t *testing.T) {
	tc := NewTestClient([]runtime.Object{configMap(map[string]string{"foo": marshal(t, "bar")})}...)
	cs := NewConfigMapKVStore(context.Background(), name, namespace, tc.clientset)
	err := cs.Init(context.Background())
	if err != nil {
		t.Errorf("Failed to Init ConfigStore: %v", err)
	}
	cs.Set(context.Background(), "jimmy", "otherbar")
	cs.Save(context.Background())
	if tc.updated == nil {
		t.Errorf("ConfigMap Not updated")
	}
	var ret string
	err = cs.Get(context.Background(), "jimmy", &ret)
	if err != nil {
		t.Errorf("failed to return string: %v", err)
	}
	if err != nil {
		t.Errorf("failed to return string: %v", err)
	}
	if ret != "otherbar" {
		t.Errorf("got back unexpected value, wanted %q got %q", "bar", ret)
	}
}

func TestLoadSaveUpdateComplex(t *testing.T) {
	ts := testStruct{
		LastThingProcessed: "somethingie",
		Stuff:              []string{"first", "second", "third"},
	}

	tc := NewTestClient([]runtime.Object{configMap(map[string]string{"foo": marshal(t, &ts)})}...)
	cs := NewConfigMapKVStore(context.Background(), name, namespace, tc.clientset)
	err := cs.Init(context.Background())
	if err != nil {
		t.Errorf("Failed to Init ConfigStore: %v", err)
	}
	ts2 := testStruct{
		LastThingProcessed: "otherthingie",
		Stuff:              []string{"fourth", "fifth", "sixth"},
	}
	cs.Set(context.Background(), "jimmy", &ts2)
	cs.Save(context.Background())
	if tc.updated == nil {
		t.Errorf("ConfigMap Not updated")
	}
	var ret testStruct
	err = cs.Get(context.Background(), "jimmy", &ret)
	if err != nil {
		t.Errorf("failed to return string: %v", err)
	}
	if err != nil {
		t.Errorf("failed to return string: %v", err)
	}
	if !reflect.DeepEqual(ret, ts2) {
		t.Errorf("got back unexpected value, wanted %+v got %+v", ts2, ret)
	}
}

func marshal(t *testing.T, value interface{}) string {
	bytes, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("Failed to Marshal %q: %v", value, err)
	}
	return string(bytes)
}

func configMap(data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
}
