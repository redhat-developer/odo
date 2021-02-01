
# Improved schema for operator backed services

## Abstract
This proposal is about improving the experience for a user of the operator backed services. Currently a user doesn't know exaustively which fields are available in a CR as we depend on optional metadata present in CRDs, to improve this we are considering using the metadata available from the cluster about the CRDs.
Getting that metadata from the cluster is challenging because a normal cluster user ( plain vanilla ) doesn't have access to that metadata and so we are going to implement two workflows - one for the user that has the privileges and one that doesn't.
This change would affect multiple service commands and the changes are described briefly below -
- `odo catalog describe service` should include the metadata fields with description so the users can provide then when doing `odo service create`
- `odo catalog list service` shouldn't change much other then listing CRDs per operator
- `odo service create` should take flags and dynamically fill the CRD structs - metadata would be used for validation
- `odo service create --input-file` would be used by adapter team or whoever wants to provide the whole CRD themselves as a file

## Implementation plan

Note - Below I would mention `cluster` manytimes and that means both openshift and kuberenetes. If there is something specific to one or the other then I would mention it.

Below is the step wise plan for implementing this feature - 

### Getting the metadata

#### Admin Access/Access to CRD API path
First we would consider the scenerio where the user has admin access or privileges to access the CRD API to the cluster.
Cluster provides an api to get metadata for any CRD needed. The url is of `<cluster-url>/api/kubernetes/apis/apiextensions.k8s.io/v1/customresourcedefinitions`. 
<details open>
    <summary> A sample output from the API looks like this </summary>

    ```
    {
    "kind": "CustomResourceDefinition",
    "apiVersion": "apiextensions.k8s.io/v1",
    "metadata": {
        "name": "servicebindings.operators.coreos.com",
        "selfLink": "/apis/apiextensions.k8s.io/v1/customresourcedefinitions/servicebindings.operators.coreos.com",
        "uid": "8d92fd7d-e982-4ad5-9805-59fbdcdf02b1",
        "resourceVersion": "39711",
        "generation": 1,
        "creationTimestamp": "2020-11-25T09:00:01Z",
        "labels": {
        "operators.coreos.com/rh-service-binding-operator.openshift-operators": ""
        },
        "managedFields": [
        {
            "manager": "catalog",
            "operation": "Update",
            "apiVersion": "apiextensions.k8s.io/v1beta1",
            "time": "2020-11-25T09:00:01Z",
            "fieldsType": "FieldsV1",
            "fieldsV1": {"f:spec":{"f:conversion":{".":{},"f:strategy":{}},"f:group":{},"f:names":{"f:kind":{},"f:listKind":{},"f:plural":{},"f:shortNames":{},"f:singular":{}},"f:preserveUnknownFields":{},"f:scope":{},"f:subresources":{".":{},"f:status":{}},"f:validation":{".":{},"f:openAPIV3Schema":{".":{},"f:description":{},"f:properties":{".":{},"f:apiVersion":{".":{},"f:description":{},"f:type":{}},"f:kind":{".":{},"f:description":{},"f:type":{}},"f:metadata":{".":{},"f:type":{}},"f:spec":{".":{},"f:description":{},"f:properties":{".":{},"f:application":{".":{},"f:description":{},"f:properties":{".":{},"f:bindingPath":{".":{},"f:description":{},"f:properties":{".":{},"f:containersPath":{".":{},"f:description":{},"f:type":{}},"f:secretPath":{".":{},"f:description":{},"f:type":{}}},"f:type":{}},"f:group":{".":{},"f:type":{}},"f:labelSelector":{".":{},"f:description":{},"f:properties":{".":{},"f:matchExpressions":{".":{},"f:description":{},"f:items":{},"f:type":{}},"f:matchLabels":{".":{},"f:additionalProperties":{},"f:description":{},"f:type":{}}},"f:type":{}},"f:name":{".":{},"f:description":{},"f:type":{}},"f:resource":{".":{},"f:type":{}},"f:version":{".":{},"f:type":{}}},"f:required":{},"f:type":{}},"f:customEnvVar":{".":{},"f:description":{},"f:items":{},"f:type":{}},"f:detectBindingResources":{".":{},"f:description":{},"f:type":{}},"f:envVarPrefix":{".":{},"f:description":{},"f:type":{}},"f:mountPathPrefix":{".":{},"f:description":{},"f:type":{}},"f:services":{".":{},"f:description":{},"f:items":{},"f:type":{}}},"f:type":{}},"f:status":{".":{},"f:description":{},"f:properties":{".":{},"f:applications":{".":{},"f:description":{},"f:items":{},"f:type":{}},"f:conditions":{".":{},"f:description":{},"f:items":{},"f:type":{}},"f:secret":{".":{},"f:description":{},"f:type":{}}},"f:required":{},"f:type":{}}},"f:type":{}}},"f:version":{},"f:versions":{}},"f:status":{"f:storedVersions":{}}}
        },
        {
            "manager": "kube-apiserver",
            "operation": "Update",
            "apiVersion": "apiextensions.k8s.io/v1",
            "time": "2020-11-25T09:00:01Z",
            "fieldsType": "FieldsV1",
            "fieldsV1": {"f:status":{"f:acceptedNames":{"f:kind":{},"f:listKind":{},"f:plural":{},"f:shortNames":{},"f:singular":{}},"f:conditions":{}}}
        },
        {
            "manager": "olm",
            "operation": "Update",
            "apiVersion": "apiextensions.k8s.io/v1",
            "time": "2020-11-25T09:00:02Z",
            "fieldsType": "FieldsV1",
            "fieldsV1": {"f:metadata":{"f:labels":{".":{},"f:operators.coreos.com/rh-service-binding-operator.openshift-operators":{}}}}
        }
        ]
    },
    "spec": {
        "group": "operators.coreos.com",
        "names": {
        "plural": "servicebindings",
        "singular": "servicebinding",
        "shortNames": [
            "sbr",
            "sbrs"
        ],
        "kind": "ServiceBinding",
        "listKind": "ServiceBindingList"
        },
        "scope": "Namespaced",
        "versions": [
        {
            "name": "v1alpha1",
            "served": true,
            "storage": true,
            "schema": {
            "openAPIV3Schema": {
                "description": "ServiceBinding expresses intent to bind an operator-backed service with an application workload.",
                "type": "object",
                "properties": {
                "apiVersion": {
    "description": "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources",
    "type": "string"
    },
                "kind": {
    "description": "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds",
    "type": "string"
    },
                "metadata": {
    "type": "object"
    },
                "spec": {
    "description": "ServiceBindingSpec defines the desired state of ServiceBinding",
    "type": "object",
    "properties": {
        "application": {
    "description": "Application is used to identify the application connecting to the backing service operator.",
    "type": "object",
    "required": [
        "group",
        "resource",
        "version"
    ],
    "properties": {
        "bindingPath": {
    "description": "BindingPath refers to the paths in the application workload's schema where the binding workload would be referenced. If BindingPath is not specified the default path locations is going to be used.  The default location for ContainersPath is going to be: \"spec.template.spec.containers\" and if SecretPath is not specified, the name of the secret object is not going to be specified.",
    "type": "object",
    "properties": {
        "containersPath": {
    "description": "ContainersPath defines the path to the corev1.Containers reference If BindingPath is not specified, the default location is going to be: \"spec.template.spec.containers\"",
    "type": "string"
    },
        "secretPath": {
    "description": "SecretPath defines the path to a string field where the name of the secret object is going to be assigned. Note: The name of the secret object is same as that of the name of SBR CR (metadata.name)",
    "type": "string"
    }
    }
    },
        "group": {
    "type": "string"
    },
        "labelSelector": {
    "description": "A label selector is a label query over a set of resources. The result of matchLabels and matchExpressions are ANDed. An empty label selector matches all objects. A null label selector matches no objects.",
    "type": "object",
    "properties": {
        "matchExpressions": {
    "description": "matchExpressions is a list of label selector requirements. The requirements are ANDed.",
    "type": "array",
    "items": {"description":"A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.","type":"object","required":["key","operator"],"properties":{"key":{"description":"key is the label key that the selector applies to.","type":"string"},"operator":{"description":"operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.","type":"string"},"values":{"description":"values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.","type":"array","items":{"type":"string"}}}}
    },
        "matchLabels": {
    "description": "matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is \"key\", the operator is \"In\", and the values array contains only \"value\". The requirements are ANDed.",
    "type": "object",
    "additionalProperties": {"type":"string"}
    }
    }
    },
        "name": {
    "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
    "type": "string"
    },
        "resource": {
    "type": "string"
    },
        "version": {
    "type": "string"
    }
    }
    },
        "customEnvVar": {
    "description": "Custom env variables",
    "type": "array",
    "items": {"description":"EnvVar represents an environment variable present in a Container.","type":"object","required":["name"],"properties":{"name":{"description":"Name of the environment variable. Must be a C_IDENTIFIER.","type":"string"},"value":{"description":"Variable references $(VAR_NAME) are expanded using the previous defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to \"\".","type":"string"},"valueFrom":{"description":"Source for the environment variable's value. Cannot be used if value is not empty.","type":"object","properties":{"configMapKeyRef":{"description":"Selects a key of a ConfigMap.","type":"object","required":["key"],"properties":{"key":{"description":"The key to select.","type":"string"},"name":{"description":"Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?","type":"string"},"optional":{"description":"Specify whether the ConfigMap or its key must be defined","type":"boolean"}}},"fieldRef":{"description":"Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP.","type":"object","required":["fieldPath"],"properties":{"apiVersion":{"description":"Version of the schema the FieldPath is written in terms of, defaults to \"v1\".","type":"string"},"fieldPath":{"description":"Path of the field to select in the specified API version.","type":"string"}}},"resourceFieldRef":{"description":"Selects a resource of the container: only resources limits and requests (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.","type":"object","required":["resource"],"properties":{"containerName":{"description":"Container name: required for volumes, optional for env vars","type":"string"},"divisor":{"description":"Specifies the output format of the exposed resources, defaults to \"1\"","type":"string"},"resource":{"description":"Required: resource to select","type":"string"}}},"secretKeyRef":{"description":"Selects a key of a secret in the pod's namespace","type":"object","required":["key"],"properties":{"key":{"description":"The key of the secret to select from.  Must be a valid secret key.","type":"string"},"name":{"description":"Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?","type":"string"},"optional":{"description":"Specify whether the Secret or its key must be defined","type":"boolean"}}}}}}}
    },
        "detectBindingResources": {
    "description": "DetectBindingResources is flag used to bind all non-bindable variables from different subresources owned by backing operator CR.",
    "type": "boolean"
    },
        "envVarPrefix": {
    "description": "EnvVarPrefix is the prefix for environment variables",
    "type": "string"
    },
        "mountPathPrefix": {
    "description": "MountPathPrefix is the prefix for volume mount",
    "type": "string"
    },
        "services": {
    "description": "Services is used to identify multiple backing services.",
    "type": "array",
    "items": {"description":"Service defines the selector based on resource name, version, and resource kind","type":"object","required":["group","kind","version"],"properties":{"envVarPrefix":{"type":"string"},"group":{"type":"string"},"id":{"type":"string"},"kind":{"type":"string"},"name":{"description":"Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?","type":"string"},"namespace":{"type":"string"},"version":{"type":"string"}}}
    }
    }
    },
                "status": {
    "description": "ServiceBindingStatus defines the observed state of ServiceBinding",
    "type": "object",
    "required": [
        "conditions",
        "secret"
    ],
    "properties": {
        "applications": {
    "description": "Applications contain all the applications filtered by name or label",
    "type": "array",
    "items": {"description":"BoundApplication defines the application workloads to which the binding secret has injected.","type":"object","required":["group","kind","version"],"properties":{"group":{"type":"string"},"kind":{"type":"string"},"name":{"description":"Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?","type":"string"},"version":{"type":"string"}}}
    },
        "conditions": {
    "description": "Conditions describes the state of the operator's reconciliation functionality.",
    "type": "array",
    "items": {"description":"Condition represents the state of the operator's reconciliation functionality.","type":"object","required":["status","type"],"properties":{"lastHeartbeatTime":{"type":"string","format":"date-time"},"lastTransitionTime":{"type":"string","format":"date-time"},"message":{"type":"string"},"reason":{"type":"string"},"status":{"type":"string"},"type":{"description":"ConditionType is the state of the operator's reconciliation functionality.","type":"string"}}}
    },
        "secret": {
    "description": "Secret is the name of the intermediate secret",
    "type": "string"
    }
    }
    }
                }
            }
            },
            "subresources": {
            "status": {}
            }
        }
        ],
        "conversion": {
        "strategy": "None"
        },
        "preserveUnknownFields": true
    },
    "status": {
        "conditions": [
        {
            "type": "NamesAccepted",
            "status": "True",
            "lastTransitionTime": "2020-11-25T09:00:01Z",
            "reason": "NoConflicts",
            "message": "no conflicts found"
        },
        {
            "type": "Established",
            "status": "True",
            "lastTransitionTime": "2020-11-25T09:00:01Z",
            "reason": "InitialNamesAccepted",
            "message": "the initial names have been accepted"
        },
        {
            "type": "NonStructuralSchema",
            "status": "True",
            "lastTransitionTime": "2020-11-25T09:00:01Z",
            "reason": "Violations",
            "message": "spec.preserveUnknownFields: Invalid value: true: must be false"
        }
        ],
        "acceptedNames": {
        "plural": "servicebindings",
        "singular": "servicebinding",
        "shortNames": [
            "sbr",
            "sbrs"
        ],
        "kind": "ServiceBinding",
        "listKind": "ServiceBindingList"
        },
        "storedVersions": [
        "v1alpha1"
        ]
    }
    }
    ```

