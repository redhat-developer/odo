---
title: JSON Output
# Put JSON output to the top
sidebar_position: 1
---

For `odo` to be used as a backend by graphical user interfaces (GUIs),
the useful commands can output their result in JSON format.

When used with the `-o json` flags, a command:
- that terminates successully, will:
  - terminate with a zero exit status,
  - will return its result in JSON format in its standard output stream.
- that terminates with an error, will:
  - terminate with a non-zero exit status,
  - will return an error message in its standard error stream, in the unique field `message` of a JSON object, as in `{ "message": "file not found" }`

The structures used to return information using JSON output are defined in [the `pkg/api` package](https://github.com/redhat-developer/odo/tree/main/pkg/api).

## odo analyze -o json

The `analyze` command analyzes the files in the current directory to select the best devfiles to use,
from the devfiles in the registries defined in the list of preferred registries with the command `odo preference view`.

The output of this command contains a list of devfile name and registry name:

```bash
odo analyze -o json
```
```json
[
	{
	    "devfile": "nodejs",
	    "devfileRegistry": "DefaultDevfileRegistry"
	}
]
```
```console
echo $?
```
```console
0
```

If the command is executed in an empty directory, it will return an error in the standard error stream and terminate with a non-zero exit status:

```bash
odo analyze -o json
```
```json
{
	"message": "No valid devfile found for project in /home/user/my/empty/directory"
}
```
```console
echo $?
```
```console
1
```

## odo init -o json

The `init` command downloads a devfile and, optionally, a starter project. The usage for this command can be found in the [odo init command reference page](init.md).

The output of this command contains the path of the downloaded devfile and its content, in JSON format.

```bash
$ odo init -o json \
    --name aname \
    --devfile go \
    --starter go-starter
```
```json
{
	"devfilePath": "/home/user/my-project/devfile.yaml",
	"devfileData": {
		"devfile": {
			"schemaVersion": "2.1.0",
      [...]
		},
		"supportedOdoFeatures": {
			"dev": true,
			"deploy": false,
			"debug": false
		}
	},
	"forwardedPorts": [],
	"runningIn": {
		"dev": false,
		"deploy": false
	},
	"managedBy": "odo"
}
```
```console
echo $?
```
```console
0
```

If the command fails, it will return an error in the standard error stream and terminate with a non-zero exit status:

```bash
# Executing the same command again will fail
$ odo init -o json \
    --name aname \
    --devfile go \
    --starter go-starter
```
```json
{
	"message": "a devfile already exists in the current directory"
}
```
```console
echo $?
```
```console
1
```

## odo describe component -o json

The `describe component` command returns information about a component, either the component
defined by a Devfile in the current directory, or a deployed component given its name and namespace.

When the `describe component` command is executed without parameter from a directory containing a Devfile, it will return:
- information about the Devfile
  - the path of the Devfile,
  - the content of the Devfile,
  - supported `odo` features, indicating if the Devfile defines necessary information to run `odo dev`, `odo dev --debug` and `odo deploy`
- the status of the component
  - the forwarded ports if odo is currently running in Dev mode,
  - the modes in which the component is deployed (either none, Dev, Deploy or both)

```bash
odo describe component -o json
```
```json
{
	"devfilePath": "/home/phmartin/Documents/tests/tmp/devfile.yaml",
	"devfileData": {
		"devfile": {
			"schemaVersion": "2.0.0",
			[ devfile.yaml file content ]
		},
		"supportedOdoFeatures": {
			"dev": true,
			"deploy": false,
			"debug": true
		}
	},
	"devForwardedPorts": [
		{
			"containerName": "runtime",
			"localAddress": "127.0.0.1",
			"localPort": 40001,
			"containerPort": 3000
		}
	],
	"runningIn": {
		"dev": true,
		"deploy": false
	},
	"managedBy": "odo"
}
```

When the `describe component` commmand is executed with a name and namespace, it will return:
- the modes in which the component is deployed (either Dev, Deploy or both)

The command with name and namespace is not able to return information about a component that has not been deployed. 

The command with name and namespace will never return information about the Devfile, even if a Devfile is present in the current directory.

The command with name and namespace will never return information about the forwarded ports, as the information resides in the directory of the Devfile.

```bash
odo describe component --name aname -o json
```
```json
{
	"runningIn": {
		"dev": true,
		"deploy": false
	},
	"managedBy": "odo"
}
```

## odo list -o json

The `odo list` command returns information about components running on a specific namespace, and defined in the local Devfile, if any.

The `components` field lists the components either deployed in the cluster, or defined in the local Devfile.

The `componentInDevfile` field gives the name of the component present in the `components` list that is defined in the local Devfile, or is empty if no local Devfile is present.

In this example, the `component2` component is running in Deploy mode, and the command has been executed from a directory containing a Devfile defining a `component1` component, not running.

```bash
odo list --namespace project1
```
```json
{
	"componentInDevfile": "component1",
	"components": [
		{
			"name": "component2",
			"managedBy": "odo",
			"runningIn": {
				"dev": false,
				"deploy": true
			},
			"projectType": "nodejs"
		},
		{
			"name": "component1",
			"managedBy": "",
			"runningIn": {
				"dev": false,
				"deploy": false
			},
			"projectType": "nodejs"
		}
	]
}
```

## odo registry -o json

The `odo registry` command lists all the Devfile stacks from Devfile registries. You can get the available flag in the [registry command reference](registry.md).

The default output will return information found into the registry index for stacks:

```shell
odo registry -o json
```
```json
[
	{
		"name": "python-django",
		"displayName": "Django",
		"description": "Python3.7 with Django",
		"registry": {
			"name": "DefaultDevfileRegistry",
			"url": "https://registry.devfile.io",
			"secure": false
		},
		"language": "python",
		"tags": [
			"Python",
			"pip",
			"Django"
		],
		"projectType": "django",
		"version": "1.0.0",
		"starterProjects": [
			"django-example"
		]
	}, [...]
]
```

Using the `--details` flag, you will also get information about the Devfile:

```shell
odo registry --details -o json
```
```json
[
	{
		"name": "python-django",
		"displayName": "Django",
		"description": "Python3.7 with Django",
		"registry": {
			"name": "DefaultDevfileRegistry",
			"url": "https://registry.devfile.io",
			"secure": false
		},
		"language": "python",
		"tags": [
			"Python",
			"pip",
			"Django"
		],
		"projectType": "django",
		"version": "1.0.0",
		"starterProjects": [
			"django-example"
		],
		"devfileData": {
			"devfile": {
				"schemaVersion": "2.0.0",
				[ devfile.yaml file content ]
			},
			"supportedOdoFeatures": {
				"dev": true,
				"deploy": false,
				"debug": true
			}
		},
	}, [...]
]
```

## odo list binding -o json

The `odo list binding` command lists all service binding resources deployed in the current namespace,
and all service binding resources declared in the Devfile, if executed from a component directory.

The names of the Service Binding resources declared in the current Devfile are listed in the `bindingsInDevfile`
field of the output.

If a Service Binding resource is found in the current namespace, it also displays the variables that can be used from
the component in the `status.bindingFiles` and/or `status.bindingEnvVars` fields.

### Examples

When a service binding resource is defined in the Devfile, and the component is not deployed, you get an output similar to:

```shell
odo list binding -o json
```
```json
{
	"bindingsInDevfile": [
		"my-nodejs-app-cluster-sample"
	],
	"bindings": [
		{
			"name": "my-nodejs-app-cluster-sample",
			"spec": {
				"application": {
					"kind": "Deployment",
					"name": "my-nodejs-app-app",
					"apiVersion": "apps/v1"
				},
				"services": [
					{
						"kind": "Cluster",
						"name": "cluster-sample",
						"apiVersion": "postgresql.k8s.enterprisedb.io/v1"
					}
				],
				"detectBindingResources": true,
				"bindAsFiles": true
			}
		}
	]
}

With the same Devfile, when `odo dev` is running, you get an output similar to
(note the `.bindings[*].status` field):


```shell
odo list binding -o json
```
```json
{
	"bindingsInDevfile": [
		"my-nodejs-app-cluster-sample"
	],
	"bindings": [
		{
			"name": "my-nodejs-app-cluster-sample",
			"spec": {
				"application": {
					"kind": "Deployment",
					"name": "my-nodejs-app-app",
					"apiVersion": "apps/v1"
				},
				"services": [
					{
						"kind": "Cluster",
						"name": "cluster-sample",
						"apiVersion": "postgresql.k8s.enterprisedb.io/v1"
					}
				],
				"detectBindingResources": true,
				"bindAsFiles": true
			},
			"status": {
				"bindingFiles": [
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/database",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/host",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/pgpass",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/provider",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/type",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/username",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/ca.crt",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/ca.key",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/clusterIP",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/password",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/tls.crt",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/tls.key"
				],
				"runningIn": {
					"dev": true,
					"deploy": false,
				}
			}
		}
	]
}
```

When `odo dev` is running, if you execute the command from a directory without Devfile,
you get an output similar to (note that the `.bindingsInDevfile` field is not present anymore):


```shell
odo list binding -o json
```
```json
{
	"bindings": [
		{
			"name": "my-nodejs-app-cluster-sample",
			"spec": {
				"application": {
					"kind": "Deployment",
					"name": "my-nodejs-app-app",
					"apiVersion": "apps/v1"
				},
				"services": [
					{
						"kind": "Cluster",
						"name": "cluster-sample",
						"apiVersion": "postgresql.k8s.enterprisedb.io/v1"
					}
				],
				"detectBindingResources": true,
				"bindAsFiles": true
			},
			"status": {
				"bindingFiles": [
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/database",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/host",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/pgpass",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/provider",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/type",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/username",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/ca.crt",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/ca.key",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/clusterIP",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/password",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/tls.crt",
					"${SERVICE_BINDING_ROOT}/my-nodejs-app-cluster-sample/tls.key"
				],
				"runningIn": {
					"dev": true,
					"deploy": false
				}
			}
		}
	]
}
```


## odo describe binding -o json

The `odo describe binding` command lists all the service binding resources declared in the devfile
and, if the resource is deployed to the cluster, also displays the variables that can be used from
the component.

If a name is given, the command does not extract information from the Devfile, but instead extracts
information from the deployed resource with the given name.

Without a name, the output of the command is a list of service binding details, for example:

```shell
odo describe binding -o json
```
```json
[
	{
		"name": "my-first-binding",
		"spec": {
			"application": {
				"kind": "Deployment",
				"name": "my-nodejs-app-app",
				"apiVersion": "apps/v1"
			},
			"services": [
				{
					"apiVersion": "postgresql.k8s.enterprisedb.io/v1",
					"kind": "Cluster",
					"name": "cluster-sample",
					"namespace": "shared-services-ns"
				}
			],
			"detectBindingResources": false,
			"bindAsFiles": true,
			"namingStrategy": "lowercase"
		},
		"status": {
			"bindingFiles": [
				"${SERVICE_BINDING_ROOT}/my-first-binding/host",
				"${SERVICE_BINDING_ROOT}/my-first-binding/password",
				"${SERVICE_BINDING_ROOT}/my-first-binding/pgpass",
				"${SERVICE_BINDING_ROOT}/my-first-binding/provider",
				"${SERVICE_BINDING_ROOT}/my-first-binding/type",
				"${SERVICE_BINDING_ROOT}/my-first-binding/username",
				"${SERVICE_BINDING_ROOT}/my-first-binding/database"
			],
			"bindingEnvVars": [
				"PASSWD"
			]
		}
	},
	{
		"name": "my-second-binding",
		"spec": {
			"application": {
				"kind": "Deployment",
				"name": "my-nodejs-app-app",
				"apiVersion": "apps/v1"
			},
			"services": [
				{
					"apiVersion": "postgresql.k8s.enterprisedb.io/v1",
					"kind": "Cluster",
					"name": "cluster-sample-2"
				}
			],
			"detectBindingResources": true,
			"bindAsFiles": true
		},
		"status": {
			"bindingFiles": [
				"${SERVICE_BINDING_ROOT}/my-second-binding/ca.crt",
				"${SERVICE_BINDING_ROOT}/my-second-binding/clusterIP",
				"${SERVICE_BINDING_ROOT}/my-second-binding/database",
				"${SERVICE_BINDING_ROOT}/my-second-binding/host",
				"${SERVICE_BINDING_ROOT}/my-second-binding/ca.key",
				"${SERVICE_BINDING_ROOT}/my-second-binding/password",
				"${SERVICE_BINDING_ROOT}/my-second-binding/pgpass",
				"${SERVICE_BINDING_ROOT}/my-second-binding/provider",
				"${SERVICE_BINDING_ROOT}/my-second-binding/tls.crt",
				"${SERVICE_BINDING_ROOT}/my-second-binding/tls.key",
				"${SERVICE_BINDING_ROOT}/my-second-binding/type",
				"${SERVICE_BINDING_ROOT}/my-second-binding/username"
			]
		}
	}
]
```

When specifying a name, the output is a unique service binding:

```shell
odo describe binding --name my-first-binding -o json
```
```json
{
	"name": "my-first-binding",
	"spec": {
			"application": {
				"kind": "Deployment",
				"name": "my-nodejs-app-app",
				"apiVersion": "apps/v1"
			},
		"services": [
			{
				"apiVersion": "postgresql.k8s.enterprisedb.io/v1",
				"kind": "Cluster",
				"name": "cluster-sample",
                "namespace": "shared-services-ns"
			}
		],
		"detectBindingResources": false,
		"bindAsFiles": true
	},
	"status": {
		"bindingFiles": [
			"${SERVICE_BINDING_ROOT}/my-first-binding/host",
			"${SERVICE_BINDING_ROOT}/my-first-binding/password",
			"${SERVICE_BINDING_ROOT}/my-first-binding/pgpass",
			"${SERVICE_BINDING_ROOT}/my-first-binding/provider",
			"${SERVICE_BINDING_ROOT}/my-first-binding/type",
			"${SERVICE_BINDING_ROOT}/my-first-binding/username",
			"${SERVICE_BINDING_ROOT}/my-first-binding/database"
		],
		"bindingEnvVars": [
			"PASSWD"
		]
	}
}
```

## odo preference view -o json

The `odo preference view` command lists all user preferences and all user Devfile registries.


```shell
odo preference view -o json
```
```json
{
	"preferences": [
		{
			"name": "UpdateNotification",
			"value": null,
			"default": true,
			"type": "bool",
			"description": "Flag to control if an update notification is shown or not (Default: true)"
		},
		{
			"name": "Timeout",
			"value": null,
			"default": 1000000000,
			"type": "int64",
			"description": "Timeout (in Duration) for cluster server connection check (Default: 1s)"
		},
		{
			"name": "PushTimeout",
			"value": null,
			"default": 240000000000,
			"type": "int64",
			"description": "PushTimeout (in Duration) for waiting for a Pod to come up (Default: 4m0s)"
		},
		{
			"name": "RegistryCacheTime",
			"value": null,
			"default": 900000000000,
			"type": "int64",
			"description": "For how long (in Duration) odo will cache information from the Devfile registry (Default: 15m0s)"
		},
		{
			"name": "ConsentTelemetry",
			"value": false,
			"default": false,
			"type": "bool",
			"description": "If true, odo will collect telemetry for the user's odo usage (Default: false)\n\t\t    For more information: https://developers.redhat.com/article/tool-data-collection"
		},
		{
			"name": "Ephemeral",
			"value": null,
			"default": false,
			"type": "bool",
			"description": "If true, odo will create an emptyDir volume to store source code (Default: false)"
		}
	],
	"registries": [
		{
			"name": "DefaultDevfileRegistry",
			"url": "https://registry.devfile.io",
			"secure": false
		}
	]
}
```

## odo list services -o json

The `odo list services` command lists all the bindable Opeartor backed services available in the current 
project/namespace.
```shell
odo list services -o json
```
```shell
$ odo list services -o json   
{
	"bindableServices": [
		{
			"name": "cluster-sample/Cluster.postgresql.k8s.enterprisedb.io",
			"namespace": "myproject"
		}
	]
}
```
You can also list all the bindable Operator backed services from a different project/namespace that you have access to:
```shell
odo list services -o json -n <project-name>
```
```shell
$ odo list services -o json -n newproject
{
	"bindableServices": [
		{
			"name": "hello-world/RabbitmqCluster.rabbitmq.com",
			"namespace": "newproject"
		}
	]
}
```
Finally, if you want to list all bindable Operator backed services from all projects/namespaces you have access to, 
use `-A` or `--all-namespaces` flag:
```shell
odo list services -o json -A
```
```shell
$ odo list services -o json -A
{
	"bindableServices": [
		{
			"name": "cluster-sample/Cluster.postgresql.k8s.enterprisedb.io",
			"namespace": "myproject"
		},
		{
			"name": "hello-world/RabbitmqCluster.rabbitmq.com",
			"namespace": "newproject"
		}
	]
}

```