
# Improved schema for operator backed services

This proposal is about improving the experience for a user of the operator backed services. Currently a user doesn't know exaustively which fields are available in a CR as we depend on optional metadata present in CRDs, to improve this we are considering using the metadata available from the cluster about the CRDs.
Getting that metadata from the cluster is challenging because a normal cluster user (which is a normal user in vanilla k8s or non-admin admin user for openshift) doesn't have access to that metadata so we are gonna follow a 3 tiered approach

- check if the user has access to the CRD, if so then use that. The reason we are doing this is because a non-admin user can still have access to CRDs because they might be given explict permissions by the admin.
- get the `swagger.json` from the openapi endpoint provided by kubernetes. a sample can be found here https://gist.github.com/girishramnani/f29949d5cb8c6547102776437e05ac19
- finally if we cannot get the openapi data as well then use the `ClusterServiceVersion` to get the parameters.

This change would affect multiple service commands and the changes are described briefly below -
- `odo catalog describe service` should include the metadata fields with description so the users can provide then when doing `odo service create`. 
- `odo catalog list service` wouldn't change at all.
- `odo service create` should take flags and dynamically fill the CRD structs - metadata would be used for validation
- `odo service create --from-file` would be used by adapter team or whoever wants to provide the whole CRD themselves as a file. this feature is already present and working.
- `json` output for all commands.

# Implementation plan

Note - Below I would mention `cluster` manytimes and that means both openshift and kuberenetes. If there is something specific to one or the other then I would mention it.
The actually order in which these features would be implemented in odo would as follows -

- [ ] `ClusterServiceVersion` approach 
- [ ] using `swagger.json` to get `openAPIv3schema`
- [ ] admin-only approach of using `CRD` api from the cluster.


Below explains the flow, that is how the code would get executed - 

## Getting the metadata

### Admin Access/Access to CRD API path
First we would consider the scenerio where the user has admin access or privileges to access the CRD API to the cluster.
Cluster provides an api to get metadata for any CRD needed. URL - `<cluster-url>/api/kubernetes/apis/apiextensions.k8s.io/v1/customresourcedefinitions`. 

Sample output - 
https://gist.github.com/girishramnani/cbb4400e463efe89c13f1386e0788793

We care about the `openAPIV3Schema` as that would be used to build the CRD struct. 

### Kuberenetes cluster Swagger has the schema

If the user doesn't have CRD access or that portion is not implemented yet then we fetch try to fetch the same `openAPIV3Schema` from the cluster's `swagger.json` from the endpoint `<cluster-url>/api/kubernetes/openapi/v2`. This is a very large document as it holds all the definitions present on the cluster.

So this needs to be cached and refreshed whenever a new operator is installed.

The caching would be per cluster as the `swagger.json` would change for different clusters. The approach would be as follows -

- when the users executes `odo catalog list services`, if there is no cache then we would fetch the `swagger.json`, odo wouldn't exit until download. We will show the user "Downloading service information from the cluster ........" while the swagger is downloaded. Also if we the feedback after releasing that the experience is not great then we will consider a non-blocking approach where the user doesn't have to wait the swagger to be downloaded. We aren't doing that for the beginning because non-blocking approach has some more complexities which might not be needed.
- we would store the `swagger.json` in the `~/.odo` as `<cluster-url>-swagger.json` to avoid conflicts.
- We would also create a `operator-listing-cache.json` in the `~/.odo` which would hold a key value mapping of `cluser-url` => `hash of the latest "odo catalog list services"`.
- now when the user runs `odo catalog list services` we would check if there is a cache present, if there is then it would compare the hash of the current `odo catalog list services` for the cluster with the one present in `operator-listing-cache.json`. If its different we redownload the swagger.json and update the hash of `odo catalog list services` in `operator-listing-cache.json`.


### Fetch ClusterServiceVersion to generate the information

If none of the above approaches work or are not implemented yet then we fallback to getting the information from `ClusterServiceVersion`

We generate the description in this approach based on `spec.customresourcedefinitions` present in the ClusterServiceVersion.
This CRD def is different from what is provided by the `CustomResourceDefinition` as it has less information.

