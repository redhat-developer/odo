#!/bin/bash

function check_installplan () {
    for i in  {1..120} ; do
        local install_plan=`kubectl get subscriptions service-binding-operator -n openshift-operators -o jsonpath='{.status.installPlanRef.name}'`
        if [[ ${install_plan} != "" ]] ; then
            return 0
        fi

        sleep 3
    done

    return 1
}


echo "# Searching for install plan."

if ! check_installplan ; then
    echo "Install plan doesn't exist"
    exit 1
fi

INSTALL_PLAN=`kubectl get subscriptions service-binding-operator -n openshift-operators -o jsonpath='{.status.installPlanRef.name}'`

echo "Install plan found: ${INSTALL_PLAN}"
