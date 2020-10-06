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

1.  Download the latest [`odo.exe`](https://mirror.openshift.com/pub/openshift-v4/clients/odo/latest/odo-windows-amd64.exe) file.

2.  Add the location of your `odo.exe` to your `GOPATH/bin` directory.

### Setting the `PATH` variable for Windows 10

Edit `Environment Variables` using search:

1.  Click **Search** and type `env` or `environment`.

2.  Select **Edit environment variables for your account**.

3.  Select **Path** from the **Variable** section and click **Edit**.

4.  Click **New** and type `C:\go-bin` into the field or click **Browse** and select the directory, and click **OK**.

### Setting the `PATH` variable for Windows 7/8

The following example demonstrates how to set up a path variable. Your binaries can be located in any location, but this example uses C:\\go-bin as the location.

1.  Create a folder at `C:\go-bin`.

2.  Right click **Start** and click **Control Panel**.

3.  Select **System and Security** and then click **System**.

4.  From the menu on the left, select the **Advanced systems settings** and click the **Environment Variables** button at the bottom.

5.  Select **Path** from the **Variable** section and click **Edit**.

6.  Click **New** and type `C:\go-bin` into the field or click **Browse** and select the directory, and click **OK**.

# Installing odo in Visual Studio Code (VSCode)

The [OpenShift VSCode extension](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-openshift-connector) uses both `odo` and the `oc` binary to interact with your Kubernetes or OpenShift cluster.

## Plugin installation

1.  Launch VS Code Quick Open (Ctrl+P)

2.  Paste the following command:
    
    ``` sh
      $ ext install redhat.vscode-openshift-connector
    ```
