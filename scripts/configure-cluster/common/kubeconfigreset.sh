#!/usr/bin/env bash

reset_kubeconfig() {
    if [[ -z $ORIGINAL_KUBECONFIG || -z $ORIGINAL_KUBECONFIG ]]; then
        echo "KUBECONFIG or ORIGINAL_KUBECONFIG not set"
        exit 1
    fi
    if [[ $KUBECONFIG != $ORIGINAL_KUBECONFIG ]]; then
        rm -rf $KUBECONFIG
    fi
    export KUBECONFIG=$ORIGINAL_KUBECONFIG
}

reset_kubeconfig