package version200

// https://raw.githubusercontent.com/devfile/api/master/schemas/devfile.json
const JsonSchema200 = `{
  "description": "Devfile schema.",
  "properties": {
    "commands": {
      "description": "Predefined, ready-to-use, workspace-related commands",
      "items": {
        "properties": {
          "apply": {
            "description": "Command that consists in applying a given component definition, typically bound to a workspace event.\n\nFor example, when an 'apply' command is bound to a 'preStart' event, and references a 'container' component, it will start the container as a K8S initContainer in the workspace POD, unless the component has its 'dedicatedPod' field set to 'true'.\n\nWhen no 'apply' command exist for a given component, it is assumed the component will be applied at workspace start by default.",
            "properties": {
              "attributes": {
                "additionalProperties": {
                  "type": "string"
                },
                "description": "Optional map of free-form additional command attributes",
                "type": "object",
                "markdownDescription": "Optional map of free-form additional command attributes"
              },
              "component": {
                "description": "Describes component that will be applied",
                "type": "string",
                "markdownDescription": "Describes component that will be applied"
              },
              "group": {
                "description": "Defines the group this command is part of",
                "properties": {
                  "isDefault": {
                    "description": "Identifies the default command for a given group kind",
                    "type": "boolean",
                    "markdownDescription": "Identifies the default command for a given group kind"
                  },
                  "kind": {
                    "description": "Kind of group the command is part of",
                    "enum": [
                      "build",
                      "run",
                      "test",
                      "debug"
                    ],
                    "type": "string",
                    "markdownDescription": "Kind of group the command is part of"
                  }
                },
                "required": [
                  "kind"
                ],
                "type": "object",
                "markdownDescription": "Defines the group this command is part of",
                "additionalProperties": false
              },
              "label": {
                "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                "type": "string",
                "markdownDescription": "Optional label that provides a label for this command to be used in Editor UI menus for example"
              }
            },
            "type": "object",
            "markdownDescription": "Command that consists in applying a given component definition, typically bound to a workspace event.\n\nFor example, when an 'apply' command is bound to a 'preStart' event, and references a 'container' component, it will start the container as a K8S initContainer in the workspace POD, unless the component has its 'dedicatedPod' field set to 'true'.\n\nWhen no 'apply' command exist for a given component, it is assumed the component will be applied at workspace start by default.",
            "additionalProperties": false
          },
          "composite": {
            "description": "Composite command that allows executing several sub-commands either sequentially or concurrently",
            "properties": {
              "attributes": {
                "additionalProperties": {
                  "type": "string"
                },
                "description": "Optional map of free-form additional command attributes",
                "type": "object",
                "markdownDescription": "Optional map of free-form additional command attributes"
              },
              "commands": {
                "description": "The commands that comprise this composite command",
                "items": {
                  "type": "string"
                },
                "type": "array",
                "markdownDescription": "The commands that comprise this composite command"
              },
              "group": {
                "description": "Defines the group this command is part of",
                "properties": {
                  "isDefault": {
                    "description": "Identifies the default command for a given group kind",
                    "type": "boolean",
                    "markdownDescription": "Identifies the default command for a given group kind"
                  },
                  "kind": {
                    "description": "Kind of group the command is part of",
                    "enum": [
                      "build",
                      "run",
                      "test",
                      "debug"
                    ],
                    "type": "string",
                    "markdownDescription": "Kind of group the command is part of"
                  }
                },
                "required": [
                  "kind"
                ],
                "type": "object",
                "markdownDescription": "Defines the group this command is part of",
                "additionalProperties": false
              },
              "label": {
                "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                "type": "string",
                "markdownDescription": "Optional label that provides a label for this command to be used in Editor UI menus for example"
              },
              "parallel": {
                "description": "Indicates if the sub-commands should be executed concurrently",
                "type": "boolean",
                "markdownDescription": "Indicates if the sub-commands should be executed concurrently"
              }
            },
            "type": "object",
            "markdownDescription": "Composite command that allows executing several sub-commands either sequentially or concurrently",
            "additionalProperties": false
          },
          "exec": {
            "description": "CLI Command executed in an existing component container",
            "properties": {
              "attributes": {
                "additionalProperties": {
                  "type": "string"
                },
                "description": "Optional map of free-form additional command attributes",
                "type": "object",
                "markdownDescription": "Optional map of free-form additional command attributes"
              },
              "commandLine": {
                "description": "The actual command-line string\n\nSpecial variables that can be used:\n\n - '$PROJECTS_ROOT': A path where projects sources are mounted\n\n - '$PROJECT_SOURCE': A path to a project source ($PROJECTS_ROOT/<project-name>). If there are multiple projects, this will point to the directory of the first one.",
                "type": "string",
                "markdownDescription": "The actual command-line string\n\nSpecial variables that can be used:\n\n - '$PROJECTS_ROOT': A path where projects sources are mounted\n\n - '$PROJECT_SOURCE': A path to a project source ($PROJECTS_ROOT/<project-name>). If there are multiple projects, this will point to the directory of the first one."
              },
              "component": {
                "description": "Describes component to which given action relates",
                "type": "string",
                "markdownDescription": "Describes component to which given action relates"
              },
              "env": {
                "description": "Optional list of environment variables that have to be set before running the command",
                "items": {
                  "properties": {
                    "name": {
                      "type": "string"
                    },
                    "value": {
                      "type": "string"
                    }
                  },
                  "required": [
                    "name",
                    "value"
                  ],
                  "type": "object",
                  "additionalProperties": false
                },
                "type": "array",
                "markdownDescription": "Optional list of environment variables that have to be set before running the command"
              },
              "group": {
                "description": "Defines the group this command is part of",
                "properties": {
                  "isDefault": {
                    "description": "Identifies the default command for a given group kind",
                    "type": "boolean",
                    "markdownDescription": "Identifies the default command for a given group kind"
                  },
                  "kind": {
                    "description": "Kind of group the command is part of",
                    "enum": [
                      "build",
                      "run",
                      "test",
                      "debug"
                    ],
                    "type": "string",
                    "markdownDescription": "Kind of group the command is part of"
                  }
                },
                "required": [
                  "kind"
                ],
                "type": "object",
                "markdownDescription": "Defines the group this command is part of",
                "additionalProperties": false
              },
              "hotReloadCapable": {
                "description": "Whether the command is capable to reload itself when source code changes. If set to 'true' the command won't be restarted and it is expected to handle file changes on its own.\n\nDefault value is 'false'",
                "type": "boolean",
                "markdownDescription": "Whether the command is capable to reload itself when source code changes. If set to 'true' the command won't be restarted and it is expected to handle file changes on its own.\n\nDefault value is 'false'"
              },
              "label": {
                "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                "type": "string",
                "markdownDescription": "Optional label that provides a label for this command to be used in Editor UI menus for example"
              },
              "workingDir": {
                "description": "Working directory where the command should be executed\n\nSpecial variables that can be used:\n\n - '${PROJECTS_ROOT}': A path where projects sources are mounted\n\n - '${PROJECT_SOURCE}': A path to a project source (${PROJECTS_ROOT}/<project-name>). If there are multiple projects, this will point to the directory of the first one.",
                "type": "string",
                "markdownDescription": "Working directory where the command should be executed\n\nSpecial variables that can be used:\n\n - '${PROJECTS_ROOT}': A path where projects sources are mounted\n\n - '${PROJECT_SOURCE}': A path to a project source (${PROJECTS_ROOT}/<project-name>). If there are multiple projects, this will point to the directory of the first one."
              }
            },
            "type": "object",
            "markdownDescription": "CLI Command executed in an existing component container",
            "additionalProperties": false,
            "required": [
              "commandLine"
            ]
          },
          "id": {
            "description": "Mandatory identifier that allows referencing this command in composite commands, from a parent, or in events.",
            "type": "string",
            "markdownDescription": "Mandatory identifier that allows referencing this command in composite commands, from a parent, or in events."
          },
          "vscodeLaunch": {
            "description": "Command providing the definition of a VsCode launch action",
            "properties": {
              "attributes": {
                "additionalProperties": {
                  "type": "string"
                },
                "description": "Optional map of free-form additional command attributes",
                "type": "object",
                "markdownDescription": "Optional map of free-form additional command attributes"
              },
              "group": {
                "description": "Defines the group this command is part of",
                "properties": {
                  "isDefault": {
                    "description": "Identifies the default command for a given group kind",
                    "type": "boolean",
                    "markdownDescription": "Identifies the default command for a given group kind"
                  },
                  "kind": {
                    "description": "Kind of group the command is part of",
                    "enum": [
                      "build",
                      "run",
                      "test",
                      "debug"
                    ],
                    "type": "string",
                    "markdownDescription": "Kind of group the command is part of"
                  }
                },
                "required": [
                  "kind"
                ],
                "type": "object",
                "markdownDescription": "Defines the group this command is part of",
                "additionalProperties": false
              },
              "inlined": {
                "description": "Inlined content of the VsCode configuration",
                "type": "string",
                "markdownDescription": "Inlined content of the VsCode configuration"
              },
              "uri": {
                "description": "Location as an absolute of relative URI the VsCode configuration will be fetched from",
                "type": "string",
                "markdownDescription": "Location as an absolute of relative URI the VsCode configuration will be fetched from"
              }
            },
            "type": "object",
            "markdownDescription": "Command providing the definition of a VsCode launch action",
            "additionalProperties": false,
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
            ]
          },
          "vscodeTask": {
            "description": "Command providing the definition of a VsCode Task",
            "properties": {
              "attributes": {
                "additionalProperties": {
                  "type": "string"
                },
                "description": "Optional map of free-form additional command attributes",
                "type": "object",
                "markdownDescription": "Optional map of free-form additional command attributes"
              },
              "group": {
                "description": "Defines the group this command is part of",
                "properties": {
                  "isDefault": {
                    "description": "Identifies the default command for a given group kind",
                    "type": "boolean",
                    "markdownDescription": "Identifies the default command for a given group kind"
                  },
                  "kind": {
                    "description": "Kind of group the command is part of",
                    "enum": [
                      "build",
                      "run",
                      "test",
                      "debug"
                    ],
                    "type": "string",
                    "markdownDescription": "Kind of group the command is part of"
                  }
                },
                "required": [
                  "kind"
                ],
                "type": "object",
                "markdownDescription": "Defines the group this command is part of",
                "additionalProperties": false
              },
              "inlined": {
                "description": "Inlined content of the VsCode configuration",
                "type": "string",
                "markdownDescription": "Inlined content of the VsCode configuration"
              },
              "uri": {
                "description": "Location as an absolute of relative URI the VsCode configuration will be fetched from",
                "type": "string",
                "markdownDescription": "Location as an absolute of relative URI the VsCode configuration will be fetched from"
              }
            },
            "type": "object",
            "markdownDescription": "Command providing the definition of a VsCode Task",
            "additionalProperties": false,
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
            ]
          }
        },
        "required": [
          "id"
        ],
        "type": "object",
        "additionalProperties": false,
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
              "vscodeTask"
            ]
          },
          {
            "required": [
              "vscodeLaunch"
            ]
          },
          {
            "required": [
              "composite"
            ]
          }
        ]
      },
      "type": "array",
      "markdownDescription": "Predefined, ready-to-use, workspace-related commands"
    },
    "components": {
      "description": "List of the workspace components, such as editor and plugins, user-provided containers, or other types of components",
      "items": {
        "properties": {
          "container": {
            "description": "Allows adding and configuring workspace-related containers",
            "properties": {
              "args": {
                "description": "The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command.\n\nDefaults to an empty array, meaning use whatever is defined in the image.",
                "items": {
                  "type": "string"
                },
                "type": "array",
                "markdownDescription": "The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command.\n\nDefaults to an empty array, meaning use whatever is defined in the image."
              },
              "command": {
                "description": "The command to run in the dockerimage component instead of the default one provided in the image.\n\nDefaults to an empty array, meaning use whatever is defined in the image.",
                "items": {
                  "type": "string"
                },
                "type": "array",
                "markdownDescription": "The command to run in the dockerimage component instead of the default one provided in the image.\n\nDefaults to an empty array, meaning use whatever is defined in the image."
              },
              "dedicatedPod": {
                "description": "Specify if a container should run in its own separated pod, instead of running as part of the main development environment pod.\n\nDefault value is 'false'",
                "type": "boolean",
                "markdownDescription": "Specify if a container should run in its own separated pod, instead of running as part of the main development environment pod.\n\nDefault value is 'false'"
              },
              "endpoints": {
                "items": {
                  "properties": {
                    "attributes": {
                      "additionalProperties": {
                        "type": "string"
                      },
                      "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                      "type": "object",
                      "markdownDescription": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\","
                    },
                    "exposure": {
                      "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                      "enum": [
                        "public",
                        "internal",
                        "none"
                      ],
                      "type": "string",
                      "markdownDescription": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'"
                    },
                    "name": {
                      "type": "string"
                    },
                    "path": {
                      "description": "Path of the endpoint URL",
                      "type": "string",
                      "markdownDescription": "Path of the endpoint URL"
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
                      ],
                      "markdownDescription": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'"
                    },
                    "secure": {
                      "description": "Describes whether the endpoint should be secured and protected by some authentication process",
                      "type": "boolean",
                      "markdownDescription": "Describes whether the endpoint should be secured and protected by some authentication process"
                    },
                    "targetPort": {
                      "type": "integer"
                    }
                  },
                  "required": [
                    "name",
                    "targetPort"
                  ],
                  "type": "object",
                  "additionalProperties": false
                },
                "type": "array"
              },
              "env": {
                "description": "Environment variables used in this container",
                "items": {
                  "properties": {
                    "name": {
                      "type": "string"
                    },
                    "value": {
                      "type": "string"
                    }
                  },
                  "required": [
                    "name",
                    "value"
                  ],
                  "type": "object",
                  "additionalProperties": false
                },
                "type": "array",
                "markdownDescription": "Environment variables used in this container"
              },
              "image": {
                "type": "string"
              },
              "memoryLimit": {
                "type": "string"
              },
              "mountSources": {
                "type": "boolean"
              },
              "sourceMapping": {
                "description": "Optional specification of the path in the container where project sources should be transferred/mounted when 'mountSources' is 'true'. When omitted, the value of the 'PROJECTS_ROOT' environment variable is used.",
                "type": "string",
                "markdownDescription": "Optional specification of the path in the container where project sources should be transferred/mounted when 'mountSources' is 'true'. When omitted, the value of the 'PROJECTS_ROOT' environment variable is used."
              },
              "volumeMounts": {
                "description": "List of volumes mounts that should be mounted is this container.",
                "items": {
                  "description": "Volume that should be mounted to a component container",
                  "properties": {
                    "name": {
                      "description": "The volume mount name is the name of an existing 'Volume' component. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files.",
                      "type": "string",
                      "markdownDescription": "The volume mount name is the name of an existing 'Volume' component. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files."
                    },
                    "path": {
                      "description": "The path in the component container where the volume should be mounted. If not path is mentioned, default path is the is '/<name>'.",
                      "type": "string",
                      "markdownDescription": "The path in the component container where the volume should be mounted. If not path is mentioned, default path is the is '/<name>'."
                    }
                  },
                  "required": [
                    "name"
                  ],
                  "type": "object",
                  "markdownDescription": "Volume that should be mounted to a component container",
                  "additionalProperties": false
                },
                "type": "array",
                "markdownDescription": "List of volumes mounts that should be mounted is this container."
              }
            },
            "type": "object",
            "markdownDescription": "Allows adding and configuring workspace-related containers",
            "additionalProperties": false,
            "required": [
              "image"
            ]
          },
          "kubernetes": {
            "description": "Allows importing into the workspace the Kubernetes resources defined in a given manifest. For example this allows reusing the Kubernetes definitions used to deploy some runtime components in production.",
            "properties": {
              "endpoints": {
                "items": {
                  "properties": {
                    "attributes": {
                      "additionalProperties": {
                        "type": "string"
                      },
                      "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                      "type": "object",
                      "markdownDescription": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\","
                    },
                    "exposure": {
                      "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                      "enum": [
                        "public",
                        "internal",
                        "none"
                      ],
                      "type": "string",
                      "markdownDescription": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'"
                    },
                    "name": {
                      "type": "string"
                    },
                    "path": {
                      "description": "Path of the endpoint URL",
                      "type": "string",
                      "markdownDescription": "Path of the endpoint URL"
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
                      ],
                      "markdownDescription": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'"
                    },
                    "secure": {
                      "description": "Describes whether the endpoint should be secured and protected by some authentication process",
                      "type": "boolean",
                      "markdownDescription": "Describes whether the endpoint should be secured and protected by some authentication process"
                    },
                    "targetPort": {
                      "type": "integer"
                    }
                  },
                  "required": [
                    "name",
                    "targetPort"
                  ],
                  "type": "object",
                  "additionalProperties": false
                },
                "type": "array"
              },
              "inlined": {
                "description": "Inlined manifest",
                "type": "string",
                "markdownDescription": "Inlined manifest"
              },
              "uri": {
                "description": "Location in a file fetched from a uri.",
                "type": "string",
                "markdownDescription": "Location in a file fetched from a uri."
              }
            },
            "type": "object",
            "markdownDescription": "Allows importing into the workspace the Kubernetes resources defined in a given manifest. For example this allows reusing the Kubernetes definitions used to deploy some runtime components in production.",
            "additionalProperties": false,
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
            ]
          },
          "name": {
            "description": "Mandatory name that allows referencing the component from other elements (such as commands) or from an external devfile that may reference this component through a parent or a plugin.",
            "type": "string",
            "markdownDescription": "Mandatory name that allows referencing the component from other elements (such as commands) or from an external devfile that may reference this component through a parent or a plugin."
          },
          "openshift": {
            "description": "Allows importing into the workspace the OpenShift resources defined in a given manifest. For example this allows reusing the OpenShift definitions used to deploy some runtime components in production.",
            "properties": {
              "endpoints": {
                "items": {
                  "properties": {
                    "attributes": {
                      "additionalProperties": {
                        "type": "string"
                      },
                      "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                      "type": "object",
                      "markdownDescription": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\","
                    },
                    "exposure": {
                      "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                      "enum": [
                        "public",
                        "internal",
                        "none"
                      ],
                      "type": "string",
                      "markdownDescription": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'"
                    },
                    "name": {
                      "type": "string"
                    },
                    "path": {
                      "description": "Path of the endpoint URL",
                      "type": "string",
                      "markdownDescription": "Path of the endpoint URL"
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
                      ],
                      "markdownDescription": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'"
                    },
                    "secure": {
                      "description": "Describes whether the endpoint should be secured and protected by some authentication process",
                      "type": "boolean",
                      "markdownDescription": "Describes whether the endpoint should be secured and protected by some authentication process"
                    },
                    "targetPort": {
                      "type": "integer"
                    }
                  },
                  "required": [
                    "name",
                    "targetPort"
                  ],
                  "type": "object",
                  "additionalProperties": false
                },
                "type": "array"
              },
              "inlined": {
                "description": "Inlined manifest",
                "type": "string",
                "markdownDescription": "Inlined manifest"
              },
              "uri": {
                "description": "Location in a file fetched from a uri.",
                "type": "string",
                "markdownDescription": "Location in a file fetched from a uri."
              }
            },
            "type": "object",
            "markdownDescription": "Allows importing into the workspace the OpenShift resources defined in a given manifest. For example this allows reusing the OpenShift definitions used to deploy some runtime components in production.",
            "additionalProperties": false,
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
            ]
          },
          "plugin": {
            "description": "Allows importing a plugin.\n\nPlugins are mainly imported devfiles that contribute components, commands and events as a consistent single unit. They are defined in either YAML files following the devfile syntax, or as 'DevWorkspaceTemplate' Kubernetes Custom Resources",
            "properties": {
              "commands": {
                "description": "Overrides of commands encapsulated in a parent devfile or a plugin. Overriding is done using a strategic merge patch",
                "items": {
                  "properties": {
                    "apply": {
                      "description": "Command that consists in applying a given component definition, typically bound to a workspace event.\n\nFor example, when an 'apply' command is bound to a 'preStart' event, and references a 'container' component, it will start the container as a K8S initContainer in the workspace POD, unless the component has its 'dedicatedPod' field set to 'true'.\n\nWhen no 'apply' command exist for a given component, it is assumed the component will be applied at workspace start by default.",
                      "properties": {
                        "attributes": {
                          "additionalProperties": {
                            "type": "string"
                          },
                          "description": "Optional map of free-form additional command attributes",
                          "type": "object",
                          "markdownDescription": "Optional map of free-form additional command attributes"
                        },
                        "component": {
                          "description": "Describes component that will be applied",
                          "type": "string",
                          "markdownDescription": "Describes component that will be applied"
                        },
                        "group": {
                          "description": "Defines the group this command is part of",
                          "properties": {
                            "isDefault": {
                              "description": "Identifies the default command for a given group kind",
                              "type": "boolean",
                              "markdownDescription": "Identifies the default command for a given group kind"
                            },
                            "kind": {
                              "description": "Kind of group the command is part of",
                              "enum": [
                                "build",
                                "run",
                                "test",
                                "debug"
                              ],
                              "type": "string",
                              "markdownDescription": "Kind of group the command is part of"
                            }
                          },
                          "required": [
                            "kind"
                          ],
                          "type": "object",
                          "markdownDescription": "Defines the group this command is part of",
                          "additionalProperties": false
                        },
                        "label": {
                          "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                          "type": "string",
                          "markdownDescription": "Optional label that provides a label for this command to be used in Editor UI menus for example"
                        }
                      },
                      "type": "object",
                      "markdownDescription": "Command that consists in applying a given component definition, typically bound to a workspace event.\n\nFor example, when an 'apply' command is bound to a 'preStart' event, and references a 'container' component, it will start the container as a K8S initContainer in the workspace POD, unless the component has its 'dedicatedPod' field set to 'true'.\n\nWhen no 'apply' command exist for a given component, it is assumed the component will be applied at workspace start by default.",
                      "additionalProperties": false
                    },
                    "composite": {
                      "description": "Composite command that allows executing several sub-commands either sequentially or concurrently",
                      "properties": {
                        "attributes": {
                          "additionalProperties": {
                            "type": "string"
                          },
                          "description": "Optional map of free-form additional command attributes",
                          "type": "object",
                          "markdownDescription": "Optional map of free-form additional command attributes"
                        },
                        "commands": {
                          "description": "The commands that comprise this composite command",
                          "items": {
                            "type": "string"
                          },
                          "type": "array",
                          "markdownDescription": "The commands that comprise this composite command"
                        },
                        "group": {
                          "description": "Defines the group this command is part of",
                          "properties": {
                            "isDefault": {
                              "description": "Identifies the default command for a given group kind",
                              "type": "boolean",
                              "markdownDescription": "Identifies the default command for a given group kind"
                            },
                            "kind": {
                              "description": "Kind of group the command is part of",
                              "enum": [
                                "build",
                                "run",
                                "test",
                                "debug"
                              ],
                              "type": "string",
                              "markdownDescription": "Kind of group the command is part of"
                            }
                          },
                          "required": [
                            "kind"
                          ],
                          "type": "object",
                          "markdownDescription": "Defines the group this command is part of",
                          "additionalProperties": false
                        },
                        "label": {
                          "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                          "type": "string",
                          "markdownDescription": "Optional label that provides a label for this command to be used in Editor UI menus for example"
                        },
                        "parallel": {
                          "description": "Indicates if the sub-commands should be executed concurrently",
                          "type": "boolean",
                          "markdownDescription": "Indicates if the sub-commands should be executed concurrently"
                        }
                      },
                      "type": "object",
                      "markdownDescription": "Composite command that allows executing several sub-commands either sequentially or concurrently",
                      "additionalProperties": false
                    },
                    "exec": {
                      "description": "CLI Command executed in an existing component container",
                      "properties": {
                        "attributes": {
                          "additionalProperties": {
                            "type": "string"
                          },
                          "description": "Optional map of free-form additional command attributes",
                          "type": "object",
                          "markdownDescription": "Optional map of free-form additional command attributes"
                        },
                        "commandLine": {
                          "description": "The actual command-line string\n\nSpecial variables that can be used:\n\n - '$PROJECTS_ROOT': A path where projects sources are mounted\n\n - '$PROJECT_SOURCE': A path to a project source ($PROJECTS_ROOT/<project-name>). If there are multiple projects, this will point to the directory of the first one.",
                          "type": "string",
                          "markdownDescription": "The actual command-line string\n\nSpecial variables that can be used:\n\n - '$PROJECTS_ROOT': A path where projects sources are mounted\n\n - '$PROJECT_SOURCE': A path to a project source ($PROJECTS_ROOT/<project-name>). If there are multiple projects, this will point to the directory of the first one."
                        },
                        "component": {
                          "description": "Describes component to which given action relates",
                          "type": "string",
                          "markdownDescription": "Describes component to which given action relates"
                        },
                        "env": {
                          "description": "Optional list of environment variables that have to be set before running the command",
                          "items": {
                            "properties": {
                              "name": {
                                "type": "string"
                              },
                              "value": {
                                "type": "string"
                              }
                            },
                            "required": [
                              "name",
                              "value"
                            ],
                            "type": "object",
                            "additionalProperties": false
                          },
                          "type": "array",
                          "markdownDescription": "Optional list of environment variables that have to be set before running the command"
                        },
                        "group": {
                          "description": "Defines the group this command is part of",
                          "properties": {
                            "isDefault": {
                              "description": "Identifies the default command for a given group kind",
                              "type": "boolean",
                              "markdownDescription": "Identifies the default command for a given group kind"
                            },
                            "kind": {
                              "description": "Kind of group the command is part of",
                              "enum": [
                                "build",
                                "run",
                                "test",
                                "debug"
                              ],
                              "type": "string",
                              "markdownDescription": "Kind of group the command is part of"
                            }
                          },
                          "required": [
                            "kind"
                          ],
                          "type": "object",
                          "markdownDescription": "Defines the group this command is part of",
                          "additionalProperties": false
                        },
                        "hotReloadCapable": {
                          "description": "Whether the command is capable to reload itself when source code changes. If set to 'true' the command won't be restarted and it is expected to handle file changes on its own.\n\nDefault value is 'false'",
                          "type": "boolean",
                          "markdownDescription": "Whether the command is capable to reload itself when source code changes. If set to 'true' the command won't be restarted and it is expected to handle file changes on its own.\n\nDefault value is 'false'"
                        },
                        "label": {
                          "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                          "type": "string",
                          "markdownDescription": "Optional label that provides a label for this command to be used in Editor UI menus for example"
                        },
                        "workingDir": {
                          "description": "Working directory where the command should be executed\n\nSpecial variables that can be used:\n\n - '${PROJECTS_ROOT}': A path where projects sources are mounted\n\n - '${PROJECT_SOURCE}': A path to a project source (${PROJECTS_ROOT}/<project-name>). If there are multiple projects, this will point to the directory of the first one.",
                          "type": "string",
                          "markdownDescription": "Working directory where the command should be executed\n\nSpecial variables that can be used:\n\n - '${PROJECTS_ROOT}': A path where projects sources are mounted\n\n - '${PROJECT_SOURCE}': A path to a project source (${PROJECTS_ROOT}/<project-name>). If there are multiple projects, this will point to the directory of the first one."
                        }
                      },
                      "type": "object",
                      "markdownDescription": "CLI Command executed in an existing component container",
                      "additionalProperties": false
                    },
                    "id": {
                      "description": "Mandatory identifier that allows referencing this command in composite commands, from a parent, or in events.",
                      "type": "string",
                      "markdownDescription": "Mandatory identifier that allows referencing this command in composite commands, from a parent, or in events."
                    },
                    "vscodeLaunch": {
                      "description": "Command providing the definition of a VsCode launch action",
                      "properties": {
                        "attributes": {
                          "additionalProperties": {
                            "type": "string"
                          },
                          "description": "Optional map of free-form additional command attributes",
                          "type": "object",
                          "markdownDescription": "Optional map of free-form additional command attributes"
                        },
                        "group": {
                          "description": "Defines the group this command is part of",
                          "properties": {
                            "isDefault": {
                              "description": "Identifies the default command for a given group kind",
                              "type": "boolean",
                              "markdownDescription": "Identifies the default command for a given group kind"
                            },
                            "kind": {
                              "description": "Kind of group the command is part of",
                              "enum": [
                                "build",
                                "run",
                                "test",
                                "debug"
                              ],
                              "type": "string",
                              "markdownDescription": "Kind of group the command is part of"
                            }
                          },
                          "required": [
                            "kind"
                          ],
                          "type": "object",
                          "markdownDescription": "Defines the group this command is part of",
                          "additionalProperties": false
                        },
                        "inlined": {
                          "description": "Inlined content of the VsCode configuration",
                          "type": "string",
                          "markdownDescription": "Inlined content of the VsCode configuration"
                        },
                        "uri": {
                          "description": "Location as an absolute of relative URI the VsCode configuration will be fetched from",
                          "type": "string",
                          "markdownDescription": "Location as an absolute of relative URI the VsCode configuration will be fetched from"
                        }
                      },
                      "type": "object",
                      "markdownDescription": "Command providing the definition of a VsCode launch action",
                      "additionalProperties": false,
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
                      ]
                    },
                    "vscodeTask": {
                      "description": "Command providing the definition of a VsCode Task",
                      "properties": {
                        "attributes": {
                          "additionalProperties": {
                            "type": "string"
                          },
                          "description": "Optional map of free-form additional command attributes",
                          "type": "object",
                          "markdownDescription": "Optional map of free-form additional command attributes"
                        },
                        "group": {
                          "description": "Defines the group this command is part of",
                          "properties": {
                            "isDefault": {
                              "description": "Identifies the default command for a given group kind",
                              "type": "boolean",
                              "markdownDescription": "Identifies the default command for a given group kind"
                            },
                            "kind": {
                              "description": "Kind of group the command is part of",
                              "enum": [
                                "build",
                                "run",
                                "test",
                                "debug"
                              ],
                              "type": "string",
                              "markdownDescription": "Kind of group the command is part of"
                            }
                          },
                          "required": [
                            "kind"
                          ],
                          "type": "object",
                          "markdownDescription": "Defines the group this command is part of",
                          "additionalProperties": false
                        },
                        "inlined": {
                          "description": "Inlined content of the VsCode configuration",
                          "type": "string",
                          "markdownDescription": "Inlined content of the VsCode configuration"
                        },
                        "uri": {
                          "description": "Location as an absolute of relative URI the VsCode configuration will be fetched from",
                          "type": "string",
                          "markdownDescription": "Location as an absolute of relative URI the VsCode configuration will be fetched from"
                        }
                      },
                      "type": "object",
                      "markdownDescription": "Command providing the definition of a VsCode Task",
                      "additionalProperties": false,
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
                      ]
                    }
                  },
                  "required": [
                    "id"
                  ],
                  "type": "object",
                  "additionalProperties": false,
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
                        "vscodeTask"
                      ]
                    },
                    {
                      "required": [
                        "vscodeLaunch"
                      ]
                    },
                    {
                      "required": [
                        "composite"
                      ]
                    }
                  ]
                },
                "type": "array",
                "markdownDescription": "Overrides of commands encapsulated in a parent devfile or a plugin. Overriding is done using a strategic merge patch"
              },
              "components": {
                "description": "Overrides of components encapsulated in a plugin. Overriding is done using a strategic merge patch. A plugin cannot override embedded plugin components.",
                "items": {
                  "properties": {
                    "container": {
                      "description": "Configuration overriding for a Container component in a plugin",
                      "properties": {
                        "args": {
                          "description": "The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command.\n\nDefaults to an empty array, meaning use whatever is defined in the image.",
                          "items": {
                            "type": "string"
                          },
                          "type": "array",
                          "markdownDescription": "The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command.\n\nDefaults to an empty array, meaning use whatever is defined in the image."
                        },
                        "command": {
                          "description": "The command to run in the dockerimage component instead of the default one provided in the image.\n\nDefaults to an empty array, meaning use whatever is defined in the image.",
                          "items": {
                            "type": "string"
                          },
                          "type": "array",
                          "markdownDescription": "The command to run in the dockerimage component instead of the default one provided in the image.\n\nDefaults to an empty array, meaning use whatever is defined in the image."
                        },
                        "dedicatedPod": {
                          "description": "Specify if a container should run in its own separated pod, instead of running as part of the main development environment pod.\n\nDefault value is 'false'",
                          "type": "boolean",
                          "markdownDescription": "Specify if a container should run in its own separated pod, instead of running as part of the main development environment pod.\n\nDefault value is 'false'"
                        },
                        "endpoints": {
                          "items": {
                            "properties": {
                              "attributes": {
                                "additionalProperties": {
                                  "type": "string"
                                },
                                "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                                "type": "object",
                                "markdownDescription": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\","
                              },
                              "exposure": {
                                "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                                "enum": [
                                  "public",
                                  "internal",
                                  "none"
                                ],
                                "type": "string",
                                "markdownDescription": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'"
                              },
                              "name": {
                                "type": "string"
                              },
                              "path": {
                                "description": "Path of the endpoint URL",
                                "type": "string",
                                "markdownDescription": "Path of the endpoint URL"
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
                                ],
                                "markdownDescription": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'"
                              },
                              "secure": {
                                "description": "Describes whether the endpoint should be secured and protected by some authentication process",
                                "type": "boolean",
                                "markdownDescription": "Describes whether the endpoint should be secured and protected by some authentication process"
                              },
                              "targetPort": {
                                "type": "integer"
                              }
                            },
                            "required": [
                              "name"
                            ],
                            "type": "object",
                            "additionalProperties": false
                          },
                          "type": "array"
                        },
                        "env": {
                          "description": "Environment variables used in this container",
                          "items": {
                            "properties": {
                              "name": {
                                "type": "string"
                              },
                              "value": {
                                "type": "string"
                              }
                            },
                            "required": [
                              "name",
                              "value"
                            ],
                            "type": "object",
                            "additionalProperties": false
                          },
                          "type": "array",
                          "markdownDescription": "Environment variables used in this container"
                        },
                        "image": {
                          "type": "string"
                        },
                        "memoryLimit": {
                          "type": "string"
                        },
                        "mountSources": {
                          "type": "boolean"
                        },
                        "sourceMapping": {
                          "description": "Optional specification of the path in the container where project sources should be transferred/mounted when 'mountSources' is 'true'. When omitted, the value of the 'PROJECTS_ROOT' environment variable is used.",
                          "type": "string",
                          "markdownDescription": "Optional specification of the path in the container where project sources should be transferred/mounted when 'mountSources' is 'true'. When omitted, the value of the 'PROJECTS_ROOT' environment variable is used."
                        },
                        "volumeMounts": {
                          "description": "List of volumes mounts that should be mounted is this container.",
                          "items": {
                            "description": "Volume that should be mounted to a component container",
                            "properties": {
                              "name": {
                                "description": "The volume mount name is the name of an existing 'Volume' component. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files.",
                                "type": "string",
                                "markdownDescription": "The volume mount name is the name of an existing 'Volume' component. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files."
                              },
                              "path": {
                                "description": "The path in the component container where the volume should be mounted. If not path is mentioned, default path is the is '/<name>'.",
                                "type": "string",
                                "markdownDescription": "The path in the component container where the volume should be mounted. If not path is mentioned, default path is the is '/<name>'."
                              }
                            },
                            "required": [
                              "name"
                            ],
                            "type": "object",
                            "markdownDescription": "Volume that should be mounted to a component container",
                            "additionalProperties": false
                          },
                          "type": "array",
                          "markdownDescription": "List of volumes mounts that should be mounted is this container."
                        }
                      },
                      "type": "object",
                      "markdownDescription": "Configuration overriding for a Container component in a plugin",
                      "additionalProperties": false
                    },
                    "kubernetes": {
                      "description": "Configuration overriding for a Kubernetes component in a plugin",
                      "properties": {
                        "endpoints": {
                          "items": {
                            "properties": {
                              "attributes": {
                                "additionalProperties": {
                                  "type": "string"
                                },
                                "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                                "type": "object",
                                "markdownDescription": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\","
                              },
                              "exposure": {
                                "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                                "enum": [
                                  "public",
                                  "internal",
                                  "none"
                                ],
                                "type": "string",
                                "markdownDescription": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'"
                              },
                              "name": {
                                "type": "string"
                              },
                              "path": {
                                "description": "Path of the endpoint URL",
                                "type": "string",
                                "markdownDescription": "Path of the endpoint URL"
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
                                ],
                                "markdownDescription": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'"
                              },
                              "secure": {
                                "description": "Describes whether the endpoint should be secured and protected by some authentication process",
                                "type": "boolean",
                                "markdownDescription": "Describes whether the endpoint should be secured and protected by some authentication process"
                              },
                              "targetPort": {
                                "type": "integer"
                              }
                            },
                            "required": [
                              "name"
                            ],
                            "type": "object",
                            "additionalProperties": false
                          },
                          "type": "array"
                        },
                        "inlined": {
                          "description": "Inlined manifest",
                          "type": "string",
                          "markdownDescription": "Inlined manifest"
                        },
                        "uri": {
                          "description": "Location in a file fetched from a uri.",
                          "type": "string",
                          "markdownDescription": "Location in a file fetched from a uri."
                        }
                      },
                      "type": "object",
                      "markdownDescription": "Configuration overriding for a Kubernetes component in a plugin",
                      "additionalProperties": false,
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
                      ]
                    },
                    "name": {
                      "description": "Mandatory name that allows referencing the Volume component in Container volume mounts or inside a parent",
                      "type": "string",
                      "markdownDescription": "Mandatory name that allows referencing the Volume component in Container volume mounts or inside a parent"
                    },
                    "openshift": {
                      "description": "Configuration overriding for an OpenShift component in a plugin",
                      "properties": {
                        "endpoints": {
                          "items": {
                            "properties": {
                              "attributes": {
                                "additionalProperties": {
                                  "type": "string"
                                },
                                "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                                "type": "object",
                                "markdownDescription": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\","
                              },
                              "exposure": {
                                "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                                "enum": [
                                  "public",
                                  "internal",
                                  "none"
                                ],
                                "type": "string",
                                "markdownDescription": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'"
                              },
                              "name": {
                                "type": "string"
                              },
                              "path": {
                                "description": "Path of the endpoint URL",
                                "type": "string",
                                "markdownDescription": "Path of the endpoint URL"
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
                                ],
                                "markdownDescription": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'"
                              },
                              "secure": {
                                "description": "Describes whether the endpoint should be secured and protected by some authentication process",
                                "type": "boolean",
                                "markdownDescription": "Describes whether the endpoint should be secured and protected by some authentication process"
                              },
                              "targetPort": {
                                "type": "integer"
                              }
                            },
                            "required": [
                              "name"
                            ],
                            "type": "object",
                            "additionalProperties": false
                          },
                          "type": "array"
                        },
                        "inlined": {
                          "description": "Inlined manifest",
                          "type": "string",
                          "markdownDescription": "Inlined manifest"
                        },
                        "uri": {
                          "description": "Location in a file fetched from a uri.",
                          "type": "string",
                          "markdownDescription": "Location in a file fetched from a uri."
                        }
                      },
                      "type": "object",
                      "markdownDescription": "Configuration overriding for an OpenShift component in a plugin",
                      "additionalProperties": false,
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
                      ]
                    },
                    "volume": {
                      "description": "Configuration overriding for a Volume component in a plugin",
                      "properties": {
                        "size": {
                          "description": "Size of the volume",
                          "type": "string",
                          "markdownDescription": "Size of the volume"
                        }
                      },
                      "type": "object",
                      "markdownDescription": "Configuration overriding for a Volume component in a plugin",
                      "additionalProperties": false
                    }
                  },
                  "required": [
                    "name"
                  ],
                  "type": "object",
                  "additionalProperties": false,
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
                    }
                  ]
                },
                "type": "array",
                "markdownDescription": "Overrides of components encapsulated in a plugin. Overriding is done using a strategic merge patch. A plugin cannot override embedded plugin components."
              },
              "id": {
                "description": "Id in a registry that contains a Devfile yaml file",
                "type": "string",
                "markdownDescription": "Id in a registry that contains a Devfile yaml file"
              },
              "kubernetes": {
                "description": "Reference to a Kubernetes CRD of type DevWorkspaceTemplate",
                "properties": {
                  "name": {
                    "type": "string"
                  },
                  "namespace": {
                    "type": "string"
                  }
                },
                "required": [
                  "name"
                ],
                "type": "object",
                "markdownDescription": "Reference to a Kubernetes CRD of type DevWorkspaceTemplate",
                "additionalProperties": false
              },
              "registryUrl": {
                "type": "string"
              },
              "uri": {
                "description": "Uri of a Devfile yaml file",
                "type": "string",
                "markdownDescription": "Uri of a Devfile yaml file"
              }
            },
            "type": "object",
            "markdownDescription": "Allows importing a plugin.\n\nPlugins are mainly imported devfiles that contribute components, commands and events as a consistent single unit. They are defined in either YAML files following the devfile syntax, or as 'DevWorkspaceTemplate' Kubernetes Custom Resources",
            "additionalProperties": false,
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
            ]
          },
          "volume": {
            "description": "Allows specifying the definition of a volume shared by several other components",
            "properties": {
              "size": {
                "description": "Size of the volume",
                "type": "string",
                "markdownDescription": "Size of the volume"
              }
            },
            "type": "object",
            "markdownDescription": "Allows specifying the definition of a volume shared by several other components",
            "additionalProperties": false
          }
        },
        "required": [
          "name"
        ],
        "type": "object",
        "additionalProperties": false,
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
              "plugin"
            ]
          }
        ]
      },
      "type": "array",
      "markdownDescription": "List of the workspace components, such as editor and plugins, user-provided containers, or other types of components"
    },
    "events": {
      "description": "Bindings of commands to events. Each command is referred-to by its name.",
      "properties": {
        "postStart": {
          "description": "Names of commands that should be executed after the workspace is completely started. In the case of Che-Theia, these commands should be executed after all plugins and extensions have started, including project cloning. This means that those commands are not triggered until the user opens the IDE in his browser.",
          "items": {
            "type": "string"
          },
          "type": "array",
          "markdownDescription": "Names of commands that should be executed after the workspace is completely started. In the case of Che-Theia, these commands should be executed after all plugins and extensions have started, including project cloning. This means that those commands are not triggered until the user opens the IDE in his browser."
        },
        "postStop": {
          "description": "Names of commands that should be executed after stopping the workspace.",
          "items": {
            "type": "string"
          },
          "type": "array",
          "markdownDescription": "Names of commands that should be executed after stopping the workspace."
        },
        "preStart": {
          "description": "Names of commands that should be executed before the workspace start. Kubernetes-wise, these commands would typically be executed in init containers of the workspace POD.",
          "items": {
            "type": "string"
          },
          "type": "array",
          "markdownDescription": "Names of commands that should be executed before the workspace start. Kubernetes-wise, these commands would typically be executed in init containers of the workspace POD."
        },
        "preStop": {
          "description": "Names of commands that should be executed before stopping the workspace.",
          "items": {
            "type": "string"
          },
          "type": "array",
          "markdownDescription": "Names of commands that should be executed before stopping the workspace."
        }
      },
      "type": "object",
      "markdownDescription": "Bindings of commands to events. Each command is referred-to by its name.",
      "additionalProperties": false
    },
    "parent": {
      "description": "Parent workspace template",
      "properties": {
        "commands": {
          "description": "Overrides of commands encapsulated in a parent devfile or a plugin. Overriding is done using a strategic merge patch",
          "items": {
            "properties": {
              "apply": {
                "description": "Command that consists in applying a given component definition, typically bound to a workspace event.\n\nFor example, when an 'apply' command is bound to a 'preStart' event, and references a 'container' component, it will start the container as a K8S initContainer in the workspace POD, unless the component has its 'dedicatedPod' field set to 'true'.\n\nWhen no 'apply' command exist for a given component, it is assumed the component will be applied at workspace start by default.",
                "properties": {
                  "attributes": {
                    "additionalProperties": {
                      "type": "string"
                    },
                    "description": "Optional map of free-form additional command attributes",
                    "type": "object",
                    "markdownDescription": "Optional map of free-form additional command attributes"
                  },
                  "component": {
                    "description": "Describes component that will be applied",
                    "type": "string",
                    "markdownDescription": "Describes component that will be applied"
                  },
                  "group": {
                    "description": "Defines the group this command is part of",
                    "properties": {
                      "isDefault": {
                        "description": "Identifies the default command for a given group kind",
                        "type": "boolean",
                        "markdownDescription": "Identifies the default command for a given group kind"
                      },
                      "kind": {
                        "description": "Kind of group the command is part of",
                        "enum": [
                          "build",
                          "run",
                          "test",
                          "debug"
                        ],
                        "type": "string",
                        "markdownDescription": "Kind of group the command is part of"
                      }
                    },
                    "required": [
                      "kind"
                    ],
                    "type": "object",
                    "markdownDescription": "Defines the group this command is part of",
                    "additionalProperties": false
                  },
                  "label": {
                    "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                    "type": "string",
                    "markdownDescription": "Optional label that provides a label for this command to be used in Editor UI menus for example"
                  }
                },
                "type": "object",
                "markdownDescription": "Command that consists in applying a given component definition, typically bound to a workspace event.\n\nFor example, when an 'apply' command is bound to a 'preStart' event, and references a 'container' component, it will start the container as a K8S initContainer in the workspace POD, unless the component has its 'dedicatedPod' field set to 'true'.\n\nWhen no 'apply' command exist for a given component, it is assumed the component will be applied at workspace start by default.",
                "additionalProperties": false
              },
              "composite": {
                "description": "Composite command that allows executing several sub-commands either sequentially or concurrently",
                "properties": {
                  "attributes": {
                    "additionalProperties": {
                      "type": "string"
                    },
                    "description": "Optional map of free-form additional command attributes",
                    "type": "object",
                    "markdownDescription": "Optional map of free-form additional command attributes"
                  },
                  "commands": {
                    "description": "The commands that comprise this composite command",
                    "items": {
                      "type": "string"
                    },
                    "type": "array",
                    "markdownDescription": "The commands that comprise this composite command"
                  },
                  "group": {
                    "description": "Defines the group this command is part of",
                    "properties": {
                      "isDefault": {
                        "description": "Identifies the default command for a given group kind",
                        "type": "boolean",
                        "markdownDescription": "Identifies the default command for a given group kind"
                      },
                      "kind": {
                        "description": "Kind of group the command is part of",
                        "enum": [
                          "build",
                          "run",
                          "test",
                          "debug"
                        ],
                        "type": "string",
                        "markdownDescription": "Kind of group the command is part of"
                      }
                    },
                    "required": [
                      "kind"
                    ],
                    "type": "object",
                    "markdownDescription": "Defines the group this command is part of",
                    "additionalProperties": false
                  },
                  "label": {
                    "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                    "type": "string",
                    "markdownDescription": "Optional label that provides a label for this command to be used in Editor UI menus for example"
                  },
                  "parallel": {
                    "description": "Indicates if the sub-commands should be executed concurrently",
                    "type": "boolean",
                    "markdownDescription": "Indicates if the sub-commands should be executed concurrently"
                  }
                },
                "type": "object",
                "markdownDescription": "Composite command that allows executing several sub-commands either sequentially or concurrently",
                "additionalProperties": false
              },
              "exec": {
                "description": "CLI Command executed in an existing component container",
                "properties": {
                  "attributes": {
                    "additionalProperties": {
                      "type": "string"
                    },
                    "description": "Optional map of free-form additional command attributes",
                    "type": "object",
                    "markdownDescription": "Optional map of free-form additional command attributes"
                  },
                  "commandLine": {
                    "description": "The actual command-line string\n\nSpecial variables that can be used:\n\n - '$PROJECTS_ROOT': A path where projects sources are mounted\n\n - '$PROJECT_SOURCE': A path to a project source ($PROJECTS_ROOT/<project-name>). If there are multiple projects, this will point to the directory of the first one.",
                    "type": "string",
                    "markdownDescription": "The actual command-line string\n\nSpecial variables that can be used:\n\n - '$PROJECTS_ROOT': A path where projects sources are mounted\n\n - '$PROJECT_SOURCE': A path to a project source ($PROJECTS_ROOT/<project-name>). If there are multiple projects, this will point to the directory of the first one."
                  },
                  "component": {
                    "description": "Describes component to which given action relates",
                    "type": "string",
                    "markdownDescription": "Describes component to which given action relates"
                  },
                  "env": {
                    "description": "Optional list of environment variables that have to be set before running the command",
                    "items": {
                      "properties": {
                        "name": {
                          "type": "string"
                        },
                        "value": {
                          "type": "string"
                        }
                      },
                      "required": [
                        "name",
                        "value"
                      ],
                      "type": "object",
                      "additionalProperties": false
                    },
                    "type": "array",
                    "markdownDescription": "Optional list of environment variables that have to be set before running the command"
                  },
                  "group": {
                    "description": "Defines the group this command is part of",
                    "properties": {
                      "isDefault": {
                        "description": "Identifies the default command for a given group kind",
                        "type": "boolean",
                        "markdownDescription": "Identifies the default command for a given group kind"
                      },
                      "kind": {
                        "description": "Kind of group the command is part of",
                        "enum": [
                          "build",
                          "run",
                          "test",
                          "debug"
                        ],
                        "type": "string",
                        "markdownDescription": "Kind of group the command is part of"
                      }
                    },
                    "required": [
                      "kind"
                    ],
                    "type": "object",
                    "markdownDescription": "Defines the group this command is part of",
                    "additionalProperties": false
                  },
                  "hotReloadCapable": {
                    "description": "Whether the command is capable to reload itself when source code changes. If set to 'true' the command won't be restarted and it is expected to handle file changes on its own.\n\nDefault value is 'false'",
                    "type": "boolean",
                    "markdownDescription": "Whether the command is capable to reload itself when source code changes. If set to 'true' the command won't be restarted and it is expected to handle file changes on its own.\n\nDefault value is 'false'"
                  },
                  "label": {
                    "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                    "type": "string",
                    "markdownDescription": "Optional label that provides a label for this command to be used in Editor UI menus for example"
                  },
                  "workingDir": {
                    "description": "Working directory where the command should be executed\n\nSpecial variables that can be used:\n\n - '${PROJECTS_ROOT}': A path where projects sources are mounted\n\n - '${PROJECT_SOURCE}': A path to a project source (${PROJECTS_ROOT}/<project-name>). If there are multiple projects, this will point to the directory of the first one.",
                    "type": "string",
                    "markdownDescription": "Working directory where the command should be executed\n\nSpecial variables that can be used:\n\n - '${PROJECTS_ROOT}': A path where projects sources are mounted\n\n - '${PROJECT_SOURCE}': A path to a project source (${PROJECTS_ROOT}/<project-name>). If there are multiple projects, this will point to the directory of the first one."
                  }
                },
                "type": "object",
                "markdownDescription": "CLI Command executed in an existing component container",
                "additionalProperties": false
              },
              "id": {
                "description": "Mandatory identifier that allows referencing this command in composite commands, from a parent, or in events.",
                "type": "string",
                "markdownDescription": "Mandatory identifier that allows referencing this command in composite commands, from a parent, or in events."
              },
              "vscodeLaunch": {
                "description": "Command providing the definition of a VsCode launch action",
                "properties": {
                  "attributes": {
                    "additionalProperties": {
                      "type": "string"
                    },
                    "description": "Optional map of free-form additional command attributes",
                    "type": "object",
                    "markdownDescription": "Optional map of free-form additional command attributes"
                  },
                  "group": {
                    "description": "Defines the group this command is part of",
                    "properties": {
                      "isDefault": {
                        "description": "Identifies the default command for a given group kind",
                        "type": "boolean",
                        "markdownDescription": "Identifies the default command for a given group kind"
                      },
                      "kind": {
                        "description": "Kind of group the command is part of",
                        "enum": [
                          "build",
                          "run",
                          "test",
                          "debug"
                        ],
                        "type": "string",
                        "markdownDescription": "Kind of group the command is part of"
                      }
                    },
                    "required": [
                      "kind"
                    ],
                    "type": "object",
                    "markdownDescription": "Defines the group this command is part of",
                    "additionalProperties": false
                  },
                  "inlined": {
                    "description": "Inlined content of the VsCode configuration",
                    "type": "string",
                    "markdownDescription": "Inlined content of the VsCode configuration"
                  },
                  "uri": {
                    "description": "Location as an absolute of relative URI the VsCode configuration will be fetched from",
                    "type": "string",
                    "markdownDescription": "Location as an absolute of relative URI the VsCode configuration will be fetched from"
                  }
                },
                "type": "object",
                "markdownDescription": "Command providing the definition of a VsCode launch action",
                "additionalProperties": false,
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
                ]
              },
              "vscodeTask": {
                "description": "Command providing the definition of a VsCode Task",
                "properties": {
                  "attributes": {
                    "additionalProperties": {
                      "type": "string"
                    },
                    "description": "Optional map of free-form additional command attributes",
                    "type": "object",
                    "markdownDescription": "Optional map of free-form additional command attributes"
                  },
                  "group": {
                    "description": "Defines the group this command is part of",
                    "properties": {
                      "isDefault": {
                        "description": "Identifies the default command for a given group kind",
                        "type": "boolean",
                        "markdownDescription": "Identifies the default command for a given group kind"
                      },
                      "kind": {
                        "description": "Kind of group the command is part of",
                        "enum": [
                          "build",
                          "run",
                          "test",
                          "debug"
                        ],
                        "type": "string",
                        "markdownDescription": "Kind of group the command is part of"
                      }
                    },
                    "required": [
                      "kind"
                    ],
                    "type": "object",
                    "markdownDescription": "Defines the group this command is part of",
                    "additionalProperties": false
                  },
                  "inlined": {
                    "description": "Inlined content of the VsCode configuration",
                    "type": "string",
                    "markdownDescription": "Inlined content of the VsCode configuration"
                  },
                  "uri": {
                    "description": "Location as an absolute of relative URI the VsCode configuration will be fetched from",
                    "type": "string",
                    "markdownDescription": "Location as an absolute of relative URI the VsCode configuration will be fetched from"
                  }
                },
                "type": "object",
                "markdownDescription": "Command providing the definition of a VsCode Task",
                "additionalProperties": false,
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
                ]
              }
            },
            "required": [
              "id"
            ],
            "type": "object",
            "additionalProperties": false,
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
                  "vscodeTask"
                ]
              },
              {
                "required": [
                  "vscodeLaunch"
                ]
              },
              {
                "required": [
                  "composite"
                ]
              }
            ]
          },
          "type": "array",
          "markdownDescription": "Overrides of commands encapsulated in a parent devfile or a plugin. Overriding is done using a strategic merge patch"
        },
        "components": {
          "description": "Overrides of components encapsulated in a parent devfile. Overriding is done using a strategic merge patch",
          "items": {
            "properties": {
              "container": {
                "description": "Allows adding and configuring workspace-related containers",
                "properties": {
                  "args": {
                    "description": "The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command.\n\nDefaults to an empty array, meaning use whatever is defined in the image.",
                    "items": {
                      "type": "string"
                    },
                    "type": "array",
                    "markdownDescription": "The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command.\n\nDefaults to an empty array, meaning use whatever is defined in the image."
                  },
                  "command": {
                    "description": "The command to run in the dockerimage component instead of the default one provided in the image.\n\nDefaults to an empty array, meaning use whatever is defined in the image.",
                    "items": {
                      "type": "string"
                    },
                    "type": "array",
                    "markdownDescription": "The command to run in the dockerimage component instead of the default one provided in the image.\n\nDefaults to an empty array, meaning use whatever is defined in the image."
                  },
                  "dedicatedPod": {
                    "description": "Specify if a container should run in its own separated pod, instead of running as part of the main development environment pod.\n\nDefault value is 'false'",
                    "type": "boolean",
                    "markdownDescription": "Specify if a container should run in its own separated pod, instead of running as part of the main development environment pod.\n\nDefault value is 'false'"
                  },
                  "endpoints": {
                    "items": {
                      "properties": {
                        "attributes": {
                          "additionalProperties": {
                            "type": "string"
                          },
                          "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                          "type": "object",
                          "markdownDescription": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\","
                        },
                        "exposure": {
                          "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                          "enum": [
                            "public",
                            "internal",
                            "none"
                          ],
                          "type": "string",
                          "markdownDescription": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'"
                        },
                        "name": {
                          "type": "string"
                        },
                        "path": {
                          "description": "Path of the endpoint URL",
                          "type": "string",
                          "markdownDescription": "Path of the endpoint URL"
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
                          ],
                          "markdownDescription": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'"
                        },
                        "secure": {
                          "description": "Describes whether the endpoint should be secured and protected by some authentication process",
                          "type": "boolean",
                          "markdownDescription": "Describes whether the endpoint should be secured and protected by some authentication process"
                        },
                        "targetPort": {
                          "type": "integer"
                        }
                      },
                      "required": [
                        "name"
                      ],
                      "type": "object",
                      "additionalProperties": false
                    },
                    "type": "array"
                  },
                  "env": {
                    "description": "Environment variables used in this container",
                    "items": {
                      "properties": {
                        "name": {
                          "type": "string"
                        },
                        "value": {
                          "type": "string"
                        }
                      },
                      "required": [
                        "name",
                        "value"
                      ],
                      "type": "object",
                      "additionalProperties": false
                    },
                    "type": "array",
                    "markdownDescription": "Environment variables used in this container"
                  },
                  "image": {
                    "type": "string"
                  },
                  "memoryLimit": {
                    "type": "string"
                  },
                  "mountSources": {
                    "type": "boolean"
                  },
                  "sourceMapping": {
                    "description": "Optional specification of the path in the container where project sources should be transferred/mounted when 'mountSources' is 'true'. When omitted, the value of the 'PROJECTS_ROOT' environment variable is used.",
                    "type": "string",
                    "markdownDescription": "Optional specification of the path in the container where project sources should be transferred/mounted when 'mountSources' is 'true'. When omitted, the value of the 'PROJECTS_ROOT' environment variable is used."
                  },
                  "volumeMounts": {
                    "description": "List of volumes mounts that should be mounted is this container.",
                    "items": {
                      "description": "Volume that should be mounted to a component container",
                      "properties": {
                        "name": {
                          "description": "The volume mount name is the name of an existing 'Volume' component. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files.",
                          "type": "string",
                          "markdownDescription": "The volume mount name is the name of an existing 'Volume' component. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files."
                        },
                        "path": {
                          "description": "The path in the component container where the volume should be mounted. If not path is mentioned, default path is the is '/<name>'.",
                          "type": "string",
                          "markdownDescription": "The path in the component container where the volume should be mounted. If not path is mentioned, default path is the is '/<name>'."
                        }
                      },
                      "required": [
                        "name"
                      ],
                      "type": "object",
                      "markdownDescription": "Volume that should be mounted to a component container",
                      "additionalProperties": false
                    },
                    "type": "array",
                    "markdownDescription": "List of volumes mounts that should be mounted is this container."
                  }
                },
                "type": "object",
                "markdownDescription": "Allows adding and configuring workspace-related containers",
                "additionalProperties": false
              },
              "kubernetes": {
                "description": "Allows importing into the workspace the Kubernetes resources defined in a given manifest. For example this allows reusing the Kubernetes definitions used to deploy some runtime components in production.",
                "properties": {
                  "endpoints": {
                    "items": {
                      "properties": {
                        "attributes": {
                          "additionalProperties": {
                            "type": "string"
                          },
                          "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                          "type": "object",
                          "markdownDescription": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\","
                        },
                        "exposure": {
                          "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                          "enum": [
                            "public",
                            "internal",
                            "none"
                          ],
                          "type": "string",
                          "markdownDescription": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'"
                        },
                        "name": {
                          "type": "string"
                        },
                        "path": {
                          "description": "Path of the endpoint URL",
                          "type": "string",
                          "markdownDescription": "Path of the endpoint URL"
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
                          ],
                          "markdownDescription": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'"
                        },
                        "secure": {
                          "description": "Describes whether the endpoint should be secured and protected by some authentication process",
                          "type": "boolean",
                          "markdownDescription": "Describes whether the endpoint should be secured and protected by some authentication process"
                        },
                        "targetPort": {
                          "type": "integer"
                        }
                      },
                      "required": [
                        "name"
                      ],
                      "type": "object",
                      "additionalProperties": false
                    },
                    "type": "array"
                  },
                  "inlined": {
                    "description": "Inlined manifest",
                    "type": "string",
                    "markdownDescription": "Inlined manifest"
                  },
                  "uri": {
                    "description": "Location in a file fetched from a uri.",
                    "type": "string",
                    "markdownDescription": "Location in a file fetched from a uri."
                  }
                },
                "type": "object",
                "markdownDescription": "Allows importing into the workspace the Kubernetes resources defined in a given manifest. For example this allows reusing the Kubernetes definitions used to deploy some runtime components in production.",
                "additionalProperties": false,
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
                ]
              },
              "name": {
                "description": "Mandatory name that allows referencing the component from other elements (such as commands) or from an external devfile that may reference this component through a parent or a plugin.",
                "type": "string",
                "markdownDescription": "Mandatory name that allows referencing the component from other elements (such as commands) or from an external devfile that may reference this component through a parent or a plugin."
              },
              "openshift": {
                "description": "Allows importing into the workspace the OpenShift resources defined in a given manifest. For example this allows reusing the OpenShift definitions used to deploy some runtime components in production.",
                "properties": {
                  "endpoints": {
                    "items": {
                      "properties": {
                        "attributes": {
                          "additionalProperties": {
                            "type": "string"
                          },
                          "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                          "type": "object",
                          "markdownDescription": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\","
                        },
                        "exposure": {
                          "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                          "enum": [
                            "public",
                            "internal",
                            "none"
                          ],
                          "type": "string",
                          "markdownDescription": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'"
                        },
                        "name": {
                          "type": "string"
                        },
                        "path": {
                          "description": "Path of the endpoint URL",
                          "type": "string",
                          "markdownDescription": "Path of the endpoint URL"
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
                          ],
                          "markdownDescription": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'"
                        },
                        "secure": {
                          "description": "Describes whether the endpoint should be secured and protected by some authentication process",
                          "type": "boolean",
                          "markdownDescription": "Describes whether the endpoint should be secured and protected by some authentication process"
                        },
                        "targetPort": {
                          "type": "integer"
                        }
                      },
                      "required": [
                        "name"
                      ],
                      "type": "object",
                      "additionalProperties": false
                    },
                    "type": "array"
                  },
                  "inlined": {
                    "description": "Inlined manifest",
                    "type": "string",
                    "markdownDescription": "Inlined manifest"
                  },
                  "uri": {
                    "description": "Location in a file fetched from a uri.",
                    "type": "string",
                    "markdownDescription": "Location in a file fetched from a uri."
                  }
                },
                "type": "object",
                "markdownDescription": "Allows importing into the workspace the OpenShift resources defined in a given manifest. For example this allows reusing the OpenShift definitions used to deploy some runtime components in production.",
                "additionalProperties": false,
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
                ]
              },
              "plugin": {
                "description": "Allows importing a plugin.\n\nPlugins are mainly imported devfiles that contribute components, commands and events as a consistent single unit. They are defined in either YAML files following the devfile syntax, or as 'DevWorkspaceTemplate' Kubernetes Custom Resources",
                "properties": {
                  "commands": {
                    "description": "Overrides of commands encapsulated in a parent devfile or a plugin. Overriding is done using a strategic merge patch",
                    "items": {
                      "properties": {
                        "apply": {
                          "description": "Command that consists in applying a given component definition, typically bound to a workspace event.\n\nFor example, when an 'apply' command is bound to a 'preStart' event, and references a 'container' component, it will start the container as a K8S initContainer in the workspace POD, unless the component has its 'dedicatedPod' field set to 'true'.\n\nWhen no 'apply' command exist for a given component, it is assumed the component will be applied at workspace start by default.",
                          "properties": {
                            "attributes": {
                              "additionalProperties": {
                                "type": "string"
                              },
                              "description": "Optional map of free-form additional command attributes",
                              "type": "object",
                              "markdownDescription": "Optional map of free-form additional command attributes"
                            },
                            "component": {
                              "description": "Describes component that will be applied",
                              "type": "string",
                              "markdownDescription": "Describes component that will be applied"
                            },
                            "group": {
                              "description": "Defines the group this command is part of",
                              "properties": {
                                "isDefault": {
                                  "description": "Identifies the default command for a given group kind",
                                  "type": "boolean",
                                  "markdownDescription": "Identifies the default command for a given group kind"
                                },
                                "kind": {
                                  "description": "Kind of group the command is part of",
                                  "enum": [
                                    "build",
                                    "run",
                                    "test",
                                    "debug"
                                  ],
                                  "type": "string",
                                  "markdownDescription": "Kind of group the command is part of"
                                }
                              },
                              "required": [
                                "kind"
                              ],
                              "type": "object",
                              "markdownDescription": "Defines the group this command is part of",
                              "additionalProperties": false
                            },
                            "label": {
                              "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                              "type": "string",
                              "markdownDescription": "Optional label that provides a label for this command to be used in Editor UI menus for example"
                            }
                          },
                          "type": "object",
                          "markdownDescription": "Command that consists in applying a given component definition, typically bound to a workspace event.\n\nFor example, when an 'apply' command is bound to a 'preStart' event, and references a 'container' component, it will start the container as a K8S initContainer in the workspace POD, unless the component has its 'dedicatedPod' field set to 'true'.\n\nWhen no 'apply' command exist for a given component, it is assumed the component will be applied at workspace start by default.",
                          "additionalProperties": false
                        },
                        "composite": {
                          "description": "Composite command that allows executing several sub-commands either sequentially or concurrently",
                          "properties": {
                            "attributes": {
                              "additionalProperties": {
                                "type": "string"
                              },
                              "description": "Optional map of free-form additional command attributes",
                              "type": "object",
                              "markdownDescription": "Optional map of free-form additional command attributes"
                            },
                            "commands": {
                              "description": "The commands that comprise this composite command",
                              "items": {
                                "type": "string"
                              },
                              "type": "array",
                              "markdownDescription": "The commands that comprise this composite command"
                            },
                            "group": {
                              "description": "Defines the group this command is part of",
                              "properties": {
                                "isDefault": {
                                  "description": "Identifies the default command for a given group kind",
                                  "type": "boolean",
                                  "markdownDescription": "Identifies the default command for a given group kind"
                                },
                                "kind": {
                                  "description": "Kind of group the command is part of",
                                  "enum": [
                                    "build",
                                    "run",
                                    "test",
                                    "debug"
                                  ],
                                  "type": "string",
                                  "markdownDescription": "Kind of group the command is part of"
                                }
                              },
                              "required": [
                                "kind"
                              ],
                              "type": "object",
                              "markdownDescription": "Defines the group this command is part of",
                              "additionalProperties": false
                            },
                            "label": {
                              "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                              "type": "string",
                              "markdownDescription": "Optional label that provides a label for this command to be used in Editor UI menus for example"
                            },
                            "parallel": {
                              "description": "Indicates if the sub-commands should be executed concurrently",
                              "type": "boolean",
                              "markdownDescription": "Indicates if the sub-commands should be executed concurrently"
                            }
                          },
                          "type": "object",
                          "markdownDescription": "Composite command that allows executing several sub-commands either sequentially or concurrently",
                          "additionalProperties": false
                        },
                        "exec": {
                          "description": "CLI Command executed in an existing component container",
                          "properties": {
                            "attributes": {
                              "additionalProperties": {
                                "type": "string"
                              },
                              "description": "Optional map of free-form additional command attributes",
                              "type": "object",
                              "markdownDescription": "Optional map of free-form additional command attributes"
                            },
                            "commandLine": {
                              "description": "The actual command-line string\n\nSpecial variables that can be used:\n\n - '$PROJECTS_ROOT': A path where projects sources are mounted\n\n - '$PROJECT_SOURCE': A path to a project source ($PROJECTS_ROOT/<project-name>). If there are multiple projects, this will point to the directory of the first one.",
                              "type": "string",
                              "markdownDescription": "The actual command-line string\n\nSpecial variables that can be used:\n\n - '$PROJECTS_ROOT': A path where projects sources are mounted\n\n - '$PROJECT_SOURCE': A path to a project source ($PROJECTS_ROOT/<project-name>). If there are multiple projects, this will point to the directory of the first one."
                            },
                            "component": {
                              "description": "Describes component to which given action relates",
                              "type": "string",
                              "markdownDescription": "Describes component to which given action relates"
                            },
                            "env": {
                              "description": "Optional list of environment variables that have to be set before running the command",
                              "items": {
                                "properties": {
                                  "name": {
                                    "type": "string"
                                  },
                                  "value": {
                                    "type": "string"
                                  }
                                },
                                "required": [
                                  "name",
                                  "value"
                                ],
                                "type": "object",
                                "additionalProperties": false
                              },
                              "type": "array",
                              "markdownDescription": "Optional list of environment variables that have to be set before running the command"
                            },
                            "group": {
                              "description": "Defines the group this command is part of",
                              "properties": {
                                "isDefault": {
                                  "description": "Identifies the default command for a given group kind",
                                  "type": "boolean",
                                  "markdownDescription": "Identifies the default command for a given group kind"
                                },
                                "kind": {
                                  "description": "Kind of group the command is part of",
                                  "enum": [
                                    "build",
                                    "run",
                                    "test",
                                    "debug"
                                  ],
                                  "type": "string",
                                  "markdownDescription": "Kind of group the command is part of"
                                }
                              },
                              "required": [
                                "kind"
                              ],
                              "type": "object",
                              "markdownDescription": "Defines the group this command is part of",
                              "additionalProperties": false
                            },
                            "hotReloadCapable": {
                              "description": "Whether the command is capable to reload itself when source code changes. If set to 'true' the command won't be restarted and it is expected to handle file changes on its own.\n\nDefault value is 'false'",
                              "type": "boolean",
                              "markdownDescription": "Whether the command is capable to reload itself when source code changes. If set to 'true' the command won't be restarted and it is expected to handle file changes on its own.\n\nDefault value is 'false'"
                            },
                            "label": {
                              "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                              "type": "string",
                              "markdownDescription": "Optional label that provides a label for this command to be used in Editor UI menus for example"
                            },
                            "workingDir": {
                              "description": "Working directory where the command should be executed\n\nSpecial variables that can be used:\n\n - '${PROJECTS_ROOT}': A path where projects sources are mounted\n\n - '${PROJECT_SOURCE}': A path to a project source (${PROJECTS_ROOT}/<project-name>). If there are multiple projects, this will point to the directory of the first one.",
                              "type": "string",
                              "markdownDescription": "Working directory where the command should be executed\n\nSpecial variables that can be used:\n\n - '${PROJECTS_ROOT}': A path where projects sources are mounted\n\n - '${PROJECT_SOURCE}': A path to a project source (${PROJECTS_ROOT}/<project-name>). If there are multiple projects, this will point to the directory of the first one."
                            }
                          },
                          "type": "object",
                          "markdownDescription": "CLI Command executed in an existing component container",
                          "additionalProperties": false
                        },
                        "id": {
                          "description": "Mandatory identifier that allows referencing this command in composite commands, from a parent, or in events.",
                          "type": "string",
                          "markdownDescription": "Mandatory identifier that allows referencing this command in composite commands, from a parent, or in events."
                        },
                        "vscodeLaunch": {
                          "description": "Command providing the definition of a VsCode launch action",
                          "properties": {
                            "attributes": {
                              "additionalProperties": {
                                "type": "string"
                              },
                              "description": "Optional map of free-form additional command attributes",
                              "type": "object",
                              "markdownDescription": "Optional map of free-form additional command attributes"
                            },
                            "group": {
                              "description": "Defines the group this command is part of",
                              "properties": {
                                "isDefault": {
                                  "description": "Identifies the default command for a given group kind",
                                  "type": "boolean",
                                  "markdownDescription": "Identifies the default command for a given group kind"
                                },
                                "kind": {
                                  "description": "Kind of group the command is part of",
                                  "enum": [
                                    "build",
                                    "run",
                                    "test",
                                    "debug"
                                  ],
                                  "type": "string",
                                  "markdownDescription": "Kind of group the command is part of"
                                }
                              },
                              "required": [
                                "kind"
                              ],
                              "type": "object",
                              "markdownDescription": "Defines the group this command is part of",
                              "additionalProperties": false
                            },
                            "inlined": {
                              "description": "Inlined content of the VsCode configuration",
                              "type": "string",
                              "markdownDescription": "Inlined content of the VsCode configuration"
                            },
                            "uri": {
                              "description": "Location as an absolute of relative URI the VsCode configuration will be fetched from",
                              "type": "string",
                              "markdownDescription": "Location as an absolute of relative URI the VsCode configuration will be fetched from"
                            }
                          },
                          "type": "object",
                          "markdownDescription": "Command providing the definition of a VsCode launch action",
                          "additionalProperties": false,
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
                          ]
                        },
                        "vscodeTask": {
                          "description": "Command providing the definition of a VsCode Task",
                          "properties": {
                            "attributes": {
                              "additionalProperties": {
                                "type": "string"
                              },
                              "description": "Optional map of free-form additional command attributes",
                              "type": "object",
                              "markdownDescription": "Optional map of free-form additional command attributes"
                            },
                            "group": {
                              "description": "Defines the group this command is part of",
                              "properties": {
                                "isDefault": {
                                  "description": "Identifies the default command for a given group kind",
                                  "type": "boolean",
                                  "markdownDescription": "Identifies the default command for a given group kind"
                                },
                                "kind": {
                                  "description": "Kind of group the command is part of",
                                  "enum": [
                                    "build",
                                    "run",
                                    "test",
                                    "debug"
                                  ],
                                  "type": "string",
                                  "markdownDescription": "Kind of group the command is part of"
                                }
                              },
                              "required": [
                                "kind"
                              ],
                              "type": "object",
                              "markdownDescription": "Defines the group this command is part of",
                              "additionalProperties": false
                            },
                            "inlined": {
                              "description": "Inlined content of the VsCode configuration",
                              "type": "string",
                              "markdownDescription": "Inlined content of the VsCode configuration"
                            },
                            "uri": {
                              "description": "Location as an absolute of relative URI the VsCode configuration will be fetched from",
                              "type": "string",
                              "markdownDescription": "Location as an absolute of relative URI the VsCode configuration will be fetched from"
                            }
                          },
                          "type": "object",
                          "markdownDescription": "Command providing the definition of a VsCode Task",
                          "additionalProperties": false,
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
                          ]
                        }
                      },
                      "required": [
                        "id"
                      ],
                      "type": "object",
                      "additionalProperties": false,
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
                            "vscodeTask"
                          ]
                        },
                        {
                          "required": [
                            "vscodeLaunch"
                          ]
                        },
                        {
                          "required": [
                            "composite"
                          ]
                        }
                      ]
                    },
                    "type": "array",
                    "markdownDescription": "Overrides of commands encapsulated in a parent devfile or a plugin. Overriding is done using a strategic merge patch"
                  },
                  "components": {
                    "description": "Overrides of components encapsulated in a plugin. Overriding is done using a strategic merge patch. A plugin cannot override embedded plugin components.",
                    "items": {
                      "properties": {
                        "container": {
                          "description": "Configuration overriding for a Container component in a plugin",
                          "properties": {
                            "args": {
                              "description": "The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command.\n\nDefaults to an empty array, meaning use whatever is defined in the image.",
                              "items": {
                                "type": "string"
                              },
                              "type": "array",
                              "markdownDescription": "The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command.\n\nDefaults to an empty array, meaning use whatever is defined in the image."
                            },
                            "command": {
                              "description": "The command to run in the dockerimage component instead of the default one provided in the image.\n\nDefaults to an empty array, meaning use whatever is defined in the image.",
                              "items": {
                                "type": "string"
                              },
                              "type": "array",
                              "markdownDescription": "The command to run in the dockerimage component instead of the default one provided in the image.\n\nDefaults to an empty array, meaning use whatever is defined in the image."
                            },
                            "dedicatedPod": {
                              "description": "Specify if a container should run in its own separated pod, instead of running as part of the main development environment pod.\n\nDefault value is 'false'",
                              "type": "boolean",
                              "markdownDescription": "Specify if a container should run in its own separated pod, instead of running as part of the main development environment pod.\n\nDefault value is 'false'"
                            },
                            "endpoints": {
                              "items": {
                                "properties": {
                                  "attributes": {
                                    "additionalProperties": {
                                      "type": "string"
                                    },
                                    "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                                    "type": "object",
                                    "markdownDescription": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\","
                                  },
                                  "exposure": {
                                    "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                                    "enum": [
                                      "public",
                                      "internal",
                                      "none"
                                    ],
                                    "type": "string",
                                    "markdownDescription": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'"
                                  },
                                  "name": {
                                    "type": "string"
                                  },
                                  "path": {
                                    "description": "Path of the endpoint URL",
                                    "type": "string",
                                    "markdownDescription": "Path of the endpoint URL"
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
                                    ],
                                    "markdownDescription": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'"
                                  },
                                  "secure": {
                                    "description": "Describes whether the endpoint should be secured and protected by some authentication process",
                                    "type": "boolean",
                                    "markdownDescription": "Describes whether the endpoint should be secured and protected by some authentication process"
                                  },
                                  "targetPort": {
                                    "type": "integer"
                                  }
                                },
                                "required": [
                                  "name"
                                ],
                                "type": "object",
                                "additionalProperties": false
                              },
                              "type": "array"
                            },
                            "env": {
                              "description": "Environment variables used in this container",
                              "items": {
                                "properties": {
                                  "name": {
                                    "type": "string"
                                  },
                                  "value": {
                                    "type": "string"
                                  }
                                },
                                "required": [
                                  "name",
                                  "value"
                                ],
                                "type": "object",
                                "additionalProperties": false
                              },
                              "type": "array",
                              "markdownDescription": "Environment variables used in this container"
                            },
                            "image": {
                              "type": "string"
                            },
                            "memoryLimit": {
                              "type": "string"
                            },
                            "mountSources": {
                              "type": "boolean"
                            },
                            "sourceMapping": {
                              "description": "Optional specification of the path in the container where project sources should be transferred/mounted when 'mountSources' is 'true'. When omitted, the value of the 'PROJECTS_ROOT' environment variable is used.",
                              "type": "string",
                              "markdownDescription": "Optional specification of the path in the container where project sources should be transferred/mounted when 'mountSources' is 'true'. When omitted, the value of the 'PROJECTS_ROOT' environment variable is used."
                            },
                            "volumeMounts": {
                              "description": "List of volumes mounts that should be mounted is this container.",
                              "items": {
                                "description": "Volume that should be mounted to a component container",
                                "properties": {
                                  "name": {
                                    "description": "The volume mount name is the name of an existing 'Volume' component. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files.",
                                    "type": "string",
                                    "markdownDescription": "The volume mount name is the name of an existing 'Volume' component. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files."
                                  },
                                  "path": {
                                    "description": "The path in the component container where the volume should be mounted. If not path is mentioned, default path is the is '/<name>'.",
                                    "type": "string",
                                    "markdownDescription": "The path in the component container where the volume should be mounted. If not path is mentioned, default path is the is '/<name>'."
                                  }
                                },
                                "required": [
                                  "name"
                                ],
                                "type": "object",
                                "markdownDescription": "Volume that should be mounted to a component container",
                                "additionalProperties": false
                              },
                              "type": "array",
                              "markdownDescription": "List of volumes mounts that should be mounted is this container."
                            }
                          },
                          "type": "object",
                          "markdownDescription": "Configuration overriding for a Container component in a plugin",
                          "additionalProperties": false
                        },
                        "kubernetes": {
                          "description": "Configuration overriding for a Kubernetes component in a plugin",
                          "properties": {
                            "endpoints": {
                              "items": {
                                "properties": {
                                  "attributes": {
                                    "additionalProperties": {
                                      "type": "string"
                                    },
                                    "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                                    "type": "object",
                                    "markdownDescription": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\","
                                  },
                                  "exposure": {
                                    "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                                    "enum": [
                                      "public",
                                      "internal",
                                      "none"
                                    ],
                                    "type": "string",
                                    "markdownDescription": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'"
                                  },
                                  "name": {
                                    "type": "string"
                                  },
                                  "path": {
                                    "description": "Path of the endpoint URL",
                                    "type": "string",
                                    "markdownDescription": "Path of the endpoint URL"
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
                                    ],
                                    "markdownDescription": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'"
                                  },
                                  "secure": {
                                    "description": "Describes whether the endpoint should be secured and protected by some authentication process",
                                    "type": "boolean",
                                    "markdownDescription": "Describes whether the endpoint should be secured and protected by some authentication process"
                                  },
                                  "targetPort": {
                                    "type": "integer"
                                  }
                                },
                                "required": [
                                  "name"
                                ],
                                "type": "object",
                                "additionalProperties": false
                              },
                              "type": "array"
                            },
                            "inlined": {
                              "description": "Inlined manifest",
                              "type": "string",
                              "markdownDescription": "Inlined manifest"
                            },
                            "uri": {
                              "description": "Location in a file fetched from a uri.",
                              "type": "string",
                              "markdownDescription": "Location in a file fetched from a uri."
                            }
                          },
                          "type": "object",
                          "markdownDescription": "Configuration overriding for a Kubernetes component in a plugin",
                          "additionalProperties": false,
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
                          ]
                        },
                        "name": {
                          "description": "Mandatory name that allows referencing the Volume component in Container volume mounts or inside a parent",
                          "type": "string",
                          "markdownDescription": "Mandatory name that allows referencing the Volume component in Container volume mounts or inside a parent"
                        },
                        "openshift": {
                          "description": "Configuration overriding for an OpenShift component in a plugin",
                          "properties": {
                            "endpoints": {
                              "items": {
                                "properties": {
                                  "attributes": {
                                    "additionalProperties": {
                                      "type": "string"
                                    },
                                    "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                                    "type": "object",
                                    "markdownDescription": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\","
                                  },
                                  "exposure": {
                                    "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                                    "enum": [
                                      "public",
                                      "internal",
                                      "none"
                                    ],
                                    "type": "string",
                                    "markdownDescription": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'"
                                  },
                                  "name": {
                                    "type": "string"
                                  },
                                  "path": {
                                    "description": "Path of the endpoint URL",
                                    "type": "string",
                                    "markdownDescription": "Path of the endpoint URL"
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
                                    ],
                                    "markdownDescription": "Describes the application and transport protocols of the traffic that will go through this endpoint.\n- 'http': Endpoint will have 'http' traffic, typically on a TCP connection. It will be automaticaly promoted to 'https' when the 'secure' field is set to 'true'.\n- 'https': Endpoint will have 'https' traffic, typically on a TCP connection.\n- 'ws': Endpoint will have 'ws' traffic, typically on a TCP connection. It will be automaticaly promoted to 'wss' when the 'secure' field is set to 'true'.\n- 'wss': Endpoint will have 'wss' traffic, typically on a TCP connection.\n- 'tcp': Endpoint will have traffic on a TCP connection, without specifying an application protocol.\n- 'udp': Endpoint will have traffic on an UDP connection, without specifying an application protocol.\n\nDefault value is 'http'"
                                  },
                                  "secure": {
                                    "description": "Describes whether the endpoint should be secured and protected by some authentication process",
                                    "type": "boolean",
                                    "markdownDescription": "Describes whether the endpoint should be secured and protected by some authentication process"
                                  },
                                  "targetPort": {
                                    "type": "integer"
                                  }
                                },
                                "required": [
                                  "name"
                                ],
                                "type": "object",
                                "additionalProperties": false
                              },
                              "type": "array"
                            },
                            "inlined": {
                              "description": "Inlined manifest",
                              "type": "string",
                              "markdownDescription": "Inlined manifest"
                            },
                            "uri": {
                              "description": "Location in a file fetched from a uri.",
                              "type": "string",
                              "markdownDescription": "Location in a file fetched from a uri."
                            }
                          },
                          "type": "object",
                          "markdownDescription": "Configuration overriding for an OpenShift component in a plugin",
                          "additionalProperties": false,
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
                          ]
                        },
                        "volume": {
                          "description": "Configuration overriding for a Volume component in a plugin",
                          "properties": {
                            "size": {
                              "description": "Size of the volume",
                              "type": "string",
                              "markdownDescription": "Size of the volume"
                            }
                          },
                          "type": "object",
                          "markdownDescription": "Configuration overriding for a Volume component in a plugin",
                          "additionalProperties": false
                        }
                      },
                      "required": [
                        "name"
                      ],
                      "type": "object",
                      "additionalProperties": false,
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
                        }
                      ]
                    },
                    "type": "array",
                    "markdownDescription": "Overrides of components encapsulated in a plugin. Overriding is done using a strategic merge patch. A plugin cannot override embedded plugin components."
                  },
                  "id": {
                    "description": "Id in a registry that contains a Devfile yaml file",
                    "type": "string",
                    "markdownDescription": "Id in a registry that contains a Devfile yaml file"
                  },
                  "kubernetes": {
                    "description": "Reference to a Kubernetes CRD of type DevWorkspaceTemplate",
                    "properties": {
                      "name": {
                        "type": "string"
                      },
                      "namespace": {
                        "type": "string"
                      }
                    },
                    "required": [
                      "name"
                    ],
                    "type": "object",
                    "markdownDescription": "Reference to a Kubernetes CRD of type DevWorkspaceTemplate",
                    "additionalProperties": false
                  },
                  "registryUrl": {
                    "type": "string"
                  },
                  "uri": {
                    "description": "Uri of a Devfile yaml file",
                    "type": "string",
                    "markdownDescription": "Uri of a Devfile yaml file"
                  }
                },
                "type": "object",
                "markdownDescription": "Allows importing a plugin.\n\nPlugins are mainly imported devfiles that contribute components, commands and events as a consistent single unit. They are defined in either YAML files following the devfile syntax, or as 'DevWorkspaceTemplate' Kubernetes Custom Resources",
                "additionalProperties": false,
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
                ]
              },
              "volume": {
                "description": "Allows specifying the definition of a volume shared by several other components",
                "properties": {
                  "size": {
                    "description": "Size of the volume",
                    "type": "string",
                    "markdownDescription": "Size of the volume"
                  }
                },
                "type": "object",
                "markdownDescription": "Allows specifying the definition of a volume shared by several other components",
                "additionalProperties": false
              }
            },
            "required": [
              "name"
            ],
            "type": "object",
            "additionalProperties": false,
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
                  "plugin"
                ]
              }
            ]
          },
          "type": "array",
          "markdownDescription": "Overrides of components encapsulated in a parent devfile. Overriding is done using a strategic merge patch"
        },
        "id": {
          "description": "Id in a registry that contains a Devfile yaml file",
          "type": "string",
          "markdownDescription": "Id in a registry that contains a Devfile yaml file"
        },
        "kubernetes": {
          "description": "Reference to a Kubernetes CRD of type DevWorkspaceTemplate",
          "properties": {
            "name": {
              "type": "string"
            },
            "namespace": {
              "type": "string"
            }
          },
          "required": [
            "name"
          ],
          "type": "object",
          "markdownDescription": "Reference to a Kubernetes CRD of type DevWorkspaceTemplate",
          "additionalProperties": false
        },
        "projects": {
          "description": "Overrides of projects encapsulated in a parent devfile. Overriding is done using a strategic merge patch.",
          "items": {
            "properties": {
              "clonePath": {
                "description": "Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name.",
                "type": "string",
                "markdownDescription": "Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name."
              },
              "git": {
                "description": "Project's Git source",
                "properties": {
                  "checkoutFrom": {
                    "description": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                    "properties": {
                      "remote": {
                        "description": "The remote name should be used as init. Required if there are more than one remote configured",
                        "type": "string",
                        "markdownDescription": "The remote name should be used as init. Required if there are more than one remote configured"
                      },
                      "revision": {
                        "description": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found.",
                        "type": "string",
                        "markdownDescription": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found."
                      }
                    },
                    "type": "object",
                    "markdownDescription": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                    "additionalProperties": false
                  },
                  "remotes": {
                    "additionalProperties": {
                      "type": "string"
                    },
                    "description": "The remotes map which should be initialized in the git project. Must have at least one remote configured",
                    "type": "object",
                    "markdownDescription": "The remotes map which should be initialized in the git project. Must have at least one remote configured"
                  },
                  "sparseCheckoutDir": {
                    "description": "Part of project to populate in the working directory.",
                    "type": "string",
                    "markdownDescription": "Part of project to populate in the working directory."
                  }
                },
                "type": "object",
                "markdownDescription": "Project's Git source",
                "additionalProperties": false
              },
              "github": {
                "description": "Project's GitHub source",
                "properties": {
                  "checkoutFrom": {
                    "description": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                    "properties": {
                      "remote": {
                        "description": "The remote name should be used as init. Required if there are more than one remote configured",
                        "type": "string",
                        "markdownDescription": "The remote name should be used as init. Required if there are more than one remote configured"
                      },
                      "revision": {
                        "description": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found.",
                        "type": "string",
                        "markdownDescription": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found."
                      }
                    },
                    "type": "object",
                    "markdownDescription": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                    "additionalProperties": false
                  },
                  "remotes": {
                    "additionalProperties": {
                      "type": "string"
                    },
                    "description": "The remotes map which should be initialized in the git project. Must have at least one remote configured",
                    "type": "object",
                    "markdownDescription": "The remotes map which should be initialized in the git project. Must have at least one remote configured"
                  },
                  "sparseCheckoutDir": {
                    "description": "Part of project to populate in the working directory.",
                    "type": "string",
                    "markdownDescription": "Part of project to populate in the working directory."
                  }
                },
                "type": "object",
                "markdownDescription": "Project's GitHub source",
                "additionalProperties": false
              },
              "name": {
                "description": "Project name",
                "type": "string",
                "markdownDescription": "Project name"
              },
              "zip": {
                "description": "Project's Zip source",
                "properties": {
                  "location": {
                    "description": "Zip project's source location address. Should be file path of the archive, e.g. file://$FILE_PATH",
                    "type": "string",
                    "markdownDescription": "Zip project's source location address. Should be file path of the archive, e.g. file://$FILE_PATH"
                  },
                  "sparseCheckoutDir": {
                    "description": "Part of project to populate in the working directory.",
                    "type": "string",
                    "markdownDescription": "Part of project to populate in the working directory."
                  }
                },
                "type": "object",
                "markdownDescription": "Project's Zip source",
                "additionalProperties": false
              }
            },
            "required": [
              "name"
            ],
            "type": "object",
            "additionalProperties": false,
            "oneOf": [
              {
                "required": [
                  "git"
                ]
              },
              {
                "required": [
                  "github"
                ]
              },
              {
                "required": [
                  "zip"
                ]
              }
            ]
          },
          "type": "array",
          "markdownDescription": "Overrides of projects encapsulated in a parent devfile. Overriding is done using a strategic merge patch."
        },
        "registryUrl": {
          "type": "string"
        },
        "starterProjects": {
          "description": "Overrides of startedProjects encapsulated in a parent devfile. Overriding is done using a strategic merge patch.",
          "items": {
            "properties": {
              "clonePath": {
                "description": "Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name.",
                "type": "string",
                "markdownDescription": "Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name."
              },
              "description": {
                "description": "Description of a starter project",
                "type": "string",
                "markdownDescription": "Description of a starter project"
              },
              "git": {
                "description": "Project's Git source",
                "properties": {
                  "checkoutFrom": {
                    "description": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                    "properties": {
                      "remote": {
                        "description": "The remote name should be used as init. Required if there are more than one remote configured",
                        "type": "string",
                        "markdownDescription": "The remote name should be used as init. Required if there are more than one remote configured"
                      },
                      "revision": {
                        "description": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found.",
                        "type": "string",
                        "markdownDescription": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found."
                      }
                    },
                    "type": "object",
                    "markdownDescription": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                    "additionalProperties": false
                  },
                  "remotes": {
                    "additionalProperties": {
                      "type": "string"
                    },
                    "description": "The remotes map which should be initialized in the git project. Must have at least one remote configured",
                    "type": "object",
                    "markdownDescription": "The remotes map which should be initialized in the git project. Must have at least one remote configured"
                  },
                  "sparseCheckoutDir": {
                    "description": "Part of project to populate in the working directory.",
                    "type": "string",
                    "markdownDescription": "Part of project to populate in the working directory."
                  }
                },
                "type": "object",
                "markdownDescription": "Project's Git source",
                "additionalProperties": false
              },
              "github": {
                "description": "Project's GitHub source",
                "properties": {
                  "checkoutFrom": {
                    "description": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                    "properties": {
                      "remote": {
                        "description": "The remote name should be used as init. Required if there are more than one remote configured",
                        "type": "string",
                        "markdownDescription": "The remote name should be used as init. Required if there are more than one remote configured"
                      },
                      "revision": {
                        "description": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found.",
                        "type": "string",
                        "markdownDescription": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found."
                      }
                    },
                    "type": "object",
                    "markdownDescription": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                    "additionalProperties": false
                  },
                  "remotes": {
                    "additionalProperties": {
                      "type": "string"
                    },
                    "description": "The remotes map which should be initialized in the git project. Must have at least one remote configured",
                    "type": "object",
                    "markdownDescription": "The remotes map which should be initialized in the git project. Must have at least one remote configured"
                  },
                  "sparseCheckoutDir": {
                    "description": "Part of project to populate in the working directory.",
                    "type": "string",
                    "markdownDescription": "Part of project to populate in the working directory."
                  }
                },
                "type": "object",
                "markdownDescription": "Project's GitHub source",
                "additionalProperties": false
              },
              "name": {
                "description": "Project name",
                "type": "string",
                "markdownDescription": "Project name"
              },
              "zip": {
                "description": "Project's Zip source",
                "properties": {
                  "location": {
                    "description": "Zip project's source location address. Should be file path of the archive, e.g. file://$FILE_PATH",
                    "type": "string",
                    "markdownDescription": "Zip project's source location address. Should be file path of the archive, e.g. file://$FILE_PATH"
                  },
                  "sparseCheckoutDir": {
                    "description": "Part of project to populate in the working directory.",
                    "type": "string",
                    "markdownDescription": "Part of project to populate in the working directory."
                  }
                },
                "type": "object",
                "markdownDescription": "Project's Zip source",
                "additionalProperties": false
              },
              "markdownDescription": {
                "description": "Description of a starter project",
                "type": "string",
                "markdownDescription": "Description of a starter project"
              }
            },
            "required": [
              "name"
            ],
            "type": "object",
            "additionalProperties": false,
            "oneOf": [
              {
                "required": [
                  "git"
                ]
              },
              {
                "required": [
                  "github"
                ]
              },
              {
                "required": [
                  "zip"
                ]
              }
            ]
          },
          "type": "array",
          "markdownDescription": "Overrides of startedProjects encapsulated in a parent devfile. Overriding is done using a strategic merge patch."
        },
        "uri": {
          "description": "Uri of a Devfile yaml file",
          "type": "string",
          "markdownDescription": "Uri of a Devfile yaml file"
        }
      },
      "type": "object",
      "markdownDescription": "Parent workspace template",
      "additionalProperties": false,
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
      ]
    },
    "projects": {
      "description": "Projects worked on in the workspace, containing names and sources locations",
      "items": {
        "properties": {
          "clonePath": {
            "description": "Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name.",
            "type": "string",
            "markdownDescription": "Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name."
          },
          "git": {
            "description": "Project's Git source",
            "properties": {
              "checkoutFrom": {
                "description": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                "properties": {
                  "remote": {
                    "description": "The remote name should be used as init. Required if there are more than one remote configured",
                    "type": "string",
                    "markdownDescription": "The remote name should be used as init. Required if there are more than one remote configured"
                  },
                  "revision": {
                    "description": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found.",
                    "type": "string",
                    "markdownDescription": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found."
                  }
                },
                "type": "object",
                "markdownDescription": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                "additionalProperties": false
              },
              "remotes": {
                "additionalProperties": {
                  "type": "string"
                },
                "description": "The remotes map which should be initialized in the git project. Must have at least one remote configured",
                "type": "object",
                "markdownDescription": "The remotes map which should be initialized in the git project. Must have at least one remote configured"
              },
              "sparseCheckoutDir": {
                "description": "Part of project to populate in the working directory.",
                "type": "string",
                "markdownDescription": "Part of project to populate in the working directory."
              }
            },
            "type": "object",
            "markdownDescription": "Project's Git source",
            "additionalProperties": false,
            "required": [
              "remotes"
            ]
          },
          "github": {
            "description": "Project's GitHub source",
            "properties": {
              "checkoutFrom": {
                "description": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                "properties": {
                  "remote": {
                    "description": "The remote name should be used as init. Required if there are more than one remote configured",
                    "type": "string",
                    "markdownDescription": "The remote name should be used as init. Required if there are more than one remote configured"
                  },
                  "revision": {
                    "description": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found.",
                    "type": "string",
                    "markdownDescription": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found."
                  }
                },
                "type": "object",
                "markdownDescription": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                "additionalProperties": false
              },
              "remotes": {
                "additionalProperties": {
                  "type": "string"
                },
                "description": "The remotes map which should be initialized in the git project. Must have at least one remote configured",
                "type": "object",
                "markdownDescription": "The remotes map which should be initialized in the git project. Must have at least one remote configured"
              },
              "sparseCheckoutDir": {
                "description": "Part of project to populate in the working directory.",
                "type": "string",
                "markdownDescription": "Part of project to populate in the working directory."
              }
            },
            "type": "object",
            "markdownDescription": "Project's GitHub source",
            "additionalProperties": false,
            "required": [
              "remotes"
            ]
          },
          "name": {
            "description": "Project name",
            "type": "string",
            "markdownDescription": "Project name"
          },
          "zip": {
            "description": "Project's Zip source",
            "properties": {
              "location": {
                "description": "Zip project's source location address. Should be file path of the archive, e.g. file://$FILE_PATH",
                "type": "string",
                "markdownDescription": "Zip project's source location address. Should be file path of the archive, e.g. file://$FILE_PATH"
              },
              "sparseCheckoutDir": {
                "description": "Part of project to populate in the working directory.",
                "type": "string",
                "markdownDescription": "Part of project to populate in the working directory."
              }
            },
            "type": "object",
            "markdownDescription": "Project's Zip source",
            "additionalProperties": false
          }
        },
        "required": [
          "name"
        ],
        "type": "object",
        "additionalProperties": false,
        "oneOf": [
          {
            "required": [
              "git"
            ]
          },
          {
            "required": [
              "github"
            ]
          },
          {
            "required": [
              "zip"
            ]
          }
        ]
      },
      "type": "array",
      "markdownDescription": "Projects worked on in the workspace, containing names and sources locations"
    },
    "starterProjects": {
      "description": "StarterProjects is a project that can be used as a starting point when bootstrapping new projects",
      "items": {
        "properties": {
          "clonePath": {
            "description": "Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name.",
            "type": "string",
            "markdownDescription": "Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name."
          },
          "description": {
            "description": "Description of a starter project",
            "type": "string",
            "markdownDescription": "Description of a starter project"
          },
          "git": {
            "description": "Project's Git source",
            "properties": {
              "checkoutFrom": {
                "description": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                "properties": {
                  "remote": {
                    "description": "The remote name should be used as init. Required if there are more than one remote configured",
                    "type": "string",
                    "markdownDescription": "The remote name should be used as init. Required if there are more than one remote configured"
                  },
                  "revision": {
                    "description": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found.",
                    "type": "string",
                    "markdownDescription": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found."
                  }
                },
                "type": "object",
                "markdownDescription": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                "additionalProperties": false
              },
              "remotes": {
                "additionalProperties": {
                  "type": "string"
                },
                "description": "The remotes map which should be initialized in the git project. Must have at least one remote configured",
                "type": "object",
                "markdownDescription": "The remotes map which should be initialized in the git project. Must have at least one remote configured"
              },
              "sparseCheckoutDir": {
                "description": "Part of project to populate in the working directory.",
                "type": "string",
                "markdownDescription": "Part of project to populate in the working directory."
              }
            },
            "type": "object",
            "markdownDescription": "Project's Git source",
            "additionalProperties": false
          },
          "github": {
            "description": "Project's GitHub source",
            "properties": {
              "checkoutFrom": {
                "description": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                "properties": {
                  "remote": {
                    "description": "The remote name should be used as init. Required if there are more than one remote configured",
                    "type": "string",
                    "markdownDescription": "The remote name should be used as init. Required if there are more than one remote configured"
                  },
                  "revision": {
                    "description": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found.",
                    "type": "string",
                    "markdownDescription": "The revision to checkout from. Should be branch name, tag or commit id. Default branch is used if missing or specified revision is not found."
                  }
                },
                "type": "object",
                "markdownDescription": "Defines from what the project should be checked out. Required if there are more than one remote configured",
                "additionalProperties": false
              },
              "remotes": {
                "additionalProperties": {
                  "type": "string"
                },
                "description": "The remotes map which should be initialized in the git project. Must have at least one remote configured",
                "type": "object",
                "markdownDescription": "The remotes map which should be initialized in the git project. Must have at least one remote configured"
              },
              "sparseCheckoutDir": {
                "description": "Part of project to populate in the working directory.",
                "type": "string",
                "markdownDescription": "Part of project to populate in the working directory."
              }
            },
            "type": "object",
            "markdownDescription": "Project's GitHub source",
            "additionalProperties": false
          },
          "name": {
            "description": "Project name",
            "type": "string",
            "markdownDescription": "Project name"
          },
          "zip": {
            "description": "Project's Zip source",
            "properties": {
              "location": {
                "description": "Zip project's source location address. Should be file path of the archive, e.g. file://$FILE_PATH",
                "type": "string",
                "markdownDescription": "Zip project's source location address. Should be file path of the archive, e.g. file://$FILE_PATH"
              },
              "sparseCheckoutDir": {
                "description": "Part of project to populate in the working directory.",
                "type": "string",
                "markdownDescription": "Part of project to populate in the working directory."
              }
            },
            "type": "object",
            "markdownDescription": "Project's Zip source",
            "additionalProperties": false
          },
          "markdownDescription": {
            "description": "Description of a starter project",
            "type": "string",
            "markdownDescription": "Description of a starter project"
          }
        },
        "required": [
          "name"
        ],
        "type": "object",
        "additionalProperties": false,
        "oneOf": [
          {
            "required": [
              "git"
            ]
          },
          {
            "required": [
              "github"
            ]
          },
          {
            "required": [
              "zip"
            ]
          }
        ]
      },
      "type": "array",
      "markdownDescription": "StarterProjects is a project that can be used as a starting point when bootstrapping new projects"
    },
    "metadata": {
      "type": "object",
      "description": "Optional metadata",
      "properties": {
        "version": {
          "type": "string",
          "description": "Optional semver-compatible version",
          "pattern": "^([0-9]+)\\.([0-9]+)\\.([0-9]+)(\\-[0-9a-z-]+(\\.[0-9a-z-]+)*)?(\\+[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?$"
        },
        "name": {
          "type": "string",
          "description": "Optional devfile name"
        },
        "alpha.build-dockerfile": {
          "type": "string",
          "description": "Optional build dockerfile link"
        },
        "alpha.deployment-manifest": {
          "type": "string",
          "description": "Optional deployment manifest link"
        }
      }
    },
    "schemaVersion": {
      "type": "string",
      "description": "Devfile schema version",
      "pattern": "^([2-9]+)\\.([0-9]+)\\.([0-9]+)(\\-[0-9a-z-]+(\\.[0-9a-z-]+)*)?(\\+[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?$"
    }
  },
  "type": "object",
  "markdownDescription": "Structure of the workspace. This is also the specification of a workspace template.",
  "additionalProperties": false,
  "required": [
    "schemaVersion"
  ]
}
`
