---
title: Installation
sidebar_position: 4
---

odo can be used as a CLI tool and as an IDE plugin; it can be run on Linux, Windows and Mac systems.

## CLI Binary installation

### Installing odo on Linux
```shell
  $ curl -L https://mirror.openshift.com/pub/openshift-v4/clients/odo/latest/odo-linux-amd64 -o /usr/local/bin/odo
  $ chmod +x /usr/local/bin/odo
```

###  Installing odo on Linux on IBM Power
```shell
  $ curl -L https://mirror.openshift.com/pub/openshift-v4/clients/odo/latest/odo-linux-ppc64le -o /usr/local/bin/odo
  $ chmod +x /usr/local/bin/odo
```

### Installing odo on Linux on IBM Z
```shell
  $ curl -L https://mirror.openshift.com/pub/openshift-v4/clients/odo/latest/odo-linux-s390x -o /usr/local/bin/odo
  $ chmod +x /usr/local/bin/odo
```

### Installing odo on macOS
```shell
  $ curl -L https://mirror.openshift.com/pub/openshift-v4/clients/odo/latest/odo-darwin-amd64 -o /usr/local/bin/odo
  $ chmod +x /usr/local/bin/odo
```

### Installing odo on Windows
1. Download the latest odo-windows-amd64.exe file.
2. Rename the downloaded file to odo.exe and move it to a folder of choice, for example C:\odo
3. Add the location of odo.exe to %PATH%.

#### Setting the PATH variable in Windows 10
1. Click **Search** and type `env` or `environment`.
2. Select **Edit environment variables for your account**.
3. Select **Path** from the **Variable** section and click **Edit**.
4. Click **New** and type `<location_of_odo_binary>` into the field or click **Browse** and select the directory, and click **OK**.

#### Setting the PATH variable in Windows 7/8
1. Click **Start** and in the Search box types `Advance System Settings`.
2. Select **Advanced systems settings** and click the **Environment Variables** button at the bottom.
3. Select the **Path** variable from the **System variable** section and click **Edit**.
4. Scroll to the end of the **Variable Value** and add `;<location_of_odo_binary>` and click **OK**.
5. Click **OK** to close the **Environment Variable** dialog.
6. Click **OK** to close the **Systems Properties** dialog.

## Installing odo in Visual Studio Code (VSCode)
The OpenShift VSCode extension uses both odo and the oc binary to interact with Kubernetes or OpenShift cluster.
1. Open VS Code.
2. Launch VS Code Quick Open (Ctrl+P)
3. Paste the following command:
    ```shell
     $ ext install redhat.vscode-openshift-connector
    ```

## Installation from source
1. Clone the repository and cd.
```shell
$ git clone https://github.com/openshift/odo.git; cd odo
```
2. Install tools used by the build and test system.
```shell
make goget-tools
```
3. Build the executable in `cmd/odo`.
```shell
make bin
```
4. Install the executable in the system's GOPATH.
```shell
make install
```
5. Run the command to verify that it was installed properly.
```shell
odo
```
