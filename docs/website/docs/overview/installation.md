---
title: Installation
sidebar_position: 4
toc_min_heading_level: 2
toc_max_heading_level: 4
---

`odo` can be used as either a [CLI tool](#cli-installation) or an [IDE plugin](#ide-installation) on [Mac](#macos), [Windows](#windows) or [Linux](#linux).

Each release is *signed*, *checksummed*, *verified*, and then pushed to our [binary mirror](https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/).

For more information on the changes of each release, they can be viewed either on [GitHub](https://github.com/redhat-developer/odo/releases) or the [blog](/blog).

## CLI Installation

### Linux

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs
defaultValue="amd64"
values={[
{label: 'Intel / AMD 64', value: 'amd64'},
{label: 'ARM 64', value: 'arm64'},
{label: 'PowerPC', value: 'ppc64le'},
{label: 'IBM Z', value: 's390x'},
]}>

<TabItem value="amd64">

Installing `odo` on `amd64` architecture:

1. Download the latest release from the mirror:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.11.0/odo-linux-amd64 -o odo
```

2. (Optional) Verify the downloaded binary with the SHA-256 sum:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.11.0/odo-linux-amd64.sha256 -o odo.sha256
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
</TabItem>

<TabItem value="arm64">

Installing `odo` on `arm64` architecture:

1. Download the latest release from the mirror:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.11.0/odo-linux-arm64 -o odo
```

2. (Optional) Verify the downloaded binary with the SHA-256 sum:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.11.0/odo-linux-arm64.sha256 -o odo.sha256
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
</TabItem>

<TabItem value="ppc64le">

Installing `odo` on `ppc64le` architecture:

1. Download the latest release from the mirror:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.11.0/odo-linux-ppc64le -o odo
```

2. (Optional) Verify the downloaded binary with the SHA-256 sum:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.11.0/odo-linux-ppc64le.sha256 -o odo.sha256
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
</TabItem>

<TabItem value="s390x">

Installing `odo` on `s390x` architecture:

1. Download the latest release from the mirror:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.11.0/odo-linux-s390x -o odo
```

2. (Optional) Verify the downloaded binary with the SHA-256 sum:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.11.0/odo-linux-s390x.sha256 -o odo.sha256
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
</TabItem>

</Tabs>

---

### MacOS

#### Homebrew

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

#### Binary

<Tabs
defaultValue="intel"
values={[
{label: 'Intel', value: 'intel'},
{label: 'Apple Silicon', value: 'arm'},
]}>

<TabItem value="intel">

Installing `odo` on `amd64` architecture:

1. Download the latest release from the mirror:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.11.0/odo-darwin-amd64 -o odo
```

2. (Optional) Verify the downloaded binary with the SHA-256 sum:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.11.0/odo-darwin-amd64.sha256 -o odo.sha256
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
</TabItem>

<TabItem value="arm">

Installing `odo` on `arm64` architecture:

1. Download the latest release from the mirror:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.11.0/odo-darwin-arm64 -o odo
```

2. (Optional) Verify the downloaded binary with the SHA-256 sum:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.11.0/odo-darwin-arm64.sha256 -o odo.sha256
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
</TabItem>

</Tabs>

---

### Windows

1. Open a PowerShell terminal

2. Download the latest release from the mirror:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.11.0/odo-windows-amd64.exe -o odo.exe
```

2. (Optional) Verify the downloaded binary with the SHA-256 sum:
```shell
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/v3.11.0/odo-windows-amd64.exe.sha256 -o odo.exe.sha256
# Visually compare the output of both files
Get-FileHash odo.exe
type odo.exe.sha256
```

4. Add the binary to your `PATH`

## IDE Installation

### Visual Studio Code (VSCode)

The [OpenShift Toolkit](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-openshift-connector) VSCode extension uses both `odo` and `oc` binaries
to interact with Kubernetes or OpenShift cluster.

To install it:

1. Open VS Code.
2. Launch VS Code **Quick Open** (`Ctrl+P`).
3. Paste the following command and press `Enter`:

```
ext install redhat.vscode-openshift-connector
```

### JetBrains IDEs
The [OpenShift Toolkit by Red Hat](https://plugins.jetbrains.com/plugin/12030-openshift-toolkit-by-red-hat/) plugin can be installed
to interact with OpenShift or Kubernetes clusters right from your JetBrains IDEs like IntelliJ IDEA, WebStorm or PyCharm.
It uses `odo` and `oc` binaries for fast iterative application development on those clusters.

To install it:

1. Press `Ctrl+Alt+S` to open the IDE settings and select **Plugins**.
2. Find the "**OpenShift Toolkit by Red Hat**" plugin in the **Marketplace** and click **Install**.

## Alternative installation methods

#### Source code
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

### Maven plugin
It is possible to integrate the `odo` binary download in a Maven project using [odo Downloader Plugin](https://github.com/tnb-software/odo-downloader).
The download can be executed using the `download` goal which automatically retrieves the version for the current architecture:
```shell
mvn software.tnb:odo-downloader-maven-plugin:0.1.3:download \
  -Dodo.target.file=$HOME/bin/odo \
  -Dodo.version=v3.11.0
```

### asdf
The [asdf version manager](https://asdf-vm.com/) is a tool for managing multiple runtime versions using a common CLI.
With `asdf` installed, the [asdf plugin for odo](https://github.com/rm3l/asdf-odo) can be used to install any released version of `odo`:
```
asdf plugin add odo
asdf install odo 3.11.0
asdf global odo 3.11.0
```
