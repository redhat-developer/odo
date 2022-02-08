package version220

// https://raw.githubusercontent.com/devfile/api/main/schemas/latest/devfile.json
const JsonSchema220 = `{
  "description": "Devfile describes the structure of a cloud-native devworkspace and development environment.",
  "type": "object",
  "title": "Devfile schema - Version 2.2.0-alpha",
  "required": [
    "schemaVersion"
  ],
  "properties": {
    "attributes": {
      "description": "Map of implementation-dependant free-form YAML attributes.",
      "type": "object",
      "additionalProperties": true
    },
    "commands": {
      "description": "Predefined, ready-to-use, devworkspace-related commands",
      "type": "array",
      "items": {
        "type": "object",
        "required": [
          "id"
        ],
        "oneOf": [
          {
            "required": [
              "exec"
            ]
          },
          {
            "required": [
              "apply"
            ]
          },
          {
            "required": [
              "composite"
            ]
          }
        ],
        "properties": {
          "apply": {
            "description": "Command that consists in applying a given component definition, typically bound to a devworkspace event.\n\nFor example, when an 'apply' command is bound to a 'preStart' event, and references a 'container' component, it will start the container as a K8S initContainer in the devworkspace POD, unless the component has its 'dedicatedPod' field set to 'true'.\n\nWhen no 'apply' command exist for a given component, it is assumed the component will be applied at devworkspace start by default, unless 'deployByDefault' for that component is set to false.",
            "type": "object",
            "required": [
              "component"
            ],
            "properties": {
              "component": {
                "description": "Describes component that will be applied",
                "type": "string"
              },
              "group": {
                "description": "Defines the group this command is part of",
                "type": "object",
                "required": [
                  "kind"
                ],
                "properties": {
                  "isDefault": {
                    "description": "Identifies the default command for a given group kind",
                    "type": "boolean"
                  },
                  "kind": {
                    "description": "Kind of group the command is part of",
                    "type": "string",
                    "enum": [
                      "build",
                      "run",
                      "test",
                      "debug",
                      "deploy"
                    ]
                  }
                },
                "additionalProperties": false
              },
              "label": {
                "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                "type": "string"
              }
            },
            "additionalProperties": false
          },
          "attributes": {
            "description": "Map of implementation-dependant free-form YAML attributes.",
            "type": "object",
            "additionalProperties": true
          },
          "composite": {
            "description": "Composite command that allows executing several sub-commands either sequentially or concurrently",
            "type": "object",
            "properties": {
              "commands": {
                "description": "The commands that comprise this composite command",
                "type": "array",
                "items": {
                  "type": "string"
                }
              },
              "group": {
                "description": "Defines the group this command is part of",
                "type": "object",
                "required": [
                  "kind"
                ],
                "properties": {
                  "isDefault": {
                    "description": "Identifies the default command for a given group kind",
                    "type": "boolean"
                  },
                  "kind": {
                    "description": "Kind of group the command is part of",
                    "type": "string",
                    "enum": [
                      "build",
                      "run",
                      "test",
                      "debug",
                      "deploy"
                    ]
                  }
                },
                "additionalProperties": false
              },
              "label": {
                "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                "type": "string"
              },
              "parallel": {
                "description": "Indicates if the sub-commands should be executed concurrently",
                "type": "boolean"
              }
            },
            "additionalProperties": false
          },
          "exec": {
            "description": "CLI Command executed in an existing component container",
            "type": "object",
            "required": [
              "commandLine",
              "component"
            ],
            "properties": {
              "commandLine": {
                "description": "The actual command-line string\n\nSpecial variables that can be used:\n\n - '$PROJECTS_ROOT': A path where projects sources are mounted as defined by container component's sourceMapping.\n\n - '$PROJECT_SOURCE': A path to a project source ($PROJECTS_ROOT/\u003cproject-name\u003e). If there are multiple projects, this will point to the directory of the first one.",
                "type": "string"
              },
              "component": {
                "description": "Describes component to which given action relates",
                "type": "string"
              },
              "env": {
                "description": "Optional list of environment variables that have to be set before running the command",
                "type": "array",
                "items": {
                  "type": "object",
                  "required": [
                    "name",
                    "value"
                  ],
                  "properties": {
                    "name": {
                      "type": "string"
                    },
                    "value": {
                      "type": "string"
                    }
                  },
                  "additionalProperties": false
                }
              },
              "group": {
                "description": "Defines the group this command is part of",
                "type": "object",
                "required": [
                  "kind"
                ],
                "properties": {
                  "isDefault": {
                    "description": "Identifies the default command for a given group kind",
                    "type": "boolean"
                  },
                  "kind": {
                    "description": "Kind of group the command is part of",
                    "type": "string",
                    "enum": [
                      "build",
                      "run",
                      "test",
                      "debug",
                      "deploy"
                    ]
                  }
                },
                "additionalProperties": false
              },
              "hotReloadCapable": {
                "description": "Whether the command is capable to reload itself when source code changes. If set to 'true' the command won't be restarted and it is expected to handle file changes on its own.\n\nDefault value is 'false'",
                "type": "boolean"
              },
              "label": {
                "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                "type": "string"
              },
              "workingDir": {
                "description": "Working directory where the command should be executed\n\nSpecial variables that can be used:\n\n - '$PROJECTS_ROOT': A path where projects sources are mounted as defined by container component's sourceMapping.\n\n - '$PROJECT_SOURCE': A path to a project source ($PROJECTS_ROOT/\u003cproject-name\u003e). If there are multiple projects, this will point to the directory of the first one.",
                "type": "string"
              }
            },
            "additionalProperties": false
          },
          "id": {
            "description": "Mandatory identifier that allows referencing this command in composite commands, from a parent, or in events.",
            "type": "string",
            "maxLength": 63,
            "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
          }
        },
        "additionalProperties": false
      }
    },
    "components": {
      "description": "List of the devworkspace components, such as editor and plugins, user-provided containers, or other types of components",
      "type": "array",
      "items": {
        "type": "object",
        "required": [
          "name"
        ],
        "oneOf": [
          {
            "required": [
              "container"
            ]
          },
          {
            "required": [
              "kubernetes"
            ]
          },
          {
            "required": [
              "openshift"
            ]
          },
          {
            "required": [
              "volume"
            ]
          },
          {
            "required": [
              "image"
            ]
          }
        ],
        "properties": {
          "attributes": {
            "description": "Map of implementation-dependant free-form YAML attributes.",
            "type": "object",
            "additionalProperties": true
          },
          "container": {
            "description": "Allows adding and configuring devworkspace-related containers",
            "type": "object",
            "required": [
              "image"
            ],
            "properties": {
              "annotation": {
                "description": "Annotations that should be added to specific resources for this container",
                "type": "object",
                "properties": {
                  "deployment": {
                    "description": "Annotations to be added to deployment",
                    "type": "object",
                    "additionalProperties": {
                      "type": "string"
                    }
                  },
                  "service": {
                    "description": "Annotations to be added to service",
                    "type": "object",
                    "additionalProperties": {
                      "type": "string"
                    }
                  }
                },
                "additionalProperties": false
              },
              "args": {
                "description": "The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command.\n\nDefaults to an empty array, meaning use whatever is defined in the image.",
                "type": "array",
                "items": {
                  "type": "string"
                }
              },
              "command": {
                "description": "The command to run in the dockerimage component instead of the default one provided in the image.\n\nDefaults to an empty array, meaning use whatever is defined in the image.",
                "type": "array",
                "items": {
                  "type": "string"
                }
              },
              "cpuLimit": {
                "type": "string"
              },
              "cpuRequest": {
                "type": "string"
              },
              "dedicatedPod": {
                "description": "Specify if a container should run in its own separated pod, instead of running as part of the main development environment pod.\n\nDefault value is 'false'",
                "type": "boolean"
              },
              "endpoints": {
                "type": "array",
                "items": {
                  "type": "object",
                  "required": [
                    "name",
                    "targetPort"
                  ],
                  "properties": {
                    "annotation": {
                      "description": "Annotations to be added to Kubernetes Ingress or Openshift Route",
                      "type": "object",
                      "additionalProperties": {
                        "type": "string"
                      }
                    },
                    "attributes": {
                      "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                      "type": "object",
                      "additionalProperties": true
                    },
                    "exposure": {
                      "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main devworkspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main devworkspace POD, on a local address.\n\nDefault value is 'public'",
                      "type": "string",
                      "default": "public",
                      "enum": [
                        "public",
                        "internal",
                        "none"
                      ]
                    },
                    "name": {
                      "type": "string",
                      "maxLength": 15,
                      "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
                    },
                    "path": {
                      "description": "Path of the endpoint URL",
                      "type": "string"
                    },
                    "protocol": {
                      "description": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'",
                      "type": "string",
                      "default": "http",
                      "enum": [
                        "http",
                        "https",
                        "ws",
                        "wss",
                        "tcp",
                        "udp"
                      ]
                    },
                    "secure": {
                      "description": "Describes whether the endpoint should be secured and protected by some authentication process. This requires a protocol of 'https' or 'wss'.",
                      "type": "boolean"
                    },
                    "targetPort": {
                      "description": "The port number should be unique.",
                      "type": "integer"
                    }
                  },
                  "additionalProperties": false
                }
              },
              "env": {
                "description": "Environment variables used in this container.\n\nThe following variables are reserved and cannot be overridden via env:\n\n - '$PROJECTS_ROOT'\n\n - '$PROJECT_SOURCE'",
                "type": "array",
                "items": {
                  "type": "object",
                  "required": [
                    "name",
                    "value"
                  ],
                  "properties": {
                    "name": {
                      "type": "string"
                    },
                    "value": {
                      "type": "string"
                    }
                  },
                  "additionalProperties": false
                }
              },
              "image": {
                "type": "string"
              },
              "memoryLimit": {
                "type": "string"
              },
              "memoryRequest": {
                "type": "string"
              },
              "mountSources": {
                "description": "Toggles whether or not the project source code should be mounted in the component.\n\nDefaults to true for all component types except plugins and components that set 'dedicatedPod' to true.",
                "type": "boolean"
              },
              "sourceMapping": {
                "description": "Optional specification of the path in the container where project sources should be transferred/mounted when 'mountSources' is 'true'. When omitted, the default value of /projects is used.",
                "type": "string",
                "default": "/projects"
              },
              "volumeMounts": {
                "description": "List of volumes mounts that should be mounted is this container.",
                "type": "array",
                "items": {
                  "description": "Volume that should be mounted to a component container",
                  "type": "object",
                  "required": [
                    "name"
                  ],
                  "properties": {
                    "name": {
                      "description": "The volume mount name is the name of an existing 'Volume' component. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files.",
                      "type": "string",
                      "maxLength": 63,
                      "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
                    },
                    "path": {
                      "description": "The path in the component container where the volume should be mounted. If not path is mentioned, default path is the is '/\u003cname\u003e'.",
                      "type": "string"
                    }
                  },
                  "additionalProperties": false
                }
              }
            },
            "additionalProperties": false
          },
          "image": {
            "description": "Allows specifying the definition of an image for outer loop builds",
            "type": "object",
            "required": [
              "imageName"
            ],
            "oneOf": [
              {
                "required": [
                  "dockerfile"
                ]
              }
            ],
            "properties": {
              "autoBuild": {
                "description": "Defines if the image should be built during startup.\n\nDefault value is 'false'",
                "type": "boolean"
              },
              "dockerfile": {
                "description": "Allows specifying dockerfile type build",
                "type": "object",
                "oneOf": [
                  {
                    "required": [
                      "uri"
                    ]
                  },
                  {
                    "required": [
                      "devfileRegistry"
                    ]
                  },
                  {
                    "required": [
                      "git"
                    ]
                  }
                ],
                "properties": {
                  "args": {
                    "description": "The arguments to supply to the dockerfile build.",
                    "type": "array",
                    "items": {
                      "type": "string"
                    }
                  },
                  "buildContext": {
                    "description": "Path of source directory to establish build context. Defaults to ${PROJECT_ROOT} in the container",
                    "type": "string"
                  },
                  "devfileRegistry": {
                    "description": "Dockerfile's Devfile Registry source",
                    "type": "object",
                    "required": [
                      "id"
                    ],
                    "properties": {
                      "id": {
                        "description": "Id in a devfile registry that contains a Dockerfile. The src in the OCI registry required for the Dockerfile build will be downloaded for building the image.",
                        "type": "string"
                      },
                      "registryUrl": {
                        "description": "Devfile Registry URL to pull the Dockerfile from when using the Devfile Registry as Dockerfile src. To ensure the Dockerfile gets resolved consistently in different environments, it is recommended to always specify the 'devfileRegistryUrl' when 'Id' is used.",
                        "type": "string"
                      }
                    },
                    "additionalProperties": false
                  },
                  "git": {
                    "description": "Dockerfile's Git source",
                    "type": "object",
                    "required": [
                      "remotes"
                    ],
                    "properties": {
                      "checkoutFrom": {
                        "description": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                        "type": "object",
                        "properties": {
                          "remote": {
                            "description": "The remote name should be used as init. Required if there are more than one remote configured",
                            "type": "string"
                          },
                          "revision": {
                            "description": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found.",
                            "type": "string"
                          }
                        },
                        "additionalProperties": false
                      },
                      "fileLocation": {
                        "description": "Location of the Dockerfile in the Git repository when using git as Dockerfile src. Defaults to Dockerfile.",
                        "type": "string"
                      },
                      "remotes": {
                        "description": "The remotes map which should be initialized in the git project. Projects must have at least one remote configured while StarterProjects \u0026 Image Component's Git source can only have at most one remote configured.",
                        "type": "object",
                        "additionalProperties": {
                          "type": "string"
                        }
                      }
                    },
                    "additionalProperties": false
                  },
                  "rootRequired": {
                    "description": "Specify if a privileged builder pod is required.\n\nDefault value is 'false'",
                    "type": "boolean"
                  },
                  "uri": {
                    "description": "URI Reference of a Dockerfile. It can be a full URL or a relative URI from the current devfile as the base URI.",
                    "type": "string"
                  }
                },
                "additionalProperties": false
              },
              "imageName": {
                "description": "Name of the image for the resulting outerloop build",
                "type": "string"
              }
            },
            "additionalProperties": false
          },
          "kubernetes": {
            "description": "Allows importing into the devworkspace the Kubernetes resources defined in a given manifest. For example this allows reusing the Kubernetes definitions used to deploy some runtime components in production.",
            "type": "object",
            "oneOf": [
              {
                "required": [
                  "uri"
                ]
              },
              {
                "required": [
                  "inlined"
                ]
              }
            ],
            "properties": {
              "deployByDefault": {
                "description": "Defines if the component should be deployed during startup.\n\nDefault value is 'false'",
                "type": "boolean"
              },
              "endpoints": {
                "type": "array",
                "items": {
                  "type": "object",
                  "required": [
                    "name",
                    "targetPort"
                  ],
                  "properties": {
                    "annotation": {
                      "description": "Annotations to be added to Kubernetes Ingress or Openshift Route",
                      "type": "object",
                      "additionalProperties": {
                        "type": "string"
                      }
                    },
                    "attributes": {
                      "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                      "type": "object",
                      "additionalProperties": true
                    },
                    "exposure": {
                      "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main devworkspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main devworkspace POD, on a local address.\n\nDefault value is 'public'",
                      "type": "string",
                      "default": "public",
                      "enum": [
                        "public",
                        "internal",
                        "none"
                      ]
                    },
                    "name": {
                      "type": "string",
                      "maxLength": 15,
                      "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
                    },
                    "path": {
                      "description": "Path of the endpoint URL",
                      "type": "string"
                    },
                    "protocol": {
                      "description": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'",
                      "type": "string",
                      "default": "http",
                      "enum": [
                        "http",
                        "https",
                        "ws",
                        "wss",
                        "tcp",
                        "udp"
                      ]
                    },
                    "secure": {
                      "description": "Describes whether the endpoint should be secured and protected by some authentication process. This requires a protocol of 'https' or 'wss'.",
                      "type": "boolean"
                    },
                    "targetPort": {
                      "description": "The port number should be unique.",
                      "type": "integer"
                    }
                  },
                  "additionalProperties": false
                }
              },
              "inlined": {
                "description": "Inlined manifest",
                "type": "string"
              },
              "uri": {
                "description": "Location in a file fetched from a uri.",
                "type": "string"
              }
            },
            "additionalProperties": false
          },
          "name": {
            "description": "Mandatory name that allows referencing the component from other elements (such as commands) or from an external devfile that may reference this component through a parent or a plugin.",
            "type": "string",
            "maxLength": 63,
            "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
          },
          "openshift": {
            "description": "Allows importing into the devworkspace the OpenShift resources defined in a given manifest. For example this allows reusing the OpenShift definitions used to deploy some runtime components in production.",
            "type": "object",
            "oneOf": [
              {
                "required": [
                  "uri"
                ]
              },
              {
                "required": [
                  "inlined"
                ]
              }
            ],
            "properties": {
              "deployByDefault": {
                "description": "Defines if the component should be deployed during startup.\n\nDefault value is 'false'",
                "type": "boolean"
              },
              "endpoints": {
                "type": "array",
                "items": {
                  "type": "object",
                  "required": [
                    "name",
                    "targetPort"
                  ],
                  "properties": {
                    "annotation": {
                      "description": "Annotations to be added to Kubernetes Ingress or Openshift Route",
                      "type": "object",
                      "additionalProperties": {
                        "type": "string"
                      }
                    },
                    "attributes": {
                      "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                      "type": "object",
                      "additionalProperties": true
                    },
                    "exposure": {
                      "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main devworkspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main devworkspace POD, on a local address.\n\nDefault value is 'public'",
                      "type": "string",
                      "default": "public",
                      "enum": [
                        "public",
                        "internal",
                        "none"
                      ]
                    },
                    "name": {
                      "type": "string",
                      "maxLength": 15,
                      "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
                    },
                    "path": {
                      "description": "Path of the endpoint URL",
                      "type": "string"
                    },
                    "protocol": {
                      "description": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'",
                      "type": "string",
                      "default": "http",
                      "enum": [
                        "http",
                        "https",
                        "ws",
                        "wss",
                        "tcp",
                        "udp"
                      ]
                    },
                    "secure": {
                      "description": "Describes whether the endpoint should be secured and protected by some authentication process. This requires a protocol of 'https' or 'wss'.",
                      "type": "boolean"
                    },
                    "targetPort": {
                      "description": "The port number should be unique.",
                      "type": "integer"
                    }
                  },
                  "additionalProperties": false
                }
              },
              "inlined": {
                "description": "Inlined manifest",
                "type": "string"
              },
              "uri": {
                "description": "Location in a file fetched from a uri.",
                "type": "string"
              }
            },
            "additionalProperties": false
          },
          "volume": {
            "description": "Allows specifying the definition of a volume shared by several other components",
            "type": "object",
            "properties": {
              "ephemeral": {
                "description": "Ephemeral volumes are not stored persistently across restarts. Defaults to false",
                "type": "boolean"
              },
              "size": {
                "description": "Size of the volume",
                "type": "string"
              }
            },
            "additionalProperties": false
          }
        },
        "additionalProperties": false
      }
    },
    "events": {
      "description": "Bindings of commands to events. Each command is referred-to by its name.",
      "type": "object",
      "properties": {
        "postStart": {
          "description": "IDs of commands that should be executed after the devworkspace is completely started. In the case of Che-Theia, these commands should be executed after all plugins and extensions have started, including project cloning. This means that those commands are not triggered until the user opens the IDE in his browser.",
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "postStop": {
          "description": "IDs of commands that should be executed after stopping the devworkspace.",
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "preStart": {
          "description": "IDs of commands that should be executed before the devworkspace start. Kubernetes-wise, these commands would typically be executed in init containers of the devworkspace POD.",
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "preStop": {
          "description": "IDs of commands that should be executed before stopping the devworkspace.",
          "type": "array",
          "items": {
            "type": "string"
          }
        }
      },
      "additionalProperties": false
    },
    "metadata": {
      "description": "Optional metadata",
      "type": "object",
      "properties": {
        "architectures": {
          "description": "Optional list of processor architectures that the devfile supports, empty list suggests that the devfile can be used on any architecture",
          "type": "array",
          "uniqueItems": true,
          "items": {
            "description": "Architecture describes the architecture type",
            "type": "string",
            "enum": [
              "amd64",
              "arm64",
              "ppc64le",
              "s390x"
            ]
          }
        },
        "attributes": {
          "description": "Map of implementation-dependant free-form YAML attributes. Deprecated, use the top-level attributes field instead.",
          "type": "object",
          "additionalProperties": true
        },
        "description": {
          "description": "Optional devfile description",
          "type": "string"
        },
        "displayName": {
          "description": "Optional devfile display name",
          "type": "string"
        },
        "globalMemoryLimit": {
          "description": "Optional devfile global memory limit",
          "type": "string"
        },
        "icon": {
          "description": "Optional devfile icon, can be a URI or a relative path in the project",
          "type": "string"
        },
        "language": {
          "description": "Optional devfile language",
          "type": "string"
        },
        "name": {
          "description": "Optional devfile name",
          "type": "string"
        },
        "projectType": {
          "description": "Optional devfile project type",
          "type": "string"
        },
        "provider": {
          "description": "Optional devfile provider information",
          "type": "string"
        },
        "supportUrl": {
          "description": "Optional link to a page that provides support information",
          "type": "string"
        },
        "tags": {
          "description": "Optional devfile tags",
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "version": {
          "description": "Optional semver-compatible version",
          "type": "string",
          "pattern": "^([0-9]+)\\.([0-9]+)\\.([0-9]+)(\\-[0-9a-z-]+(\\.[0-9a-z-]+)*)?(\\+[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?$"
        },
        "website": {
          "description": "Optional devfile website",
          "type": "string"
        }
      },
      "additionalProperties": true
    },
    "parent": {
      "description": "Parent devworkspace template",
      "type": "object",
      "oneOf": [
        {
          "required": [
            "uri"
          ]
        },
        {
          "required": [
            "id"
          ]
        },
        {
          "required": [
            "kubernetes"
          ]
        }
      ],
      "properties": {
        "attributes": {
          "description": "Overrides of attributes encapsulated in a parent devfile. Overriding is done according to K8S strategic merge patch standard rules.",
          "type": "object",
          "additionalProperties": true
        },
        "commands": {
          "description": "Overrides of commands encapsulated in a parent devfile or a plugin. Overriding is done according to K8S strategic merge patch standard rules.",
          "type": "array",
          "items": {
            "type": "object",
            "required": [
              "id"
            ],
            "oneOf": [
              {
                "required": [
                  "exec"
                ]
              },
              {
                "required": [
                  "apply"
                ]
              },
              {
                "required": [
                  "composite"
                ]
              }
            ],
            "properties": {
              "apply": {
                "description": "Command that consists in applying a given component definition, typically bound to a devworkspace event.\n\nFor example, when an 'apply' command is bound to a 'preStart' event, and references a 'container' component, it will start the container as a K8S initContainer in the devworkspace POD, unless the component has its 'dedicatedPod' field set to 'true'.\n\nWhen no 'apply' command exist for a given component, it is assumed the component will be applied at devworkspace start by default, unless 'deployByDefault' for that component is set to false.",
                "type": "object",
                "properties": {
                  "component": {
                    "description": "Describes component that will be applied",
                    "type": "string"
                  },
                  "group": {
                    "description": "Defines the group this command is part of",
                    "type": "object",
                    "properties": {
                      "isDefault": {
                        "description": "Identifies the default command for a given group kind",
                        "type": "boolean"
                      },
                      "kind": {
                        "description": "Kind of group the command is part of",
                        "type": "string",
                        "enum": [
                          "build",
                          "run",
                          "test",
                          "debug",
                          "deploy"
                        ]
                      }
                    },
                    "additionalProperties": false
                  },
                  "label": {
                    "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                    "type": "string"
                  }
                },
                "additionalProperties": false
              },
              "attributes": {
                "description": "Map of implementation-dependant free-form YAML attributes.",
                "type": "object",
                "additionalProperties": true
              },
              "composite": {
                "description": "Composite command that allows executing several sub-commands either sequentially or concurrently",
                "type": "object",
                "properties": {
                  "commands": {
                    "description": "The commands that comprise this composite command",
                    "type": "array",
                    "items": {
                      "type": "string"
                    }
                  },
                  "group": {
                    "description": "Defines the group this command is part of",
                    "type": "object",
                    "properties": {
                      "isDefault": {
                        "description": "Identifies the default command for a given group kind",
                        "type": "boolean"
                      },
                      "kind": {
                        "description": "Kind of group the command is part of",
                        "type": "string",
                        "enum": [
                          "build",
                          "run",
                          "test",
                          "debug",
                          "deploy"
                        ]
                      }
                    },
                    "additionalProperties": false
                  },
                  "label": {
                    "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                    "type": "string"
                  },
                  "parallel": {
                    "description": "Indicates if the sub-commands should be executed concurrently",
                    "type": "boolean"
                  }
                },
                "additionalProperties": false
              },
              "exec": {
                "description": "CLI Command executed in an existing component container",
                "type": "object",
                "properties": {
                  "commandLine": {
                    "description": "The actual command-line string\n\nSpecial variables that can be used:\n\n - '$PROJECTS_ROOT': A path where projects sources are mounted as defined by container component's sourceMapping.\n\n - '$PROJECT_SOURCE': A path to a project source ($PROJECTS_ROOT/\u003cproject-name\u003e). If there are multiple projects, this will point to the directory of the first one.",
                    "type": "string"
                  },
                  "component": {
                    "description": "Describes component to which given action relates",
                    "type": "string"
                  },
                  "env": {
                    "description": "Optional list of environment variables that have to be set before running the command",
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
                        }
                      },
                      "additionalProperties": false
                    }
                  },
                  "group": {
                    "description": "Defines the group this command is part of",
                    "type": "object",
                    "properties": {
                      "isDefault": {
                        "description": "Identifies the default command for a given group kind",
                        "type": "boolean"
                      },
                      "kind": {
                        "description": "Kind of group the command is part of",
                        "type": "string",
                        "enum": [
                          "build",
                          "run",
                          "test",
                          "debug",
                          "deploy"
                        ]
                      }
                    },
                    "additionalProperties": false
                  },
                  "hotReloadCapable": {
                    "description": "Whether the command is capable to reload itself when source code changes. If set to 'true' the command won't be restarted and it is expected to handle file changes on its own.\n\nDefault value is 'false'",
                    "type": "boolean"
                  },
                  "label": {
                    "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                    "type": "string"
                  },
                  "workingDir": {
                    "description": "Working directory where the command should be executed\n\nSpecial variables that can be used:\n\n - '$PROJECTS_ROOT': A path where projects sources are mounted as defined by container component's sourceMapping.\n\n - '$PROJECT_SOURCE': A path to a project source ($PROJECTS_ROOT/\u003cproject-name\u003e). If there are multiple projects, this will point to the directory of the first one.",
                    "type": "string"
                  }
                },
                "additionalProperties": false
              },
              "id": {
                "description": "Mandatory identifier that allows referencing this command in composite commands, from a parent, or in events.",
                "type": "string",
                "maxLength": 63,
                "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
              }
            },
            "additionalProperties": false
          }
        },
        "components": {
          "description": "Overrides of components encapsulated in a parent devfile or a plugin. Overriding is done according to K8S strategic merge patch standard rules.",
          "type": "array",
          "items": {
            "type": "object",
            "required": [
              "name"
            ],
            "oneOf": [
              {
                "required": [
                  "container"
                ]
              },
              {
                "required": [
                  "kubernetes"
                ]
              },
              {
                "required": [
                  "openshift"
                ]
              },
              {
                "required": [
                  "volume"
                ]
              },
              {
                "required": [
                  "image"
                ]
              }
            ],
            "properties": {
              "attributes": {
                "description": "Map of implementation-dependant free-form YAML attributes.",
                "type": "object",
                "additionalProperties": true
              },
              "container": {
                "description": "Allows adding and configuring devworkspace-related containers",
                "type": "object",
                "properties": {
                  "annotation": {
                    "description": "Annotations that should be added to specific resources for this container",
                    "type": "object",
                    "properties": {
                      "deployment": {
                        "description": "Annotations to be added to deployment",
                        "type": "object",
                        "additionalProperties": {
                          "type": "string"
                        }
                      },
                      "service": {
                        "description": "Annotations to be added to service",
                        "type": "object",
                        "additionalProperties": {
                          "type": "string"
                        }
                      }
                    },
                    "additionalProperties": false
                  },
                  "args": {
                    "description": "The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command.\n\nDefaults to an empty array, meaning use whatever is defined in the image.",
                    "type": "array",
                    "items": {
                      "type": "string"
                    }
                  },
                  "command": {
                    "description": "The command to run in the dockerimage component instead of the default one provided in the image.\n\nDefaults to an empty array, meaning use whatever is defined in the image.",
                    "type": "array",
                    "items": {
                      "type": "string"
                    }
                  },
                  "cpuLimit": {
                    "type": "string"
                  },
                  "cpuRequest": {
                    "type": "string"
                  },
                  "dedicatedPod": {
                    "description": "Specify if a container should run in its own separated pod, instead of running as part of the main development environment pod.\n\nDefault value is 'false'",
                    "type": "boolean"
                  },
                  "endpoints": {
                    "type": "array",
                    "items": {
                      "type": "object",
                      "required": [
                        "name"
                      ],
                      "properties": {
                        "annotation": {
                          "description": "Annotations to be added to Kubernetes Ingress or Openshift Route",
                          "type": "object",
                          "additionalProperties": {
                            "type": "string"
                          }
                        },
                        "attributes": {
                          "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                          "type": "object",
                          "additionalProperties": true
                        },
                        "exposure": {
                          "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main devworkspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main devworkspace POD, on a local address.\n\nDefault value is 'public'",
                          "type": "string",
                          "enum": [
                            "public",
                            "internal",
                            "none"
                          ]
                        },
                        "name": {
                          "type": "string",
                          "maxLength": 15,
                          "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
                        },
                        "path": {
                          "description": "Path of the endpoint URL",
                          "type": "string"
                        },
                        "protocol": {
                          "description": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'",
                          "type": "string",
                          "enum": [
                            "http",
                            "https",
                            "ws",
                            "wss",
                            "tcp",
                            "udp"
                          ]
                        },
                        "secure": {
                          "description": "Describes whether the endpoint should be secured and protected by some authentication process. This requires a protocol of 'https' or 'wss'.",
                          "type": "boolean"
                        },
                        "targetPort": {
                          "description": "The port number should be unique.",
                          "type": "integer"
                        }
                      },
                      "additionalProperties": false
                    }
                  },
                  "env": {
                    "description": "Environment variables used in this container.\n\nThe following variables are reserved and cannot be overridden via env:\n\n - '$PROJECTS_ROOT'\n\n - '$PROJECT_SOURCE'",
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
                        }
                      },
                      "additionalProperties": false
                    }
                  },
                  "image": {
                    "type": "string"
                  },
                  "memoryLimit": {
                    "type": "string"
                  },
                  "memoryRequest": {
                    "type": "string"
                  },
                  "mountSources": {
                    "description": "Toggles whether or not the project source code should be mounted in the component.\n\nDefaults to true for all component types except plugins and components that set 'dedicatedPod' to true.",
                    "type": "boolean"
                  },
                  "sourceMapping": {
                    "description": "Optional specification of the path in the container where project sources should be transferred/mounted when 'mountSources' is 'true'. When omitted, the default value of /projects is used.",
                    "type": "string"
                  },
                  "volumeMounts": {
                    "description": "List of volumes mounts that should be mounted is this container.",
                    "type": "array",
                    "items": {
                      "description": "Volume that should be mounted to a component container",
                      "type": "object",
                      "required": [
                        "name"
                      ],
                      "properties": {
                        "name": {
                          "description": "The volume mount name is the name of an existing 'Volume' component. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files.",
                          "type": "string",
                          "maxLength": 63,
                          "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
                        },
                        "path": {
                          "description": "The path in the component container where the volume should be mounted. If not path is mentioned, default path is the is '/\u003cname\u003e'.",
                          "type": "string"
                        }
                      },
                      "additionalProperties": false
                    }
                  }
                },
                "additionalProperties": false
              },
              "image": {
                "description": "Allows specifying the definition of an image for outer loop builds",
                "type": "object",
                "oneOf": [
                  {
                    "required": [
                      "dockerfile"
                    ]
                  },
                  {
                    "required": [
                      "autoBuild"
                    ]
                  }
                ],
                "properties": {
                  "autoBuild": {
                    "description": "Defines if the image should be built during startup.\n\nDefault value is 'false'",
                    "type": "boolean"
                  },
                  "dockerfile": {
                    "description": "Allows specifying dockerfile type build",
                    "type": "object",
                    "oneOf": [
                      {
                        "required": [
                          "uri"
                        ]
                      },
                      {
                        "required": [
                          "devfileRegistry"
                        ]
                      },
                      {
                        "required": [
                          "git"
                        ]
                      }
                    ],
                    "properties": {
                      "args": {
                        "description": "The arguments to supply to the dockerfile build.",
                        "type": "array",
                        "items": {
                          "type": "string"
                        }
                      },
                      "buildContext": {
                        "description": "Path of source directory to establish build context. Defaults to ${PROJECT_ROOT} in the container",
                        "type": "string"
                      },
                      "devfileRegistry": {
                        "description": "Dockerfile's Devfile Registry source",
                        "type": "object",
                        "properties": {
                          "id": {
                            "description": "Id in a devfile registry that contains a Dockerfile. The src in the OCI registry required for the Dockerfile build will be downloaded for building the image.",
                            "type": "string"
                          },
                          "registryUrl": {
                            "description": "Devfile Registry URL to pull the Dockerfile from when using the Devfile Registry as Dockerfile src. To ensure the Dockerfile gets resolved consistently in different environments, it is recommended to always specify the 'devfileRegistryUrl' when 'Id' is used.",
                            "type": "string"
                          }
                        },
                        "additionalProperties": false
                      },
                      "git": {
                        "description": "Dockerfile's Git source",
                        "type": "object",
                        "properties": {
                          "checkoutFrom": {
                            "description": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                            "type": "object",
                            "properties": {
                              "remote": {
                                "description": "The remote name should be used as init. Required if there are more than one remote configured",
                                "type": "string"
                              },
                              "revision": {
                                "description": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found.",
                                "type": "string"
                              }
                            },
                            "additionalProperties": false
                          },
                          "fileLocation": {
                            "description": "Location of the Dockerfile in the Git repository when using git as Dockerfile src. Defaults to Dockerfile.",
                            "type": "string"
                          },
                          "remotes": {
                            "description": "The remotes map which should be initialized in the git project. Projects must have at least one remote configured while StarterProjects \u0026 Image Component's Git source can only have at most one remote configured.",
                            "type": "object",
                            "additionalProperties": {
                              "type": "string"
                            }
                          }
                        },
                        "additionalProperties": false
                      },
                      "rootRequired": {
                        "description": "Specify if a privileged builder pod is required.\n\nDefault value is 'false'",
                        "type": "boolean"
                      },
                      "uri": {
                        "description": "URI Reference of a Dockerfile. It can be a full URL or a relative URI from the current devfile as the base URI.",
                        "type": "string"
                      }
                    },
                    "additionalProperties": false
                  },
                  "imageName": {
                    "description": "Name of the image for the resulting outerloop build",
                    "type": "string"
                  }
                },
                "additionalProperties": false
              },
              "kubernetes": {
                "description": "Allows importing into the devworkspace the Kubernetes resources defined in a given manifest. For example this allows reusing the Kubernetes definitions used to deploy some runtime components in production.",
                "type": "object",
                "oneOf": [
                  {
                    "required": [
                      "uri"
                    ]
                  },
                  {
                    "required": [
                      "inlined"
                    ]
                  }
                ],
                "properties": {
                  "deployByDefault": {
                    "description": "Defines if the component should be deployed during startup.\n\nDefault value is 'false'",
                    "type": "boolean"
                  },
                  "endpoints": {
                    "type": "array",
                    "items": {
                      "type": "object",
                      "required": [
                        "name"
                      ],
                      "properties": {
                        "annotation": {
                          "description": "Annotations to be added to Kubernetes Ingress or Openshift Route",
                          "type": "object",
                          "additionalProperties": {
                            "type": "string"
                          }
                        },
                        "attributes": {
                          "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                          "type": "object",
                          "additionalProperties": true
                        },
                        "exposure": {
                          "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main devworkspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main devworkspace POD, on a local address.\n\nDefault value is 'public'",
                          "type": "string",
                          "enum": [
                            "public",
                            "internal",
                            "none"
                          ]
                        },
                        "name": {
                          "type": "string",
                          "maxLength": 15,
                          "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
                        },
                        "path": {
                          "description": "Path of the endpoint URL",
                          "type": "string"
                        },
                        "protocol": {
                          "description": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'",
                          "type": "string",
                          "enum": [
                            "http",
                            "https",
                            "ws",
                            "wss",
                            "tcp",
                            "udp"
                          ]
                        },
                        "secure": {
                          "description": "Describes whether the endpoint should be secured and protected by some authentication process. This requires a protocol of 'https' or 'wss'.",
                          "type": "boolean"
                        },
                        "targetPort": {
                          "description": "The port number should be unique.",
                          "type": "integer"
                        }
                      },
                      "additionalProperties": false
                    }
                  },
                  "inlined": {
                    "description": "Inlined manifest",
                    "type": "string"
                  },
                  "uri": {
                    "description": "Location in a file fetched from a uri.",
                    "type": "string"
                  }
                },
                "additionalProperties": false
              },
              "name": {
                "description": "Mandatory name that allows referencing the component from other elements (such as commands) or from an external devfile that may reference this component through a parent or a plugin.",
                "type": "string",
                "maxLength": 63,
                "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
              },
              "openshift": {
                "description": "Allows importing into the devworkspace the OpenShift resources defined in a given manifest. For example this allows reusing the OpenShift definitions used to deploy some runtime components in production.",
                "type": "object",
                "oneOf": [
                  {
                    "required": [
                      "uri"
                    ]
                  },
                  {
                    "required": [
                      "inlined"
                    ]
                  }
                ],
                "properties": {
                  "deployByDefault": {
                    "description": "Defines if the component should be deployed during startup.\n\nDefault value is 'false'",
                    "type": "boolean"
                  },
                  "endpoints": {
                    "type": "array",
                    "items": {
                      "type": "object",
                      "required": [
                        "name"
                      ],
                      "properties": {
                        "annotation": {
                          "description": "Annotations to be added to Kubernetes Ingress or Openshift Route",
                          "type": "object",
                          "additionalProperties": {
                            "type": "string"
                          }
                        },
                        "attributes": {
                          "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                          "type": "object",
                          "additionalProperties": true
                        },
                        "exposure": {
                          "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main devworkspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main devworkspace POD, on a local address.\n\nDefault value is 'public'",
                          "type": "string",
                          "enum": [
                            "public",
                            "internal",
                            "none"
                          ]
                        },
                        "name": {
                          "type": "string",
                          "maxLength": 15,
                          "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
                        },
                        "path": {
                          "description": "Path of the endpoint URL",
                          "type": "string"
                        },
                        "protocol": {
                          "description": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'",
                          "type": "string",
                          "enum": [
                            "http",
                            "https",
                            "ws",
                            "wss",
                            "tcp",
                            "udp"
                          ]
                        },
                        "secure": {
                          "description": "Describes whether the endpoint should be secured and protected by some authentication process. This requires a protocol of 'https' or 'wss'.",
                          "type": "boolean"
                        },
                        "targetPort": {
                          "description": "The port number should be unique.",
                          "type": "integer"
                        }
                      },
                      "additionalProperties": false
                    }
                  },
                  "inlined": {
                    "description": "Inlined manifest",
                    "type": "string"
                  },
                  "uri": {
                    "description": "Location in a file fetched from a uri.",
                    "type": "string"
                  }
                },
                "additionalProperties": false
              },
              "volume": {
                "description": "Allows specifying the definition of a volume shared by several other components",
                "type": "object",
                "properties": {
                  "ephemeral": {
                    "description": "Ephemeral volumes are not stored persistently across restarts. Defaults to false",
                    "type": "boolean"
                  },
                  "size": {
                    "description": "Size of the volume",
                    "type": "string"
                  }
                },
                "additionalProperties": false
              }
            },
            "additionalProperties": false
          }
        },
        "id": {
          "description": "Id in a registry that contains a Devfile yaml file",
          "type": "string"
        },
        "kubernetes": {
          "description": "Reference to a Kubernetes CRD of type DevWorkspaceTemplate",
          "type": "object",
          "required": [
            "name"
          ],
          "properties": {
            "name": {
              "type": "string"
            },
            "namespace": {
              "type": "string"
            }
          },
          "additionalProperties": false
        },
        "projects": {
          "description": "Overrides of projects encapsulated in a parent devfile. Overriding is done according to K8S strategic merge patch standard rules.",
          "type": "array",
          "items": {
            "type": "object",
            "required": [
              "name"
            ],
            "oneOf": [
              {
                "required": [
                  "git"
                ]
              },
              {
                "required": [
                  "zip"
                ]
              }
            ],
            "properties": {
              "attributes": {
                "description": "Map of implementation-dependant free-form YAML attributes.",
                "type": "object",
                "additionalProperties": true
              },
              "clonePath": {
                "description": "Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name.",
                "type": "string"
              },
              "git": {
                "description": "Project's Git source",
                "type": "object",
                "properties": {
                  "checkoutFrom": {
                    "description": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                    "type": "object",
                    "properties": {
                      "remote": {
                        "description": "The remote name should be used as init. Required if there are more than one remote configured",
                        "type": "string"
                      },
                      "revision": {
                        "description": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found.",
                        "type": "string"
                      }
                    },
                    "additionalProperties": false
                  },
                  "remotes": {
                    "description": "The remotes map which should be initialized in the git project. Projects must have at least one remote configured while StarterProjects \u0026 Image Component's Git source can only have at most one remote configured.",
                    "type": "object",
                    "additionalProperties": {
                      "type": "string"
                    }
                  }
                },
                "additionalProperties": false
              },
              "name": {
                "description": "Project name",
                "type": "string",
                "maxLength": 63,
                "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
              },
              "zip": {
                "description": "Project's Zip source",
                "type": "object",
                "properties": {
                  "location": {
                    "description": "Zip project's source location address. Should be file path of the archive, e.g. file://$FILE_PATH",
                    "type": "string"
                  }
                },
                "additionalProperties": false
              }
            },
            "additionalProperties": false
          }
        },
        "registryUrl": {
          "description": "Registry URL to pull the parent devfile from when using id in the parent reference. To ensure the parent devfile gets resolved consistently in different environments, it is recommended to always specify the 'registryUrl' when 'id' is used.",
          "type": "string"
        },
        "starterProjects": {
          "description": "Overrides of starterProjects encapsulated in a parent devfile. Overriding is done according to K8S strategic merge patch standard rules.",
          "type": "array",
          "items": {
            "type": "object",
            "required": [
              "name"
            ],
            "oneOf": [
              {
                "required": [
                  "git"
                ]
              },
              {
                "required": [
                  "zip"
                ]
              }
            ],
            "properties": {
              "attributes": {
                "description": "Map of implementation-dependant free-form YAML attributes.",
                "type": "object",
                "additionalProperties": true
              },
              "description": {
                "description": "Description of a starter project",
                "type": "string"
              },
              "git": {
                "description": "Project's Git source",
                "type": "object",
                "properties": {
                  "checkoutFrom": {
                    "description": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                    "type": "object",
                    "properties": {
                      "remote": {
                        "description": "The remote name should be used as init. Required if there are more than one remote configured",
                        "type": "string"
                      },
                      "revision": {
                        "description": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found.",
                        "type": "string"
                      }
                    },
                    "additionalProperties": false
                  },
                  "remotes": {
                    "description": "The remotes map which should be initialized in the git project. Projects must have at least one remote configured while StarterProjects \u0026 Image Component's Git source can only have at most one remote configured.",
                    "type": "object",
                    "additionalProperties": {
                      "type": "string"
                    }
                  }
                },
                "additionalProperties": false
              },
              "name": {
                "description": "Project name",
                "type": "string",
                "maxLength": 63,
                "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
              },
              "subDir": {
                "description": "Sub-directory from a starter project to be used as root for starter project.",
                "type": "string"
              },
              "zip": {
                "description": "Project's Zip source",
                "type": "object",
                "properties": {
                  "location": {
                    "description": "Zip project's source location address. Should be file path of the archive, e.g. file://$FILE_PATH",
                    "type": "string"
                  }
                },
                "additionalProperties": false
              }
            },
            "additionalProperties": false
          }
        },
        "uri": {
          "description": "URI Reference of a parent devfile YAML file. It can be a full URL or a relative URI with the current devfile as the base URI.",
          "type": "string"
        },
        "variables": {
          "description": "Overrides of variables encapsulated in a parent devfile. Overriding is done according to K8S strategic merge patch standard rules.",
          "type": "object",
          "additionalProperties": {
            "type": "string"
          }
        }
      },
      "additionalProperties": false
    },
    "projects": {
      "description": "Projects worked on in the devworkspace, containing names and sources locations",
      "type": "array",
      "items": {
        "type": "object",
        "required": [
          "name"
        ],
        "oneOf": [
          {
            "required": [
              "git"
            ]
          },
          {
            "required": [
              "zip"
            ]
          }
        ],
        "properties": {
          "attributes": {
            "description": "Map of implementation-dependant free-form YAML attributes.",
            "type": "object",
            "additionalProperties": true
          },
          "clonePath": {
            "description": "Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name.",
            "type": "string"
          },
          "git": {
            "description": "Project's Git source",
            "type": "object",
            "required": [
              "remotes"
            ],
            "properties": {
              "checkoutFrom": {
                "description": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                "type": "object",
                "properties": {
                  "remote": {
                    "description": "The remote name should be used as init. Required if there are more than one remote configured",
                    "type": "string"
                  },
                  "revision": {
                    "description": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found.",
                    "type": "string"
                  }
                },
                "additionalProperties": false
              },
              "remotes": {
                "description": "The remotes map which should be initialized in the git project. Projects must have at least one remote configured while StarterProjects \u0026 Image Component's Git source can only have at most one remote configured.",
                "type": "object",
                "additionalProperties": {
                  "type": "string"
                }
              }
            },
            "additionalProperties": false
          },
          "name": {
            "description": "Project name",
            "type": "string",
            "maxLength": 63,
            "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
          },
          "zip": {
            "description": "Project's Zip source",
            "type": "object",
            "properties": {
              "location": {
                "description": "Zip project's source location address. Should be file path of the archive, e.g. file://$FILE_PATH",
                "type": "string"
              }
            },
            "additionalProperties": false
          }
        },
        "additionalProperties": false
      }
    },
    "schemaVersion": {
      "description": "Devfile schema version",
      "type": "string",
      "pattern": "^([2-9])\\.([0-9]+)\\.([0-9]+)(\\-[0-9a-z-]+(\\.[0-9a-z-]+)*)?(\\+[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?$"
    },
    "starterProjects": {
      "description": "StarterProjects is a project that can be used as a starting point when bootstrapping new projects",
      "type": "array",
      "items": {
        "type": "object",
        "required": [
          "name"
        ],
        "oneOf": [
          {
            "required": [
              "git"
            ]
          },
          {
            "required": [
              "zip"
            ]
          }
        ],
        "properties": {
          "attributes": {
            "description": "Map of implementation-dependant free-form YAML attributes.",
            "type": "object",
            "additionalProperties": true
          },
          "description": {
            "description": "Description of a starter project",
            "type": "string"
          },
          "git": {
            "description": "Project's Git source",
            "type": "object",
            "required": [
              "remotes"
            ],
            "properties": {
              "checkoutFrom": {
                "description": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                "type": "object",
                "properties": {
                  "remote": {
                    "description": "The remote name should be used as init. Required if there are more than one remote configured",
                    "type": "string"
                  },
                  "revision": {
                    "description": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found.",
                    "type": "string"
                  }
                },
                "additionalProperties": false
              },
              "remotes": {
                "description": "The remotes map which should be initialized in the git project. Projects must have at least one remote configured while StarterProjects \u0026 Image Component's Git source can only have at most one remote configured.",
                "type": "object",
                "additionalProperties": {
                  "type": "string"
                }
              }
            },
            "additionalProperties": false
          },
          "name": {
            "description": "Project name",
            "type": "string",
            "maxLength": 63,
            "pattern": "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
          },
          "subDir": {
            "description": "Sub-directory from a starter project to be used as root for starter project.",
            "type": "string"
          },
          "zip": {
            "description": "Project's Zip source",
            "type": "object",
            "properties": {
              "location": {
                "description": "Zip project's source location address. Should be file path of the archive, e.g. file://$FILE_PATH",
                "type": "string"
              }
            },
            "additionalProperties": false
          }
        },
        "additionalProperties": false
      }
    },
    "variables": {
      "description": "Map of key-value variables used for string replacement in the devfile. Values can be referenced via {{variable-key}} to replace the corresponding value in string fields in the devfile. Replacement cannot be used for\n\n - schemaVersion, metadata, parent source\n\n - element identifiers, e.g. command id, component name, endpoint name, project name\n\n - references to identifiers, e.g. in events, a command's component, container's volume mount name\n\n - string enums, e.g. command group kind, endpoint exposure",
      "type": "object",
      "additionalProperties": {
        "type": "string"
      }
    }
  },
  "additionalProperties": false
}
`
