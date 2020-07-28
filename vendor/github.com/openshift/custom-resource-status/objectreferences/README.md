Object References
=================

The `ObjectReference` type is provided by Kubernetes Core API
`"k8s.io/api/core/v1"` but the functions to set and find an `ObjectReference`
are provided in this package. This is useful if you would like
to include in the Status of your Custom Resource a list of objects
that are managed by your operator (ie. Deployments, Services, other
Custom Resources, etc.).

For example, we can add `RelatedObjects` to our Status struct:

```
// ExampleAppStatus defines the observed state of ExampleApp
type ExampleAppStatus struct {
  ...
  // RelatedObjects is a list of objects that are "interesting" or related to this operator.
  RelatedObjects []corev1.ObjectReference `json:"relatedObjects,omitempty"`
}
```

Then, through Reconcile, when an object we manage has been found we can add it to
the `RelatedObjects` slice.

```
found := &someAPI.SomeObject{}
err := r.client.Get(context.TODO(), types.NamespacedName{Name: object.Name, Namespace: object.Namespace}, found)
...handle err

// Add it to the list of RelatedObjects if found
// import "k8s.io/client-go/tools/reference"
objectRef, err := reference.GetReference(r.scheme, found)
if err != nil {
  return err
}
objectreferencesv1.SetObjectReference(&instance.Status.RelatedObjects, *objectRef)

// Update the status
err = r.client.Status().Update(context.TODO(), instance)
...handle err
```

**NOTE**: This package specifies a minimum for what constitutes a valid object
reference. The minimum valid object reference consists of non-empty strings
for the object's:

* APIVersion
* Kind
* Name
