#!/usr/bin/env bash

###########################################
# Setup Kubeconfig and login as kubeadmin #
###########################################

setup_kubeadmin() {
    export DEFAULT_INSTALLER_ASSETS_DIR=${DEFAULT_INSTALLER_ASSETS_DIR:-$(pwd)}
    export KUBEADMIN_USER=${KUBEADMIN_USER:-"kubeadmin"}
    export KUBEADMIN_PASSWORD_FILE=${KUBEADMIN_PASSWORD_FILE:-"${DEFAULT_INSTALLER_ASSETS_DIR}/auth/kubeadmin-password"}
    if [[ -z $CI && ! -f $KUBEADMIN_PASSWORD_FILE ]]; then
        echo "Could not find kubeadmin password file"
        exit 1
    fi
    KUBEADMIN_PASSWORD=`cat $KUBEADMIN_PASSWORD_FILE`
    # Login as admin user
    oc login -u $KUBEADMIN_USER -p $KUBEADMIN_PASSWORD
}

setup_kubeconfig() {
    export ORIGINAL_KUBECONFIG=${KUBECONFIG:-"${DEFAULT_INSTALLER_ASSETS_DIR}/auth/kubeconfig"}
    export KUBECONFIG=$ORIGINAL_KUBECONFIG
    if [[ -z $CI && ! -f $KUBECONFIG ]]; then
        echo "Could not find kubeconfig file"
        exit 1
    fi
    if [[ $CI == "openshift" && ! -z $KUBECONFIG ]]; then
        # Copy kubeconfig to temporary kubeconfig file
        # Read and Write permission to temporary kubeconfig file
        TMP_DIR=$(mktemp -d)
        cp $KUBECONFIG $TMP_DIR/kubeconfig
        chmod 640 $TMP_DIR/kubeconfig
        export KUBECONFIG=$TMP_DIR/kubeconfig
    fi
}

setup_kubeconfig
setup_kubeadmin
echo "Call `scripts/configure-cluster/common/kubeconfigreset.sh` separately if you want to reset the kubeconfig"