</details>

We care about the `openAPIV3Schema` as that would be used to build the CRD struct. 

#### Non-admin/restriction to CRDs API access path

If the user doesnt't have admin access then we cannot provide the enhancement metadata for the CRDs and hence we would show a warning that `odo service creation would provide the best User Experience when user has these ...privileges`.
The fall back is to use the `alm-examples` field from the `ClusterServiceVersion` provided by the operator which has very limited information and are optional. Some samples of what the alm examples hold

```

apiVersion: etcd.database.coreos.com/v1beta2
kind: EtcdCluster
metadata:
  annotations:
    etcd.database.coreos.com/scope: clusterwide
  name: example
spec:
  size: 3
  version: 3.2.13

```

```
apiVersion: operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  name: example-servicebinding
spec:
  application:
    group: apps
    name: nodejs-rest-http-crud
    resource: deployments
    version: v1
  mountPathPrefix: /var/credentials
  services:
  - group: postgresql.example.dev
    kind: Database
    name: pg-instance
    version: v1alpha1
```


### Sample used for all explaination below

<details open>
<summary> Below is extract of an example `openAPIV3Schema` which would be used for explaination </summary>
```
{
  "application": {
    "description": "Application is used to identify the application connecting to the backing service operator.",
    "type": "object",
    "required": [
      "group",
      "resource",
      "version"
    ],
    "properties": {
      "bindingPath": {
        "description": "BindingPath refers to the paths in the application workload's schema where the binding workload would be referenced. If BindingPath is not specified the default path locations is going to be used.  The default location for ContainersPath is going to be: \"spec.template.spec.containers\" and if SecretPath is not specified, the name of the secret object is not going to be specified.",
        "type": "object",
        "properties": {
          "containersPath": {
            "description": "ContainersPath defines the path to the corev1.Containers reference If BindingPath is not specified, the default location is going to be: \"spec.template.spec.containers\"",
            "type": "string"
          },
          "secretPath": {
            "description": "SecretPath defines the path to a string field where the name of the secret object is going to be assigned. Note: The name of the secret object is same as that of the name of SBR CR (metadata.name)",
            "type": "string"
          }
        }
      },
      "group": {
        "type": "string"
      },
      "labelSelector": {
        "description": "A label selector is a label query over a set of resources. The result of matchLabels and matchExpressions are ANDed. An empty label selector matches all objects. A null label selector matches no objects.",
        "type": "object",
        "properties": {
          "matchExpressions": {
            "description": "matchExpressions is a list of label selector requirements. The requirements are ANDed.",
            "type": "array",
            "items": {
              "description": "A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
              "type": "object",
              "required": [
                "key",
                "operator"
              ],
              "properties": {
                "key": {
                  "description": "key is the label key that the selector applies to.",
                  "type": "string"
                },
                "operator": {
                  "description": "operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.",
                  "type": "string"
                },
                "values": {
                  "description": "values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.",
                  "type": "array",
                  "items": {
                    "type": "string"
                  }
                }
              }
            }
          },
          "matchLabels": {
            "description": "matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is \"key\", the operator is \"In\", and the values array contains only \"value\". The requirements are ANDed.",
            "type": "object",
            "additionalProperties": {
              "type": "string"
            }
          }
        }
      },
      "name": {
        "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
        "type": "string"
      },
      "resource": {
        "type": "string"
      },
      "version": {
        "type": "string"
      }
    }
  },
  "customEnvVar": {
    "description": "Custom env variables",
    "type": "array",
    "items": {
      "description": "EnvVar represents an environment variable present in a Container.",
      "type": "object",
      "required": [
        "name"
      ],
      "properties": {
        "name": {
          "description": "Name of the environment variable. Must be a C_IDENTIFIER.",
          "type": "string"
        },
        "value": {
          "description": "Variable references $(VAR_NAME) are expanded using the previous defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to \"\".",
          "type": "string"
        },
        "valueFrom": {
          "description": "Source for the environment variable's value. Cannot be used if value is not empty.",
          "type": "object",
          "properties": {
            "configMapKeyRef": {
              "description": "Selects a key of a ConfigMap.",
              "type": "object",
              "required": [
                "key"
              ],
              "properties": {
                "key": {
                  "description": "The key to select.",
                  "type": "string"
                },
                "name": {
                  "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
                  "type": "string"
                },
                "optional": {
                  "description": "Specify whether the ConfigMap or its key must be defined",
                  "type": "boolean"
                }
              }
            },
            "fieldRef": {
              "description": "Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP.",
              "type": "object",
              "required": [
                "fieldPath"
              ],
              "properties": {
                "apiVersion": {
                  "description": "Version of the schema the FieldPath is written in terms of, defaults to \"v1\".",
                  "type": "string"
                },
                "fieldPath": {
                  "description": "Path of the field to select in the specified API version.",
                  "type": "string"
                }
              }
            },
            "resourceFieldRef": {
              "description": "Selects a resource of the container: only resources limits and requests (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and requests.ephemeral-storage) are currently supported.",
              "type": "object",
              "required": [
                "resource"
              ],
              "properties": {
                "containerName": {
                  "description": "Container name: required for volumes, optional for env vars",
                  "type": "string"
                },
                "divisor": {
                  "description": "Specifies the output format of the exposed resources, defaults to \"1\"",
                  "type": "string"
                },
                "resource": {
                  "description": "Required: resource to select",
                  "type": "string"
                }
              }
            },
            "secretKeyRef": {
              "description": "Selects a key of a secret in the pod's namespace",
              "type": "object",
              "required": [
                "key"
              ],
              "properties": {
                "key": {
                  "description": "The key of the secret to select from.  Must be a valid secret key.",
                  "type": "string"
                },
                "name": {
                  "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
                  "type": "string"
                },
                "optional": {
                  "description": "Specify whether the Secret or its key must be defined",
                  "type": "boolean"
                }
              }
            }
          }
        }
      }
    }
  },
  "detectBindingResources": {
    "description": "DetectBindingResources is flag used to bind all non-bindable variables from different subresources owned by backing operator CR.",
    "type": "boolean"
  },
  "envVarPrefix": {
    "description": "EnvVarPrefix is the prefix for environment variables",
    "type": "string"
  },
  "mountPathPrefix": {
    "description": "MountPathPrefix is the prefix for volume mount",
    "type": "string"
  },
  "services": {
    "description": "Services is used to identify multiple backing services.",
    "type": "array",
    "items": {
      "description": "Service defines the selector based on resource name, version, and resource kind",
      "type": "object",
      "required": [
        "group",
        "kind",
        "version"
      ],
      "properties": {
        "envVarPrefix": {
          "type": "string"
        },
        "group": {
          "type": "string"
        },
        "id": {
          "type": "string"
        },
        "kind": {
          "type": "string"
        },
        "name": {
          "description": "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
          "type": "string"
        },
        "namespace": {
          "type": "string"
        },
        "version": {
          "type": "string"
        }
      }
    }
  }
}
```
</details>

- There are 5 top level fields in the metadata -
    - `application`
    - `customEnvVar`
    - `detectBindingResources`
    - `envVarPrefix`
    - `mountPathPrefix`
    - `services`

### Conversion of openAPIV3Schema to json schema

We after discussions with adapters team have decided to convert the openAPIV3Schema to json schema so that it would be easy to consume and hence we would be working on the converted json schema version to derive information from it.

TODO: find a good enough library to convert openAPIV3Schema to json schema reliably 

### cobra flags to golang map conversion

We need to finalise a mechanism/syntax so that a user can set any property for a CR using the command line, for e.g. if we considered `jsonpath` something like this `odo service create servicebinding.coreos.io/Servicebinding -services[0].envVarPrefix "SVC" -services[0].namespace "openshift"`. 

Currently the decision is to use cobra's syntax for passing flags but we need to finalise and consider all if cobra's syntax will support all scenerios.

### "odo catalog list services"

- No changes needed to `odo catalog list services` as it already shows the `Operators` and the respective CRs they provide for the user to `describe` on.

### "odo catalog describe service"

#### JSON output

### "odo catalog search services"

#### JSON output


### "odo service create"


#### JSON output


#### --input-file flag




