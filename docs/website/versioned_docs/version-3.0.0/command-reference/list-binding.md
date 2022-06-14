---
title: odo list binding
---

## Description

You can use `odo list binding` to list all the Service Bindings declared in the current namespace and, if present, 
in the Devfile of the curretn directory.

This command supports the service bindings added with the command `odo add binding` and bindings added manually
to the Devfile, using a `ServiceBinding` resource from one of these apiVersion:
- `binding.operators.coreos.com/v1alpha1`
- `servicebinding.io/v1alpha3`

## Running the Command

To list all the service bindings, you can run `odo list binding`:
```shell
odo list binding
```

Example:

```sh
$ odo list bindings
[TODO]
```
