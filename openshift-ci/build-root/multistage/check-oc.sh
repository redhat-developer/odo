#!/bin/bash
set -x

## Constants
OC_BINARY="./oc"

# Copy oc binary to bin path
if [ -f $OC_BINARY ]; then
    cp ./oc /usr/bin/oc
fi
