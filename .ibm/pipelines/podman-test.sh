#!/bin/bash

# Script to run podman test on IBM Cloud

# variables for tests
# NOTE: values of VSI is taken from `$odo/script/ansible/VM/vars.yaml`
LOGFILE="pr-${GIT_PR_NUMBER}-unit-tests-${BUILD_NUMBER}"
export REPO=${REPO:-"https://github.com/redhat-developer/odo"}
VSI_NAME="odo-podman-vsi-"${BUILD_NUMBER}
VSI_FIP_NAME="odo-podman-FIP-"${BUILD_NUMBER}
VSI_SPEC="bx2-2x8"
VSI_KEYS="automation-key"
VSI_SUBNET="odo-test-automation-subnet"
VSI_VPC="odo-test-automation-vpc"
IBM_REGION_ZONE="${IBM_REGION}-2"

source .ibm/pipelines/functions.sh

set -e
# create new vsi for podman test
VSI_NIC_ID=$(ibmcloud is inc $VSI_NAME $VSI_VPC $IBM_REGION_ZONE $VSI_SPEC $VSI_SUBNET --image $IMAGE --keys $VSI_KEYS) | jq -r '.primary_network_interface.id'
# create Floating IP for ssh
VSI_FIP=$(ibmcloud is floating-ip-reserve $VSI_FIP_NAME --nic $VSI_NIC_ID) | jq -r '.address'

#start testing
echo $SSH_KEY > key
#copy test script
scp -o StrictHostKeyChecking=no -i ./key ./.ibm/pipelines/podman-test-script.sh root@VSI_FIP:/tmp/podman-test-script.sh
# execute test script
ssh -i ./key root@VSI_FIP -o StrictHostKeyChecking=no bash /tmp/podman-test-script.sh "${BUILD_NUMBER}" "${LOGFILE}" "${REPO}" "${GIT_PR_NUMBER}" "${TEST_EXEC_NODES}"
#get and print test result
RESULT=$?
echo "RESULT: $RESULT"

# fetch logs from remote vsi
scp -o StrictHostKeyChecking=no -i ./key root@VSI_FIP:/tmp/${LOGFILE} /tmp/${LOGFILE}

RESULT=${PIPESTATUS[0]}

# delete VSI and related resource 
ibmcloud is floating-ip-release --force $VSI_FIP_NAME || error=true
ibmcloud is instance-delete --force $VSI_NAME || error=true
# add info if cleanup above commands failed 
if [ error == true ]; then
    echo -e "==================================\nCLEAN UP FAILED WITH ERROR\n==================================" >> "/tmp/${LOGFILE}"
fi

ibmcloud login --apikey "${API_KEY}" -r "${IBM_REGION}"
save_logs "${LOGFILE}" "podman Tests" ${RESULT}

exit ${RESULT}
