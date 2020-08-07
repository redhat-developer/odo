# Kubernetes API Mocks

This folder contains all mocks and machinery to interact with a fake Kubernetes API.

## Usage

### 1. Instantiate Fake

Inside a testing function, instantiate `Fake` by sharing the test context `t` and namespace name:

``` go
import "github.com/redhat-developer/service-binding-operator/test/mocks"

f := mocks.NewFake(t, "namespace")
```

### 2. Add Mocked Objects

Add mocked objects are you need.

``` go
f.AddMockedUnstructuredSecret("db-credentials")
```

### 3. Instantiate API Clients

Instantiate a fake API client, with:

``` go
fakeClient := f.FakeClient()
fakeDynamicClient := f.FakeDynClient()
```

## Unstructured List vs. Typed Resource

As you may notice, in [`mocks.go`](./mocks.go) we have methods returning typed Kubernetes objects,
and sometimes returning `unstructured.Unstructured`. That happens because when using `List` with the
dynamic client, it fails on parsing objects inside:

```
item[0]: can't assign or convert v1alpha1.ClusterServiceVersion into unstructured.Unstructured
```

However, using `Unstructured` it does not fail during testing.
