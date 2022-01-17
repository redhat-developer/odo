#!/bin/bash
set -x
# Setup to find nessasary data from cluster setup
## Constants
LIBDIR="./scripts/configure-cluster"
LIBCOMMON="$LIBDIR/common"
SETUP_OPERATORS="$LIBCOMMON/setup-operators.sh"
AUTH_SCRIPT="$LIBCOMMON/auth.sh"
KUBEADMIN_SCRIPT="$LIBCOMMON/kubeconfigandadmin.sh"

#CI_OPERATOR_HUB_PROJECT="ci-operator-hub-project"
# Exported to current env

# list of namespace to create
IMAGE_TEST_NAMESPACES="openjdk-11-rhel8 nodejs-12-rhel7 nodejs-12 openjdk-11 nodejs-14"

. $KUBEADMIN_SCRIPT

# Setup the cluster for Operator tests

## Create a new namesapce which will be used for OperatorHub checks
#oc new-project $CI_OPERATOR_HUB_PROJECT
## Let developer user have access to the project
##oc adm policy add-role-to-user edit developer

sh $SETUP_OPERATORS
# OperatorHub setup complete

# Create the namespace for e2e image test apply pull secret to the namespace
for i in `echo $IMAGE_TEST_NAMESPACES`; do
    # create the namespace
    oc new-project $i
    # Applying pull secret to the namespace which will be used for pulling images from authenticated registry
    oc get secret pull-secret -n openshift-config -o yaml | sed "s/openshift-config/$i/g" | oc apply -f -
    # Let developer user have access to the project
    oc adm policy add-role-to-user edit developer
done

#Missing required images in OpenShift and Adding it manually to cluster
oc --request-timeout 5m apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/arch/s390x/official/nodejs/imagestreams/nodejs-rhel.json
oc --request-timeout 5m delete istag nodejs:latest -n openshift
oc --request-timeout 5m import-image nodejs:latest --from=registry.redhat.io/rhscl/nodejs-12-rhel7 --confirm -n openshift
oc annotate istag/nodejs:latest tags=builder -n openshift --overwrite
oc --request-timeout 5m import-image java:8 --namespace=openshift --from=registry.redhat.io/redhat-openjdk-18/openjdk18-openshift --confirm
oc annotate istag/java:8 --namespace=openshift tags=builder --overwrite
oc --request-timeout 5m import-image java:latest --namespace=openshift --from=registry.redhat.io/redhat-openjdk-18/openjdk18-openshift --confirm
oc annotate istag/java:latest --namespace=openshift tags=builder --overwrite
oc --request-timeout 5m apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/arch/s390x/official/ruby/imagestreams/ruby-rhel.json
oc annotate istag/ruby:latest --namespace=openshift tags=builder --overwrite
oc --request-timeout 5m import-image wildfly --confirm \--from docker.io/clefos/wildfly-120-centos7:latest --insecure -n openshift
oc annotate istag/wildfly:latest --namespace=openshift tags=builder --overwrite
oc --request-timeout 5m apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/arch/s390x/official/nginx/imagestreams/nginx-rhel.json
oc annotate istag/nginx:latest --namespace=openshift tags=builder --overwrite
oc --request-timeout 5m apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/community/dotnet/imagestreams/dotnet-centos.json
oc annotate istag/dotnet:latest --namespace=openshift tags=builder --overwrite
oc --request-timeout 5m apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/arch/s390x/official/php/imagestreams/php-rhel.json
oc annotate istag/php:latest --namespace=openshift tags=builder --overwrite
oc --request-timeout 5m apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/arch/s390x/official/python/imagestreams/python-rhel.json
oc annotate istag/python:latest --namespace=openshift tags=builder --overwrite

sh $AUTH_SCRIPT

# Project list
oc projects

