#!/bin/bash

set -e

# this script updates version number in odo source code
# run this script from root source with new version as an argument (./scripts/bump-version.sh v0.0.2 )

NEW_VERSION=$1

if [[ -z "${NEW_VERSION}" ]]; then
    echo "Version number is missing."
    echo "One argument required."
    echo "example: $0 v0.0.2"
    exit 1
fi

check_version(){
    file=$1
    
    grep ${NEW_VERSION} $file
    echo ""
}


echo "* Bumping version in pkg/odo/cli/version/version.go"
sed -i "s/\(VERSION = \)\"v[0-9]*\.[0-9]*\.[0-9]*\(?:-\w+\)\?\"/\1\"${NEW_VERSION}\"/g" pkg/odo/cli/version/version.go
check_version pkg/odo/cli/version/version.go

echo "* Bumping version in scripts/install.sh"
sed -i "s/\(LATEST_VERSION=\)\"v[0-9]*\.[0-9]*\.[0-9]*\(?:-\w+\)\?\"/\1\"${NEW_VERSION}\"/g" scripts/install.sh
check_version scripts/install.sh

echo "****************************************************************************************"
echo "* Don't forget to update homebrew package at https://github.com/kadel/homebrew-odo ! *"
echo "****************************************************************************************"

echo "****************************************************************************************"
echo "* Don't forget to update build/VERSION once the binaries become available !            *"
echo "****************************************************************************************"

