#!/bin/bash

# fail if some commands fails
set -e
# show commands
set -x

CHECKOUT_GINKGO_VERSION="v1.14.0"

# Delete existing ginkgo repo
if [ -d "$GOPATH/src/github.com/onsi/ginkgo" ]; then
    rm -rf $GOPATH/src/github.com/onsi/ginkgo
fi

git clone https://github.com/onsi/ginkgo $GOPATH/src/github.com/onsi/ginkgo
pushd $GOPATH/src/github.com/onsi/ginkgo
git checkout $CHECKOUT_GINKGO_VERSION
go install -mod=“” github.com/onsi/ginkgo/ginkgo
popd

INSTALLED_GINKGO_VERSION=`ginkgo version`

if [[ "$INSTALLED_GINKGO_VERSION" == *"${CHECKOUT_GINKGO_VERSION:1}"* ]]; then
    echo "Pinned to $INSTALLED_GINKGO_VERSION"
else
    echo "Unable to pin down $INSTALLED_GINKGO_VERSION"
    exit 1
fi

# Clean up
rm -rf $GOPATH/src/github.com/onsi/ginkgo
