# Installation Guide

List of various installation methods:

- [macOS](#macOS)
- [Deb (Debian / Ubuntu)](#deb)
- [RPM (Fedora / CentOS / RHEL)](#rpm)
- [Windows](#windows)

Latest binaries are available on [Bintray](https://dl.bintray.com/odo/odo/latest/).

Latest releases are available on [Github](https://github.com/openshift/odo/releases/latest).

## macOS

#### Enable via Homebrew Tap

In order to access the repo:

```sh
brew tap kadel/odo
```

#### Install

The latest MASTER build:

```sh
brew install kadel/odo/odo --HEAD
```

Latest released version:

```sh
brew install kadel/odo/odo
```

## Deb

Ubuntu / Debian

#### Add GPG Key

Add the GPG key from Bintray used to sign repositories:

```sh
curl -L https://bintray.com/user/downloadSubjectPublicKey?username=bintray | apt-key add -
```

Add the repository to `/etc/apt/sources.list`:

```sh
# For latest Master builds
echo "deb https://dl.bintray.com/odo/odo-deb-dev stretch main" | sudo tee -a /etc/apt/sources.list

# For latest signed releases
echo "deb https://dl.bintray.com/odo/odo-deb-releases stretch main" | sudo tee -a /etc/apt/sources.list
```

Now install odo:

```sh
sudo apt-get update
sudo apt-get install odo
```

## RPM

Fedora / CentOS / RHEL

#### Add the odo repository to `/etc/yum.repods.d/`:

For latest builds:

```sh
$ vim /etc/yum.repos.d/bintray-odo-odo-rpm-dev.repo
```

```
# /etc/yum.repos.d/bintray-odo-odo-rpm-dev.repo
[bintraybintray-odo-odo-rpm-dev]
name=bintray-odo-odo-rpm-dev
baseurl=https://dl.bintray.com/odo/odo-rpm-dev
gpgcheck=0
repo_gpgcheck=0
enabled=1
```

Or you can download it using following command:

```sh
sudo curl -L https://bintray.com/odo/odo-rpm-dev/rpm -o /etc/yum.repos.d/bintray-odo-odo-rpm-dev.repo
```

For the latest release:

```sh
$ vim /etc/yum.repos.d/bintray-odo-odo-rpm-releases.repo
```

```
# /etc/yum.repos.d/bintray-odo-odo-rpm-releases.repo
[bintraybintray-odo-odo-rpm-releases]
name=bintray-odo-odo-rpm-releases
baseurl=https://dl.bintray.com/odo/odo-rpm-releases
gpgcheck=0
repo_gpgcheck=0
enabled=1
```

Or you can download it using following command:

```sh
sudo curl -L https://bintray.com/odo/odo-rpm-releases/rpm -o /etc/yum.repos.d/bintray-odo-odo-rpm-releases.repo
```

#### Install Odo

```sh
# CentOS / RHEL
yum install odo

# Fedora
dnf install odo
 ```

## Windows

1. Download the latest  file from Bintray ([odo.exe](https://dl.bintray.com/odo/odo/latest/windows-amd64/odo.exe)) or from the latest release page on [GitHub](https://github.com/openshift/odo/releases).
2. Extract the file
3. Add the location of extracted binary to your GOPATH/bin directory (see below if you have yet to create a Go binary directory)

#### Setting a PATH variable for Windows 7/8

Your binaries can be located wherever you like,
but we'll use `C:\go-bin` in this example.

* Create folder at `C:\go-bin`.
* Right click on "Start" and click on "Control Panel". Select "System and Security", then click on "System".
* From the menu on the left, select the "Advanced systems settings".
* Click the "Environment Variables" button at the bottom.
* Select "Path" from the "Variable" section & Click "Edit"
* Click "New" 
* Type `C:\go-bin` into the field or Click "Browse" and select the directory.
* Click OK.

#### Setting a PATH variable for Windows 10

There is a faster way to edit `Environment Variables` with search
* Left click on "Search" and type `env` or `environment`. select `Edit environment variables for your account`
* and follow step above
