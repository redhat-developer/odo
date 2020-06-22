#!/bin/bash

function check_crds () {
    local crd_name="$1"

    for i in  {1..120} ; do
        if ( kubectl get crds |grep ${crd_name} 2>&1 > /dev/null ) ; then
            return 0
        fi

        sleep 3
    done

    return 1
}

CRD_NAME="servicebindingrequests.apps.openshift.io"

echo "# Searching for '${CRD_NAME}'..."

if ! check_crds ${CRD_NAME} ; then
    echo "CRD doesn't exist: ${CRD_NAME}"
    exit 1
fi

echo "CRD is found: ${CRD_NAME}"