This is how one of the `customresourcedefinition` looks like ( Kafka from strimzi )
```
{
  "parameters": [
    {
      "description": "Kafka version",
      "displayName": "Version",
      "path": "kafka.version",
      "x-descriptors": [
        "urn:alm:descriptor:com.tectonic.ui:text"
      ]
    },
    {
      "description": "The desired number of Kafka brokers.",
      "displayName": "Kafka Brokers",
      "path": "kafka.replicas",
      "x-descriptors": [
        "urn:alm:descriptor:com.tectonic.ui:podCount"
      ]
    },
    {
      "description": "The type of storage used by Kafka brokers",
      "displayName": "Kafka storage",
      "path": "kafka.storage.type",
      "x-descriptors": [
        "urn:alm:descriptor:com.tectonic.ui:select:ephemeral",
        "urn:alm:descriptor:com.tectonic.ui:select:persistent-claim",
        "urn:alm:descriptor:com.tectonic.ui:select:jbod"
      ]
    },
    {
      "description": "Limits describes the minimum/maximum amount of compute resources required/allowed",
      "displayName": "Kafka Resource Requirements",
      "path": "kafka.resources",
      "x-descriptors": [
        "urn:alm:descriptor:com.tectonic.ui:resourceRequirements"
      ]
    },
    {
      "description": "The desired number of Zookeeper nodes.",
      "displayName": "Zookeeper Nodes",
      "path": "zookeeper.replicas",
      "x-descriptors": [
        "urn:alm:descriptor:com.tectonic.ui:podCount"
      ]
    },
    {
      "description": "The type of storage used by Zookeeper nodes",
      "displayName": "Zookeeper storage",
      "path": "zookeeper.storage.type",
      "x-descriptors": [
        "urn:alm:descriptor:com.tectonic.ui:select:ephemeral",
        "urn:alm:descriptor:com.tectonic.ui:select:persistent-claim",
        "urn:alm:descriptor:com.tectonic.ui:select:jbod"
      ]
    },
    {
      "description": "Limits describes the minimum/maximum amount of compute resources required/allowed",
      "displayName": "Zookeeper Resource Requirements",
      "path": "zookeeper.resources",
      "x-descriptors": [
        "urn:alm:descriptor:com.tectonic.ui:resourceRequirements"
      ]
    }
  ]
}
```

## Allowing user to set the parameters 

The current approach of sending map values via cli are as follows - 

- we would add a `-p` cobra param as a list
- each of the parameter would represent the key in a map and value in a map `e.g. -p "key"="value"`
- we would allow json path in the key for the user to specific any field in the map that they want to set e.g. `services[0].namespace`.
- array params are gonna be passed using `[i]` syntax. `-p spec.env[0].name FOO -p spec.env[0].value BAR`
- we optimistically try to convert the type of param values in the below order
    - try to parse as integer, if it goes through then we assume it as an integer
    - try to parse as a float and if it goes through the we assume the value as float
    - try to parse as a boolean and if goes through then we assume the value as boolean
    - then everything else becomes string.

Sample - `odo service create servicebinding.coreos.io/Servicebinding/<version> -p  "services[0].envVarPrefix"="SVC" -p "services[0].namespace"="openshift"`

We would using https://github.com/tidwall/sjson to map the keys of the json.

this would yield into a map that looks like this

```

{
  "services":[
    {
      "envVarPrefix": "SVC",
      "namespace": "openshift"
    }
  ]
}

```

- odo also needs to have smart auto-complete which auto selects the version if the CR only has one version.


## Using the metadata to validate the user input

At this stage the user either has access to the `openAPIV3Schema` or `ClusterServiceVersion` and also the user has provided the service parameters they want to set as well. To hide the difference between these implementation from the user the `catalog describe` output would follow the same format. e.g.
```
- FieldPath: zookeeper.resources
  DisplayName: Zookeeper Resource Requirements
  Description: Limits describes the minimum/maximum amount of compute resources required/allowed (optional)
  Type: <type> (optional)
   
- FieldPath: zookeeper.storage.type
  Type: <type> (optional)
```

Observe that `Description` and `Type` are optional as for some scenerios that information is not present.

### User has access to openAPIV3Schema

We would use a json schema validator to validate the user provided params with the schema.
	
A similar approach is followed when validating devfile against devfile schema and we would use the same package https://github.com/santhosh-tekuri/jsonschema for validation for consistency.

<details>
<summary> Below is an extract of an example `openAPIV3Schema` which would be used for explaination </summary>

