#!/bin/sh

## Script for installing and running `oc cluster up`
## Inspired by https://github.com/radanalyticsio/oshinko-cli/blob/master/.travis.yml

sudo service docker stop

sudo sed -i -e 's/sock/sock --insecure-registry 172.30.0.0\/16/' /etc/default/docker
sudo cat /etc/default/docker

sudo service docker start
sudo service docker status

## download oc binaries
sudo wget https://github.com/openshift/origin/releases/download/v3.10.0/openshift-origin-client-tools-v3.10.0-dd10d17-linux-64bit.tar.gz -O /tmp/openshift-origin-client-tools.tar.gz 2> /dev/null > /dev/null

sudo tar -xvzf /tmp/openshift-origin-client-tools.tar.gz --strip-components=1 -C /usr/local/bin

## Get oc version
oc version

## below cmd is important to get oc working in ubuntu
sudo docker run -v /:/rootfs -ti --rm --entrypoint=/bin/bash --privileged openshift/origin:v3.10.0 -c "mv /rootfs/bin/findmnt /rootfs/bin/findmnt.backup"

while true; do
    oc cluster up
    if [ "$?" -eq 0 ]; then
	./scripts/travis-check-pods.sh
	if [ "$?" -eq 0 ]; then
            break
        fi
    fi
    echo "Retrying oc cluster up after failure"
    oc cluster down
    sleep 5
done
