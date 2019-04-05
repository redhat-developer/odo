# Installation Guide

This document provides information on the installation of `odo` on various operating systems and the location of the binaries and the releases.

* The latest `odo` binaries are available on [Bintray](https://dl.bintray.com/odo/odo/latest/).
* The latest `odo` releases are available on [Github](https://github.com/openshift/odo/releases/latest).

## Installing `odo`

### macOS

1. Access the repository using the Homebrew Tap as follows:

```sh
brew tap kadel/odo
```

2. Install from the latest Master build or the latest released version:

    * For the latest Master build use:

    ```sh
    brew install kadel/odo/odo --HEAD
    ```

    * For the latest released version use:

    ```sh
    brew install kadel/odo/odo
    ```

### Ubuntu/Debian

1. Add the GPG key from the Bintray used to sign repositories:

```sh
curl -L https://bintray.com/user/downloadSubjectPublicKey?username=bintray | apt-key add -
```
2. Add the repository to `/etc/apt/sources.list`:

    * For the latest Master builds use:

    ```sh
    echo "deb https://dl.bintray.com/odo/odo-deb-dev stretch main" | sudo tee -a /etc/apt/sources.list
    ```
    * For the latest signed releases use:

    ```sh
    echo "deb https://dl.bintray.com/odo/odo-deb-releases stretch main" | sudo tee -a /etc/apt/sources.list
    ```
3. Install odo:

```sh
sudo apt-get update
sudo apt-get install odo
```

### Fedora/CentOS/RHEL
1. Access the `odo` repository:

    * From the latest builds:
        1. Download the odo repository to `/etc/yum.repods.d/` using:

        ```sh
        sudo curl -L https://bintray.com/odo/odo-rpm-dev/rpm -o /etc/yum.repos.d/bintray-odo-odo-rpm-dev.repo
        ```
        2. Verify the content of the file and ensure that you see the following:

        ```
        # cat/etc/yum.repos.d/bintray-odo-odo-rpm-dev.repo
        [bintraybintray-odo-odo-rpm-dev]
        name=bintray-odo-odo-rpm-dev
        baseurl=https://dl.bintray.com/odo/odo-rpm-dev
        gpgcheck=0
        repo_gpgcheck=0
        enabled=1
        ```

    * From the latest release:
        1. Download the odo repository to `/etc/yum.repods.d/` using:

        ```sh
        sudo curl -L https://bintray.com/odo/odo-rpm-releases/rpm -o /etc/yum.repos.d/bintray-odo-odo-rpm-releases.repo
        ```
        2. Verify the content of the file and ensure that you see the following:

        ```
        # cat/etc/yum.repos.d/bintray-odo-odo-rpm-releases.repo
        [bintraybintray-odo-odo-rpm-releases]
        name=bintray-odo-odo-rpm-releases
        baseurl=https://dl.bintray.com/odo/odo-rpm-releases
        gpgcheck=0
        repo_gpgcheck=0
        enabled=1
        ```

2. Install Odo:

    * For CentOS or RHEL use `yum install odo`.
    * For Fedora use `dnf install odo`.

### Windows

1. Download the latest file from the Bintray ([odo.exe](https://dl.bintray.com/odo/odo/latest/windows-amd64/odo.exe)) or from the latest release page on [GitHub](https://github.com/openshift/odo/releases).
2. Extract the file and add the location of extracted binary to your `GOPATH/bin` directory (see below to create a Go binary directory)

#### Setting a PATH variable for Windows 7/8

The following example demonstrates how to set up a path variable. Your binaries can be located in any location but for the purpose of this example we will use `C:\go-bin` as the location.

1. Create a folder at `C:\go-bin`.
2. Right click on **Start** and click on **Control Panel**.
3. Select **System and Security** and then click on **System**.
4. From the menu on the left, select the **Advanced systems settings** and click the **Environment Variables** button at the bottom.
5. Select **Path** from the **Variable** section and click **Edit**.
6. Click **New** and type `C:\go-bin` into the field or click **Browse** and select the directory and click **OK**.

#### Setting a PATH variable for Windows 10
You can edit `Environment Variables` faster using search as follows:

1. Left click on **Search** and type `env` or `environment`.
2. Select **Edit environment variables for your account** and follow steps 5 and 6 listed above.
