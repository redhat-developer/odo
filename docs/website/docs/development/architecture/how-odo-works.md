---
title: How odo works
sidebar_position: 30
# Display h2 to h5 headings
toc_min_heading_level: 2
toc_max_heading_level: 6
---

## How `odo dev` works

In a nutshell, when running [`odo dev`](../../command-reference/dev), `odo` will do the following:

1. It reads and validates the Devfile in the current directory.
   For example, it makes sure a `command` of the right kind (`run` or `debug` if running `odo dev --debug`) is defined.
2. It determines the resources it needs to create (or update) and manage, and instructs the Kubernetes API Server to create all necessary resources.
   Specifically, `odo` creates the following resources:
   - [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) for running the containers. See [below](#deployment) for further details.
   - [Service](https://kubernetes.io/docs/concepts/services-networking/service/) for accessibility. See [below](#service) for further details.

   If the Devfile contains any [`Kubernetes` components](../../development/devfile#components) not referenced by any [`apply` command](https://devfile.io/docs/2.2.0-alpha/adding-apply-commands),
   `odo dev` automatically applies them.
   This is the mechanism used by [`odo add binding`](../../command-reference/add-binding) to have `ServiceBinding` resources created by `odo dev`. 
3. Once the resources are ready, `odo` executes any `build` (optional) and `run` (or `debug`) commands defined in the Devfile into the right containers.   
4. It reacts to events occurring in the cluster and that might affect the resources managed.
5. Unless told otherwise (when running `odo dev --no-watch`), `odo` watches for local changes
   and synchronizes changed files into the running container.
   If the local Devfile is modified, `odo` may need to change the resources it previously created, which might result in recreating the
   running containers.
   Note that synchronization and push to the cluster can also be triggered on demand by pressing `p` at any time.
   See [this page](../../command-reference/dev#applying-local-changes-to-the-application-on-the-cluster) for more details.
6. `odo` optionally restarts the running application (if the command is not marked as `hotReloadCapable` in the Devfile).
7. `odo` then sets up port-forwarding, starting from 40001 as the local port and incrementing by 1 until it finds a free local port.
8. When `odo dev` is stopped (via `Ctrl+C`), it deletes all the resources created previously and stops port-forwarding and code synchronization.

:::caution
It is strongly discouraged to run multiple `odo dev` processes in parallel from the same component directory.
Otherwise, such processes will compete with each other in trying to manage the same Kubernetes resources,
and you would end up with several instances of port-forwarding and code synchronization.
:::

## How `odo dev` translates a `container` component into Kubernetes resources

Given a component (aptly named `my-component-name` in the `metadata.name` field) in the Devfile excerpt below: 

```yaml showLineNumbers
metadata:
  # highlight-next-line
  name: my-component-name
```

`odo` will create the following Kubernetes resources in the current namespace:
- a Deployment named `my-component-name-app`
- a Service named `my-component-name-app`

:::info NOTE
Per the Devfile specification, the `metadata.name` field is optional.
If it is not defined, `odo` will try to autodetect the name from the project source code (based on information from files like `package.json` or `pom.xml`).
As a last resort, if a name cannot be determined automatically, it will use the current directory name.
:::

### Deployment

`odo` will create a Deployment with the characteristics below.

#### Annotations

By default, `odo` adds the following annotations to the Deployment:

| Key                                    | Description                                                                                                                                           | Example Value       |
|----------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------|
| `odo.dev/project-type `                | the application runtime, if available. Optional. Value is read in order from the `metadata.projectType` or `metadata.language` fields in the Devfile  | `spring`            |

Notes:
- Any additional annotations defined via the `components[].container.annotation.deployment` field will also be added to this resource.
- All those annotations are also added to the underlying Pods managed by this Deployment.

<details>
<summary>Example</summary>

<table>
<thead>
<tr>
<td>Devfile</td>
<td></td>
<td>Kubernetes Deployment</td>
</tr>
</thead>
<tbody>
<tr>
<td>

```yaml
metadata:
  # highlight-start
  name: my-sample-java-springboot
  projectType: spring
  # highlight-end
  language: java
components:
- name: tools
  container:
     annotation:
        deployment:
           # highlight-start
           example.com/my-annotation: value-1
           # highlight-end
```
</td>
<td> => </td>
<td>

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
   # highlight-next-line
   name: my-sample-java-springboot-app
   annotations:
      # highlight-start
      odo.dev/project-type: spring
      example.com/my-annotation: value-1
      # highlight-end
spec:
   template:
      metadata:
         # highlight-next-line
         name: my-sample-java-springboot-app
         annotations:
            # highlight-start
            odo.dev/project-type: spring
            example.com/my-annotation: value-1
            # highlight-end
```
</td>
</tr>
</tbody>
</table>

</details>

#### Labels

By default, `odo` adds the following labels to the Deployment:

| Key                                    | Description                                                                                                                                          | Example Value       |
|----------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------|
| `app`                                  | the application; always `app`                                                                                                                        | `app`               |
| `app.kubernetes.io/instance`           | the component name                                                                                                                                   | `my-component-name` |
| `app.kubernetes.io/managed-by`         | the tool used to create this resource; always `odo`                                                                                                  | `odo`               |
| `app.kubernetes.io/managed-by-version` | the version of the odo binary used to create this resource                                                                                           | `v3.0.0-rc1`        |
| `app.kubernetes.io/part-of`            | the higher-level application using this resource; always `app`                                                                                       | `app`               |
| `app.openshift.io/runtime`             | the application runtime, if available. Optional. Value is read in order from the `metadata.projectType` or `metadata.language` fields in the Devfile | `spring`            |
| `component`                            | the component name                                                                                                                                   | `my-component-name` |
| `odo.dev/mode`                         | in which mode the component is running. Possible values: `Dev`, `Deploy`                                                                             | `Dev`               |

Note that the same labels are added to the underlying Pods managed by this Deployment.

<details>
<summary>Example</summary>

<table>
<thead>
<tr>
<td>Devfile</td>
<td></td>
<td>Kubernetes Deployment</td>
</tr>
</thead>
<tbody>
<tr>
<td>

```yaml
metadata:
  # highlight-start
  name: my-sample-java-springboot
  projectType: spring
  # highlight-end
  language: java
components:
- name: tools
  container:
    ...
```
</td>
<td> => </td>
<td>

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
   # highlight-next-line
   name: my-sample-java-springboot-app
   labels:
      # highlight-start
      app: app
      app.kubernetes.io/instance: my-sample-java-springboot
      app.kubernetes.io/managed-by: odo
      app.kubernetes.io/managed-by-version: v3.0.0-rc1
      app.kubernetes.io/part-of: app
      app.openshift.io/runtime: spring
      component: my-sample-java-springboot
      odo.dev/mode: Dev
      # highlight-end
spec:
    template:
        metadata:
            # highlight-next-line
            name: my-sample-java-springboot-app
            labels:
                # highlight-start
                app: app
                app.kubernetes.io/instance: my-sample-java-springboot
                app.kubernetes.io/managed-by: odo
                app.kubernetes.io/managed-by-version: v3.0.0-rc1
                app.kubernetes.io/part-of: app
                app.openshift.io/runtime: spring
                component: my-sample-java-springboot
                odo.dev/mode: Dev
                # highlight-end
    selector:
      matchLabels:
        component: my-sample-java-springboot
```
</td>
</tr>
</tbody>
</table>

</details>

#### Replicas

The number of Replicas for this Deployment is explicitly set to 1 and is expected to always have this value.

#### Pods and Containers

Each `components[].container` block is translated into a dedicated `container` definition in the Pod template.

##### Environment variables

Each `components[].container.env` section defined in the Devfile translates into the same container environment variable in the Pod container.

Additionally, the following environment variables are reserved and injected into the container definition: 

| Key              | Description                                                                                                                        | Example Value  |
|------------------|------------------------------------------------------------------------------------------------------------------------------------|----------------|
| `PROJECTS_ROOT`  | A path where projects sources are mounted as defined by container component's `sourceMapping`                                      | `/projects`    |
| `PROJECT_SOURCE` | A path to a project source (`$PROJECTS_ROOT/`). If there are multiple projects, this will point to the directory of the first one  | `/projects`    |


<details>
<summary>Example</summary>

<table>
<thead>
<tr>
<td>Devfile</td>
<td></td>
<td>Kubernetes Deployment</td>
</tr>
</thead>
<tbody>
<tr>
<td>

```yaml
metadata:
  # highlight-next-line
  name: my-sample-java-springboot
components:
- name: tools
  container:
     image: quay.io/eclipse/che-java11-maven
     env:
     # highlight-start
     - name: DEBUG_PORT
       value: "5858"
     # highlight-end
```
</td>
<td> => </td>
<td>

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
   # highlight-next-line
   name: my-sample-java-springboot-app
spec:
    template:
        metadata:
            # highlight-next-line
            name: my-sample-java-springboot-app
        spec:
            containers:
               - name: tools
                 env:
                  # highlight-start
                  - name: DEBUG_PORT
                    value: "5858"
                  - name: PROJECTS_ROOT
                    value: /projects
                  - name: PROJECT_SOURCE
                    value: /projects
                  # highlight-end
                 image: quay.io/eclipse/che-java11-maven
                 imagePullPolicy: Always
```
</td>
</tr>
</tbody>
</table>

</details>

##### Command and Args

`odo` will use the specified `components[].container.command` or `components[].container.args` fields as is 
for the Kubernetes container `command` and `args` definitions.
The only requirement is that those fields should result in a non-terminating container, 
so `odo` can execute the commands it needs to manage the application. 

If both fields are missing, `odo` defaults to:
- setting the container `command` to `tail`.
  This assumes that the container image (set in the Devfile) contains the `tail` executable.
- setting the container `args` to `[-f, /dev/null]`


<details>
<summary>Example</summary>

<table>
<thead>
<tr>
<td>Devfile</td>
<td></td>
<td>Kubernetes Deployment</td>
</tr>
</thead>
<tbody>
<tr>
<td>

```yaml
metadata:
  # highlight-next-line
  name: my-sample-java-springboot
  projectType: spring
  language: java
components:
- name: tools
  container:
     image: quay.io/eclipse/che-java11-maven
     # no command or args fields set
```
</td>
<td> => </td>
<td>

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
   # highlight-next-line
   name: my-sample-java-springboot-app
spec:
    template:
        metadata:
            # highlight-next-line
            name: my-sample-java-springboot-app
        spec:
            containers:
               - name: tools
                 # highlight-start
                 command: ['tail']
                 args: ['-f', '/dev/null']
                 # highlight-end
                 image: quay.io/eclipse/che-java11-maven
                 imagePullPolicy: Always
```
</td>
</tr>
<tr>
<td>

```yaml
metadata:
  # highlight-next-line
  name: my-sample-java-springboot
components:
- name: tools
  container:
     image: quay.io/eclipse/che-java11-maven
     # highlight-start
     command: ['/bin/my-entrypoint.sh']
     args: ['arg1', 'arg2']
     # highlight-end
```
</td>
<td> => </td>
<td>

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
   # highlight-next-line
   name: my-sample-java-springboot-app
spec:
    template:
        metadata:
            # highlight-next-line
            name: my-sample-java-springboot-app
        spec:
            containers:
               - name: tools
                 # highlight-start
                 command: ['/bin/my-entrypoint.sh']
                 args: ['arg1', 'arg2']
                 # highlight-end
                 image: quay.io/eclipse/che-java11-maven
                 imagePullPolicy: Always
```
</td>
</tr>
</tbody>
</table>

</details>

##### Image Pull Policy

At this time, the image pull policy for all containers is `Always`.

##### Resources Limits and Requests

`odo` maps each `components[].container.{cpu,memory}{Limit,Request}` to corresponding `resources.{limits,requests}.{cpu,memory}` fields with the respective values.


<details>
<summary>Example</summary>

<table>
<thead>
<tr>
<td>Devfile</td>
<td></td>
<td>Kubernetes Deployment</td>
</tr>
</thead>
<tbody>
<tr>
<td>

```yaml
metadata:
  # highlight-next-line
  name: my-sample-java-springboot
components:
- name: tools
  container:
     image: quay.io/eclipse/che-java11-maven
     # highlight-start
     cpuLimit: 500m
     cpuRequest: 250m
     memoryLimit: 512Mi
     memoryRequest: 256Mi
     # highlight-end
```
</td>
<td> => </td>
<td>

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
   # highlight-next-line
   name: my-sample-java-springboot-app
spec:
    template:
        metadata:
            # highlight-next-line
            name: my-sample-java-springboot-app
        spec:
            containers:
               - name: tools
                 # highlight-start
                 resources:
                    limits:
                       cpu: 500m
                       memory: 512Mi
                    requests:
                       cpu: 250m
                       memory: 256Mi
                 # highlight-end
                 image: quay.io/eclipse/che-java11-maven
                 imagePullPolicy: Always
```
</td>
</tr>
</tbody>
</table>

</details>

##### Ports

`odo` translates each element in the `components[].container.endpoints[]` block into a dedicated `containerPort` with the same name and port.

<details>
<summary>Example</summary>

<table>
<thead>
<tr>
<td>Devfile</td>
<td></td>
<td>Kubernetes Deployment</td>
</tr>
</thead>
<tbody>
<tr>
<td>

```yaml
metadata:
  # highlight-next-line
  name: my-sample-java-springboot
components:
- name: tools
  container:
     image: quay.io/eclipse/che-java11-maven
     # highlight-start
     endpoints:
     - name: http-springboot
       targetPort: 8080
     - name: my-custom-ep
       targetPort: 3000
       protocol: http
       exposure: internal
     - name: debug
       targetPort: 5858
       exposure: none
     # highlight-end
```
</td>
<td> => </td>
<td>

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
   # highlight-next-line
   name: my-sample-java-springboot-app
spec:
    template:
        metadata:
            # highlight-next-line
            name: my-sample-java-springboot-app
        spec:
            containers:
               - name: tools
                 # highlight-start
                 ports:
                 - containerPort: 8080
                   name: http-springboot
                   protocol: TCP
                 - containerPort: 3000
                   name: my-custom-ep
                   protocol: TCP
                 - containerPort: 5858
                   name: debug
                   protocol: TCP
                 # highlight-end
                 image: quay.io/eclipse/che-java11-maven
                 imagePullPolicy: Always
```
</td>
</tr>
</tbody>
</table>

</details>

#### Volumes

`odo` creates the following volumes and mounts them in the containers:

| Volume name       | Volume Type                                                                                                                                                                                                                                                                                             | Mount Path   | Description                                                                              |
|-------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------|------------------------------------------------------------------------------------------|
| `odo-shared-data` | [`emptyDir`](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir)                                                                                                                                                                                                                             | `/opt/odo`   | Internal. Contain files (like PIDs for commands) necessary for `odo`.                    |
| `odo-projects`    | [PersistentVolumeClaim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims) (PVC) if [`Ephemeral`](../../overview/configure#preference-key-table) preference is `false`.<br/>[`emptyDir`](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) otherwise. | `/projects`  | Used for syncing files. Mounted only if container component has `mountSources: true` set |

Every `volume` component in the Devfile is translated into a [PersistentVolumeClaim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims) (PVC) (with the default storage class) 
and a Volume referencing that PVC in the Deployment.
If the Volume is mounted in the Devfile container component, the corresponding Kubernetes container will also contain a `VolumeMount` referencing that Volume.


<details>
<summary>Example</summary>

<table>
<thead>
<tr>
<td>Devfile</td>
<td></td>
<td>Kubernetes Deployment</td>
</tr>
</thead>
<tbody>
<tr>
<td>

```yaml
metadata:
  # highlight-next-line
  name: my-sample-java-springboot
components:
- name: tools
  container:
     image: quay.io/eclipse/che-java11-maven
     # highlight-next-line
     mountSources: true
```
</td>
<td> => </td>
<td>

* With `Ephemeral` Setting set to `false` (default)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
   # highlight-next-line
   name: my-sample-java-springboot-app
spec:
    template:
        metadata:
            # highlight-next-line
            name: my-sample-java-springboot-app
        spec:
            volumes:
               # highlight-start
               - name: odo-shared-data
                 emptyDir: {}
               - name: odo-projects
                 persistentVolumeClaim:
                    # odo also creates and manages this PVC
                    claimName: odo-projects-my-sample-java-springboot-app
               # highlight-end
            containers:
               - name: tools
                 # highlight-start
                 volumeMounts:
                    - name: odo-shared-data
                      mountPath: /opt/odo/
                    - name: odo-projects
                      mountPath: /projects
                 # highlight-end
```

---

* With `Ephemeral` Setting set to `true`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
   # highlight-next-line
   name: my-sample-java-springboot-app
spec:
    template:
        metadata:
            # highlight-next-line
            name: my-sample-java-springboot-app
        spec:
            volumes:
               # highlight-start
               - name: odo-shared-data
                 emptyDir: {}
               - name: odo-projects
                 emptyDir: {}
               # highlight-end
            containers:
               - name: tools
                 # highlight-start
                 volumeMounts:
                    - name: odo-shared-data
                      mountPath: /opt/odo/
                    - name: odo-projects
                      mountPath: /projects
                 # highlight-end
```

</td>
</tr>

<tr>
<td>

* With a `volume` component

```yaml
metadata:
  # highlight-next-line
  name: my-sample-java-springboot
components:
- name: tools
  container:
     # highlight-start
     volumeMounts:
        - name: m2
          path: /home/user/.m2
     # highlight-end
     image: quay.io/eclipse/che-java11-maven
     mountSources: true
- name: m2
  # highlight-start
  volume:
     size: 3Gi
  # highlight-end
```
</td>
<td> => </td>
<td>

* With `Ephemeral` Setting set to `false` (default)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
   # highlight-next-line
   name: my-sample-java-springboot-app
spec:
    template:
        metadata:
            # highlight-next-line
            name: my-sample-java-springboot-app
        spec:
            volumes:
               # highlight-start
               - name: odo-shared-data
                 emptyDir: {}
               - name: odo-projects
                 persistentVolumeClaim:
                    # odo also creates and manages this PVC
                    claimName: odo-projects-my-sample-java-springboot-app
               - name: m2-my-sample-java-springboot-app-vol
                 persistentVolumeClaim:
                    # odo also creates and manages this PVC
                    claimName: m2-my-sample-java-springboot-app
               # highlight-end
            containers:
               - name: tools
                 # highlight-start
                 volumeMounts:
                    - name: odo-shared-data
                      mountPath: /opt/odo/
                    - name: odo-projects
                      mountPath: /projects
                    - name: m2-my-sample-java-springboot-app-vol
                      mountPath: /home/user/.m2
                 # highlight-end
```

---

* With `Ephemeral` Setting set to `true`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
   # highlight-next-line
   name: my-sample-java-springboot-app
spec:
    template:
        metadata:
            # highlight-next-line
            name: my-sample-java-springboot-app
        spec:
            volumes:
               # highlight-start
               - name: odo-shared-data
                 emptyDir: {}
               - name: odo-projects
                 emptyDir: {}
               - name: m2-my-sample-java-springboot-app-vol
                 persistentVolumeClaim:
                    # odo also creates and manages this PVC
                    claimName: m2-my-sample-java-springboot-app
               # highlight-end
            containers:
               - name: tools
                 # highlight-start
                 volumeMounts:
                    - name: odo-shared-data
                      mountPath: /opt/odo/
                    - name: odo-projects
                      mountPath: /projects
                    - name: m2-my-sample-java-springboot-app-vol
                      mountPath: /home/user/.m2
                 # highlight-end
```

</td>
</tr>

</tbody>
</table>

</details>

### Service

`odo` will create a Service of type `ClusterIP` with the characteristics below.

#### Annotations

By default, `odo` adds the following annotations to the Service:

| Key                                 | Value                                                                                             | Description                                                                                                                                                                                                                                                                                       |
|-------------------------------------|---------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `service.binding/backend_ip`        | `path={.spec.clusterIP}`                                                                          | exposes the Service `clusterIP` address as binding data, so this can be used as a backing service via the Service Binding Operator (SBO). More details on [SBO documentation](https://redhat-developer.github.io/service-binding-operator/userguide/exposing-binding-data/adding-annotation.html) |
| `service.binding/backend_port`      | `path={.spec.ports},`<br/>`elementType=sliceOfMaps,`<br/>`sourceKey=name,`<br/>`sourceValue=port` | exposes the Service ports as binding data, so this can be used as a backing service via the Service Binding Operator (SBO). More details on [SBO documentation](https://redhat-developer.github.io/service-binding-operator/userguide/exposing-binding-data/adding-annotation.html)               |

Note that any additional annotations defined via the `components[].container.annotation.service` Devfile field will also be added to this resource.


<details>
<summary>Example</summary>

<table>
<thead>
<tr>
<td>Devfile</td>
<td></td>
<td>Kubernetes Service</td>
</tr>
</thead>
<tbody>
<tr>
<td>

```yaml
metadata:
  # highlight-start
  name: my-sample-java-springboot
  projectType: spring
  # highlight-end
  language: java
components:
- name: tools
  container:
     annotation:
        service:
           # highlight-start
           example.com/my-svc-annotation: value-1
           # highlight-end
```
</td>
<td> => </td>
<td>

```yaml
apiVersion: v1
kind: Service
metadata:
   # highlight-next-line
   name: my-sample-java-springboot-app
   annotations:
      # highlight-start
      service.binding/backend_ip: path={.spec.clusterIP}
      service.binding/backend_port: path={.spec.ports},elementType=sliceOfMaps,sourceKey=name,sourceValue=port
      example.com/my-svc-annotation: value-1
      # highlight-end
spec:
   type: ClusterIP
   selector:
      # highlight-next-line
      component: my-sample-java-springboot-app
```
</td>
</tr>
</tbody>
</table>

</details>


#### Labels

By default, `odo` adds the following labels to the Service:

| Key                                    | Description                                                                                                                                          | Example Value       |
|----------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------|
| `app`                                  | the application; always `app`                                                                                                                        | `app`               |
| `app.kubernetes.io/instance`           | the component name                                                                                                                                   | `my-component-name` |
| `app.kubernetes.io/managed-by`         | the tool used to create this resource; always `odo`                                                                                                  | `odo`               |
| `app.kubernetes.io/managed-by-version` | the version of the odo binary used to create this resource                                                                                           | `v3.0.0-rc1`        |
| `app.kubernetes.io/part-of`            | the higher-level application using this resource; always `app`                                                                                       | `app`               |
| `app.openshift.io/runtime`             | the application runtime, if available. Optional. Value is read in order from the `metadata.projectType` or `metadata.language` fields in the Devfile | `spring`            |
| `component`                            | the component name                                                                                                                                   | `my-component-name` |
| `odo.dev/mode`                         | in which mode the component is running. Possible values: `Dev`, `Deploy`                                                                             | `Dev`               |


<details>
<summary>Example</summary>

<table>
<thead>
<tr>
<td>Devfile</td>
<td></td>
<td>Kubernetes Service</td>
</tr>
</thead>
<tbody>
<tr>
<td>

```yaml
metadata:
  # highlight-start
  name: my-sample-java-springboot
  projectType: spring
  # highlight-end
  language: java
components:
- name: tools
  container:
    ...
```
</td>
<td> => </td>
<td>

```yaml
apiVersion: v1
kind: Service
metadata:
   # highlight-next-line
   name: my-sample-java-springboot-app
   labels:
      # highlight-start
      app: app
      app.kubernetes.io/instance: my-sample-java-springboot
      app.kubernetes.io/managed-by: odo
      app.kubernetes.io/managed-by-version: v3.0.0-rc1
      app.kubernetes.io/part-of: app
      app.openshift.io/runtime: spring
      component: my-sample-java-springboot
      odo.dev/mode: Dev
      # highlight-end
spec:
   type: ClusterIP
   selector:
      # highlight-next-line
      component: my-sample-java-springboot-app
```
</td>
</tr>
</tbody>
</table>

</details>

#### Ports

For each endpoint with an `exposure` other than `none` defined in the `components[].container.endpoints` Devfile block, 
`odo` adds a `port` to the Service `spec`.

<details>
<summary>Example</summary>

<table>
<thead>
<tr>
<td>Devfile</td>
<td></td>
<td>Kubernetes Service</td>
</tr>
</thead>
<tbody>
<tr>
<td>

```yaml
metadata:
  # highlight-next-line
  name: my-sample-java-springboot
components:
   - id: my-container1
     container:
        endpoints:
           # highlight-start
           - name: http-springboot
             targetPort: 8080
           - name: my-custom-ep
             targetPort: 3000
             exposure: internal
           - name: debug
             targetPort: 5005
             exposure: none
        # highlight-end
   - id: my-container2
     container:
        endpoints:
           # highlight-start
           - name: another-ep
             targetPort: 9090
           # highlight-end
```
</td>
<td> => </td>
<td>

```yaml
apiVersion: v1
kind: Service
metadata:
   # highlight-next-line
   name: my-sample-java-springboot-app
spec:
   type: ClusterIP
   ports:
   # highlight-start
   - name: http-springboot
     port: 8080
     protocol: TCP
     targetPort: 8080
   - name: my-custom-ep
     port: 3000
     protocol: TCP
     targetPort: 3000
   - name: another-ep
     port: 9090
     protocol: TCP
     targetPort: 3000
   # highlight-end
```

</td>
</tr>
</tbody>
</table>

</details>


### Full example

<details>
<summary>Example of Devfile and resulting Kubernetes resources</summary>

Given this Devfile:

```yaml
schemaVersion: 2.2.0
metadata:
   description: Spring Boot® using Java
   displayName: Spring Boot®
   globalMemoryLimit: 2674Mi
   icon: https://spring.io/images/projects/spring-edf462fec682b9d48cf628eaf9e19521.svg
   language: java
   name: my-sample-java-springboot
   projectType: spring
   tags:
      - Java
      - Spring
   version: 1.1.0

commands:
- exec:
    commandLine: mvn clean -Dmaven.repo.local=/home/user/.m2/repository package -Dmaven.test.skip=true
    component: tools
    group:
      isDefault: true
      kind: build
    workingDir: ${PROJECT_SOURCE}
  id: build
- exec:
    commandLine: mvn -Dmaven.repo.local=/home/user/.m2/repository spring-boot:run
    component: tools
    group:
      isDefault: true
      kind: run
    workingDir: ${PROJECT_SOURCE}
  id: run
- exec:
    commandLine: java -Xdebug -Xrunjdwp:server=y,transport=dt_socket,address=${DEBUG_PORT},suspend=n
      -jar target/*.jar
    component: tools
    group:
      isDefault: true
      kind: debug
    workingDir: ${PROJECT_SOURCE}
  id: debug

components:
- container:
    annotation:
      deployment:
        example.com/my-deploy-annotation1: my-deploy-annotation-val1
      service:
        example.com/my-svc-annotation1: my-svc-annotation-val1
    endpoints:
    - name: http-springboot
      targetPort: 8080
    - name: debug
      targetPort: 5858
      exposure: none
    env:
    - name: DEBUG_PORT
      value: "5858"
    image: quay.io/eclipse/che-java11-maven:next
    memoryLimit: 768Mi
    mountSources: true
    volumeMounts:
    - name: m2
      path: /home/user/.m2
  name: tools

- container:
    annotation:
      deployment:
        example.com/my-deploy-annotation-echo1: my-deploy-annotation-val1
      service:
        example.com/my-svc-annotation-echo1: my-svc-annotation-val1
    endpoints:
    - name: echo-ep1
      targetPort: 18080
    env:
    - name: MY_ENV_VAR
      value: "some value"
    image: alpine:latest
    mountSources: false
    command: [tail]
    args: [-f, /dev/null]
  name: echo-container

- name: m2
  volume:
    size: 3Gi

```

`odo` will generate the following Kubernetes Resources:

* Deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    example.com/my-deploy-annotation-echo1: my-deploy-annotation-val1
    example.com/my-deploy-annotation1: my-deploy-annotation-val1
    odo.dev/project-type: spring
  labels:
    app: app
    app.kubernetes.io/instance: my-sample-java-springboot
    app.kubernetes.io/managed-by: odo
    app.kubernetes.io/managed-by-version: v3.0.0-rc1
    app.kubernetes.io/part-of: app
    app.openshift.io/runtime: spring
    component: my-sample-java-springboot
    odo.dev/mode: Dev
  name: my-sample-java-springboot-app
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      component: my-sample-java-springboot
  strategy:
    type: Recreate
  template:
    metadata:
      annotations:
        example.com/my-deploy-annotation-echo1: my-deploy-annotation-val1
        example.com/my-deploy-annotation1: my-deploy-annotation-val1
        odo.dev/project-type: spring
      labels:
        app: app
        app.kubernetes.io/instance: my-sample-java-springboot
        app.kubernetes.io/managed-by: odo
        app.kubernetes.io/managed-by-version: v3.0.0-rc1
        app.kubernetes.io/part-of: app
        app.openshift.io/runtime: spring
        component: my-sample-java-springboot
        odo.dev/mode: Dev
      name: my-sample-java-springboot-app
      namespace: default
    spec:
      containers:
      - args:
        - -f
        - /dev/null
        command:
        - tail
        env:
        - name: DEBUG_PORT
          value: "5858"
        - name: PROJECTS_ROOT
          value: /projects
        - name: PROJECT_SOURCE
          value: /projects
        image: quay.io/eclipse/che-java11-maven:next
        imagePullPolicy: Always
        name: tools
        ports:
        - containerPort: 8080
          name: http-springboot
          protocol: TCP
        - containerPort: 5858
          name: debug
          protocol: TCP
        resources:
          limits:
            memory: 768Mi
        volumeMounts:
        - mountPath: /projects
          name: odo-projects
        - mountPath: /opt/odo/
          name: odo-shared-data
        - mountPath: /home/user/.m2
          name: m2-my-sample-java-springboot-app-vol
      - args:
        - -f
        - /dev/null
        command:
        - tail
        env:
        - name: MY_ENV_VAR
          value: some value
        image: alpine:latest
        imagePullPolicy: Always
        name: echo-container
        ports:    
        - containerPort: 18080
          name: echo-ep1
          protocol: TCP
        resources: {}
        volumeMounts:
        - mountPath: /opt/odo/
          name: odo-shared-data
      restartPolicy: Always
      securityContext: {}
      volumes:
      - name: m2-my-sample-java-springboot-app-vol
        persistentVolumeClaim:
          claimName: m2-my-sample-java-springboot-app
      - name: odo-projects
        persistentVolumeClaim:
          claimName: odo-projects-my-sample-java-springboot-app
      - emptyDir: {}
        name: odo-shared-data

```

* Service:

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    example.com/my-svc-annotation-echo1: my-svc-annotation-val1
    example.com/my-svc-annotation1: my-svc-annotation-val1
    service.binding/backend_ip: path={.spec.clusterIP}
    service.binding/backend_port: path={.spec.ports},elementType=sliceOfMaps,sourceKey=name,sourceValue=port
  labels:
    app: app
    app.kubernetes.io/instance: my-sample-java-springboot
    app.kubernetes.io/managed-by: odo
    app.kubernetes.io/managed-by-version: v3.0.0-rc1
    app.kubernetes.io/part-of: app
    app.openshift.io/runtime: spring
    component: my-sample-java-springboot
    odo.dev/mode: Dev
  namespace: default
  name: my-sample-java-springboot-app
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    kind: Deployment
    name: my-sample-java-springboot-app
spec:
  ports:
  - name: http-springboot
    port: 8080
    protocol: TCP
    targetPort: 8080
  - name: echo-ep1
    port: 18080
    protocol: TCP
    targetPort: 18080
  selector:
    component: my-sample-java-springboot
  sessionAffinity: None
  type: ClusterIP
```

* PersistentVolumeClaim for the project source code (because of `Ephemeral` Setting set to `false` (default)):

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  finalizers:
  - kubernetes.io/pvc-protection
  labels:
    app: app
    app.kubernetes.io/instance: my-sample-java-springboot
    app.kubernetes.io/managed-by: odo
    app.kubernetes.io/managed-by-version: v3.0.0-rc1
    app.kubernetes.io/part-of: app
    app.kubernetes.io/storage-name: odo-projects
    app.openshift.io/runtime: spring
    component: my-sample-java-springboot
    odo-source-pvc: odo-projects
    odo.dev/mode: Dev
    storage-name: odo-projects
  name: odo-projects-my-sample-java-springboot-app
  namespace: default
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    kind: Deployment
    name: my-sample-java-springboot-app
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 2Gi
```

* PersistentVolumeClaim for the `volume` component:
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  finalizers:
  - kubernetes.io/pvc-protection
  labels:
    app: app
    app.kubernetes.io/instance: my-sample-java-springboot
    app.kubernetes.io/managed-by: odo
    app.kubernetes.io/managed-by-version: v3.0.0-rc1
    app.kubernetes.io/part-of: app
    app.kubernetes.io/storage-name: m2
    app.openshift.io/runtime: spring
    component: my-sample-java-springboot
    odo.dev/mode: Dev
    storage-name: m2
  name: m2-my-sample-java-springboot-app
  namespace: default
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    kind: Deployment
    name: my-sample-java-springboot-app
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 3Gi
```

</details>

