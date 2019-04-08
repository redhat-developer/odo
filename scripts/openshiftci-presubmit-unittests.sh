#!/bin/sh

# fail if some commands fails
set -e
# show commands
set -x

export ARTIFACTS_DIR=/tmp/artifacts
export CUSTOM_HOMEDIR=$ARTIFACTS_DIR

make test

make cross
cp -r dist $ARTIFACTS_DIR