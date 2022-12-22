---
title: odo init
---

The `odo init` command is the first command to be executed when you want to bootstrap a new component, using `odo`. If sources already exist,
the command `odo dev` should be considered instead.

This command must be executed from a directory with no `devfile.yaml` file.

The command can be executed in two flavors, either interactive or non-interactive.

## Running the command
### Interactive mode

In interactive mode, the behavior of `odo init` depends on whether the current directory already contains source code or not.

#### Empty directory

If the directory is empty, you will be guided to:
- choose a devfile from the list of devfiles present in the registry or registries referenced (using the `odo registry` command),
- configure the devfile
- choose a starter project referenced by the selected devfile,
- choose a name for the component present in the devfile; this name must follow the [Kubernetes naming convention](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names) and not be all-numeric.

```console
odo init
```

<details>
<summary>Example</summary>

import EmptyDirOutput from './docs-mdx/init/empty_directory_output.mdx';

<EmptyDirOutput />

</details>

#### Directory with sources

If the current directory is not empty, `odo init` will make its best to autodetect the type of application and propose you a Devfile that should suit your project.
It will try to detect the following, based on the files in the current directory:
- Language
- Project Type
- Ports used in your application
- A Devfile that should help you start with `odo` 

If the information detected does not seem correct to you, you are able to select a different Devfile.

In all cases, you will be guided to:
- configure the devfile
- choose a name for the component present in the devfile; this name must follow the [Kubernetes naming convention](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names) and not be all-numeric.

```console
odo init
```

<details>
<summary>Example</summary>

```console
$ odo init                                                                                                                                                                                                          
  __                                                                                                                                                                                                                
 /  \__     Initializing a new component                                                                                                                                                                            
 \__/  \    Files: Source code detected, a Devfile will be determined based upon source code autodetection                                                                                                          
 /  \__/    odo version: v3.3.0                                                                                                                                                                                     
 \__/                                                                                                                                                                                                               
                                                                                                                                                                                                                    
Interactive mode enabled, please answer the following questions:                                                                                                                                                    
Based on the files in the current directory odo detected                                                                                                                                                            
Language: JavaScript                                                                                                                                                                                                
Project type: Node.js                                                                                                                                                                                               
Application ports: 3000                                                                                                                                                                                             
The devfile "nodejs" from the registry "DefaultDevfileRegistry" will be downloaded.                                                                                                                                 
? Is this correct? Yes                                                                                                                                                                                              
 ✓  Downloading devfile "nodejs" from registry "DefaultDevfileRegistry" [1s]                                                                                                                                        
                                                                                                                                                                                                                    
↪ Container Configuration "runtime":                                                                                                                                                                                
  OPEN PORTS:                                                                                                                                                                                                       
    - 5858                                                                                                                                                                                                          
    - 3000                                                                                                                                                                                                          
  ENVIRONMENT VARIABLES:                                                                                                                                                                                            
    - DEBUG_PORT = 5858                                                                                                                                                                                             
                                                                                                                                                                                                                    
? Select container for which you want to change configuration? NONE - configuration is correct                                                                                                                      
? Enter component name: nodejs                                                                                                                                                                                      
                                                                                                                                                                                                                    
You can automate this command by executing:
   odo init --name nodejs --devfile nodejs --devfile-registry DefaultDevfileRegistry

Your new component 'nodejs' is ready in the current directory.
To start editing your component, use 'odo dev' and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
```
</details>

### Non-interactive mode

In non-interactive mode, you will have to specify from the command-line the information needed to get a devfile.

