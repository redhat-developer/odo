#!/bin/bash
# Setup to find nessasary data from cluster setup
## Constants
HTPASSWD_FILE="./htpass"
USERPASS="developer"
HTPASSWD_SECRET="htpasswd-secret"
# Overrideable information
DEFAULT_INSTALLER_ASSETS_DIR=${DEFAULT_INSTALLER_ASSETS_DIR:-$(pwd)}
KUBEADMIN_USER=${KUBEADMIN_USER:-"kubeadmin"}
KUBEADMIN_PASSWORD_FILE=${KUBEADMIN_PASSWORD_FILE:-"${DEFAULT_INSTALLER_ASSETS_DIR}/auth/kubeadmin-password"}
SLEEP_AFTER_SECRET_CREATION=${SLEEP_AFTER_SECRET_CREATION:-18}
# Exported to current env
export KUBECONFIG=${KUBECONFIG:-"${DEFAULT_INSTALLER_ASSETS_DIR}/auth/kubeconfig"}

# List of users to create
USERS="developer odonoprojectattemptscreateproject odosingleprojectattemptscreate odologinnoproject odologinsingleproject1"

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

# Login as admin user
oc login -u $KUBEADMIN_USER -p $KUBEADMIN_PASSWORD

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
  - name: htpassidp
    challenge: true
    login: true
    mappingMethod: claim
    type: HTPasswd
    htpasswd:
      fileData:
        name: ${HTPASSWD_SECRET}
EOF

# TODO : Find better way to check application of settings on cluster
sleep ${SLEEP_AFTER_SECRET_CREATION}

# Login as developer and setup project
oc login -u developer -p $USERPASS
oc new-project myproject
sleep 4
