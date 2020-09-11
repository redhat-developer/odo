#!/bin/bash
set -x
# Setup to find nessasary data from cluster setup
## Constants
HTPASSWD_FILE="./htpass"
USERPASS="developer"
HTPASSWD_SECRET="htpasswd-secret"
#SETUP_OPERATORS="./scripts/setup-operators.sh"
# Overrideable information
DEFAULT_INSTALLER_ASSETS_DIR=${DEFAULT_INSTALLER_ASSETS_DIR:-$(pwd)}
KUBEADMIN_USER=${KUBEADMIN_USER:-"kubeadmin"}
KUBEADMIN_PASSWORD_FILE=${KUBEADMIN_PASSWORD_FILE:-"${DEFAULT_INSTALLER_ASSETS_DIR}/auth/kubeadmin-password"}
# Default values
OC_STABLE_LOGIN="false"
#CI_OPERATOR_HUB_PROJECT="ci-operator-hub-project"
# Exported to current env
ORIGINAL_KUBECONFIG=${KUBECONFIG:-"${DEFAULT_INSTALLER_ASSETS_DIR}/auth/kubeconfig"}
export KUBECONFIG=$ORIGINAL_KUBECONFIG

# List of users to create
USERS="developer odonoprojectattemptscreate odosingleprojectattemptscreate odologinnoproject odologinsingleproject1"

# list of namespace to create
IMAGE_TEST_NAMESPACES="openjdk-11-rhel8 nodejs-12-rhel7"

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
oc adm policy add-role-to-user edit developer

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

# Remove existing htpasswd file, if any
if [ -f $HTPASSWD_FILE ]; then
    rm -rf $HTPASSWD_FILE
fi

# Set so first time -c parameter gets applied to htpasswd
HTPASSWD_CREATED=" -c "

# Create htpasswd entries for all listed users
for i in `echo $USERS`; do
    htpasswd -b $HTPASSWD_CREATED $HTPASSWD_FILE $i $USERPASS
    HTPASSWD_CREATED=""
done

# Workarounds - Note we should find better soulutions asap
## Missing wildfly in OpenShift Adding it manually to cluster Please remove once wildfly is again visible
#oc apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/arch/x86_64/community/wildfly/imagestreams/wildfly-centos7.json
oc import-image nodejs --from=registry.redhat.io/rhscl/nodejs-12-rhel7 --confirm -n openshift
sleep 15
oc annotate istag/nodejs:latest tags=builder -n openshift --overwrite
oc import-image java:8 --namespace=openshift --from=registry.redhat.io/redhat-openjdk-18/openjdk18-openshift --confirm
sleep 15
oc annotate istag/java:8 --namespace=openshift tags=builder --overwrite
oc import-image java:latest --namespace=openshift --from=registry.redhat.io/redhat-openjdk-18/openjdk18-openshift --confirm
sleep 15
oc annotate istag/java:latest --namespace=openshift tags=builder --overwrite
oc import-image ruby --from=registry.redhat.io/ubi8/ruby-26 -n openshift --confirm
sleep 15
oc annotate istag/ruby:latest --namespace=openshift tags=builder --overwrite
oc import-image wildfly:latest --confirm \--from docker.io/saomany/wildfly-120-centos7:ppc64le --insecure -n openshift
sleep 15
oc annotate istag/wildfly:latest --namespace=openshift tags=builder --overwrite
oc apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/arch/ppc64le/official/nginx/imagestreams/nginx-rhel.json
sleep 15
oc annotate istag/nginx:latest --namespace=openshift tags=builder --overwrite
oc apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/community/dotnet/imagestreams/dotnet-centos7.json
sleep 15
oc annotate istag/dotnet:latest --namespace=openshift tags=builder --overwrite
oc apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/arch/ppc64le/official/php/imagestreams/php-rhel.json
sleep 15
oc annotate istag/php:latest --namespace=openshift tags=builder --overwrite
oc import-image python:latest --from=registry.redhat.io/rhel8/python-27 --confirm -n openshift
sleep 15
oc annotate istag/python:latest --namespace=openshift tags=builder --overwrite

# Create secret in cluster, removing if it already exists
oc get secret $HTPASSWD_SECRET -n openshift-config &> /dev/null
if [ $? -eq 0 ]; then
    oc delete secret $HTPASSWD_SECRET -n openshift-config &> /dev/null
fi
oc create secret generic ${HTPASSWD_SECRET} --from-file=htpasswd=${HTPASSWD_FILE} -n openshift-config

# Upload htpasswd as new login config
oc apply -f - <<EOF
apiVersion: config.openshift.io/v1
kind: OAuth
metadata:
  name: cluster
spec:
  identityProviders:
  - name: htpassidp1
    challenge: true
    login: true
    mappingMethod: claim
    type: HTPasswd
    htpasswd:
      fileData:
        name: ${HTPASSWD_SECRET}
EOF

# Login as developer and check for stable server
for i in {1..40}; do
    # Try logging in as developer
    #oc login -u developer -p $USERPASS &> /dev/null
    KUBEADMIN_PASSWORD=`cat $KUBEADMIN_PASSWORD_FILE`
    oc login -u $KUBEADMIN_USER -p $KUBEADMIN_PASSWORD &> /dev/null
    if [ $? -eq 0 ]; then
        # If login succeeds, assume success
	    OC_STABLE_LOGIN="true"
        # Attempt failure of `oc whoami`
        for j in {1..25}; do
            oc whoami &> /dev/null
            if [ $? -ne 0 ]; then
                # If `oc whoami` fails, assume fail and break out of trying `oc whoami`
                OC_STABLE_LOGIN="false"
                break
            fi
            sleep 2
        done
        # If `oc whoami` never failed, break out trying to login again
        if [ $OC_STABLE_LOGIN == "true" ]; then
            break
        fi
    fi
    sleep 3
done

if [ $OC_STABLE_LOGIN == "false" ]; then
    echo "Failed to login as developer"
    exit 1
fi

# Setup project
oc new-project myproject
oc get secret pull-secret -n openshift-config -o yaml | sed "s/openshift-config/myproject/g" | oc apply -f -
oc adm policy add-role-to-user edit developer
sleep 4
oc version

# Project list
oc projects

# KUBECONFIG cleanup only if CI is set
if [ ! -f $CI ]; then
    rm -rf $KUBECONFIG
    export KUBECONFIG=$ORIGINAL_KUBECONFIG
fi
