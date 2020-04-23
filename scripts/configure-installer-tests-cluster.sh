#!/bin/bash
set -x
# Setup to find nessasary data from cluster setup
## Constants
HTPASSWD_FILE="./htpass"
USERPASS="developer"
HTPASSWD_SECRET="htpasswd-secret"
SETUP_OPERATORS="./scripts/setup-operators.sh"
# Overrideable information
DEFAULT_INSTALLER_ASSETS_DIR=${DEFAULT_INSTALLER_ASSETS_DIR:-$(pwd)}
KUBEADMIN_USER=${KUBEADMIN_USER:-"kubeadmin"}
KUBEADMIN_PASSWORD_FILE=${KUBEADMIN_PASSWORD_FILE:-"${DEFAULT_INSTALLER_ASSETS_DIR}/auth/kubeadmin-password"}
# Default values
OC_STABLE_LOGIN="false"
CI_OPERATOR_HUB_PROJECT="ci-operator-hub-project"
# Exported to current env
export KUBECONFIG=${KUBECONFIG:-"${DEFAULT_INSTALLER_ASSETS_DIR}/auth/kubeconfig"}

# List of users to create
USERS="developer odonoprojectattemptscreate odosingleprojectattemptscreate odologinnoproject odologinsingleproject1"

# Attempt resolution of kubeadmin, only if a CI is set
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
fi

# Setup the cluster for Operator tests

## Create a new namesapce which will be used for OperatorHub checks
oc new-project $CI_OPERATOR_HUB_PROJECT
## Let developer user have access to the project
oc adm policy add-role-to-user edit developer

sh $SETUP_OPERATORS
# OperatorHub setup complete

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
oc apply -n openshift -f https://raw.githubusercontent.com/openshift/library/master/arch/x86_64/community/wildfly/imagestreams/wildfly-centos7.json

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
    oc login -u developer -p $USERPASS &> /dev/null
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
sleep 4
oc version
