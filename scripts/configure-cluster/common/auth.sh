#!/usr/bin/env bash

set -e

HTPASSWD_FILE="./htpass"
HTPASSWD_SECRET="htpasswd-secret"

createhtpasswd() {
    # List of users to create
    USERS="developer odonoprojectattemptscreate odosingleprojectattemptscreate odologinnoproject odologinsingleproject1"
    USERPASS="password@123"
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
}

createclustersecret() {
    # Create secret in cluster, removing if it already exists
    oc get secret $HTPASSWD_SECRET -n openshift-config &> /dev/null
    if [ $? -eq 0 ]; then
        oc delete secret $HTPASSWD_SECRET -n openshift-config &> /dev/null
    fi
    oc create secret generic ${HTPASSWD_SECRET} --from-file=htpasswd=${HTPASSWD_FILE} -n openshift-config
}

configureclusterauth() {
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
}

waitforstablelogin() {
    OC_STABLE_LOGIN="false"
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
}

setupfirstproject() {
    # Setup project
    oc new-project myproject
    sleep 4
    oc version
    # Project list
    oc projects
}

createhtpasswd
createclustersecret
configureclusterauth
waitforstablelogin
setupfirstproject