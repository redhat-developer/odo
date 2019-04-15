#!/bin/bash

# Code from https://github.com/radanalyticsio/oshinko-cli/blob/master/travis-check-pods.sh

oc login -u system:admin
oc project default

while true; do
    V=$(oc get dc docker-registry --template='{{index .status "latestVersion"}}')
    P=$(oc get pod docker-registry-$V-deploy --template='{{index .status "phase"}}')
    if [ "$?" -eq 0 ]; then
        echo phase is $P for docker-registry deploy $V
        if [ "$P" == "Failed" ]; then
            echo "registry deploy failed, try again"
            oc get pods
            oc rollout retry dc/docker-registry
            sleep 10
            continue
        fi
    fi
    REG=$(oc get pod -l deploymentconfig=docker-registry --template='{{index .items 0 "status" "phase"}}')
    if [ "$?" -eq 0 ]; then
        break
    fi
    oc get pods
    echo "Waiting for registry pod"
    sleep 10
done

while true; do
    REG=$(oc get pod -l deploymentconfig=docker-registry --template='{{index .items 0 "status" "phase"}}')
    if [ "$?" -ne 0 -o "$REG" == "Error" ]; then
        echo "Registy pod is in error state..."
        exit 1
    fi
    if [ "$REG" == "Running" ]; then
        break
    fi
    sleep 5
done

while true; do
    V=$(oc get dc router --template='{{index .status "latestVersion"}}')
    P=$(oc get pod router-$V-deploy --template='{{index .status "phase"}}')
    if [ "$?" -eq 0 ]; then
        echo phase is $P for router deploy $V
        if [ "$P" == "Failed" ]; then
            echo "router deploy failed, try again"
            oc get pods
            oc rollout retry dc/router
            sleep 10
            continue
        fi
    fi
    REG=$(oc get pod -l deploymentconfig=router --template='{{index .items 0 "status" "phase"}}')
    if [ "$?" -eq 0 ]; then
        break
    fi
    oc get pods
    echo "Waiting for router pod"
    sleep 10
done


while true; do
    REG=$(oc get pod -l deploymentconfig=router --template='{{index .items 0 "status" "phase"}}')
    if [ "$?" -ne 0 -o "$REG" == "Error" ]; then
        echo "Router pod is in error state..."
        exit 1
    fi
    if [ "$REG" == "Running" ]; then
        break
    fi
    sleep 5
done

echo "Registry and router pods are okay"

if [ "$1" = "service-catalog" ]; then
    echo "Waiting for template-service-broker"

    while true; do
        status=$(oc get clusterservicebroker template-service-broker -o jsonpath='{.status.conditions[0].status}')
        if [ "$status" == "True" ]; then
            break
        fi
        sleep 5
    done
    
    echo "Waiting for openshift-automation-service-broker"
    while true; do
        status=$(oc get clusterservicebroker openshift-automation-service-broker -o jsonpath='{.status.conditions[0].status}')
        if [ "$status" == "True" ]; then
            break
        fi
        sleep 5
    done
fi