#!/bin/bash
set -e

# The version of odo to install. Possible values - "master" and "latest"
# master - builds from git master branch
# latest - released versions specified by LATEST_VERSION variable
ODO_VERSION="latest"

# Latest released odo version
LATEST_VERSION="v1.2.2"

GITHUB_RELEASES_URL="https://github.com/openshift/odo/releases/download/${LATEST_VERSION}"
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

# Used to determine whether to install or uninstall odo
INSTALLER_ACTION=""

parse_installer_action_flag ()
{
    case "$@" in

    # Set INSTALLER_ACTION to uninstall odo or install odo
    # Include --uninstall flag when running installer.sh to uninstall and simply run installer.sh to install latest version of odo
    --uninstall)
        INSTALLER_ACTION="uninstall"
    ;;

    *)
        INSTALLER_ACTION="install"
    ;;
    esac
}

echo_stderr ()
{
    echo "$@" >&2
}

command_exists() {
    distribution=$(get_distribution)

    case "$distribution" in

    ubuntu|debian)
        # Use which to verify install/uninstall on ubuntu and debian distributions
        which "$@" > /dev/null 2>&1
    ;;

    *)
        command -v "$@" > /dev/null 2>&1
    ;;
    esac
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
# https://github.com/openshift/odo/#installation"
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
    echo_stderr "# Invalid value of odo version provided. Provide master or latest."
    exit 1
}

installer_odo() {
    echo "# Detecting distribution..."

    platform="$(check_platform)"
    echo "# Detected platform: $platform"

    if [ "$INSTALLER_ACTION" == "install" ] && command_exists odo; then
        echo_stderr echo_stderr "#
odo version \"$(odo version --client)\" is already installed on your system. Running this installer script might cause issues with your current installation. If you want to install odo using this script, please remove the current installation of odo from you system.
Aborting now!
"
        exit 1
    elif [ "$INSTALLER_ACTION" == "uninstall" ] && ! command_exists odo; then
        echo_stderr "# odo is not installed on your system. Ending execution of uninstall script."
        exit 1
    fi

    # macOS specific steps
    if [ $platform = "darwin-amd64" ]; then
        if ! command_exists brew; then
            echo_stderr "# brew command does not exist. Please install brew and run the installer again."
        fi

        if [ "$INSTALLER_ACTION" == "install" ]; then
            brew tap kadel/odo
            echo "# Installing odo ${ODO_VERSION} on macOS"
            case $ODO_VERSION in
                master)
                  brew install kadel/odo/odo -- HEAD
                  ;;
                latest)
                  brew install kadel/odo/odo
            esac
        elif [ "$INSTALLER_ACTION" == "uninstall" ]; then
            echo "# Uninstalling odo on macOS"
            brew uninstall odo
        fi

        return 0
    fi

    set_privileged_execution

    distribution=$(get_distribution)
  	echo "# Detected distribution: $distribution"

    case "$distribution" in

    ubuntu|debian)
        if [ "$INSTALLER_ACTION" == "install" ]; then
            echo "# Installing odo version: $ODO_VERSION on $distribution"
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
        elif [ "$INSTALLER_ACTION" == "uninstall" ]; then
            echo "# Uninstalling odo..."
            $PRIVILEGED_EXECUTION "apt-get remove -y odo"
        fi
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

        if [ "$INSTALLER_ACTION" == "install" ]; then
            echo "# Installing odo version $ODO_VERSION on $distribution"

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
        elif [ "$INSTALLER_ACTION" == "uninstall" ]; then
            echo "# Uninstalling odo..."
            $PRIVILEGED_EXECUTION "$package_manager remove -y odo"
        fi
    ;;

    *)
        if [ "$INSTALLER_ACTION" == "install" ]; then
            echo "# Could not identify distribution. Proceeding with a binary install..."

            BINARY_URL=""
            TMP_DIR=$(mktemp -d)
            case "$ODO_VERSION" in

            master)
                BINARY_URL="$BINTRAY_URL/$platform/odo"
                echo "# Downloading odo from $BINARY_URL"
                curl -Lo $TMP_DIR/odo "$BINARY_URL"
            ;;

            latest)
                BINARY_URL="$GITHUB_RELEASES_URL/odo-$platform.tar.gz"
                echo "# Downloading odo from $BINARY_URL"
                curl -Lo $TMP_DIR/odo.tar.gz "$BINARY_URL"
                echo "# Extracting odo.tar.gz"
                tar -xvzf $TMP_DIR/odo.tar.gz
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

        elif [ "$INSTALLER_ACTION" == "uninstall" ]; then
            echo "# Proceeding with removing binary..."
            rm -r $INSTALLATION_PATH/odo
        fi
    ;;
    esac
}

verify_odo() {
    if [ $INSTALLER_ACTION == "install" ] && command_exists odo; then
        echo "
# Verification complete!
# odo version \"$(odo version --client)\" has been installed at $(type -P odo)
"
    elif [ "$INSTALLER_ACTION" == "uninstall" ] && ! command_exists odo; then
        echo "
# Verification complete!
# odo has been uninstalled"
    else
        echo_stderr "
# Something is wrong with odo installer. Please run the installer script again. If the issue persists, please create an issue at https://github.com/openshift/odo/issues"
        exit 1
    fi
}

parse_installer_action_flag "$@"
installer_odo
verify_odo
