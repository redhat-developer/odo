#!/bin/sh

# fail if some commands fails
set -e
# show commands
set -x

export CI="openshift"
export TIMEOUT="30m"
make configure-installer-tests-cluster
make bin
export PATH="$PATH:$(pwd)"
export CUSTOM_HOMEDIR="/tmp/artifacts"

make test-e2e-scenarios