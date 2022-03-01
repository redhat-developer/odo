### Id and Name:
`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`

The restriction is added to allow easy translation to K8s resource names, and also to have consistent rules for both name and id fields.

The validation will be done as part of schema validation, the rule will be introduced as a regex in schema definition, any objection of the rule in devfile will result in a failure.

- limit to lowercase characters i.e., no uppercase allowed
- limit within 63 characters
- no special characters allowed except dash(-)
- start with an alphanumeric character
- end with an alphanumeric character


### Endpoints:
- all the endpoint names are unique across components
- all the endpoint ports are unique across components. Only exception: container component with `dedicatedpod=true`

### Commands:
1. id must be unique
2. composite command:
    - Should not reference itself via a subcommand
    - Should not indirectly reference itself via a subcommand which is a composite command
    - Should reference a valid devfile command
3. exec command should: map to a valid container component
4. apply command should: map to a valid container/kubernetes/openshift/image component
5. `{build, run, test, debug, deploy}`, each kind of group can only have one default command associated with it. If there are multiple commands of the same kind without a default, a warning will be displayed.

### Components:
Common rules for all components types:
- Name must be unique

#### Container component 
1. the container components must reference a valid volume component if it uses volume mounts, and the volume components are unique
2. `PROJECT_SOURCE` or `PROJECTS_ROOT` are reserved environment variables defined under env, cannot be defined again in `env`
3. the annotations should not have conflict values for same key, except deployment annotations and service annotations set for a container with `dedicatedPod=true`
4. resource requirements, e.g. `cpuLimit`, `cpuRequest`, `memoryLimit`, `memoryRequest`, must be in valid quantity format; and the resource requested must be less than the resource limit (if specified).

#### Plugin Component
- Commands in plugins components share the same commands validation rules as listed above. Validation occurs after overriding and merging, in flattened devfile
- Registry URL needs to be in valid format

#### Kubernetes & Openshift component 
- URI needs to be in valid URI format

#### Image component 
- A Dockerfile Image component's git source cannot have more than one remote defined. If checkout remote is mentioned, validate it against the remote configured map


### Events:
1. preStart and postStop events can only be Apply commands
2. postStart and preStop events can only be Exec commands
3. if preStart and postStop events refer to a composite command, then all containing commands need to be Apply commands.
4. if postStart and preStop events refer to a composite command, then all containing commands need to be Exec commands.


### Parent:
- Share the same validation rules as listed above. Validation occurs after overriding and merging, in flattened devfile


### starterProjects:
- Starter project entries cannot have more than one remote defined
- if checkout remote is mentioned, validate it against the starter project remote configured map

### projects
- if more than one remote is configured, a checkout remote is mandatory
- if checkout remote is mentioned, validate it against the starter project remote configured map
