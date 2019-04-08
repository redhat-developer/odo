#!/bin/sh

# fail if some commands fails
set -e
# show commands
set -x

export CUSTOM_HOMEDIR="/tmp/artifacts"
make test