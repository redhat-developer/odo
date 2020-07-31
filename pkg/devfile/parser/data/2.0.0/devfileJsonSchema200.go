package version200

const JsonSchema200 = `{
  "description": "Devfile schema.",
  "properties": {
    "commands": {
      "description": "Predefined, ready-to-use, workspace-related commands",
      "items": {
        "properties": {
          "composite": {
            "description": "Composite command that allows executing several sub-commands either sequentially or concurrently",
            "properties": {
              "attributes": {
                "additionalProperties": {
                  "type": "string"
                },
                "description": "Optional map of free-form additional command attributes",
                "type": "object"
              },
              "commands": {
                "description": "The commands that comprise this composite command",
                "items": {
                  "type": "string"
                },
                "type": "array"
              },
              "group": {
                "description": "Defines the group this command is part of",
                "properties": {
                  "isDefault": {
                    "description": "Identifies the default command for a given group kind",
                    "type": "boolean"
                  },
                  "kind": {
                    "description": "Kind of group the command is part of",
                    "enum": [
                      "build",
                      "run",
                      "test",
                      "debug"
                    ],
                    "type": "string"
                  }
                },
                "required": [
                  "kind"
                ],
                "type": "object",
                "additionalProperties": false
              },
              "id": {
                "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
                "type": "string"
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
            "required": [
              "id"
            ],
            "type": "object",
            "additionalProperties": false
          },
          "exec": {
            "description": "CLI Command executed in a component container",
            "properties": {
              "attributes": {
                "additionalProperties": {
                  "type": "string"
                },
                "description": "Optional map of free-form additional command attributes",
                "type": "object"
              },
              "commandLine": {
                "description": "The actual command-line string",
                "type": "string"
              },
              "component": {
                "description": "Describes component to which given action relates",
                "type": "string"
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
                "type": "array"
              },
              "group": {
                "description": "Defines the group this command is part of",
                "properties": {
                  "isDefault": {
                    "description": "Identifies the default command for a given group kind",
                    "type": "boolean"
                  },
                  "kind": {
                    "description": "Kind of group the command is part of",
                    "enum": [
                      "build",
                      "run",
                      "test",
                      "debug"
                    ],
                    "type": "string"
                  }
                },
                "required": [
                  "kind"
                ],
                "type": "object",
                "additionalProperties": false
              },
              "id": {
                "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
                "type": "string"
              },
              "label": {
                "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                "type": "string"
              },
              "workingDir": {
                "description": "Working directory where the command should be executed",
                "type": "string"
              }
            },
            "required": [
              "id",
              "commandLine"
            ],
            "type": "object",
            "additionalProperties": false
          },
          "vscodeLaunch": {
            "description": "Command providing the definition of a VsCode launch action",
            "properties": {
              "attributes": {
                "additionalProperties": {
                  "type": "string"
                },
                "description": "Optional map of free-form additional command attributes",
                "type": "object"
              },
              "group": {
                "description": "Defines the group this command is part of",
                "properties": {
                  "isDefault": {
                    "description": "Identifies the default command for a given group kind",
                    "type": "boolean"
                  },
                  "kind": {
                    "description": "Kind of group the command is part of",
                    "enum": [
                      "build",
                      "run",
                      "test",
                      "debug"
                    ],
                    "type": "string"
                  }
                },
                "required": [
                  "kind"
                ],
                "type": "object",
                "additionalProperties": false
              },
              "id": {
                "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
                "type": "string"
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
            "required": [
              "id"
            ],
            "type": "object",
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
                "type": "object"
              },
              "group": {
                "description": "Defines the group this command is part of",
                "properties": {
                  "isDefault": {
                    "description": "Identifies the default command for a given group kind",
                    "type": "boolean"
                  },
                  "kind": {
                    "description": "Kind of group the command is part of",
                    "enum": [
                      "build",
                      "run",
                      "test",
                      "debug"
                    ],
                    "type": "string"
                  }
                },
                "required": [
                  "kind"
                ],
                "type": "object",
                "additionalProperties": false
              },
              "id": {
                "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
                "type": "string"
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
            "required": [
              "id"
            ],
            "type": "object",
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
      "type": "array"
    },
    "components": {
      "description": "List of the workspace components, such as editor and plugins, user-provided containers, or other types of components",
      "items": {
        "properties": {
          "container": {
            "description": "Allows adding and configuring workspace-related containers",
            "properties": {
              "args": {
                "description": "The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command. Defaults to an empty array, meaning use whatever is defined in the image.",
                "items": {
                  "type": "string"
                },
                "type": "array"
              },
              "command": {
                "description": "The command to run in the dockerimage component instead of the default one provided in the image. Defaults to an empty array, meaning use whatever is defined in the image.",
                "items": {
                  "type": "string"
                },
                "type": "array"
              },
              "endpoints": {
                "items": {
                  "properties": {
                    "attributes": {
                      "additionalProperties": {
                        "type": "string"
                      },
                      "type": "object"
                    },
                    "configuration": {
                      "properties": {
                        "cookiesAuthEnabled": {
                          "type": "boolean"
                        },
                        "discoverable": {
                          "type": "boolean"
                        },
                        "path": {
                          "type": "string"
                        },
                        "protocol": {
                          "description": "The is the low-level protocol of traffic coming through this endpoint. Default value is \"tcp\"",
                          "type": "string"
                        },
                        "public": {
                          "type": "boolean"
                        },
                        "scheme": {
                          "description": "The is the URL scheme to use when accessing the endpoint. Default value is \"http\"",
                          "type": "string"
                        },
                        "secure": {
                          "type": "boolean"
                        },
                        "type": {
                          "enum": [
                            "ide",
                            "terminal",
                            "ide-dev"
                          ],
                          "type": "string"
                        }
                      },
                      "type": "object",
                      "additionalProperties": false
                    },
                    "name": {
                      "type": "string"
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
                "type": "array"
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
              "name": {
                "type": "string"
              },
              "sourceMapping": {
                "description": "Optional specification of the path in the container where project sources should be transferred/mounted when ` + `mountSources` + ` is ` + `true` + `. When omitted, the value of the ` + `PROJECTS_ROOT` + ` environment variable is used.",
                "type": "string"
              },
              "volumeMounts": {
                "description": "List of volumes mounts that should be mounted is this container.",
                "items": {
                  "description": "Volume that should be mounted to a component container",
                  "properties": {
                    "name": {
                      "description": "The volume mount name is the name of an existing ` + `Volume` + ` component. If no corresponding ` + `Volume` + ` component exist it is implicitly added. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files.",
                      "type": "string"
                    },
                    "path": {
                      "description": "The path in the component container where the volume should be mounted. If not path is mentioned, default path is the is ` + `/<name>` + `.",
                      "type": "string"
                    }
                  },
                  "required": [
                    "name"
                  ],
                  "type": "object",
                  "additionalProperties": false
                },
                "type": "array"
              }
            },
            "required": [
              "name",
              "image"
            ],
            "type": "object",
            "additionalProperties": false
          },
          "kubernetes": {
            "description": "Allows importing into the workspace the Kubernetes resources defined in a given manifest. For example this allows reusing the Kubernetes definitions used to deploy some runtime components in production.",
            "properties": {
              "inlined": {
                "description": "Inlined manifest",
                "type": "string"
              },
              "name": {
                "description": "Mandatory name that allows referencing the component in commands, or inside a parent",
                "type": "string"
              },
              "uri": {
                "description": "Location in a file fetched from a uri.",
                "type": "string"
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
          "openshift": {
            "description": "Allows importing into the workspace the OpenShift resources defined in a given manifest. For example this allows reusing the OpenShift definitions used to deploy some runtime components in production.",
            "properties": {
              "inlined": {
                "description": "Inlined manifest",
                "type": "string"
              },
              "name": {
                "description": "Mandatory name that allows referencing the component in commands, or inside a parent",
                "type": "string"
              },
              "uri": {
                "description": "Location in a file fetched from a uri.",
                "type": "string"
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
            "description": "Allows importing a plugin. Plugins are mainly imported devfiles that contribute components, commands and events as a consistent single unit. They are defined in either YAML files following the devfile syntax, or as ` + `DevWorkspaceTemplate` + ` Kubernetes Custom Resources",
            "properties": {
              "commands": {
                "description": "Overrides of commands encapsulated in a plugin. Overriding is done using a strategic merge",
                "items": {
                  "properties": {
                    "composite": {
                      "description": "Composite command that allows executing several sub-commands either sequentially or concurrently",
                      "properties": {
                        "attributes": {
                          "additionalProperties": {
                            "type": "string"
                          },
                          "description": "Optional map of free-form additional command attributes",
                          "type": "object"
                        },
                        "commands": {
                          "description": "The commands that comprise this composite command",
                          "items": {
                            "type": "string"
                          },
                          "type": "array"
                        },
                        "group": {
                          "description": "Defines the group this command is part of",
                          "properties": {
                            "isDefault": {
                              "description": "Identifies the default command for a given group kind",
                              "type": "boolean"
                            },
                            "kind": {
                              "description": "Kind of group the command is part of",
                              "enum": [
                                "build",
                                "run",
                                "test",
                                "debug"
                              ],
                              "type": "string"
                            }
                          },
                          "required": [
                            "kind"
                          ],
                          "type": "object",
                          "additionalProperties": false
                        },
                        "id": {
                          "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
                          "type": "string"
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
                      "required": [
                        "id"
                      ],
                      "type": "object",
                      "additionalProperties": false
                    },
                    "exec": {
                      "description": "CLI Command executed in a component container",
                      "properties": {
                        "attributes": {
                          "additionalProperties": {
                            "type": "string"
                          },
                          "description": "Optional map of free-form additional command attributes",
                          "type": "object"
                        },
                        "commandLine": {
                          "description": "The actual command-line string",
                          "type": "string"
                        },
                        "component": {
                          "description": "Describes component to which given action relates",
                          "type": "string"
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
                          "type": "array"
                        },
                        "group": {
                          "description": "Defines the group this command is part of",
                          "properties": {
                            "isDefault": {
                              "description": "Identifies the default command for a given group kind",
                              "type": "boolean"
                            },
                            "kind": {
                              "description": "Kind of group the command is part of",
                              "enum": [
                                "build",
                                "run",
                                "test",
                                "debug"
                              ],
                              "type": "string"
                            }
                          },
                          "required": [
                            "kind"
                          ],
                          "type": "object",
                          "additionalProperties": false
                        },
                        "id": {
                          "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
                          "type": "string"
                        },
                        "label": {
                          "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                          "type": "string"
                        },
                        "workingDir": {
                          "description": "Working directory where the command should be executed",
                          "type": "string"
                        }
                      },
                      "required": [
                        "id"
                      ],
                      "type": "object",
                      "additionalProperties": false
                    },
                    "vscodeLaunch": {
                      "description": "Command providing the definition of a VsCode launch action",
                      "properties": {
                        "attributes": {
                          "additionalProperties": {
                            "type": "string"
                          },
                          "description": "Optional map of free-form additional command attributes",
                          "type": "object"
                        },
                        "group": {
                          "description": "Defines the group this command is part of",
                          "properties": {
                            "isDefault": {
                              "description": "Identifies the default command for a given group kind",
                              "type": "boolean"
                            },
                            "kind": {
                              "description": "Kind of group the command is part of",
                              "enum": [
                                "build",
                                "run",
                                "test",
                                "debug"
                              ],
                              "type": "string"
                            }
                          },
                          "required": [
                            "kind"
                          ],
                          "type": "object",
                          "additionalProperties": false
                        },
                        "id": {
                          "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
                          "type": "string"
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
                      "required": [
                        "id"
                      ],
                      "type": "object",
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
                          "type": "object"
                        },
                        "group": {
                          "description": "Defines the group this command is part of",
                          "properties": {
                            "isDefault": {
                              "description": "Identifies the default command for a given group kind",
                              "type": "boolean"
                            },
                            "kind": {
                              "description": "Kind of group the command is part of",
                              "enum": [
                                "build",
                                "run",
                                "test",
                                "debug"
                              ],
                              "type": "string"
                            }
                          },
                          "required": [
                            "kind"
                          ],
                          "type": "object",
                          "additionalProperties": false
                        },
                        "id": {
                          "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
                          "type": "string"
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
                      "required": [
                        "id"
                      ],
                      "type": "object",
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
                "type": "array"
              },
              "components": {
                "description": "Overrides of components encapsulated in a plugin. Overriding is done using a strategic merge",
                "items": {
                  "properties": {
                    "container": {
                      "description": "Configuration overriding for a Container component",
                      "properties": {
                        "args": {
                          "description": "The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command. Defaults to an empty array, meaning use whatever is defined in the image.",
                          "items": {
                            "type": "string"
                          },
                          "type": "array"
                        },
                        "command": {
                          "description": "The command to run in the dockerimage component instead of the default one provided in the image. Defaults to an empty array, meaning use whatever is defined in the image.",
                          "items": {
                            "type": "string"
                          },
                          "type": "array"
                        },
                        "endpoints": {
                          "items": {
                            "properties": {
                              "attributes": {
                                "additionalProperties": {
                                  "type": "string"
                                },
                                "type": "object"
                              },
                              "configuration": {
                                "properties": {
                                  "cookiesAuthEnabled": {
                                    "type": "boolean"
                                  },
                                  "discoverable": {
                                    "type": "boolean"
                                  },
                                  "path": {
                                    "type": "string"
                                  },
                                  "protocol": {
                                    "description": "The is the low-level protocol of traffic coming through this endpoint. Default value is \"tcp\"",
                                    "type": "string"
                                  },
                                  "public": {
                                    "type": "boolean"
                                  },
                                  "scheme": {
                                    "description": "The is the URL scheme to use when accessing the endpoint. Default value is \"http\"",
                                    "type": "string"
                                  },
                                  "secure": {
                                    "type": "boolean"
                                  },
                                  "type": {
                                    "enum": [
                                      "ide",
                                      "terminal",
                                      "ide-dev"
                                    ],
                                    "type": "string"
                                  }
                                },
                                "type": "object",
                                "additionalProperties": false
                              },
                              "name": {
                                "type": "string"
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
                          "type": "array"
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
                        "name": {
                          "type": "string"
                        },
                        "sourceMapping": {
                          "description": "Optional specification of the path in the container where project sources should be transferred/mounted when ` + `mountSources` + ` is ` + `true` + `. When omitted, the value of the ` + `PROJECTS_ROOT` + ` environment variable is used.",
                          "type": "string"
                        },
                        "volumeMounts": {
                          "description": "List of volumes mounts that should be mounted is this container.",
                          "items": {
                            "description": "Volume that should be mounted to a component container",
                            "properties": {
                              "name": {
                                "description": "The volume mount name is the name of an existing ` + `Volume` + ` component. If no corresponding ` + `Volume` + ` component exist it is implicitly added. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files.",
                                "type": "string"
                              },
                              "path": {
                                "description": "The path in the component container where the volume should be mounted. If not path is mentioned, default path is the is ` + `/<name>` + `.",
                                "type": "string"
                              }
                            },
                            "required": [
                              "name"
                            ],
                            "type": "object",
                            "additionalProperties": false
                          },
                          "type": "array"
                        }
                      },
                      "required": [
                        "name"
                      ],
                      "type": "object",
                      "additionalProperties": false
                    },
                    "kubernetes": {
                      "description": "Configuration overriding for a Kubernetes component",
                      "properties": {
                        "inlined": {
                          "description": "Inlined manifest",
                          "type": "string"
                        },
                        "name": {
                          "description": "Mandatory name that allows referencing the component in commands, or inside a parent",
                          "type": "string"
                        },
                        "uri": {
                          "description": "Location in a file fetched from a uri.",
                          "type": "string"
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
                    "openshift": {
                      "description": "Configuration overriding for an OpenShift component",
                      "properties": {
                        "inlined": {
                          "description": "Inlined manifest",
                          "type": "string"
                        },
                        "name": {
                          "description": "Mandatory name that allows referencing the component in commands, or inside a parent",
                          "type": "string"
                        },
                        "uri": {
                          "description": "Location in a file fetched from a uri.",
                          "type": "string"
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
                      "description": "Configuration overriding for a Volume component",
                      "properties": {
                        "name": {
                          "description": "Mandatory name that allows referencing the Volume component in Container volume mounts or inside a parent",
                          "type": "string"
                        },
                        "size": {
                          "description": "Size of the volume",
                          "type": "string"
                        }
                      },
                      "required": [
                        "name"
                      ],
                      "type": "object",
                      "additionalProperties": false
                    }
                  },
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
                "type": "array"
              },
              "id": {
                "description": "Id in a registry that contains a Devfile yaml file",
                "type": "string"
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
                "additionalProperties": false
              },
              "name": {
                "description": "Optional name that allows referencing the component in commands, or inside a parent If omitted it will be infered from the location (uri or registryEntry)",
                "type": "string"
              },
              "registryUrl": {
                "type": "string"
              },
              "uri": {
                "description": "Uri of a Devfile yaml file",
                "type": "string"
              }
            },
            "type": "object",
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
              "name": {
                "description": "Mandatory name that allows referencing the Volume component in Container volume mounts or inside a parent",
                "type": "string"
              },
              "size": {
                "description": "Size of the volume",
                "type": "string"
              }
            },
            "required": [
              "name"
            ],
            "type": "object",
            "additionalProperties": false
          }
        },
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
      "type": "array"
    },
    "events": {
      "description": "Bindings of commands to events. Each command is referred-to by its name.",
      "properties": {
        "postStart": {
          "description": "Names of commands that should be executed after the workspace is completely started. In the case of Che-Theia, these commands should be executed after all plugins and extensions have started, including project cloning. This means that those commands are not triggered until the user opens the IDE in his browser.",
          "items": {
            "type": "string"
          },
          "type": "array"
        },
        "postStop": {
          "description": "Names of commands that should be executed after stopping the workspace.",
          "items": {
            "type": "string"
          },
          "type": "array"
        },
        "preStart": {
          "description": "Names of commands that should be executed before the workspace start. Kubernetes-wise, these commands would typically be executed in init containers of the workspace POD.",
          "items": {
            "type": "string"
          },
          "type": "array"
        },
        "preStop": {
          "description": "Names of commands that should be executed before stopping the workspace.",
          "items": {
            "type": "string"
          },
          "type": "array"
        }
      },
      "type": "object",
      "additionalProperties": false
    },
    "parent": {
      "description": "Parent workspace template",
      "properties": {
        "commands": {
          "description": "Predefined, ready-to-use, workspace-related commands",
          "items": {
            "properties": {
              "composite": {
                "description": "Composite command that allows executing several sub-commands either sequentially or concurrently",
                "properties": {
                  "attributes": {
                    "additionalProperties": {
                      "type": "string"
                    },
                    "description": "Optional map of free-form additional command attributes",
                    "type": "object"
                  },
                  "commands": {
                    "description": "The commands that comprise this composite command",
                    "items": {
                      "type": "string"
                    },
                    "type": "array"
                  },
                  "group": {
                    "description": "Defines the group this command is part of",
                    "properties": {
                      "isDefault": {
                        "description": "Identifies the default command for a given group kind",
                        "type": "boolean"
                      },
                      "kind": {
                        "description": "Kind of group the command is part of",
                        "enum": [
                          "build",
                          "run",
                          "test",
                          "debug"
                        ],
                        "type": "string"
                      }
                    },
                    "required": [
                      "kind"
                    ],
                    "type": "object",
                    "additionalProperties": false
                  },
                  "id": {
                    "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
                    "type": "string"
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
                "required": [
                  "id"
                ],
                "type": "object",
                "additionalProperties": false
              },
              "exec": {
                "description": "CLI Command executed in a component container",
                "properties": {
                  "attributes": {
                    "additionalProperties": {
                      "type": "string"
                    },
                    "description": "Optional map of free-form additional command attributes",
                    "type": "object"
                  },
                  "commandLine": {
                    "description": "The actual command-line string",
                    "type": "string"
                  },
                  "component": {
                    "description": "Describes component to which given action relates",
                    "type": "string"
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
                    "type": "array"
                  },
                  "group": {
                    "description": "Defines the group this command is part of",
                    "properties": {
                      "isDefault": {
                        "description": "Identifies the default command for a given group kind",
                        "type": "boolean"
                      },
                      "kind": {
                        "description": "Kind of group the command is part of",
                        "enum": [
                          "build",
                          "run",
                          "test",
                          "debug"
                        ],
                        "type": "string"
                      }
                    },
                    "required": [
                      "kind"
                    ],
                    "type": "object",
                    "additionalProperties": false
                  },
                  "id": {
                    "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
                    "type": "string"
                  },
                  "label": {
                    "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                    "type": "string"
                  },
                  "workingDir": {
                    "description": "Working directory where the command should be executed",
                    "type": "string"
                  }
                },
                "required": [
                  "id"
                ],
                "type": "object",
                "additionalProperties": false
              },
              "vscodeLaunch": {
                "description": "Command providing the definition of a VsCode launch action",
                "properties": {
                  "attributes": {
                    "additionalProperties": {
                      "type": "string"
                    },
                    "description": "Optional map of free-form additional command attributes",
                    "type": "object"
                  },
                  "group": {
                    "description": "Defines the group this command is part of",
                    "properties": {
                      "isDefault": {
                        "description": "Identifies the default command for a given group kind",
                        "type": "boolean"
                      },
                      "kind": {
                        "description": "Kind of group the command is part of",
                        "enum": [
                          "build",
                          "run",
                          "test",
                          "debug"
                        ],
                        "type": "string"
                      }
                    },
                    "required": [
                      "kind"
                    ],
                    "type": "object",
                    "additionalProperties": false
                  },
                  "id": {
                    "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
                    "type": "string"
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
                "required": [
                  "id"
                ],
                "type": "object",
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
                    "type": "object"
                  },
                  "group": {
                    "description": "Defines the group this command is part of",
                    "properties": {
                      "isDefault": {
                        "description": "Identifies the default command for a given group kind",
                        "type": "boolean"
                      },
                      "kind": {
                        "description": "Kind of group the command is part of",
                        "enum": [
                          "build",
                          "run",
                          "test",
                          "debug"
                        ],
                        "type": "string"
                      }
                    },
                    "required": [
                      "kind"
                    ],
                    "type": "object",
                    "additionalProperties": false
                  },
                  "id": {
                    "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
                    "type": "string"
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
                "required": [
                  "id"
                ],
                "type": "object",
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
          "type": "array"
        },
        "components": {
          "description": "List of the workspace components, such as editor and plugins, user-provided containers, or other types of components",
          "items": {
            "properties": {
              "container": {
                "description": "Allows adding and configuring workspace-related containers",
                "properties": {
                  "args": {
                    "description": "The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command. Defaults to an empty array, meaning use whatever is defined in the image.",
                    "items": {
                      "type": "string"
                    },
                    "type": "array"
                  },
                  "command": {
                    "description": "The command to run in the dockerimage component instead of the default one provided in the image. Defaults to an empty array, meaning use whatever is defined in the image.",
                    "items": {
                      "type": "string"
                    },
                    "type": "array"
                  },
                  "endpoints": {
                    "items": {
                      "properties": {
                        "attributes": {
                          "additionalProperties": {
                            "type": "string"
                          },
                          "type": "object"
                        },
                        "configuration": {
                          "properties": {
                            "cookiesAuthEnabled": {
                              "type": "boolean"
                            },
                            "discoverable": {
                              "type": "boolean"
                            },
                            "path": {
                              "type": "string"
                            },
                            "protocol": {
                              "description": "The is the low-level protocol of traffic coming through this endpoint. Default value is \"tcp\"",
                              "type": "string"
                            },
                            "public": {
                              "type": "boolean"
                            },
                            "scheme": {
                              "description": "The is the URL scheme to use when accessing the endpoint. Default value is \"http\"",
                              "type": "string"
                            },
                            "secure": {
                              "type": "boolean"
                            },
                            "type": {
                              "enum": [
                                "ide",
                                "terminal",
                                "ide-dev"
                              ],
                              "type": "string"
                            }
                          },
                          "type": "object",
                          "additionalProperties": false
                        },
                        "name": {
                          "type": "string"
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
                    "type": "array"
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
                  "name": {
                    "type": "string"
                  },
                  "sourceMapping": {
                    "description": "Optional specification of the path in the container where project sources should be transferred/mounted when ` + `mountSources` + ` is ` + `true` + `. When omitted, the value of the ` + `PROJECTS_ROOT` + ` environment variable is used.",
                    "type": "string"
                  },
                  "volumeMounts": {
                    "description": "List of volumes mounts that should be mounted is this container.",
                    "items": {
                      "description": "Volume that should be mounted to a component container",
                      "properties": {
                        "name": {
                          "description": "The volume mount name is the name of an existing ` + `Volume` + ` component. If no corresponding ` + `Volume` + ` component exist it is implicitly added. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files.",
                          "type": "string"
                        },
                        "path": {
                          "description": "The path in the component container where the volume should be mounted. If not path is mentioned, default path is the is ` + `/<name>` + `.",
                          "type": "string"
                        }
                      },
                      "required": [
                        "name"
                      ],
                      "type": "object",
                      "additionalProperties": false
                    },
                    "type": "array"
                  }
                },
                "required": [
                  "name"
                ],
                "type": "object",
                "additionalProperties": false
              },
              "kubernetes": {
                "description": "Allows importing into the workspace the Kubernetes resources defined in a given manifest. For example this allows reusing the Kubernetes definitions used to deploy some runtime components in production.",
                "properties": {
                  "inlined": {
                    "description": "Inlined manifest",
                    "type": "string"
                  },
                  "name": {
                    "description": "Mandatory name that allows referencing the component in commands, or inside a parent",
                    "type": "string"
                  },
                  "uri": {
                    "description": "Location in a file fetched from a uri.",
                    "type": "string"
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
              "openshift": {
                "description": "Allows importing into the workspace the OpenShift resources defined in a given manifest. For example this allows reusing the OpenShift definitions used to deploy some runtime components in production.",
                "properties": {
                  "inlined": {
                    "description": "Inlined manifest",
                    "type": "string"
                  },
                  "name": {
                    "description": "Mandatory name that allows referencing the component in commands, or inside a parent",
                    "type": "string"
                  },
                  "uri": {
                    "description": "Location in a file fetched from a uri.",
                    "type": "string"
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
                "description": "Allows importing a plugin. Plugins are mainly imported devfiles that contribute components, commands and events as a consistent single unit. They are defined in either YAML files following the devfile syntax, or as ` + `DevWorkspaceTemplate` + ` Kubernetes Custom Resources",
                "properties": {
                  "commands": {
                    "description": "Overrides of commands encapsulated in a plugin. Overriding is done using a strategic merge",
                    "items": {
                      "properties": {
                        "composite": {
                          "description": "Composite command that allows executing several sub-commands either sequentially or concurrently",
                          "properties": {
                            "attributes": {
                              "additionalProperties": {
                                "type": "string"
                              },
                              "description": "Optional map of free-form additional command attributes",
                              "type": "object"
                            },
                            "commands": {
                              "description": "The commands that comprise this composite command",
                              "items": {
                                "type": "string"
                              },
                              "type": "array"
                            },
                            "group": {
                              "description": "Defines the group this command is part of",
                              "properties": {
                                "isDefault": {
                                  "description": "Identifies the default command for a given group kind",
                                  "type": "boolean"
                                },
                                "kind": {
                                  "description": "Kind of group the command is part of",
                                  "enum": [
                                    "build",
                                    "run",
                                    "test",
                                    "debug"
                                  ],
                                  "type": "string"
                                }
                              },
                              "required": [
                                "kind"
                              ],
                              "type": "object",
                              "additionalProperties": false
                            },
                            "id": {
                              "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
                              "type": "string"
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
                          "required": [
                            "id"
                          ],
                          "type": "object",
                          "additionalProperties": false
                        },
                        "exec": {
                          "description": "CLI Command executed in a component container",
                          "properties": {
                            "attributes": {
                              "additionalProperties": {
                                "type": "string"
                              },
                              "description": "Optional map of free-form additional command attributes",
                              "type": "object"
                            },
                            "commandLine": {
                              "description": "The actual command-line string",
                              "type": "string"
                            },
                            "component": {
                              "description": "Describes component to which given action relates",
                              "type": "string"
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
                              "type": "array"
                            },
                            "group": {
                              "description": "Defines the group this command is part of",
                              "properties": {
                                "isDefault": {
                                  "description": "Identifies the default command for a given group kind",
                                  "type": "boolean"
                                },
                                "kind": {
                                  "description": "Kind of group the command is part of",
                                  "enum": [
                                    "build",
                                    "run",
                                    "test",
                                    "debug"
                                  ],
                                  "type": "string"
                                }
                              },
                              "required": [
                                "kind"
                              ],
                              "type": "object",
                              "additionalProperties": false
                            },
                            "id": {
                              "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
                              "type": "string"
                            },
                            "label": {
                              "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
                              "type": "string"
                            },
                            "workingDir": {
                              "description": "Working directory where the command should be executed",
                              "type": "string"
                            }
                          },
                          "required": [
                            "id"
                          ],
                          "type": "object",
                          "additionalProperties": false
                        },
                        "vscodeLaunch": {
                          "description": "Command providing the definition of a VsCode launch action",
                          "properties": {
                            "attributes": {
                              "additionalProperties": {
                                "type": "string"
                              },
                              "description": "Optional map of free-form additional command attributes",
                              "type": "object"
                            },
                            "group": {
                              "description": "Defines the group this command is part of",
                              "properties": {
                                "isDefault": {
                                  "description": "Identifies the default command for a given group kind",
                                  "type": "boolean"
                                },
                                "kind": {
                                  "description": "Kind of group the command is part of",
                                  "enum": [
                                    "build",
                                    "run",
                                    "test",
                                    "debug"
                                  ],
                                  "type": "string"
                                }
                              },
                              "required": [
                                "kind"
                              ],
                              "type": "object",
                              "additionalProperties": false
                            },
                            "id": {
                              "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
                              "type": "string"
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
                          "required": [
                            "id"
                          ],
                          "type": "object",
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
                              "type": "object"
                            },
                            "group": {
                              "description": "Defines the group this command is part of",
                              "properties": {
                                "isDefault": {
                                  "description": "Identifies the default command for a given group kind",
                                  "type": "boolean"
                                },
                                "kind": {
                                  "description": "Kind of group the command is part of",
                                  "enum": [
                                    "build",
                                    "run",
                                    "test",
                                    "debug"
                                  ],
                                  "type": "string"
                                }
                              },
                              "required": [
                                "kind"
                              ],
                              "type": "object",
                              "additionalProperties": false
                            },
                            "id": {
                              "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
                              "type": "string"
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
                          "required": [
                            "id"
                          ],
                          "type": "object",
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
                    "type": "array"
                  },
                  "components": {
                    "description": "Overrides of components encapsulated in a plugin. Overriding is done using a strategic merge",
                    "items": {
                      "properties": {
                        "container": {
                          "description": "Configuration overriding for a Container component",
                          "properties": {
                            "args": {
                              "description": "The arguments to supply to the command running the dockerimage component. The arguments are supplied either to the default command provided in the image or to the overridden command. Defaults to an empty array, meaning use whatever is defined in the image.",
                              "items": {
                                "type": "string"
                              },
                              "type": "array"
                            },
                            "command": {
                              "description": "The command to run in the dockerimage component instead of the default one provided in the image. Defaults to an empty array, meaning use whatever is defined in the image.",
                              "items": {
                                "type": "string"
                              },
                              "type": "array"
                            },
                            "endpoints": {
                              "items": {
                                "properties": {
                                  "attributes": {
                                    "additionalProperties": {
                                      "type": "string"
                                    },
                                    "type": "object"
                                  },
                                  "configuration": {
                                    "properties": {
                                      "cookiesAuthEnabled": {
                                        "type": "boolean"
                                      },
                                      "discoverable": {
                                        "type": "boolean"
                                      },
                                      "path": {
                                        "type": "string"
                                      },
                                      "protocol": {
                                        "description": "The is the low-level protocol of traffic coming through this endpoint. Default value is \"tcp\"",
                                        "type": "string"
                                      },
                                      "public": {
                                        "type": "boolean"
                                      },
                                      "scheme": {
                                        "description": "The is the URL scheme to use when accessing the endpoint. Default value is \"http\"",
                                        "type": "string"
                                      },
                                      "secure": {
                                        "type": "boolean"
                                      },
                                      "type": {
                                        "enum": [
                                          "ide",
                                          "terminal",
                                          "ide-dev"
                                        ],
                                        "type": "string"
                                      }
                                    },
                                    "type": "object",
                                    "additionalProperties": false
                                  },
                                  "name": {
                                    "type": "string"
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
                              "type": "array"
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
                            "name": {
                              "type": "string"
                            },
                            "sourceMapping": {
                              "description": "Optional specification of the path in the container where project sources should be transferred/mounted when ` + `mountSources` + ` is ` + `true` + `. When omitted, the value of the ` + `PROJECTS_ROOT` + ` environment variable is used.",
                              "type": "string"
                            },
                            "volumeMounts": {
                              "description": "List of volumes mounts that should be mounted is this container.",
                              "items": {
                                "description": "Volume that should be mounted to a component container",
                                "properties": {
                                  "name": {
                                    "description": "The volume mount name is the name of an existing ` + `Volume` + ` component. If no corresponding ` + `Volume` + ` component exist it is implicitly added. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files.",
                                    "type": "string"
                                  },
                                  "path": {
                                    "description": "The path in the component container where the volume should be mounted. If not path is mentioned, default path is the is ` + `/<name>` + `.",
                                    "type": "string"
                                  }
                                },
                                "required": [
                                  "name"
                                ],
                                "type": "object",
                                "additionalProperties": false
                              },
                              "type": "array"
                            }
                          },
                          "required": [
                            "name"
                          ],
                          "type": "object",
                          "additionalProperties": false
                        },
                        "kubernetes": {
                          "description": "Configuration overriding for a Kubernetes component",
                          "properties": {
                            "inlined": {
                              "description": "Inlined manifest",
                              "type": "string"
                            },
                            "name": {
                              "description": "Mandatory name that allows referencing the component in commands, or inside a parent",
                              "type": "string"
                            },
                            "uri": {
                              "description": "Location in a file fetched from a uri.",
                              "type": "string"
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
                        "openshift": {
                          "description": "Configuration overriding for an OpenShift component",
                          "properties": {
                            "inlined": {
                              "description": "Inlined manifest",
                              "type": "string"
                            },
                            "name": {
                              "description": "Mandatory name that allows referencing the component in commands, or inside a parent",
                              "type": "string"
                            },
                            "uri": {
                              "description": "Location in a file fetched from a uri.",
                              "type": "string"
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
                          "description": "Configuration overriding for a Volume component",
                          "properties": {
                            "name": {
                              "description": "Mandatory name that allows referencing the Volume component in Container volume mounts or inside a parent",
                              "type": "string"
                            },
                            "size": {
                              "description": "Size of the volume",
                              "type": "string"
                            }
                          },
                          "required": [
                            "name"
                          ],
                          "type": "object",
                          "additionalProperties": false
                        }
                      },
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
                    "type": "array"
                  },
                  "id": {
                    "description": "Id in a registry that contains a Devfile yaml file",
                    "type": "string"
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
                    "additionalProperties": false
                  },
                  "name": {
                    "description": "Optional name that allows referencing the component in commands, or inside a parent If omitted it will be infered from the location (uri or registryEntry)",
                    "type": "string"
                  },
                  "registryUrl": {
                    "type": "string"
                  },
                  "uri": {
                    "description": "Uri of a Devfile yaml file",
                    "type": "string"
                  }
                },
                "type": "object",
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
                  "name": {
                    "description": "Mandatory name that allows referencing the Volume component in Container volume mounts or inside a parent",
                    "type": "string"
                  },
                  "size": {
                    "description": "Size of the volume",
                    "type": "string"
                  }
                },
                "required": [
                  "name"
                ],
                "type": "object",
                "additionalProperties": false
              }
            },
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
          "type": "array"
        },
        "events": {
          "description": "Bindings of commands to events. Each command is referred-to by its name.",
          "properties": {
            "postStart": {
              "description": "Names of commands that should be executed after the workspace is completely started. In the case of Che-Theia, these commands should be executed after all plugins and extensions have started, including project cloning. This means that those commands are not triggered until the user opens the IDE in his browser.",
              "items": {
                "type": "string"
              },
              "type": "array"
            },
            "postStop": {
              "description": "Names of commands that should be executed after stopping the workspace.",
              "items": {
                "type": "string"
              },
              "type": "array"
            },
            "preStart": {
              "description": "Names of commands that should be executed before the workspace start. Kubernetes-wise, these commands would typically be executed in init containers of the workspace POD.",
              "items": {
                "type": "string"
              },
              "type": "array"
            },
            "preStop": {
              "description": "Names of commands that should be executed before stopping the workspace.",
              "items": {
                "type": "string"
              },
              "type": "array"
            }
          },
          "type": "object",
          "additionalProperties": false
        },
        "id": {
          "description": "Id in a registry that contains a Devfile yaml file",
          "type": "string"
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
          "additionalProperties": false
        },
        "projects": {
          "description": "Projects worked on in the workspace, containing names and sources locations",
          "items": {
            "properties": {
              "clonePath": {
                "description": "Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name.",
                "type": "string"
              },
              "git": {
                "description": "Project's Git source",
                "properties": {
                  "branch": {
                    "description": "The branch to check",
                    "type": "string"
                  },
                  "location": {
                    "description": "Project's source location address. Should be URL for git and github located projects, or; file:// for zip",
                    "type": "string"
                  },
                  "sparseCheckoutDir": {
                    "description": "Part of project to populate in the working directory.",
                    "type": "string"
                  },
                  "startPoint": {
                    "description": "The tag or commit id to reset the checked out branch to",
                    "type": "string"
                  }
                },
                "type": "object",
                "additionalProperties": false
              },
              "github": {
                "description": "Project's GitHub source",
                "properties": {
                  "branch": {
                    "description": "The branch to check",
                    "type": "string"
                  },
                  "location": {
                    "description": "Project's source location address. Should be URL for git and github located projects, or; file:// for zip",
                    "type": "string"
                  },
                  "sparseCheckoutDir": {
                    "description": "Part of project to populate in the working directory.",
                    "type": "string"
                  },
                  "startPoint": {
                    "description": "The tag or commit id to reset the checked out branch to",
                    "type": "string"
                  }
                },
                "type": "object",
                "additionalProperties": false
              },
              "name": {
                "description": "Project name",
                "type": "string"
              },
              "zip": {
                "description": "Project's Zip source",
                "properties": {
                  "location": {
                    "description": "Project's source location address. Should be URL for git and github located projects, or; file:// for zip",
                    "type": "string"
                  },
                  "sparseCheckoutDir": {
                    "description": "Part of project to populate in the working directory.",
                    "type": "string"
                  }
                },
                "type": "object",
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
          "type": "array"
        },
        "registryUrl": {
          "type": "string"
        },
        "uri": {
          "description": "Uri of a Devfile yaml file",
          "type": "string"
        }
      },
      "type": "object",
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
            "type": "string"
          },
          "git": {
            "description": "Project's Git source",
            "properties": {
              "branch": {
                "description": "The branch to check",
                "type": "string"
              },
              "location": {
                "description": "Project's source location address. Should be URL for git and github located projects, or; file:// for zip",
                "type": "string"
              },
              "sparseCheckoutDir": {
                "description": "Part of project to populate in the working directory.",
                "type": "string"
              },
              "startPoint": {
                "description": "The tag or commit id to reset the checked out branch to",
                "type": "string"
              }
            },
            "type": "object",
            "additionalProperties": false
          },
          "github": {
            "description": "Project's GitHub source",
            "properties": {
              "branch": {
                "description": "The branch to check",
                "type": "string"
              },
              "location": {
                "description": "Project's source location address. Should be URL for git and github located projects, or; file:// for zip",
                "type": "string"
              },
              "sparseCheckoutDir": {
                "description": "Part of project to populate in the working directory.",
                "type": "string"
              },
              "startPoint": {
                "description": "The tag or commit id to reset the checked out branch to",
                "type": "string"
              }
            },
            "type": "object",
            "additionalProperties": false
          },
          "name": {
            "description": "Project name",
            "type": "string"
          },
          "zip": {
            "description": "Project's Zip source",
            "properties": {
              "location": {
                "description": "Project's source location address. Should be URL for git and github located projects, or; file:// for zip",
                "type": "string"
              },
              "sparseCheckoutDir": {
                "description": "Part of project to populate in the working directory.",
                "type": "string"
              }
            },
            "type": "object",
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
      "type": "array"
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
			"type":"string",
			"description": "Optional URL to remote Dockerfile"
		},
		"alpha.deployment-manifest":  {
			"type":"string",
			"description": "Optional URL to remote Deployment Manifest"
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
  "additionalProperties": false,
  "required": [
    "schemaVersion"
  ]
}`
