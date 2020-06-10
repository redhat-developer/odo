#!/bin/sh

## Script for installing and running `oc cluster up`
## Inspired by https://github.com/radanalyticsio/oshinko-cli/blob/master/.travis.yml

## Use this variable to get more control over downloading client binary
OPENSHIFT_CLIENT_BINARY_URL=${OPENSHIFT_CLIENT_BINARY_URL:-'https://github.com/openshift/origin/releases/download/v3.11.0/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit.tar.gz'}

sudo service docker stop

sudo sed -i -e 's/"mtu": 1460/"mtu": 1460, "insecure-registries": ["172.30.0.0\/16"]/' /etc/docker/daemon.json
sudo cat /etc/docker/daemon.json

sudo service docker start
sudo service docker status

# Docker version that oc cluster up uses
docker version

## download oc binaries
sudo wget $OPENSHIFT_CLIENT_BINARY_URL -O /tmp/openshift-origin-client-tools.tar.gz 2> /dev/null > /dev/null

sudo tar -xvzf /tmp/openshift-origin-client-tools.tar.gz --strip-components=1 -C /usr/local/bin

## Get oc version
oc version

## below cmd is important to get oc working in ubuntu
OPENSHIFT_CLIENT_VERSION=`echo $OPENSHIFT_CLIENT_BINARY_URL | awk -F '//' '{print $2}' | cut -d '/' -f 6`
sudo docker run -v /:/rootfs -ti --rm --entrypoint=/bin/bash --privileged openshift/origin:$OPENSHIFT_CLIENT_VERSION -c "mv /rootfs/bin/findmnt /rootfs/bin/findmnt.backup"

while true; do
    if [ "$1" = "service-catalog" ]; then
        oc cluster up --base-dir=$HOME/oscluster
        oc cluster add --base-dir=$HOME/oscluster service-catalog
        oc cluster add --base-dir=$HOME/oscluster template-service-broker
        oc cluster add --base-dir=$HOME/oscluster automation-service-broker
    else
        oc cluster up
    fi
    if [ "$?" -eq 0 ]; then
        ./scripts/travis-check-pods.sh $1
        if [ "$?" -eq 0 ]; then
                break
            fi
    fi
    echo "Retrying oc cluster up after failure"
    oc cluster down
    sleep 5
done
