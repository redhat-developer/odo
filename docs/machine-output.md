# Machine Readable Output

This document outlines all the machine readable output options and examples.

`odo app describe -o json | jq`
 
```json
{
  "kind": "Application",
  "apiVersion": "odo.openshift.io/v1alpha1",
  "metadata": {
    "name": "app",
    "namespace": "myproject",
    "creationTimestamp": null
  },
  "spec": {},
  "status": {
    "active": false
  }
}
```


`odo app list -o json | jq`
 
```json
{
  "kind": "ApplicationList",
  "apiVersion": "odo.openshift.io/v1alpha1",
  "metadata": {},
  "items": [
    {
      "kind": "Application",
      "apiVersion": "odo.openshift.io/v1alpha1",
      "metadata": {
        "name": "app",
        "namespace": "myproject",
        "creationTimestamp": null
      },
      "spec": {
        "components": [
          "nodejs-nvnh"
        ]
      },
      "status": {
        "active": false
      }
    }
  ]
}
```


`odo app list --project myproject -o json | jq`
 
```json
{
  "kind": "ApplicationList",
  "apiVersion": "odo.openshift.io/v1alpha1",
  "metadata": {},
  "items": [
    {
      "kind": "app",
      "apiVersion": "odo.openshift.io/v1alpha1",
      "metadata": {
        "name": "app",
        "namespace": "myproject",
        "creationTimestamp": null
      },
      "spec": {
        "components": [
          "app-nodejs-komz"
        ]
      }
    }
  ]
}

```


`odo component describe -o json | jq`
 
```json
{
  "kind": "Component",
  "apiVersion": "odo.openshift.io/v1alpha1",
  "metadata": {
    "name": "nodejs-nvnh",
    "creationTimestamp": null
  },
  "spec": {
    "type": "nodejs",
    "source": "file://./",
    "url": [
      "example",
      "json",
      "nodejs-nvnh-8080"
    ],
    "storage": [
      "mystorage"
    ]
  },
  "status": {
    "state": "Pushed"
  }
}
```


`odo component list --output json | jq`
 
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
        "name": "nodejs-vzae",
        "creationTimestamp": null
      },
      "spec": {
        "type": "nodejs",
        "source": "file://./"
      },
      "status": {
        "state": "Pushed"
      }
    }
  ]
}
```


`odo describe -o json | jq`
 
```json
{
  "kind": "Component",
  "apiVersion": "odo.openshift.io/v1alpha1",
  "metadata": {
    "name": "nodejs-nvnh",
    "creationTimestamp": null
  },
  "spec": {
    "type": "nodejs",
    "source": "file://./",
    "url": [
      "example",
      "json",
      "nodejs-nvnh-8080"
    ],
    "storage": [
      "mystorage"
    ]
  },
  "status": {
    "state": "Pushed"
  }
}
```


`odo list -o json | jq`
 
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
        "name": "nodejs-nvnh",
        "creationTimestamp": null
      },
      "spec": {
        "type": "nodejs",
        "source": "file://./",
        "url": [
          "example",
          "json",
          "nodejs-nvnh-8080"
        ],
        "storage": [
          "mystorage"
        ]
      },
      "status": {
        "state": "Pushed"
      }
    }
  ]
}
```


`odo project list -o json | jq`
 
```json
{
  "kind": "List",
  "apiVersion": "odo.openshift.io/v1alpha1",
  "metadata": {},
  "items": [
    {
      "kind": "Project",
      "apiVersion": "odo.openshift.io/v1alpha1",
      "metadata": {
        "name": "myproject",
        "creationTimestamp": null
      },
      "spec": {
        "apps": [
          "app"
        ]
      },
      "status": {
        "active": true
      }
    }
  ]
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
