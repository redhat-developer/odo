---
# Page settings
layout: default
keywords:
comments: false

# Hero section
title: Installing Odo
description: Installing odo on macOS, Linux and Windows

# Micro navigation
micro_nav: true

# Page navigation
page_nav:
    next:
        content: Understanding odo
        url: '/docs/understanding-odo'
---
The following section describes how to install `odo` on different platforms via CLI as well as IDEs.

# Installing the odo CLI tool (latest)

## Installing odo on Linux

### Binary installation

``` sh
  $ curl -L https://mirror.openshift.com/pub/openshift-v4/clients/odo/latest/odo-linux-amd64 -o /usr/local/bin/odo
  $ chmod +x /usr/local/bin/odo
```

## Installing odo on Linux on IBM Z

### Binary installation

``` 
  $ curl -L https://mirror.openshift.com/pub/openshift-v4/clients/odo/latest/odo-linux-s390x -o /usr/local/bin/odo
  $ chmod +x /usr/local/bin/odo
```

## Installing odo on Linux on IBM Power

### Binary installation

``` 
  $ curl -L https://mirror.openshift.com/pub/openshift-v4/clients/odo/latest/odo-linux-ppc64le -o /usr/local/bin/odo
  $ chmod +x /usr/local/bin/odo
```

## Installing odo on macOS

### Binary installation

``` sh
  $ curl -L https://mirror.openshift.com/pub/openshift-v4/clients/odo/latest/odo-darwin-amd64 -o /usr/local/bin/odo
  $ chmod +x /usr/local/bin/odo
```

## Installing odo on Windows

### Binary installation

1.  Download the latest [`odo-windows-amd64.exe`](https://mirror.openshift.com/pub/openshift-v4/clients/odo/latest/odo-windows-amd64.exe) file.

2.  Rename the downloaded file to `odo.exe` and move it to a folder of your choice, for example `C:\odo`

3.  Add the location of your `odo.exe` to your `%PATH%`.

### Setting the `PATH` variable for Windows 10

Edit `Environment Variables` using search: The following example demonstrates how to set up a path variable. Your binary can be located in any location, but this example uses C:\\odo as the location.

1.  Click **Search** and type `env` or `environment`.

2.  Select **Edit environment variables for your account**.

3.  Select **Path** from the **Variable** section and click **Edit**.

4.  Click **New** and type `C:\odo` into the field or click **Browse** and select the directory, and click **OK**.

### Setting the `PATH` variable for Windows 7/8

1.  Click **Start** and in the `Search` box types `Advance System Settings`.

2.  Select **Advanced systems settings** and click the **Environment Variables** button at the bottom.

3.  Select the **Path** variable from the **System variable** section and click **Edit**.

4.  Scroll to the end of the **Variable Value** and add `;C:\odo` and click **OK**.

5.  Click **OK** to close the **Environment Variable** dialog.

6.  Click **OK** to close the **Systems Properties** dialog.

# Installing odo in Visual Studio Code (VSCode)

The [OpenShift VSCode extension](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-openshift-connector) uses both `odo` and the `oc` binary to interact with your Kubernetes or OpenShift cluster.

## Plugin installation

1.  Launch VS Code Quick Open (Ctrl+P)

2.  Paste the following command:
    
    ``` sh
      $ ext install redhat.vscode-openshift-connector
    ```
