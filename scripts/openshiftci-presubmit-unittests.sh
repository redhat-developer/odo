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

go version
yum whatprovides go

make goget-tools
make validate
make test

# crosscompile and publish artifacts
make cross
cp -r dist $ARTIFACT_DIR

# RPM Tests
scripts/rpm-x86_64-test.sh
