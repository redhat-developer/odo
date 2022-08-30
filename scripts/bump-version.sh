#!/bin/bash

set -e

# this script updates version number in odo source code
# run this script from root source with new version as an argument (./scripts/bump-version.sh v0.0.2 )

NEW_VERSION=$1

if [[ -z "${NEW_VERSION}" ]]; then
    echo "Version number is missing."
    echo "One argument required."
    echo "example: $0 0.0.2"
    exit 1
fi

check_version(){
    file=$1

    grep ${NEW_VERSION} $file
    echo ""
}


echo "* Bumping version in pkg/version/version.go"
sed -i "s/\(VERSION = \)\".*\"/\1\"v${NEW_VERSION}\"/g" pkg/version/version.go
check_version pkg/version/version.go

echo "* Bumping version in scripts/rpm-prepare.sh"
sed -i "s/\(ODO_VERSION:=\).*}/\1${NEW_VERSION}}/g" scripts/rpm-prepare.sh
check_version scripts/rpm-prepare.sh

echo "* Bumping version in Dockerfile.rhel"
sed -i "s/\(version=\).*/\1${NEW_VERSION}/g" Dockerfile.rhel
check_version Dockerfile.rhel

echo "****************************************************************************************"
echo "* Don't forget to update homebrew package at https://github.com/kadel/homebrew-odo ! *"
echo "****************************************************************************************"

echo "****************************************************************************************"
echo "* Don't forget to update build/VERSION once the binaries become available !            *"
echo "****************************************************************************************"
