#!/bin/bash

function check_pod () {
    for i in  {1..120} ; do
        local running_status=`kubectl get pods -n openshift-marketplace | grep "example-operators" | awk '{print $3}'`
        if [[ ${running_status} == "Running" ]] ; then
            return 0
        fi

        sleep 3
    done

    return 1
}


echo "# Searching for operator source pod."

if ! check_pod ; then
    echo "Operator source pod is not running"
    exit 1
fi

RUNNING_STATUS=`kubectl get pods -n openshift-marketplace | grep "example-operators" | awk '{print $3}'`

echo "Operator source pod found with status: ${RUNNING_STATUS}"
