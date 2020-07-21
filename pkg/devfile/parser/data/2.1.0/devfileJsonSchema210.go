package version210

const JsonSchema210 = `{
	"description": "Devfile schema.",
	"properties": {
	  "commands": {
		"description": "Predefined, ready-to-use, workspace-related commands",
		"items": {
		  "properties": {
			"composite": {
			  "description": "Composite command",
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
				  "type": "object"
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
				  "type": "boolean"
				}
			  },
			  "required": [
				"id"
			  ],
			  "type": "object"
			},
			"custom": {
			  "description": "Custom command",
			  "properties": {
				"attributes": {
				  "additionalProperties": {
					"type": "string"
				  },
				  "description": "Optional map of free-form additional command attributes",
				  "type": "object"
				},
				"commandClass": {
				  "type": "string"
				},
				"embeddedResource": {
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
				  "type": "object"
				},
				"id": {
				  "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
				  "type": "string"
				},
				"label": {
				  "description": "Optional label that provides a label for this command to be used in Editor UI menus for example",
				  "type": "string"
				}
			  },
			  "required": [
				"commandClass",
				"embeddedResource",
				"id"
			  ],
			  "type": "object"
			},
			"exec": {
			  "description": "Exec command",
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
					"type": "object"
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
				  "type": "object"
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
				"commandLine",
				"id"
			  ],
			  "type": "object"
			},
			"type": {
			  "description": "Type of workspace command",
			  "enum": [
				"Exec",
				"VscodeTask",
				"VscodeLaunch",
				"Custom"
			  ],
			  "type": "string"
			},
			"vscodeLaunch": {
			  "description": "VscodeLaunch command",
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
				  "type": "object"
				},
				"id": {
				  "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
				  "type": "string"
				},
				"inlined": {
				  "description": "Embedded content of the vscode configuration file",
				  "type": "string"
				},
				"locationType": {
				  "description": "Type of Vscode configuration command location",
				  "type": "string"
				},
				"url": {
				  "description": "Location as an absolute of relative URL",
				  "type": "string"
				}
			  },
			  "required": [
				"id"
			  ],
			  "type": "object"
			},
			"vscodeTask": {
			  "description": "VscodeTask command",
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
				  "type": "object"
				},
				"id": {
				  "description": "Mandatory identifier that allows referencing this command in composite commands, or from a parent, or in events.",
				  "type": "string"
				},
				"inlined": {
				  "description": "Embedded content of the vscode configuration file",
				  "type": "string"
				},
				"locationType": {
				  "description": "Type of Vscode configuration command location",
				  "type": "string"
				},
				"url": {
				  "description": "Location as an absolute of relative URL",
				  "type": "string"
				}
			  },
			  "required": [
				"id"
			  ],
			  "type": "object"
			}
		  },
		  "type": "object"
		},
		"type": "array"
	  },
	  "components": {
		"description": "List of the workspace components, such as editor and plugins, user-provided containers, or other types of components",
		"items": {
		  "properties": {
			"cheEditor": {
			  "description": "CheEditor component",
			  "properties": {
				"locationType": {
				  "description": "Type of plugin location",
				  "enum": [
					"RegistryEntry",
					"Uri"
				  ],
				  "type": "string"
				},
				"memoryLimit": {
				  "type": "string"
				},
				"name": {
				  "description": "Optional name that allows referencing the component in commands, or inside a parent If omitted it will be infered from the location (uri or registryEntry)",
				  "type": "string"
				},
				"registryEntry": {
				  "description": "Location of an entry inside a plugin registry",
				  "properties": {
					"baseUrl": {
					  "type": "string"
					},
					"id": {
					  "type": "string"
					}
				  },
				  "required": [
					"id"
				  ],
				  "type": "object"
				},
				"uri": {
				  "description": "Location defined as an URI",
				  "type": "string"
				}
			  },
			  "type": "object"
			},
			"chePlugin": {
			  "description": "ChePlugin component",
			  "properties": {
				"locationType": {
				  "description": "Type of plugin location",
				  "enum": [
					"RegistryEntry",
					"Uri"
				  ],
				  "type": "string"
				},
				"memoryLimit": {
				  "type": "string"
				},
				"name": {
				  "description": "Optional name that allows referencing the component in commands, or inside a parent If omitted it will be infered from the location (uri or registryEntry)",
				  "type": "string"
				},
				"registryEntry": {
				  "description": "Location of an entry inside a plugin registry",
				  "properties": {
					"baseUrl": {
					  "type": "string"
					},
					"id": {
					  "type": "string"
					}
				  },
				  "required": [
					"id"
				  ],
				  "type": "object"
				},
				"uri": {
				  "description": "Location defined as an URI",
				  "type": "string"
				}
			  },
			  "type": "object"
			},
			"container": {
			  "description": "Container component",
			  "properties": {
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
						"type": "object"
					  },
					  "name": {
						"type": "string"
					  },
					  "targetPort": {
						"type": "integer"
					  }
					},
					"required": [
					  "configuration",
					  "name",
					  "targetPort"
					],
					"type": "object"
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
					"type": "object"
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
				  "description": "Optional specification of the path in the container where project sources should be transferred/mounted when mountSources is true. When omitted, the value of the PROJECTS_ROOT environment variable is used.",
				  "type": "string"
				},
				"volumeMounts": {
				  "description": "List of volumes mounts that should be mounted is this container.",
				  "items": {
					"description": "Volume that should be mounted to a component container",
					"properties": {
					  "name": {
						"description": "The volume mount name is the name of an existing Volume component. If no corresponding Volume component exist it is implicitly added. If several containers mount the same volume name then they will reuse the same volume and will be able to access to the same files.",
						"type": "string"
					  },
					  "path": {
						"description": "The path in the component container where the volume should be mounted",
						"type": "string"
					  }
					},
					"required": [
					  "name",
					  "path"
					],
					"type": "object"
				  },
				  "type": "array"
				}
			  },
			  "required": [
				"image",
				"name"
			  ],
			  "type": "object"
			},
			"custom": {
			  "description": "Custom component",
			  "properties": {
				"componentClass": {
				  "type": "string"
				},
				"embeddedResource": {
				  "type": "object"
				},
				"name": {
				  "type": "string"
				}
			  },
			  "required": [
				"componentClass",
				"embeddedResource",
				"name"
			  ],
			  "type": "object"
			},
			"kubernetes": {
			  "description": "Kubernetes component",
			  "properties": {
				"inlined": {
				  "description": "Reference to the plugin definition",
				  "type": "string"
				},
				"locationType": {
				  "description": "Type of Kubernetes-like location",
				  "type": "string"
				},
				"name": {
				  "description": "Mandatory name that allows referencing the component in commands, or inside a parent",
				  "type": "string"
				},
				"url": {
				  "description": "Location in a plugin registry",
				  "type": "string"
				}
			  },
			  "required": [
				"name"
			  ],
			  "type": "object"
			},
			"openshift": {
			  "description": "Openshift component",
			  "properties": {
				"inlined": {
				  "description": "Reference to the plugin definition",
				  "type": "string"
				},
				"locationType": {
				  "description": "Type of Kubernetes-like location",
				  "type": "string"
				},
				"name": {
				  "description": "Mandatory name that allows referencing the component in commands, or inside a parent",
				  "type": "string"
				},
				"url": {
				  "description": "Location in a plugin registry",
				  "type": "string"
				}
			  },
			  "required": [
				"name"
			  ],
			  "type": "object"
			},
			"type": {
			  "description": "Type of project source",
			  "enum": [
				"Container",
				"Kubernetes",
				"Openshift",
				"CheEditor",
				"Volume",
				"ChePlugin",
				"Custom",
				"Dockerfile"
			  ],
			  "type": "string"
			},
			"volume": {
			  "description": "Volume component",
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
			  "type": "object"
			},
			"Dockerfile":{
				"description":"Dockerfile component",
				"properties":{
					"name":{
						"description":"Mandatory name that allows referencing the dockerfile component",
						"type":"string"
				   	},
				   	"source":{
						"sourceDir":{
							"description":"path of source directory to establish build context",
							"type":"string"
						},
						"location":{
							"description":"location of the source code repostory",
						 	"type":"string"
						},
					  	"type":"object"
				   	},
					"dockerfileLocation":{
						"description":"path to dockerfile",
						"type":"string"
					},
					"destination":{
						"description":"path to registry where the build image is to be pushed",
						"type":"string"
					}
				},
				"required":[
					"name",
					"dockerfileLocation",
					"source"
				],
				"type":"object"
			}
		  },
		  "type": "object"
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
		"type": "object"
	  },
	  "parent": {
		"description": "Parent workspace template",
		"properties": {
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
			"type": "object"
		  },
		  "locationType": {
			"description": "Type of parent location",
			"enum": [
			  "Uri",
			  "RegistryEntry",
			  "Kubernetes"
			],
			"type": "string"
		  },
		  "registryEntry": {
			"description": "Entry in a registry (base URL + ID) that contains a Devfile yaml file",
			"properties": {
			  "baseUrl": {
				"type": "string"
			  },
			  "id": {
				"type": "string"
			  }
			},
			"required": [
			  "id"
			],
			"type": "object"
		  },
		  "uri": {
			"description": "Uri of a Devfile yaml file",
			"type": "string"
		  }
		},
		"type": "object"
	  },
	  "projects": {
		"description": "Projects worked on in the workspace, containing names and sources locations",
		"items": {
		  "properties": {
			"clonePath": {
			  "description": "Path relative to the root of the projects to which this project should be cloned into. This is a unix-style relative path (i.e. uses forward slashes). The path is invalid if it is absolute or tries to escape the project root through the usage of '..'. If not specified, defaults to the project name.",
			  "type": "string"
			},
			"custom": {
			  "description": "Project's Custom source",
			  "properties": {
				"embeddedResource": {
				  "type": "object"
				},
				"projectSourceClass": {
				  "type": "string"
				}
			  },
			  "required": [
				"embeddedResource",
				"projectSourceClass"
			  ],
			  "type": "object"
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
			  "required": [
				"location"
			  ],
			  "type": "object"
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
			  "required": [
				"location"
			  ],
			  "type": "object"
			},
			"name": {
			  "description": "Project name",
			  "type": "string"
			},
			"sourceType": {
			  "description": "Type of project source",
			  "enum": [
				"Git",
				"Github",
				"Zip",
				"Custom"
			  ],
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
			  "required": [
				"location"
			  ],
			  "type": "object"
			}
		  },
		  "required": [
			"name"
		  ],
		  "type": "object"
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
	"required": [
	  "schemaVersion"
	]
  }`
