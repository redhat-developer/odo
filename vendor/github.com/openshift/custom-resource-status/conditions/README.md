Conditions
==========

Provides:

* `Condition` type as specified in the [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
* `ConditionType` and generally useful constants for this type (ie. "Available",
    "Progressing", "Degraded", and "Upgradeable")
* Functions for setting, removing, finding, and evaluating conditions.

To use, simply add `Conditions` to your Custom Resource Status struct like:

```
// ExampleAppStatus defines the observed state of ExampleApp
type ExampleAppStatus struct {
  ...
  // conditions describes the state of the operator's reconciliation functionality.
  // +patchMergeKey=type
  // +patchStrategy=merge
  // +optional
  // Conditions is a list of conditions related to operator reconciliation
  Conditions []conditions.Condition `json:"conditions,omitempty"  patchStrategy:"merge" patchMergeKey:"type"`
}
```

Then, as appropriate in your Reconcile function, use
`conditions.SetStatusConditions` like:

```
instance := &examplev1alpha1.ExampleApp{}
err := r.client.Get(context.TODO(), request.NamespacedName, instance)
...handle err

conditions.SetStatusCondition(&instance.Status.Conditions, conditions.Condition{
  Type:   conditions.ConditionAvailable,
  Status: corev1.ConditionFalse,
  Reason: "ReconcileStarted",
  Message: "Reconciling resource"
})

// Update the status
err = r.client.Status().Update(context.TODO(), instance)
...handle err
```
