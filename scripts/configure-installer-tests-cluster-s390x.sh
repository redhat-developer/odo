#!/bin/bash
set -x
# Setup to find nessasary data from cluster setup
## Constants
LIBDIR="./scripts/configure-cluster"
LIBCOMMON="$LIBDIR/common"
SETUP_OPERATORS="$LIBCOMMON/setup-operators.sh"
AUTH_SCRIPT="$LIBCOMMON/auth.sh"
# Overrideable information
DEFAULT_INSTALLER_ASSETS_DIR=${DEFAULT_INSTALLER_ASSETS_DIR:-$(pwd)}
KUBEADMIN_USER=${KUBEADMIN_USER:-"kubeadmin"}
KUBEADMIN_PASSWORD_FILE=${KUBEADMIN_PASSWORD_FILE:-"${DEFAULT_INSTALLER_ASSETS_DIR}/auth/kubeadmin-password"}

#CI_OPERATOR_HUB_PROJECT="ci-operator-hub-project"
# Exported to current env
export KUBECONFIG=${KUBECONFIG:-"${DEFAULT_INSTALLER_ASSETS_DIR}/auth/kubeconfig"}

# list of namespace to create
IMAGE_TEST_NAMESPACES="openjdk-11-rhel8 nodejs-12-rhel7 nodejs-12"

# Attempt resolution of kubeadmin, only if a CI is not set
if [ -z $CI ]; then
    # Check if nessasary files exist
    if [ ! -f $KUBEADMIN_PASSWORD_FILE ]; then
        echo "Could not find kubeadmin password file"
        exit 1
    fi

    if [ ! -f $KUBECONFIG ]; then
        echo "Could not find kubeconfig file"
        exit 1
    fi

    # Get kubeadmin password from file
    KUBEADMIN_PASSWORD=`cat $KUBEADMIN_PASSWORD_FILE`

    # Login as admin user
    oc login -u $KUBEADMIN_USER -p $KUBEADMIN_PASSWORD
else
    # Copy kubeconfig to temporary kubeconfig file
    # Read and Write permission to temporary kubeconfig file
    TMP_DIR=$(mktemp -d)
    cp $KUBECONFIG $TMP_DIR/kubeconfig
    chmod 640 $TMP_DIR/kubeconfig
    export KUBECONFIG=$TMP_DIR/kubeconfig
fi

# Setup the cluster for Operator tests

## Create a new namesapce which will be used for OperatorHub checks
#oc new-project $CI_OPERATOR_HUB_PROJECT
## Let developer user have access to the project
##oc adm policy add-role-to-user edit developer

#sh $SETUP_OPERATORS
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
oc apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/arch/s390x/official/nodejs/imagestreams/nodejs-rhel.json
sleep 5
oc delete istag nodejs:latest -n openshift
sleep 5
oc import-image nodejs:latest --from=registry.redhat.io/rhscl/nodejs-12-rhel7 --confirm -n openshift
sleep 5
oc annotate istag/nodejs:latest tags=builder -n openshift --overwrite
oc import-image java:8 --namespace=openshift --from=registry.redhat.io/redhat-openjdk-18/openjdk18-openshift --confirm
sleep 5
oc annotate istag/java:8 --namespace=openshift tags=builder --overwrite
oc import-image java:latest --namespace=openshift --from=registry.redhat.io/redhat-openjdk-18/openjdk18-openshift --confirm
sleep 5
oc annotate istag/java:latest --namespace=openshift tags=builder --overwrite
oc apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/arch/s390x/official/ruby/imagestreams/ruby-rhel.json
sleep 5
oc annotate istag/ruby:latest --namespace=openshift tags=builder --overwrite
oc import-image wildfly --confirm \--from docker.io/clefos/wildfly-120-centos7:latest --insecure -n openshift
sleep 5
oc annotate istag/wildfly:latest --namespace=openshift tags=builder --overwrite
oc apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/arch/s390x/official/nginx/imagestreams/nginx-rhel.json
sleep 5
oc annotate istag/nginx:latest --namespace=openshift tags=builder --overwrite
oc apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/community/dotnet/imagestreams/dotnet-centos.json
sleep 5
oc annotate istag/dotnet:latest --namespace=openshift tags=builder --overwrite
oc apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/arch/s390x/official/php/imagestreams/php-rhel.json
sleep 5
oc annotate istag/php:latest --namespace=openshift tags=builder --overwrite
oc apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/arch/s390x/official/python/imagestreams/python-rhel.json
sleep 5
oc annotate istag/python:latest --namespace=openshift tags=builder --overwrite

sh $AUTH_SCRIPT

KUBEADMIN_PASSWORD=`cat $KUBEADMIN_PASSWORD_FILE`
oc login -u $KUBEADMIN_USER -p $KUBEADMIN_PASSWORD &> /dev/null
oc get secret pull-secret -n openshift-config -o yaml | sed "s/openshift-config/myproject/g" | oc apply -f -

# Project list
oc projects

# KUBECONFIG cleanup only if CI is set
if [ ! -f $CI ]; then
    rm -rf $KUBECONFIG
    export KUBECONFIG=$ORIGINAL_KUBECONFIG
fi