```

{
  "application": {
    "type": "object",
    "required": [
      "group",
      "resource",
      "version"
    ],
    "properties": {
      "bindingPath": {
        "type": "object",
        "properties": {
          "containersPath": {
            "type": "string"
          },
          "secretPath": {
            "type": "string"
          }
        }
      },
      "group": {
        "type": "string"
      },
      "labelSelector": {
        "type": "object",
        "properties": {
          "matchExpressions": {
            "type": "array",
            "items": {
              "type": "object",
              "required": [
                "key",
                "operator"
              ],
              "properties": {
                "key": {
                  "type": "string"
                },
                "operator": {
                  "type": "string"
                },
                "values": {
                  "type": "array",
                  "items": {
                    "type": "string"
                  }
                }
              }
            }
          },
          "matchLabels": {
            "type": "object",
            "additionalProperties": {
              "type": "string"
            }
          }
        }
      },
      "name": {
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
    "type": "array",
    "items": {
      "type": "object",
      "required": [
        "name"
      ],
      "properties": {
        "name": {
          "type": "string"
        },
        "value": {
          "type": "string"
        },
        "valueFrom": {
          "type": "object",
          "properties": {
            "configMapKeyRef": {
              "type": "object",
              "required": [
                "key"
              ],
              "properties": {
                "key": {
                  "type": "string"
                },
                "name": {
                  "type": "string"
                },
                "optional": {
                  "type": "boolean"
                }
              }
            },
            "fieldRef": {
              "type": "object",
              "required": [
                "fieldPath"
              ],
              "properties": {
                "apiVersion": {
                  "type": "string"
                },
                "fieldPath": {
                  "type": "string"
                }
              }
            },
            "resourceFieldRef": {
              "type": "object",
              "required": [
                "resource"
              ],
              "properties": {
                "containerName": {
                  "type": "string"
                },
                "divisor": {
                  "type": "string"
                },
                "resource": {
                  "type": "string"
                }
              }
            },
            "secretKeyRef": {
              "type": "object",
              "required": [
                "key"
              ],
              "properties": {
                "key": {
                  "type": "string"
                },
                "name": {
                  "type": "string"
                },
                "optional": {
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
    "type": "boolean"
  },
  "envVarPrefix": {
    "type": "string"
  },
  "mountPathPrefix": {
    "type": "string"
  },
  "services": {
    "type": "array",
    "items": {
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

Note - removed `description` fields to make the sample concise. 
</details>



### User has access to ClusterServiceVersion

The approach is to go through the keys provided by the user against the ones present in the ClusterServiceVersion's CRDDescription. if the user has provided parameters which aren't present in the Description ( SpecDescriptors ) then we show an error with all the parameters that are incorrectly provided.

## "odo catalog describe service"

The json output for all the approaches below would be -

```
{
  "kind": ...

  "spec":[
    {
      "fieldPath": "zookeeper.resources.limit.min",
      "displayName": "Limit Min",
      "Description": "<description>",
      "type":"string"
    },
    ....
  ],
}

```
Note - this is not "openAPIv3schema" format.

### Approach where user has access to the CustomResourceDefinition
`odo catalog service describe strimzi-cluster-operator.v0.21.1/Kafka -o json` would show a flat converted version of `openAPIv3schema` with same structure as `ClusterServiceVersion` with an extra addition of `type` information being present. 

The conversion approach would be 
- traverse the `openAPIv3schema` and if we find standard types (like string, integer) we add it to the descriptor listing with its path
- while traversing if we find any complex types like `object` or `array` we dont add then to the listing but traverse deeper.
- if we find an array we add `[*]` as a suffix of the path traversed so far, so that its obvious to the user that they need to add an index there. i.e. `zookeeper.env[*].key`

For human readable output -
```
- FieldPath: zookeeper.resources.limit.min
  DisplayName: Limit Min
  Description: <description>
  Type: string
- FieldPath: zookeeper.storage.type
  Type: string
- FieldPath: zookeeper.env[*].key
  Type: string
```

### Approach where user only has access to ClusterServiceVersion 

`odo catalog service describe strimzi-cluster-operator.v0.21.1/Kafka -o json` would show the `ClusterServiceVersion`'s `CRDDescription` shown in the `Fetch ClusterServiceVersion to generate the information` section

For human readable output a non tablular approach would be used.

```
- FieldPath: zookeeper.resources
   DisplayName: Zookeeper Resource Requirements
   Description: Limits describes the minimum/maximum amount of compute resources required/allowed
- FieldPath: zookeeper.storage.type
   DisplayName: ....
   ...

```



## "odo service create"

For any scenerio the user would provide the parameters in flat format i.e. `odo service create -p zookeeper.storage.type=ephemeral`. We would build the full parameters map using these flat key values then in case of `openapiv3schema`, validate it against the jsonschema. For `ClusterServiceVersion` we already have the flat path values so we validate them before building the parameters map.

### --from-file flag

`from-file` would work as it does now with just a difference where we would validate the file against jsonschema using the validator implemented.


