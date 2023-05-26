---
title: Troubleshoot Storage Permission issues on managed cloud providers clusters
sidebar_position: 9
---

Using `odo` to run an application on a managed [Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine), [Azure Kubernetes Service (AKS)](https://azure.microsoft.com/en-us/products/kubernetes-service), or [Amazon Elatic Kubernetes Service (EKS)](https://aws.amazon.com/eks/) cluster does not always work out of the box, especially while using Devfiles from the [Devfile Registry](https://registry.devfile.io); users often encounter issues while syncing local files into the container due to insufficient permissions on mounted volumes.

<details>
<summary>For example, while running a Java Maven application using a <code>java-maven</code> devfile on an Amazon Elastic Kubernetes Service, a sample error may look like this.</summary>

```shell
$ odo dev
  __
 /  \__     Developing using the "java-springboot-starter" Devfile
 \__/  \    Namespace: default
 /  \__/    odo version: v3.6.0
 \__/

↪ Running on the cluster in Dev mode
 •  Waiting for Kubernetes resources  ...
 ✓  Added storage m2 to component
 ⚠  Pod is Pending
 ✓  Pod is Running
 ◑  Syncing files into the container ✗  Command 'tar xf - -C /projects --no-same-owner' in container failed.

 ✗  stdout:

 ✗  stderr: tar: src: Cannot mkdir: Permission denied
tar: src/main/resources/application.properties: Cannot open: No such file or directory
tar: HELP.md: Cannot open: Permission denied
tar: mvnw: Cannot open: Permission denied
tar: devfile.yaml: Cannot open: Permission denied
tar: mvnw.cmd: Cannot open: Permission denied
tar: pom.xml: Cannot open: Permission denied
tar: src: Cannot mkdir: Permission denied
tar: src/main/java/com/example/demo/DemoApplication.java: Cannot open: No such file or directory
tar: .gitignore: Cannot open: Permission denied
tar: src: Cannot mkdir: Permission denied
tar: src/test/java/com/example/demo/DemoApplicationTests.java: Cannot open: No such file or directory
tar: Exiting with failure status due to previous errors


 ✗  err: error while streaming command: command terminated with exit code 2

 ✗  Syncing files into the container [610ms]
Error occurred on Push - watch command was unable to push component: failed to sync to component with name java-springboot-starter: failed to sync to component with name java-springboot-starter: unable push files to pod: error while streaming command: command terminated with exit code 2

 ◐  Syncing files into the container ✗  Command 'tar xf - -C /projects --no-same-owner' in container failed.

 ✗  stdout:

 ✗  stderr: tar: src: Cannot mkdir: Permission denied
tar: src/main/resources/application.properties: Cannot open: No such file or directory
tar: src: Cannot mkdir: Permission denied
tar: src/test/java/com/example/demo/DemoApplicationTests.java: Cannot open: No such file or directory
tar: devfile.yaml: Cannot open: Permission denied
tar: src: Cannot mkdir: Permission denied
tar: src/main/java/com/example/demo/DemoApplication.java: Cannot open: No such file or directory
tar: pom.xml: Cannot open: Permission denied
tar: .gitignore: Cannot open: Permission denied
tar: mvnw.cmd: Cannot open: Permission denied
tar: HELP.md: Cannot open: Permission denied
tar: mvnw: Cannot open: Permission denied
tar: Exiting with failure status due to previous errors


 ✗  err: error while streaming command: command terminated with exit code 2
```

</details>

<details>
<summary>Or, while running a Go application with <code>go</code> devfile on an Azure Kubernetes Service may end up in an error like this.</summary>

```shell
$ odo dev
  __
 /  \__     Developing using the "places" Devfile
 \__/  \    Namespace: default
 /  \__/    odo version: v3.10.0
 \__/

 ⚠  You are using "default" namespace, odo may not work as expected in the default namespace.
 ⚠  You may set a new namespace by running `odo create namespace <name>`, or set an existing one by running `odo set namespace <name>`

↪ Running on the cluster in Dev mode
 •  Waiting for Kubernetes resources  ...
 ⚠  Pod is Pending
 ✓  Pod is Running
 ◐  Syncing files into the container ✗  Command 'tar xf - -C /projects --no-same-owner' in container failed.

 ✗  stdout:

 ✗  stderr: tar: main.go: Cannot open: Permission denied
tar: .gitignore: Cannot open: Permission denied
tar: README.md: Cannot open: Permission denied
tar: devfile.yaml: Cannot open: Permission denied
tar: go.mod: Cannot open: Permission denied
tar: Exiting with failure status due to previous errors


 ✗  err: error while streaming command: command terminated with exit code 2

 ✗  Syncing files into the container [4s]
Error occurred on Push - watch command was unable to push component: failed to sync to component with name places: failed to sync to component with name places: unable push files to pod: error while streaming command: command terminated with exit code 2


↪ Dev mode
 Status:
 Watching for changes in the current directory /tmp/go-app

 Keyboard Commands:
[Ctrl+c] - Exit and delete resources from the cluster
     [p] - Manually apply local changes to the application on the cluster
^CCleaning resources, please wait
 ✗  Dev mode interrupted by user
```

</details>

Various factors are responsible for this:
* Storage Provisioner used for the cluster
* User set by the container image
* Location on the container where the files are to be synced
* Using Ephemeral vs Non-Ephemeral Volumes

Users may encounter storage related permissions issues even while working on a standard Kubernetes or OpenShift cluster.

This guide will discuss some workarounds that can be used to fix these issues.

### Using Ephemeral Volumes
This is the simplest way to overcome this issue. There are 2 parts to this solution:
1. Set `odo` preference `Ephemeral` to _true_.

   ```shell
    odo preference set Ephemeral true -f
    ```
2. If the Devfile contains a `volume` component, then set its `ephemeral` property to `true`.

   The above configuration will use the [`emptyDir` Ephemeral volumes](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) instead of creating Persistent Volumes to mount the source files; it also ensures the current user can read/write to the directories.

### Setting `fsGroup` to the PodSecurityContext
By setting `fsGroup` in the PodSecurityContext, all processes of the container are also made part of the supplementary group ID set in the field. The owner for volume mount location and any files created in that volume will be Group ID set in the field. This solution is quite common when looking for permission related issues on a mounted volume, [example](https://stackoverflow.com/questions/50156124/kubernetes-nfs-persistent-volumes-permission-denied#50187723).

This solution can be implemented by setting a [`pod-overrides`](https://devfile.io/docs/2.2.0/overriding-pod-and-container-attributes#pod-overrides) attribute to the Devfile `container` component.

<details>
<summary>Example <code>java-maven</code> Devfile with a <code>fsGroup</code> set in PodSecurityContext.</summary>

```yaml showLineNumbers
commands:
- exec:
    commandLine: mvn -Dmaven.repo.local=/home/user/.m2/repository package
    component: tools
    group:
      isDefault: true
      kind: build
    workingDir: ${PROJECT_SOURCE}
  id: mvn-package
- exec:
    commandLine: java -jar target/*.jar
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
    command:
    - tail
    - -f
    - /dev/null
    endpoints:
    - name: http-maven
      targetPort: 8080
    - exposure: none
      name: debug
      targetPort: 5858
    env:
    - name: DEBUG_PORT
      value: "5858"
    image: registry.access.redhat.com/ubi8/openjdk-11:latest
    memoryLimit: 512Mi
    mountSources: true
    volumeMounts:
    - name: m2
      path: /home/user/.m2
  name: tools
#  highlight-start
  attributes:
    pod-overrides:
      spec:
        securityContext:
          fsGroup: 2000
#  highlight-end
- name: m2
  volume: {}
metadata:
  description: Java application based on Maven 3.6 and OpenJDK 11
  displayName: Maven Java
  icon: https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/java-maven.jpg
  language: Java
  name: jmaven-app
  projectType: Maven
  tags:
  - Java
  - Maven
  version: 1.2.0
schemaVersion: 2.1.0
starterProjects:
- git:
    remotes:
      origin: https://github.com/odo-devfiles/springboot-ex.git
  name: springbootproject
```
</details>


But this solution may not always be feasible, especially while dealing with large filesystems.
>Be cautious with the use of fsGroup. The changing of group ownership of an entire volume can cause pod startup delays for slow and/or large filesystems. It can also be detrimental to other processes that share the same volume if their processes do not have access permissions to the new GID. For this reason, some providers for shared file systems such as NFS do not implement this functionality. These settings also do not affect ephemeral volumes.
>
> Read these articles by [Synk](https://snyk.io/blog/10-kubernetes-security-context-settings-you-should-understand/) and [Google Cloud](https://cloud.google.com/kubernetes-engine/docs/troubleshooting/troubleshooting-gke-storage#mounting_a_volume_stops_responding_due_to_the_fsgroup_setting) to learn more about it.
