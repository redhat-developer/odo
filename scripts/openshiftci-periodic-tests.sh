#!/bin/sh

# fail if some commands fails
set -e
# show commands
set -x

# check if devfiles are updated and if so, update and test them
check_devfile() {
    NOTEQUAL= false
    # Languages for which devfiles preset in examples dir
    LANGUAGES=('nodejs' 'python' 'springboot')

    for LANGUAGE in "${LANGUAGES[@]}"; do
        Example_devfile_path=./tests/examples/source/devfiles/$LANGUAGE/devfile-registry.yaml
        TEMPDIR=$(mktemp -d)
# download devfiles with odo
        if [[ $LANGUAGE == "springboot" ]]; then
            odo create java-$LANGUAGE language --context $TEMPDIR
        else
            odo create $LANGUAGE language --context $TEMPDIR
        fi

        Devfile_path=$TEMPDIR/devfile.yaml
# check if devfiles differ then the one in examples dir
        diff $Devfile_path $Example_devfile_path
        if [ $? != 0 ] ; then # $status is 0 if the files are the same
            echo "$LANGUAGE Devfile is not same as example"
            # copy if devfile differs
            cp $Devfile_path $Example_devfile_path
            NOTEQUAL= true
        fi

    done

    if [ $NOTEQUAL == true ]; then
        return -1
    else
        return 0
    fi
}

export CI="openshift"
make configure-installer-tests-cluster
make bin
mkdir -p $GOPATH/bin
make goget-ginkgo
export PATH="$PATH:$(pwd):$GOPATH/bin"
export CUSTOM_HOMEDIR=$ARTIFACT_DIR

# Copy kubeconfig to temporary kubeconfig file
# Read and Write permission to temporary kubeconfig file
TMP_DIR=$(mktemp -d)
cp $KUBECONFIG $TMP_DIR/kubeconfig
chmod 640 $TMP_DIR/kubeconfig
export KUBECONFIG=$TMP_DIR/kubeconfig

# Login as developer
oc login -u developer -p password@123

# Check login user name for debugging purpose
oc whoami

# Integration tests
check_devfile
if [[ $? == -1 ]]; then
    make test-integration-devfile || error=true
fi

make test-operator-hub || error=true

if [ $error ]; then
    exit -1
fi

cp -r reports tests/reports $ARTIFACT_DIR

oc logout
