---
title: Installation
sidebar_position: 4
---

odo can be used as a CLI tool and as an IDE plugin; it can be run on Linux, Windows and Mac systems.

## CLI Binary installation
odo supports amd64 architecture for Linux, Mac and Windows.
Additionally, it also supports amd64, arm64, s390x, and ppc64le architectures for Linux.

See the [release page](https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/latest/) for more information.

### Installing odo on Linux/Mac
```shell
OS="$(uname | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')"
curl -L https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/latest/odo-$OS-$ARCH -o odo
sudo install odo /usr/local/bin/
```

### Installing odo on Windows
1. Download the [odo-windows-amd64.exe](https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/odo/latest/odo-windows-amd64.exe) file.
2. Rename the downloaded file to odo.exe and move it to a folder of choice, for example `C:\odo`.
3. Add the location of odo.exe to `%PATH%` variable (refer to the steps below).

#### Setting the PATH variable in Windows 10
1. Click **Search** and type `env` or `environment`.
2. Select **Edit environment variables for your account**.
3. Select **Path** from the **Variable** section and click **Edit**.
4. Click **New**, add the location where you copied the odo binary (e.g. `C:\odo` in [Step 2 of Installation](#installing-odo-on-windows) into the field or click **Browse** and select the directory, and click **OK**.

#### Setting the PATH variable in Windows 7/8
1. Click **Start** and in the Search box types `Advanced System Settings`.
2. Select **Advanced systems settings** and click the **Environment Variables** button at the bottom.
3. Select the **Path** variable from the **System variables** section and click **Edit**.
4. Scroll to the end of the **Variable value** and add `;` followed by the location where you copied the odo binary (e.g. `C:\odo` in [Step 2 of Installation](#installing-odo-on-windows) and click **OK**.
5. Click **OK** to close the **Environment Variables** dialog.
6. Click **OK** to close the **System Properties** dialog.

## Installing odo in Visual Studio Code (VSCode)
The [OpenShift VSCode extension](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-openshift-connector) uses both odo and oc binary to interact with Kubernetes or OpenShift cluster.
1. Open VS Code.
2. Launch VS Code Quick Open (Ctrl+P)
3. Paste the following command:
    ```shell
     ext install redhat.vscode-openshift-connector
    ```

## Installing from source
1. Clone the repository and cd into it.
   ```shell
   git clone https://github.com/openshift/odo.git
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
