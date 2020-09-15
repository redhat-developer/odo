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
The following section describes how to install `odo` on different
platforms using the CLI.

# Installing v2.0.0-beta-1 of odo

v2.0.0-beta-1 of odo uses **devfile** as its default deployment
mechanism.

## Installing odo on Linux

### Binary installation

    # curl -L https://github.com/openshift/odo/releases/download/v2.0.0-beta-1/odo-linux-amd64 -o /usr/local/bin/odo
    # chmod +x /usr/local/bin/odo

## Installing odo on macOS

### Binary installation

    # curl -L https://github.com/openshift/odo/releases/download/v2.0.0-beta-1/odo-darwin-amd64 -o /usr/local/bin/odo
    # chmod +x /usr/local/bin/odo

## Installing odo on Windows

### Binary installation

1.  Download the latest
    [`odo.exe`](https://github.com/openshift/odo/releases/download/v2.0.0-beta-1/odo-windows-amd64.exe)
    file.

2.  Add the location of your `odo.exe` to your `GOPATH/bin` directory.

### Setting the `PATH` variable for Windows 10

Edit `Environment Variables` using search:

1.  Click **Search** and type `env` or `environment`.

2.  Select **Edit environment variables for your account**.

3.  Select **Path** from the **Variable** section and click **Edit**.

4.  Click **New** and type `C:\go-bin` into the field or click
    **Browse** and select the directory, and click **OK**.

### Setting the `PATH` variable for Windows 7/8

The following example demonstrates how to set up a path variable. Your
binaries can be located in any location, but this example uses
C:\\go-bin as the location.

1.  Create a folder at `C:\go-bin`.

2.  Right click **Start** and click **Control Panel**.

3.  Select **System and Security** and then click **System**.

4.  From the menu on the left, select the **Advanced systems settings**
    and click the **Environment Variables** button at the bottom.

5.  Select **Path** from the **Variable** section and click **Edit**.

6.  Click **New** and type `C:\go-bin` into the field or click
    **Browse** and select the directory, and click **OK**.

# Installing v1.2.5 of odo

v1.2.5 of odo uses **Source-to-Image** as its default deployment
mechanism.

Visit the latest [OpenShift
documentation](https://docs.openshift.com/container-platform/4.5/cli_reference/developer_cli_odo/installing-odo.html)
for installation details.
