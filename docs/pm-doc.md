## Introduction

### What is odo?

`odo` is a fast, iterative and straightforward CLI tool for developers who write, build, and deploy applications on Kubernetes.

We abstract the complex concepts of Kubernetes so you can focus on one thing: `code`.

Choose your favourite framework and `odo` will deploy it *fast* and *often* to your container orchestrator cluster.

`odo` is focused on [inner loop](./introduction#what-is-inner-loop-and-outer-loop) development as well as tooling that would help users transition to the [outer loop](./introduction#what-is-inner-loop-and-outer-loop).

Brendan Burns, one of the co-founders of Kubernetes, said in the [book Kubernetes Patterns](https://www.redhat.com/cms/managed-files/cm-oreilly-kubernetes-patterns-ebook-f19824-201910-en.pdf):

> It (Kubernetes) is the foundation on which applications will be built, and it provides a large library of APIs and tools for building these applications, but it does little to provide the application or container developer with any hints or guidance for how these various pieces can be combined into a complete, reliable system that satisfies their business needs and goals.

`odo` satisfies that need by making Kubernetes development *super easy* for application developers and cloud engineers.

### What is "inner loop" and "outer loop"?

The **inner loop** consists of local coding, building, running, and testing the application -- all activities that you, as a developer, can control. 

The **outer loop** consists of the larger team processes that your code flows through on its way to the cluster: code reviews, integration tests, security and compliance, and so on. 

The inner loop could happen mostly on your laptop. The outer loop happens on shared servers and runs in containers, and is often automated with continuous integration/continuous delivery (CI/CD) pipelines. 

Usually, a code commit to source control is the transition point between the inner and outer loops.

*([Source](https://developers.redhat.com/blog/2020/06/16/enterprise-kubernetes-development-with-odo-the-cli-tool-for-developers#improving_the_developer_workflow))*

### Why should I use `odo`?

You should use `odo` if:
* You love frameworks such as Node.js, Spring Boot or dotNet
* Your application is intended to run in a Kubernetes-like infrastructure
* You don't want to spend time fighting with DevOps and learning Kubernetes in order to deploy to your enterprise infrastructure

If you are an application developer wishing to deploy to Kubernetes easily, then `odo` is for you.

### How is odo different from `kubectl` and `oc`?

Both [`kubectl`](https://github.com/kubernetes/kubectl) and [`oc`](https://github.com/openshift/oc/) require deep understanding of Kubernetes and OpenShift concepts.

`odo` is different as it focuses on application developers and cloud engineers. Both `kubectl` and `oc` are DevOps oriented tools and help in deploying applications to and maintaining a Kubernetes cluster provided you know Kubernetes well.

`odo` is not meant to:
* Maintain a production Kubernetes cluster
* Perform sysadmin tasks against a Kubernetes cluster

## Installation

`odo` can be used as either a [CLI tool](#cli-installation) or an [IDE plugin](#ide-installation) on [Mac](#macos), [Windows](#windows) or [Linux](#linux).

Each release is *signed*, *checksummed*, *verified*, and then pushed to our [binary mirror](https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/).

For more information on the changes of each release, they can be viewed either on [GitHub](https://github.com/redhat-developer/odo/releases) or the [blog](/blog).

### CLI Installation

#### Linux

Installing `odo` on `amd64` architecture:

1. Download the latest release from the mirror:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.14.0/odo-linux-amd64 -o odo
```

2. (Optional) Verify the downloaded binary with the SHA-256 sum:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.14.0/odo-linux-amd64.sha256 -o odo.sha256
echo "$(<odo.sha256)  odo" | shasum -a 256 --check
```

3. Install odo:
```shell
sudo install -o root -g root -m 0755 odo /usr/local/bin/odo
```

4. (Optional) If you do not have root access, you can install `odo` to the local directory and add it to your `$PATH`:

```shell
mkdir -p $HOME/bin 
cp ./odo $HOME/bin/odo
export PATH=$PATH:$HOME/bin
# (Optional) Add the $HOME/bin to your shell initialization file
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc
```

Installing `odo` on `arm64` architecture:

1. Download the latest release from the mirror:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.14.0/odo-linux-arm64 -o odo
```

2. (Optional) Verify the downloaded binary with the SHA-256 sum:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.14.0/odo-linux-arm64.sha256 -o odo.sha256
echo "$(<odo.sha256)  odo" | shasum -a 256 --check
```

3. Install odo:
```shell
sudo install -o root -g root -m 0755 odo /usr/local/bin/odo
```

4. (Optional) If you do not have root access, you can install `odo` to the local directory and add it to your `$PATH`:

```shell
mkdir -p $HOME/bin 
cp ./odo $HOME/bin/odo
export PATH=$PATH:$HOME/bin
# (Optional) Add the $HOME/bin to your shell initialization file
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc
```

Installing `odo` on `ppc64le` architecture:

1. Download the latest release from the mirror:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.14.0/odo-linux-ppc64le -o odo
```

2. (Optional) Verify the downloaded binary with the SHA-256 sum:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.14.0/odo-linux-ppc64le.sha256 -o odo.sha256
echo "$(<odo.sha256)  odo" | shasum -a 256 --check
```

3. Install odo:
```shell
sudo install -o root -g root -m 0755 odo /usr/local/bin/odo
```

4. (Optional) If you do not have root access, you can install `odo` to the local directory and add it to your `$PATH`:

```shell
mkdir -p $HOME/bin 
cp ./odo $HOME/bin/odo
export PATH=$PATH:$HOME/bin
# (Optional) Add the $HOME/bin to your shell initialization file
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc
```

Installing `odo` on `s390x` architecture:

1. Download the latest release from the mirror:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.14.0/odo-linux-s390x -o odo
```

2. (Optional) Verify the downloaded binary with the SHA-256 sum:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.14.0/odo-linux-s390x.sha256 -o odo.sha256
echo "$(<odo.sha256)  odo" | shasum -a 256 --check
```

3. Install odo:
```shell
sudo install -o root -g root -m 0755 odo /usr/local/bin/odo
```

4. (Optional) If you do not have root access, you can install `odo` to the local directory and add it to your `$PATH`:

```shell
mkdir -p $HOME/bin 
cp ./odo $HOME/bin/odo
export PATH=$PATH:$HOME/bin
# (Optional) Add the $HOME/bin to your shell initialization file
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc
```

---

#### MacOS

##### Homebrew

**NOTE:** This will install from the *main* branch on GitHub

Installing `odo` using [Homebrew](https://brew.sh/):

1. Install odo:

```shell
brew install odo-dev
```

2. Verify the version you installed is up-to-date:

```shell
odo version
```

##### Binary

Installing `odo` on `amd64` architecture:

1. Download the latest release from the mirror:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.14.0/odo-darwin-amd64 -o odo
```

2. (Optional) Verify the downloaded binary with the SHA-256 sum:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.14.0/odo-darwin-amd64.sha256 -o odo.sha256
echo "$(<odo.sha256)  odo" | shasum -a 256 --check
```

3. Install odo:
```shell
chmod +x ./odo
sudo mv ./odo /usr/local/bin/odo
```

4. (Optional) If you do not have root access, you can install `odo` to the local directory and add it to your `$PATH`:

```shell
mkdir -p $HOME/bin 
cp ./odo $HOME/bin/odo
export PATH=$PATH:$HOME/bin
# (Optional) Add the $HOME/bin to your shell initialization file
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc
```

Installing `odo` on `arm64` architecture:

1. Download the latest release from the mirror:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.14.0/odo-darwin-arm64 -o odo
```

2. (Optional) Verify the downloaded binary with the SHA-256 sum:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.14.0/odo-darwin-arm64.sha256 -o odo.sha256
echo "$(<odo.sha256)  odo" | shasum -a 256 --check
```

3. Install odo:
```shell
chmod +x ./odo
sudo mv ./odo /usr/local/bin/odo
```

4. (Optional) If you do not have root access, you can install `odo` to the local directory and add it to your `$PATH`:

```shell
mkdir -p $HOME/bin 
cp ./odo $HOME/bin/odo
export PATH=$PATH:$HOME/bin
# (Optional) Add the $HOME/bin to your shell initialization file
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc
```

---

#### Windows

1. Open a PowerShell terminal

2. Download the latest release from the mirror:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.14.0/odo-windows-amd64.exe -o odo.exe
```

2. (Optional) Verify the downloaded binary with the SHA-256 sum:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.14.0/odo-windows-amd64.exe.sha256 -o odo.exe.sha256
# Visually compare the output of both files
Get-FileHash odo.exe
type odo.exe.sha256
```

4. Add the binary to your `PATH`

### IDE Installation

#### Visual Studio Code (VSCode)

The [OpenShift Toolkit](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-openshift-connector) VSCode extension uses both `odo` and `oc` binaries
to interact with Kubernetes or OpenShift cluster.

To install it:

1. Open VS Code.
2. Launch VS Code **Quick Open** (`Ctrl+P`).
3. Paste the following command and press `Enter`:

```
ext install redhat.vscode-openshift-connector
```

#### JetBrains IDEs
The [OpenShift Toolkit by Red Hat](https://plugins.jetbrains.com/plugin/12030-openshift-toolkit-by-red-hat/) plugin can be installed
to interact with OpenShift or Kubernetes clusters right from your JetBrains IDEs like IntelliJ IDEA, WebStorm or PyCharm.
It uses `odo` and `oc` binaries for fast iterative application development on those clusters.

To install it:

1. Press `Ctrl+Alt+S` to open the IDE settings and select **Plugins**.
2. Find the "**OpenShift Toolkit by Red Hat**" plugin in the **Marketplace** and click **Install**.

### Alternative installation methods

##### Source code
1. Clone the repository and cd into it.
   ```shell
   git clone https://github.com/redhat-developer/odo.git
   cd odo
   ```
2. Install tools used by the build and test system.
   ```shell
   make goget-tools
   ```
3. Build the executable from the sources in `cmd/odo`.
   ```shell
   make bin
   ```
4. Check the build version to verify that it was built properly.
   ```shell
   ./odo version
   ```
5. Install the executable in the system's GOPATH.
   ```shell
   make install
   ```
6. Check the binary version to verify that it was installed properly; verify that it is same as the build version.
   ```shell
   odo version
   ```

#### Maven plugin
It is possible to integrate the `odo` binary download in a Maven project using [odo Downloader Plugin](https://github.com/tnb-software/odo-downloader).
The download can be executed using the `download` goal which automatically retrieves the version for the current architecture:
```shell
mvn software.tnb:odo-downloader-maven-plugin:0.1.3:download \
  -Dodo.target.file=$HOME/bin/odo \
  -Dodo.version=v3.14.0
```

#### asdf
The [asdf version manager](https://asdf-vm.com/) is a tool for managing multiple runtime versions using a common CLI.
With `asdf` installed, the [asdf plugin for odo](https://github.com/rm3l/asdf-odo) can be used to install any released version of `odo`:
```
asdf plugin add odo
asdf install odo 3.14.0
asdf global odo 3.14.0
```

### Nightly builds

Nightly builds of `odo` are also available. Note that these builds are provided as is and can be highly unstable.

#### Linux

Installing `odo` on `amd64` architecture:

1. Download the latest nightly build:
```shell
curl -L https://s3.us-south.cloud-object-storage.appdomain.cloud/odo-nightly/odo-linux-amd64 -o odo
```

2. Install odo:
```shell
sudo install -o root -g root -m 0755 odo /usr/local/bin/odo
```

3. (Optional) If you do not have root access, you can install `odo` to the local directory and add it to your `$PATH`:

```shell
mkdir -p $HOME/bin 
cp ./odo $HOME/bin/odo
export PATH=$PATH:$HOME/bin
# (Optional) Add the $HOME/bin to your shell initialization file
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc
```

Installing `odo` on `arm64` architecture:

1. Download the latest nightly build:
```shell
curl -L https://s3.us-south.cloud-object-storage.appdomain.cloud/odo-nightly/odo-linux-arm64 -o odo
```

2. Install odo:
```shell
sudo install -o root -g root -m 0755 odo /usr/local/bin/odo
```

3(Optional) If you do not have root access, you can install `odo` to the local directory and add it to your `$PATH`:

```shell
mkdir -p $HOME/bin 
cp ./odo $HOME/bin/odo
export PATH=$PATH:$HOME/bin
# (Optional) Add the $HOME/bin to your shell initialization file
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc
```

Installing `odo` on `ppc64le` architecture:

1. Download the latest nightly build:
```shell
curl -L https://s3.us-south.cloud-object-storage.appdomain.cloud/odo-nightly/odo-linux-ppc64le -o odo
```

2. Install odo:
```shell
sudo install -o root -g root -m 0755 odo /usr/local/bin/odo
```

3(Optional) If you do not have root access, you can install `odo` to the local directory and add it to your `$PATH`:

```shell
mkdir -p $HOME/bin 
cp ./odo $HOME/bin/odo
export PATH=$PATH:$HOME/bin
# (Optional) Add the $HOME/bin to your shell initialization file
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc
```

Installing `odo` on `s390x` architecture:

1. Download the latest nightly build:
```shell
curl -L https://s3.us-south.cloud-object-storage.appdomain.cloud/odo-nightly/odo-linux-s390x -o odo
```

2. Install odo:
```shell
sudo install -o root -g root -m 0755 odo /usr/local/bin/odo
```

3. (Optional) If you do not have root access, you can install `odo` to the local directory and add it to your `$PATH`:

```shell
mkdir -p $HOME/bin 
cp ./odo $HOME/bin/odo
export PATH=$PATH:$HOME/bin
# (Optional) Add the $HOME/bin to your shell initialization file
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc
```

---

#### MacOS

Installing `odo` on `amd64` architecture:

1. Download the latest nightly build:
```shell
curl -L https://s3.us-south.cloud-object-storage.appdomain.cloud/odo-nightly/odo-darwin-amd64 -o odo
```

2. Install odo:
```shell
chmod +x ./odo
sudo mv ./odo /usr/local/bin/odo
```

3(Optional) If you do not have root access, you can install `odo` to the local directory and add it to your `$PATH`:

```shell
mkdir -p $HOME/bin 
cp ./odo $HOME/bin/odo
export PATH=$PATH:$HOME/bin
# (Optional) Add the $HOME/bin to your shell initialization file
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc
```

Installing `odo` on `arm64` architecture:

1. Download the latest nightly build:
```shell
curl -L https://s3.us-south.cloud-object-storage.appdomain.cloud/odo-nightly/odo-darwin-arm64 -o odo
```

2. Install odo:
```shell
chmod +x ./odo
sudo mv ./odo /usr/local/bin/odo
```

3. (Optional) If you do not have root access, you can install `odo` to the local directory and add it to your `$PATH`:

```shell
mkdir -p $HOME/bin 
cp ./odo $HOME/bin/odo
export PATH=$PATH:$HOME/bin
# (Optional) Add the $HOME/bin to your shell initialization file
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc
```

---

#### Windows

1. Open a PowerShell terminal

2. Download the latest nightly build:
```shell
curl -L https://s3.us-south.cloud-object-storage.appdomain.cloud/odo-nightly/odo-windows-amd64.exe -o odo.exe
```

3. Add the binary to your `PATH`

## Developing with Java (Spring Boot)

### Step 0. Creating the initial source code (optional)

We will create the example source code by using some popular frameworks.

Before we begin, we will create a new directory and cd into it.
```shell
mkdir quickstart-demo && cd quickstart-demo
```

This is *optional* and you may use an existing project instead (make sure you cd into the project directory before running any odo commands) or a starter project from `odo init`.

For Java, we will use the [Spring Initializr](https://start.spring.io/) to generate the example source code:

1. Navigate to [start.spring.io](https://start.spring.io/).
2. Select **Maven** under **Project**.
3. Click on "Add" under "Dependencies".
4. Select "Spring Web".
5. Click "Generate" to generate and download the source code.

Finally, extract the downloaded source code archive in the 'quickstart-demo' directory.

Your source code has now been generated and created in the directory.

### Step 1. Preparing the target platform

Before starting, you should make sure that odo is connected to your cluster and that you have created a new project.

#### Login to OpenShift Cluster

The easiest way to connect `odo` to an OpenShift cluster is use copy "Copy login command" function in OpenShift Web Console.

1. Login to OpenShift Web Console.
2. At the top right corner click on your username and then on "Copy login command".
3. You will be prompted to enter your login credentials again.
4. After login, open "Display Token" link.
5. Copy whole `oc login --token ...` command and paste it into the terminal, **before executing the command replace `oc` with `odo`.**

#### Create a new project

If you are using OpenShift, you can create a new namespace with the `odo create project` command.

```console
odo create project odo-dev
```

Sample output:
```console
$ odo create project odo-dev
 ✓  Creating the project "odo-dev" [1s]
 ✓  Project "odo-dev" is ready for use
 ✓  New project created and now using project: odo-dev
```

### Step 2. Initializing your application (`odo init`)

Now we'll initialize your application by creating a `devfile.yaml` to be deployed.

`odo` handles this automatically with the `odo init` command by autodetecting your source code and downloading the appropriate Devfile.

**Note:** If you skipped *Step 0*, select a "starter project" when running `odo init`.

<p>Let's run <code>odo init</code> and select <span>Java (Spring Boot)</span>:</p>

```console
odo init
```

Sample Output:

```console
$ odo init
  __
 /  \__     Initializing a new component
 \__/  \    Files: Source code detected, a Devfile will be determined based upon source code autodetection
 /  \__/    odo version: v3.14.0
 \__/

Interactive mode enabled, please answer the following questions:
 ✓  Determining a Devfile for the current directory [1s]
Based on the files in the current directory odo detected
Supported architectures: all
Language: Java
Project type: springboot
The devfile "java-springboot:1.2.0" from the registry "DefaultDevfileRegistry" will be downloaded.
? Is this correct? Yes
 ✓  Downloading devfile "java-springboot:1.2.0" from registry "DefaultDevfileRegistry" [3s]

↪ Container Configuration "tools":
  OPEN PORTS:
    - 8080
    - 5858
  ENVIRONMENT VARIABLES:
    - DEBUG_PORT = 5858

? Select container for which you want to change configuration? NONE - configuration is correct
? Enter component name: my-java-app

You can automate this command by executing:
   odo init --name my-java-app --devfile java-springboot --devfile-registry DefaultDevfileRegistry --devfile-version 1.2.0

Your new component 'my-java-app' is ready in the current directory.
To start editing your component, use 'odo dev' and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
```

> If you skipped Step 0 and selected "starter project", your output will be slightly different.

### Step 3. Developing your application continuously (`odo dev`)

Now that we've generated our code as well as our Devfile, let's start on development.

`odo` uses [inner loop development](/docs/introduction#what-is-inner-loop-and-outer-loop) and allows you to code,
build, run and test the application in a continuous workflow.

Once you run `odo dev`, you can freely edit code in your favourite IDE and watch as `odo` rebuilds and redeploys it.

<p>Let's run <code>odo dev</code> to start development on your <span>Java (Spring Boot)</span> application:</p>

```console
odo dev
```

Sample Output:

```console
$ odo dev
  __
 /  \__     Developing using the "my-java-app" Devfile
 \__/  \    Namespace: odo-dev
 /  \__/    odo version: v3.14.0
 \__/

↪ Running on the cluster in Dev mode
 •  Waiting for Kubernetes resources  ...
 ✓  Added storage m2 to component
 ⚠  Pod is Pending
 ✓  Pod is Running
 ✓  Syncing files into the container [167ms]
 ✓  Building your application in container (command: build) [3m]
 •  Executing the application (command: run)  ...
 ✓  Waiting for the application to be ready [1s]
 -  Forwarding from 127.0.0.1:20001 -> 8080


↪ Dev mode
 Status:
 Watching for changes in the current directory /home/user/quickstart-demo

 Keyboard Commands:
[Ctrl+c] - Exit and delete resources from the cluster
     [p] - Manually apply local changes to the application on the cluster
```

Then wait a few seconds until `odo dev` displays `Forwarding from 127.0.0.1:...` in its output,
meaning that `odo` has successfully set up port forwarding to reach the application running in the container.

You can now access the application via the local port displayed by `odo dev` ([127.0.0.1:20001](http://127.0.0.1:20001) in the sample output above) and start your development loop.
`odo` will watch for changes and push the code for real-time updates.
