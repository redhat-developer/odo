#!/bin/bash
# Setup to find nessasary data from cluster setup
AUTH_DIR=${AUTH_DIR:-`pwd`}
KUBEADMIN_USER=${KUBEADMIN_USER:-"kubeadmin"}
KUBEADMIN_PASSWORD_FILE="${AUTH_DIR}/auth/kubeadmin-password"
HTPASSWD_FILE="./htpass"
USERPASS="developer"
HTPASSWD_SECRET="htpasswd-secret"

export KUBECONFIG="${AUTH_DIR}/auth/kubeconfig"

# List of users to create
USERS="developer odonoprojectattemptscreateproject odosingleprojectattemptscreate odologinnoproject odologinsingleproject1"

# Check if nessasary files exist
if [ ! -f $KUBEADMIN_PASSWORD_FILE ]; then
    echo "Could not find kubeadmin password file"
    exit 1
fi

if [ ! -f $KUBECONFIG ]; then
    echo "Could not find kubeadm password file"
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

# Create secret in cluster, removing if it already exists
oc get secret $HTPASSWD_SECRET
if [ $? -eq 0 ]; then
    oc delete secret $HTPASSWD_SECRET
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

sleep 20

# Login as developer and setup project
oc login -u developer -p $USERPASS
oc new-project myproject
sleep 5