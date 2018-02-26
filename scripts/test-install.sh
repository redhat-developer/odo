#!/bin/bash

# this script test install.sh in different distributions using docker.

# docker images where install script will be tested in
DOCKER_IMAGES="ubuntu:latest debian:latest fedora:latest base/archlinux:latest"

# save tests that failed to this variable
FAILED_INSTALL=""

for image in $DOCKER_IMAGES; do
    echo "******************************************************"
    echo "*** Testing install.sh in $image"
    echo "******************************************************"
    
    docker run -it --rm -v `pwd`:/opt/ocdev $image /opt/ocdev/scripts/install.sh
    if [ $? -eq 0 ]; then
        echo "******************************************************"
        echo "**** PASSED for $image"
        echo "******************************************************"
    else
        echo "******************************************************"
        echo "**** FAILED for $image"
        echo "******************************************************"
        FAILED_INSTALL="$FAILED_INSTALL $image"
    fi
    echo ""
done


if [ -n "$FAILED_INSTALL" ]; then
    echo "TEST FAILED!!"
    echo "Instalation script failed in following images:"
    echo "$FAILED_INSTALL"
else
    echo "ALL TESTS SUCCEEDED"
fi