If you want to download a devfile from a registry, you must specify the devfile name with the `--devfile` flag. The devfile with the specified name will be searched in the registries referenced (using `odo preference view`), and the first one matching will be downloaded.
If you want to download the devfile from a specific registry in the list or referenced registries, you can use the `--devfile-registry` flag to specify the name of this registry. By default odo uses official devfile registry [registry.devfile.io](https://registry.devfile.io). You can use registry's [web interface](https://registry.devfile.io/viewer) to view its content.
If you want to download a version devfile, you must specify the version with `--devfile-version` flag.

If you prefer to download a devfile from an URL or from the local filesystem, you can use the `--devfile-path` instead.

The `--starter` flag indicates the name of the starter project (as referenced in the selected devfile), that you want to use to start your development. To see the available starter projects for devfile stacks in the official devfile registry use its [web interface](https://registry.devfile.io/viewer) to view its content.  

The required `--name` flag indicates how the component initialized by this command should be named. The name must follow the [Kubernetes naming convention](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names) and not be all-numeric.

#### Fetch Devfile from any registry of the list

In this example, the devfile will be downloaded from the **Staging** registry, which is the first one in the list containing the `nodejs-react` devfile.
```shell
odo init --name <component-name> --devfile <devfile> [--starter STARTER]
```
<details>
<summary>Example</summary>

```console
$ odo preference view
[...]

Devfile registries:
 NAME                       URL                                   SECURE
 Staging                    https://registry.stage.devfile.io     No
 DefaultDevfileRegistry     https://registry.devfile.io           No

$  odo registry --devfile nodejs-react
 NAME          REGISTRY                DESCRIPTION                                  VERSIONS 
 nodejs-react  StagingRegistry         React is a free and open-source front-en...  2.0.2    
 nodejs-react  DefaultDevfileRegistry  React is a free and open-source front-en...  2.0.2   

$ odo init --devfile nodejs-react --name my-nr-app 
  __
 /  \__     Initializing a new component
 \__/  \    
 /  \__/    odo version: v3.4.0
 \__/

 ✓  Downloading devfile "nodejs-react" [3s]

Your new component 'my-nr-app' is ready in the current directory.
To start editing your component, use 'odo dev' and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.

```
</details>


#### Fetch Devfile from a specific registry of the list

In this example, the devfile will be downloaded from the **DefaultDevfileRegistry** registry, as explicitly indicated by the `--devfile-registry` flag.
<details>
<summary>Example</summary>

```console
$ odo preference view
[...]

Devfile registries:
 NAME                       URL                                   SECURE
 Staging                    https://registry.stage.devfile.io     No
 DefaultDevfileRegistry     https://registry.devfile.io           No

$ odo init --name my-spring-app --devfile java-springboot --devfile-registry DefaultDevfileRegistry --starter springbootproject
 ✓  Downloading devfile "java-springboot" from registry "DefaultDevfileRegistry" [980ms]
 ✓  Downloading starter project "springbootproject" [399ms]

Your new component "my-spring-app" is ready in the current directory.
To start editing your component, use "odo dev" and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
To deploy your component to a cluster use "odo deploy".
```
</details>


#### Fetch Devfile from a URL

```console
odo init --devfile-path <URL> --name <component-name> [--starter STARTER]
```
<details>
<summary>Example</summary>

```console
$ odo init --devfile-path https://registry.devfile.io/devfiles/nodejs-angular --name my-nodejs-app --starter nodejs-angular-starter
 ✓  Downloading devfile from "https://registry.devfile.io/devfiles/nodejs-angular" [415ms]
 ✓  Downloading starter project "nodejs-angular-starter" [484ms]

Your new component "my-nodejs-app" is ready in the current directory.
To start editing your component, use "odo dev" and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
To deploy your component to a cluster use "odo deploy".
```
</details>

#### Fetch Devfile of a specific version

```console
odo init --devfile <devfile-name> --devfile-version <devfile-version> --name <component-name> [--starter STARTER]
```

<details>
<summary>Examples</summary>

import VersionedOutput from './docs-mdx/init/versioned_devfile_output.mdx';

<VersionedOutput />

</details>

:::note
Use "latest" as the version name to fetch the latest version of a given Devfile.

<details>
<summary>Example</summary>

```console
$ odo init --devfile go --name my-go-app  --devfile-version latest
  __
 /  \__     Initializing a new component
 \__/  \    
 /  \__/    odo version: v3.4.0
 \__/

 ✓  Downloading devfile "go:latest" [4s]

Your new component 'my-go-app' is ready in the current directory.
To start editing your component, use 'odo dev' and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
To deploy your component to a cluster use "odo deploy".
```
</details>
:::