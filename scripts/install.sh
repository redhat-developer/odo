#!/bin/bash
set -e

# The version of odo to install. Possible values - "master" and "latest"
# master - builds from git master branch
# latest - released versions specified by LATEST_VERSION variable
ODO_VERSION="latest"

# Latest released odo version
LATEST_VERSION="v0.0.12"

GITHUB_RELEASES_URL="https://github.com/redhat-developer/odo/releases/download/${LATEST_VERSION}"
BINTRAY_URL="https://dl.bintray.com/odo/odo/latest"

INSTALLATION_PATH="/usr/local/bin/"
PRIVILEGED_EXECUTION="sh -c"

DEBIAN_GPG_PUBLIC_KEY="https://bintray.com/user/downloadSubjectPublicKey?username=bintray"
DEBIAN_MASTER_REPOSITORY="https://dl.bintray.com/odo/odo-deb-dev"
DEBIAN_LATEST_REPOSITORY="https://dl.bintray.com/odo/odo-deb-releases"

RPM_MASTER_YUM_REPO="https://bintray.com/odo/odo-rpm-dev/rpm"
RPM_LATEST_YUM_REPO="https://bintray.com/odo/odo-rpm-releases/rpm"

SUPPORTED_PLATFORMS="
darwin-amd64
linux-amd64
linux-arm
"

echo_stderr ()
{
    echo "$@" >&2
}

command_exists() {
	command -v "$@" > /dev/null 2>&1
}

check_platform() {
    kernel="$(uname -s)"
    if [ "$(uname -m)" = "x86_64" ]; then
        arch="amd64"
    fi

    platform_type=$(echo "${kernel}-${arch}" | tr '[:upper:]' '[:lower:]')

    if ! echo "# $SUPPORTED_PLATFORMS" | grep "$platform_type" > /dev/null; then
        echo_stderr "
# The installer has detected your platform to be $platform_type, which is
# currently not supported by this installer script.

# Please visit the following URL for detailed installation steps:
# https://github.com/redhat-developer/odo/#installation

        "
        exit 1
    fi
    echo "$platform_type"
}

get_distribution() {
	lsb_dist=""
	if [ -r /etc/os-release ]; then
		lsb_dist="$(. /etc/os-release && echo "$ID")"
	fi
	echo "$lsb_dist"
}

set_privileged_execution() {
	if [ "$(id -u)" != "0" ]; then
        if command_exists sudo; then
            echo "# Installer will run privileged commands with sudo"
            PRIVILEGED_EXECUTION='sudo -E sh -c'
        elif command_exists su ; then
            echo "# Installer will run privileged commands with \"su -c\""
            PRIVILEGED_EXECUTION='su -c'
        else
    	    echo_stderr "# 
This installer needs to run as root. The current user is not root, and we could not find "sudo" or "su" installed on the system. Please run again with root privileges, or install "sudo" or "su" packages.
"
        fi
    else
        echo "# Installer is being run as root"
	fi
}

invalid_odo_version_error() {
    echo_stderr "# Invalid value of odo version provided, provide master or latest."
    exit 1
}

install_odo() {
    echo "# Starting odo installation..."
    echo "# Detecting distribution..."

    platform="$(check_platform)"
    echo "# Detected platform: $platform"

    if command_exists odo; then
        echo_stderr "# 
odo version \"$(odo version)\" is already installed on your system, running this installer script might case issues with your current installation. If you want to install odo using this script, please remove the current installation of odo from you system.
Aborting now!
"
        exit 1
    fi

    # macOS specific steps
    if [ "$platform" = "darwin-amd64" ]; then
        if ! command_exists brew; then
            echo_stderr "# brew command does not exist. Please install and run the installer again."
        fi

        echo "# Enabling kadel/odo... "
        brew tap kadel/odo

        echo "Installing odo..."
        case "$ODO_VERSION" in
        master)
            brew install kadel/odo/odo -- HEAD
            ;;
        latest)
            brew install kadel/odo/odo
        esac

        return 0
    fi

    set_privileged_execution

    distribution=$(get_distribution)
  	echo "# Detected distribution: $distribution"
  	echo "# Installing odo version: $ODO_VERSION"

    case "$distribution" in

    ubuntu|debian)
        echo "# Installing pre-requisites..."
        $PRIVILEGED_EXECUTION "apt-get update"
        $PRIVILEGED_EXECUTION "apt-get install -y gnupg apt-transport-https curl"

        echo "# "Adding GPG public key...
        $PRIVILEGED_EXECUTION "curl -L \"$DEBIAN_GPG_PUBLIC_KEY\" |  apt-key add -"

        echo "# Adding repository to /etc/apt/sources.list"
        case "$ODO_VERSION" in
        master)
            $PRIVILEGED_EXECUTION "echo \"deb $DEBIAN_MASTER_REPOSITORY stretch main\" |  tee -a /etc/apt/sources.list"
            ;;
        latest)
            $PRIVILEGED_EXECUTION "echo \"deb $DEBIAN_LATEST_REPOSITORY stretch main\" | tee -a /etc/apt/sources.list"
            ;;
        *)
            invalid_odo_version_error
        esac

        $PRIVILEGED_EXECUTION "apt-get update"
        $PRIVILEGED_EXECUTION "apt-get install -y odo"
        ;;

    centos|fedora)
        package_manager=""
        case "$distribution" in
        fedora)
            package_manager="dnf"
            ;;
        centos)
            package_manager="yum"
            ;;
        esac

        echo "# Adding odo repo under /etc/yum.repos.d/"
        case "$ODO_VERSION" in

        master)
            $PRIVILEGED_EXECUTION "curl -L $RPM_MASTER_YUM_REPO -o /etc/yum.repos.d/bintray-odo-odo-rpm-dev.repo"
            ;;
        latest)
            $PRIVILEGED_EXECUTION "curl -L $RPM_LATEST_YUM_REPO -o /etc/yum.repos.d/bintray-odo-odo-rpm-releases.repo"
            ;;
        *)
            invalid_odo_version_error
        esac

        $PRIVILEGED_EXECUTION "$package_manager install -y odo"
        ;;

    *)
        echo "# Could not identify distribution, proceeding with a binary install..."

        BINARY_URL=""
        TMP_DIR=$(mktemp -d)
        case "$ODO_VERSION" in
        master)
            BINARY_URL="$BINTRAY_URL/$platform/odo"
            echo "# Downloading odo from $BINARY_URL"
            curl -Lo $TMP_DIR/odo "$BINARY_URL"
            ;;
        latest)
            BINARY_URL="$GITHUB_RELEASES_URL/odo-$platform.gz"
            echo "# Downloading odo from $BINARY_URL"
            curl -Lo $TMP_DIR/odo.gz "$BINARY_URL"
            echo "# Extracting odo.gz"
            gunzip -d $TMP_DIR/odo.gz
            ;;
        *)
            invalid_odo_version_error
        esac

        echo "# Setting execute permissions on odo"
        chmod +x $TMP_DIR/odo
        echo "# Moving odo binary to $INSTALLATION_PATH"
        $PRIVILEGED_EXECUTION "mv $TMP_DIR/odo $INSTALLATION_PATH"
        echo "# odo has been successfully installed on your machine"
        rm -r $TMP_DIR
        ;;
    esac
}

verify_odo() {
    if command_exists odo; then
        echo "
# Verification complete!
# odo version \"$(odo version)\" has been installed at $(type -P odo)
"
    else
        echo_stderr "
# Something is wrong with odo installation, please run the installaer script again. If the issue persists, please create an issue at https://github.com/redhat-developer/odo/issues"
        exit 1
    fi
}

install_odo
verify_odo