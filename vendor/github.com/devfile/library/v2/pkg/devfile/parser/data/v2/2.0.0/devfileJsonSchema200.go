//
// Copyright 2022 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package version200

// https://raw.githubusercontent.com/devfile/api/2.0.x/schemas/latest/devfile.json
const JsonSchema200 = `{
  "description": "Devfile describes the structure of a cloud-native workspace and development environment.",
  "type": "object",
  "title": "Devfile schema - Version 2.0.0",
  "required": [
    "schemaVersion"
  ],
  "properties": {
    "commands": {
      "description": "Predefined, ready-to-use, workspace-related commands",
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
        ],
        "properties": {
          "apply": {
            "description": "Command that consists in applying a given component definition, typically bound to a workspace event.\n\nFor example, when an 'apply' command is bound to a 'preStart' event, and references a 'container' component, it will start the container as a K8S initContainer in the workspace POD, unless the component has its 'dedicatedPod' field set to 'true'.\n\nWhen no 'apply' command exist for a given component, it is assumed the component will be applied at workspace start by default.",
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
                      "debug"
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
                      "debug"
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
                      "debug"
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
          },
          "vscodeLaunch": {
            "description": "Command providing the definition of a VsCode launch action",
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
                      "debug"
                    ]
                  }
                },
                "additionalProperties": false
              },
              "inlined": {
                "description": "Inlined content of the VsCode configuration",
                "type": "string"
              },
              "uri": {
                "description": "Location as an absolute of relative URI the VsCode configuration will be fetched from",
                "type": "string"
              }
            },
            "additionalProperties": false
          },
          "vscodeTask": {
            "description": "Command providing the definition of a VsCode Task",
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
                      "debug"
                    ]
                  }
                },
                "additionalProperties": false
              },
              "inlined": {
                "description": "Inlined content of the VsCode configuration",
                "type": "string"
              },
              "uri": {
                "description": "Location as an absolute of relative URI the VsCode configuration will be fetched from",
                "type": "string"
              }
            },
            "additionalProperties": false
          }
        },
        "additionalProperties": false
      }
    },
    "components": {
      "description": "List of the workspace components, such as editor and plugins, user-provided containers, or other types of components",
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
              "plugin"
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
            "description": "Allows adding and configuring workspace-related containers",
            "type": "object",
            "required": [
              "image"
            ],
            "properties": {
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
                    "attributes": {
                      "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                      "type": "object",
                      "additionalProperties": true
                    },
                    "exposure": {
                      "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
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
                      "maxLength": 63,
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
          "kubernetes": {
            "description": "Allows importing into the workspace the Kubernetes resources defined in a given manifest. For example this allows reusing the Kubernetes definitions used to deploy some runtime components in production.",
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
              "endpoints": {
                "type": "array",
                "items": {
                  "type": "object",
                  "required": [
                    "name",
                    "targetPort"
                  ],
                  "properties": {
                    "attributes": {
                      "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                      "type": "object",
                      "additionalProperties": true
                    },
                    "exposure": {
                      "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
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
                      "maxLength": 63,
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
            "description": "Allows importing into the workspace the OpenShift resources defined in a given manifest. For example this allows reusing the OpenShift definitions used to deploy some runtime components in production.",
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
              "endpoints": {
                "type": "array",
                "items": {
                  "type": "object",
                  "required": [
                    "name",
                    "targetPort"
                  ],
                  "properties": {
                    "attributes": {
                      "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                      "type": "object",
                      "additionalProperties": true
                    },
                    "exposure": {
                      "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
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
                      "maxLength": 63,
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
          "plugin": {
            "description": "Allows importing a plugin.\n\nPlugins are mainly imported devfiles that contribute components, commands and events as a consistent single unit. They are defined in either YAML files following the devfile syntax, or as 'DevWorkspaceTemplate' Kubernetes Custom Resources",
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
                  ],
                  "properties": {
                    "apply": {
                      "description": "Command that consists in applying a given component definition, typically bound to a workspace event.\n\nFor example, when an 'apply' command is bound to a 'preStart' event, and references a 'container' component, it will start the container as a K8S initContainer in the workspace POD, unless the component has its 'dedicatedPod' field set to 'true'.\n\nWhen no 'apply' command exist for a given component, it is assumed the component will be applied at workspace start by default.",
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
                                "debug"
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
                                "debug"
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
                                "debug"
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
                    },
                    "vscodeLaunch": {
                      "description": "Command providing the definition of a VsCode launch action",
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
                                "debug"
                              ]
                            }
                          },
                          "additionalProperties": false
                        },
                        "inlined": {
                          "description": "Inlined content of the VsCode configuration",
                          "type": "string"
                        },
                        "uri": {
                          "description": "Location as an absolute of relative URI the VsCode configuration will be fetched from",
                          "type": "string"
                        }
                      },
                      "additionalProperties": false
                    },
                    "vscodeTask": {
                      "description": "Command providing the definition of a VsCode Task",
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
                                "debug"
                              ]
                            }
                          },
                          "additionalProperties": false
                        },
                        "inlined": {
                          "description": "Inlined content of the VsCode configuration",
                          "type": "string"
                        },
                        "uri": {
                          "description": "Location as an absolute of relative URI the VsCode configuration will be fetched from",
                          "type": "string"
                        }
                      },
                      "additionalProperties": false
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
                    }
                  ],
                  "properties": {
                    "attributes": {
                      "description": "Map of implementation-dependant free-form YAML attributes.",
                      "type": "object",
                      "additionalProperties": true
                    },
                    "container": {
                      "description": "Allows adding and configuring workspace-related containers",
                      "type": "object",
                      "properties": {
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
                              "attributes": {
                                "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                                "type": "object",
                                "additionalProperties": true
                              },
                              "exposure": {
                                "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                                "type": "string",
                                "enum": [
                                  "public",
                                  "internal",
                                  "none"
                                ]
                              },
                              "name": {
                                "type": "string",
                                "maxLength": 63,
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
                    "kubernetes": {
                      "description": "Allows importing into the workspace the Kubernetes resources defined in a given manifest. For example this allows reusing the Kubernetes definitions used to deploy some runtime components in production.",
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
                        "endpoints": {
                          "type": "array",
                          "items": {
                            "type": "object",
                            "required": [
                              "name"
                            ],
                            "properties": {
                              "attributes": {
                                "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                                "type": "object",
                                "additionalProperties": true
                              },
                              "exposure": {
                                "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                                "type": "string",
                                "enum": [
                                  "public",
                                  "internal",
                                  "none"
                                ]
                              },
                              "name": {
                                "type": "string",
                                "maxLength": 63,
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
                      "description": "Allows importing into the workspace the OpenShift resources defined in a given manifest. For example this allows reusing the OpenShift definitions used to deploy some runtime components in production.",
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
                        "endpoints": {
                          "type": "array",
                          "items": {
                            "type": "object",
                            "required": [
                              "name"
                            ],
                            "properties": {
                              "attributes": {
                                "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                                "type": "object",
                                "additionalProperties": true
                              },
                              "exposure": {
                                "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                                "type": "string",
                                "enum": [
                                  "public",
                                  "internal",
                                  "none"
                                ]
                              },
                              "name": {
                                "type": "string",
                                "maxLength": 63,
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
              "registryUrl": {
                "type": "string"
              },
              "uri": {
                "description": "Uri of a Devfile yaml file",
                "type": "string"
              }
            },
            "additionalProperties": false
          },
          "volume": {
            "description": "Allows specifying the definition of a volume shared by several other components",
            "type": "object",
            "properties": {
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
          "description": "IDs of commands that should be executed after the workspace is completely started. In the case of Che-Theia, these commands should be executed after all plugins and extensions have started, including project cloning. This means that those commands are not triggered until the user opens the IDE in his browser.",
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "postStop": {
          "description": "IDs of commands that should be executed after stopping the workspace.",
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "preStart": {
          "description": "IDs of commands that should be executed before the workspace start. Kubernetes-wise, these commands would typically be executed in init containers of the workspace POD.",
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "preStop": {
          "description": "IDs of commands that should be executed before stopping the workspace.",
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
        "attributes": {
          "description": "Map of implementation-dependant free-form YAML attributes.",
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
          "description": "Optional devfile icon",
          "type": "string"
        },
        "name": {
          "description": "Optional devfile name",
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
        }
      },
      "additionalProperties": true
    },
    "parent": {
      "description": "Parent workspace template",
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
            ],
            "properties": {
              "apply": {
                "description": "Command that consists in applying a given component definition, typically bound to a workspace event.\n\nFor example, when an 'apply' command is bound to a 'preStart' event, and references a 'container' component, it will start the container as a K8S initContainer in the workspace POD, unless the component has its 'dedicatedPod' field set to 'true'.\n\nWhen no 'apply' command exist for a given component, it is assumed the component will be applied at workspace start by default.",
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
                          "debug"
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
                          "debug"
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
                          "debug"
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
              },
              "vscodeLaunch": {
                "description": "Command providing the definition of a VsCode launch action",
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
                          "debug"
                        ]
                      }
                    },
                    "additionalProperties": false
                  },
                  "inlined": {
                    "description": "Inlined content of the VsCode configuration",
                    "type": "string"
                  },
                  "uri": {
                    "description": "Location as an absolute of relative URI the VsCode configuration will be fetched from",
                    "type": "string"
                  }
                },
                "additionalProperties": false
              },
              "vscodeTask": {
                "description": "Command providing the definition of a VsCode Task",
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
                          "debug"
                        ]
                      }
                    },
                    "additionalProperties": false
                  },
                  "inlined": {
                    "description": "Inlined content of the VsCode configuration",
                    "type": "string"
                  },
                  "uri": {
                    "description": "Location as an absolute of relative URI the VsCode configuration will be fetched from",
                    "type": "string"
                  }
                },
                "additionalProperties": false
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
                  "plugin"
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
                "description": "Allows adding and configuring workspace-related containers",
                "type": "object",
                "properties": {
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
                        "attributes": {
                          "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                          "type": "object",
                          "additionalProperties": true
                        },
                        "exposure": {
                          "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                          "type": "string",
                          "enum": [
                            "public",
                            "internal",
                            "none"
                          ]
                        },
                        "name": {
                          "type": "string",
                          "maxLength": 63,
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
              "kubernetes": {
                "description": "Allows importing into the workspace the Kubernetes resources defined in a given manifest. For example this allows reusing the Kubernetes definitions used to deploy some runtime components in production.",
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
                  "endpoints": {
                    "type": "array",
                    "items": {
                      "type": "object",
                      "required": [
                        "name"
                      ],
                      "properties": {
                        "attributes": {
                          "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                          "type": "object",
                          "additionalProperties": true
                        },
                        "exposure": {
                          "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                          "type": "string",
                          "enum": [
                            "public",
                            "internal",
                            "none"
                          ]
                        },
                        "name": {
                          "type": "string",
                          "maxLength": 63,
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
                "description": "Allows importing into the workspace the OpenShift resources defined in a given manifest. For example this allows reusing the OpenShift definitions used to deploy some runtime components in production.",
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
                  "endpoints": {
                    "type": "array",
                    "items": {
                      "type": "object",
                      "required": [
                        "name"
                      ],
                      "properties": {
                        "attributes": {
                          "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                          "type": "object",
                          "additionalProperties": true
                        },
                        "exposure": {
                          "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                          "type": "string",
                          "enum": [
                            "public",
                            "internal",
                            "none"
                          ]
                        },
                        "name": {
                          "type": "string",
                          "maxLength": 63,
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
              "plugin": {
                "description": "Allows importing a plugin.\n\nPlugins are mainly imported devfiles that contribute components, commands and events as a consistent single unit. They are defined in either YAML files following the devfile syntax, or as 'DevWorkspaceTemplate' Kubernetes Custom Resources",
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
                      ],
                      "properties": {
                        "apply": {
                          "description": "Command that consists in applying a given component definition, typically bound to a workspace event.\n\nFor example, when an 'apply' command is bound to a 'preStart' event, and references a 'container' component, it will start the container as a K8S initContainer in the workspace POD, unless the component has its 'dedicatedPod' field set to 'true'.\n\nWhen no 'apply' command exist for a given component, it is assumed the component will be applied at workspace start by default.",
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
                                    "debug"
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
                                    "debug"
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
                                    "debug"
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
                        },
                        "vscodeLaunch": {
                          "description": "Command providing the definition of a VsCode launch action",
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
                                    "debug"
                                  ]
                                }
                              },
                              "additionalProperties": false
                            },
                            "inlined": {
                              "description": "Inlined content of the VsCode configuration",
                              "type": "string"
                            },
                            "uri": {
                              "description": "Location as an absolute of relative URI the VsCode configuration will be fetched from",
                              "type": "string"
                            }
                          },
                          "additionalProperties": false
                        },
                        "vscodeTask": {
                          "description": "Command providing the definition of a VsCode Task",
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
                                    "debug"
                                  ]
                                }
                              },
                              "additionalProperties": false
                            },
                            "inlined": {
                              "description": "Inlined content of the VsCode configuration",
                              "type": "string"
                            },
                            "uri": {
                              "description": "Location as an absolute of relative URI the VsCode configuration will be fetched from",
                              "type": "string"
                            }
                          },
                          "additionalProperties": false
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
                        }
                      ],
                      "properties": {
                        "attributes": {
                          "description": "Map of implementation-dependant free-form YAML attributes.",
                          "type": "object",
                          "additionalProperties": true
                        },
                        "container": {
                          "description": "Allows adding and configuring workspace-related containers",
                          "type": "object",
                          "properties": {
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
                                  "attributes": {
                                    "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                                    "type": "object",
                                    "additionalProperties": true
                                  },
                                  "exposure": {
                                    "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                                    "type": "string",
                                    "enum": [
                                      "public",
                                      "internal",
                                      "none"
                                    ]
                                  },
                                  "name": {
                                    "type": "string",
                                    "maxLength": 63,
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
                        "kubernetes": {
                          "description": "Allows importing into the workspace the Kubernetes resources defined in a given manifest. For example this allows reusing the Kubernetes definitions used to deploy some runtime components in production.",
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
                            "endpoints": {
                              "type": "array",
                              "items": {
                                "type": "object",
                                "required": [
                                  "name"
                                ],
                                "properties": {
                                  "attributes": {
                                    "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                                    "type": "object",
                                    "additionalProperties": true
                                  },
                                  "exposure": {
                                    "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                                    "type": "string",
                                    "enum": [
                                      "public",
                                      "internal",
                                      "none"
                                    ]
                                  },
                                  "name": {
                                    "type": "string",
                                    "maxLength": 63,
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
                          "description": "Allows importing into the workspace the OpenShift resources defined in a given manifest. For example this allows reusing the OpenShift definitions used to deploy some runtime components in production.",
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
                            "endpoints": {
                              "type": "array",
                              "items": {
                                "type": "object",
                                "required": [
                                  "name"
                                ],
                                "properties": {
                                  "attributes": {
                                    "description": "Map of implementation-dependant string-based free-form attributes.\n\nExamples of Che-specific attributes:\n- cookiesAuthEnabled: \"true\" / \"false\",\n- type: \"terminal\" / \"ide\" / \"ide-dev\",",
                                    "type": "object",
                                    "additionalProperties": true
                                  },
                                  "exposure": {
                                    "description": "Describes how the endpoint should be exposed on the network.\n- 'public' means that the endpoint will be exposed on the public network, typically through a K8S ingress or an OpenShift route.\n- 'internal' means that the endpoint will be exposed internally outside of the main workspace POD, typically by K8S services, to be consumed by other elements running on the same cloud internal network.\n- 'none' means that the endpoint will not be exposed and will only be accessible inside the main workspace POD, on a local address.\n\nDefault value is 'public'",
                                    "type": "string",
                                    "enum": [
                                      "public",
                                      "internal",
                                      "none"
                                    ]
                                  },
                                  "name": {
                                    "type": "string",
                                    "maxLength": 63,
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
                  "registryUrl": {
                    "type": "string"
                  },
                  "uri": {
                    "description": "Uri of a Devfile yaml file",
                    "type": "string"
                  }
                },
                "additionalProperties": false
              },
              "volume": {
                "description": "Allows specifying the definition of a volume shared by several other components",
                "type": "object",
                "properties": {
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
                  "github"
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
                    "description": "The remotes map which should be initialized in the git project. Must have at least one remote configured",
                    "type": "object",
                    "additionalProperties": {
                      "type": "string"
                    }
                  }
                },
                "additionalProperties": false
              },
              "github": {
                "description": "Project's GitHub source",
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
                    "description": "The remotes map which should be initialized in the git project. Must have at least one remote configured",
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
              "sparseCheckoutDirs": {
                "description": "Populate the project sparsely with selected directories.",
                "type": "array",
                "items": {
                  "type": "string"
                }
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
                  "github"
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
                    "description": "The remotes map which should be initialized in the git project. Must have at least one remote configured",
                    "type": "object",
                    "additionalProperties": {
                      "type": "string"
                    }
                  }
                },
                "additionalProperties": false
              },
              "github": {
                "description": "Project's GitHub source",
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
                    "description": "The remotes map which should be initialized in the git project. Must have at least one remote configured",
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
          "description": "Uri of a Devfile yaml file",
          "type": "string"
        }
      },
      "additionalProperties": false
    },
    "projects": {
      "description": "Projects worked on in the workspace, containing names and sources locations",
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
              "github"
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
                "description": "The remotes map which should be initialized in the git project. Must have at least one remote configured",
                "type": "object",
                "additionalProperties": {
                  "type": "string"
                }
              }
            },
            "additionalProperties": false
          },
          "github": {
            "description": "Project's GitHub source",
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
                "description": "The remotes map which should be initialized in the git project. Must have at least one remote configured",
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
          "sparseCheckoutDirs": {
            "description": "Populate the project sparsely with selected directories.",
            "type": "array",
            "items": {
              "type": "string"
            }
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
              "github"
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
                "description": "The remotes map which should be initialized in the git project. Must have at least one remote configured",
                "type": "object",
                "additionalProperties": {
                  "type": "string"
                }
              }
            },
            "additionalProperties": false
          },
          "github": {
            "description": "Project's GitHub source",
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
                "description": "The remotes map which should be initialized in the git project. Must have at least one remote configured",
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
    }
  },
  "additionalProperties": false
}
`
