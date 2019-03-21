# Machine Readable Output

This document outlines all the machine readable output options and examples.

`jq` is used in order to "beautify" the JSON-output.

`$ odo url create -o json | jq`

```json
{
  "kind": "url",
  "apiVersion": "odo.openshift.io/v1alpha1",
  "metadata": {
    "name": "foobar-8080",
    "creationTimestamp": null
  },
  "spec": {
    "host": "foobar-8080-odo-cmac-foobar.e8ca.engint.openshiftapps.com",
    "protocol": "http",
    "port": 8080
  }
}
```


`$ odo storage create mystorage --path /opt/foobar --size=1Gi -o json | jq`

```json
{
  "kind": "storage",
  "apiVersion": "odo.openshift.io/v1alpha1",
  "metadata": {
    "name": "mystorage",
    "creationTimestamp": null
  },
  "spec": {
    "size": "1Gi"
  },
  "status": {
    "path": "/opt/foobar"
  }
}
```

`$ odo list -o json | jq`

```json
{
  "kind": "List",
  "apiVersion": "odo.openshift.io/v1alpha1",
  "metadata": {},
  "items": [
    {
      "kind": "Component",
      "apiVersion": "odo.openshift.io/v1alpha1",
      "metadata": {
        "name": "foobar",
        "creationTimestamp": null
      },
      "spec": {
        "type": "nodejs",
        "source": "file:///home/wikus/nodejs-ex",
        "url": [
          "foobar-8080"
        ],
        "storage": [
          "mystorage"
        ]
      },
      "status": {
        "active": true
      }
    }
  ]
}
```

`$ odo url list -o json | jq`

```json
{
  "kind": "List",
  "apiVersion": "odo.openshift.io/v1alpha1",
  "metadata": {},
  "items": [
    {
      "kind": "url",
      "apiVersion": "odo.openshift.io/v1alpha1",
      "metadata": {
        "name": "foobar-8080",
        "creationTimestamp": null
      },
      "spec": {
        "host": "foobar-8080-odo-cmac-foobar.e8ca.engint.openshiftapps.com",
        "protocol": "http",
        "port": 8080
      }
    }
  ]
}
```

`$ odo storage list -o json | jq`

```json
{
  "kind": "List",
  "apiVersion": "odo.openshift.io/v1aplha1",
  "metadata": {},
  "items": [
    {
      "kind": "Storage",
      "apiVersion": "odo.openshift.io/v1alpha1",
      "metadata": {
        "name": "mystorage",
        "creationTimestamp": null
      },
      "spec": {
        "size": "1Gi"
      },
      "status": {
        "path": "/opt/foobar"
      }
    }
  ]
}
```
