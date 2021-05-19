#!/bin/bash
set -x
# Setup to find nessasary data from cluster setup
## Constants
LIBDIR="./scripts/configure-cluster"
LIBCOMMON="$LIBDIR/common"
SETUP_OPERATORS="$LIBCOMMON/setup-operators.sh"
AUTH_SCRIPT="$LIBCOMMON/auth.sh"
KUBEADMIN_SCRIPT="$LIBCOMMON/kubeconfigandadmin.sh"

CI_OPERATOR_HUB_PROJECT="ci-operator-hub-project"

. $KUBEADMIN_SCRIPT

# Setup the cluster for Operator tests

# Create a new namesapce which will be used for OperatorHub checks
oc new-project $CI_OPERATOR_HUB_PROJECT
# Let developer user have access to the project
oc adm policy add-role-to-user edit developer

sh $SETUP_OPERATORS
# OperatorHub setup complete

# Workarounds - Note we should find better soulutions asap
# Missing wildfly in OpenShift Adding it manually to cluster Please remove once wildfly is again visible
oc apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/arch/x86_64/community/wildfly/imagestreams/wildfly-centos7.json

sh $AUTH_SCRIPT
