#!/bin/bash

# this script test installer.sh in different distributions using docker.

# docker images where install/uninstall script will be tested in
DOCKER_IMAGES="ubuntu:latest debian:latest fedora:latest base/archlinux:latest"

# save tests that failed to this variable
FAILED_INSTALL=""
FAILED_UNINSTALL=""

for image in $DOCKER_IMAGES; do
    echo "******************************************************"
    echo "*** Testing installer.sh in $image"
    echo "******************************************************"

    docker run -it -v `pwd`:/opt/odo $image /opt/odo/scripts/installer.sh
    if [ $? -eq 0 ]; then
        echo "******************************************************"
        echo "**** Install PASSED for $image"
        echo "******************************************************"
    else
        echo "******************************************************"
        echo "**** Install FAILED for $image"
        echo "******************************************************"
        FAILED_INSTALL="$FAILED_INSTALL $image"
    fi

    docker run -it --rm -v `pwd`:/opt/odo $image /bin/bash -c "/opt/odo/scripts/installer.sh; /opt/odo/scripts/installer.sh --uninstall"
    if [ $? -eq 0 ]; then
        echo "******************************************************"
        echo "**** Uninstall PASSED for $image"
        echo "******************************************************"
    else
        echo "******************************************************"
        echo "**** Uninstall FAILED for $image"
        echo "******************************************************"
        FAILED_UNINSTALL="$FAILED_UNINSTALL $image"
    fi
    echo ""
done

if [ -n "$FAILED_INSTALL" ] || [ -n "$FAILED_UNINSTALL" ]; then
    echo "TEST FAILED!!"
    echo "Installation script failed in following images:"
    echo "Install test failures: $FAILED_INSTALL"
    echo "Uninstall test failures: $FAILED_UNINSTALL"
    exit 1
else
    echo "ALL TESTS SUCCEEDED"
fi
