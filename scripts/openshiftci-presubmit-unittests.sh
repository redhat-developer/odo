#!/bin/sh

# fail if some commands fails
set -e
# show commands
set -x

export CUSTOM_HOMEDIR=$ARTIFACT_DIR
export PATH=$PATH:$GOPATH/bin
# set location for golangci-lint cache
# otherwise /.cache is used, and it fails on permission denied
export GOLANGCI_LINT_CACHE="/tmp/.cache"

ARCH=`uname -i`
if [ "$ARCH" = "x86_64" ]; then
  export GOARCH=amd64
else
  export GOARCH=$ARCH
fi

wget https://github.com/golangci/golangci-lint/releases/download/v1.37.0/golangci-lint-1.37.0-linux-$GOARCH.tar.gz
tar --no-same-owner -xzf golangci-lint-1.37.0-linux-amd64.tar.gz
mv golangci-lint-1.37.0-linux-amd64/golangci-lint $(go env GOPATH)/bin/golangci-lint
chmod +x $(go env GOPATH)/bin/golangci-lint

make validate
make test

# crosscompile and publish artifacts
make cross
cp -r dist $ARTIFACT_DIR

# RPM Tests
#scripts/rpm-x86_64-test.sh
