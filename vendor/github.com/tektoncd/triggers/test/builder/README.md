# Builder package for tests

This package holds `Builder` functions that can be used to create struct in
tests with less noise.

One of the most important characteristic of a unit test (and any type of test
really) is **readability**. This means it should be easy to read and clearly
show the intent of the test. The setup (and cleanup) of the tests should be as
small as possible to avoid the noise. Those builders exists to help with that.

There are two types of functions defined in this package:

    *Builders*: Create and return a struct
    *Modifiers*: Return a function that will operate on a given struct.

```go
    // Definition
    type TriggerBindingOp func(*v1alpha1.TriggerBinding)
    // Builder
    func TriggerBinding(name, namespace string, ops ...TriggerBindingOp) *v1alpha1.TriggerBinding {
        // […]
    }
    // Modifier
    func TriggerBindingSpec(ops ...TriggerBindingSpecOp) TriggerBindingOp {
        // […]
    }
```

The main reason to define the `Op` type, and using it in the methods signatures
is to group Modifier function together. It makes it easier to see what is a
Modifier (or Builder) and on what it operates.

The go tests in this package exemplify the consolidation that can be achieved by
using the builders:

- [`EventListener`](eventlistener_test.go)
- [`TriggerBinding`](triggerbinding_test.go)
- [`TriggerTemplate`](triggertemplate_test.go)
