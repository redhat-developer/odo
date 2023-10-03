#!/bin/bash

###################################################
# This script is used to run the test on windows
# using the IBM DevOps Services.
#

source .ibm/pipelines/functions.sh

skip_if_only

LOGFILE="pr-${GIT_PR_NUMBER}-windows-tests-${BUILD_NUMBER}"
TEST_NAME="Windows Tests (OCP)"

export REPO=${REPO:-"https://github.com/redhat-developer/odo"}
#copy test script inside /tmp/
sshpass -p $WINDOWS_PASSWORD scp -o StrictHostKeyChecking=no ./.ibm/pipelines/windows-test-script.ps1 Administrator@$WINDOWS_IP:/tmp/windows-test-script.ps1

#execute test from the test script
export TEST_EXEC_NODES=${TEST_EXEC_NODES:-"16"}
sshpass -p $WINDOWS_PASSWORD ssh Administrator@$WINDOWS_IP -o StrictHostKeyChecking=no powershell /tmp/windows-test-script.ps1 "${GIT_PR_NUMBER}" "${BUILD_NUMBER}" "${API_KEY_QE}" "${IBM_OPENSHIFT_ENDPOINT}" "${LOGFILE}" "${REPO}" "${CLUSTER_ID}" "${TEST_EXEC_NODES}" "${DEVFILE_REGISTRY}" "${SKIP_SERVICE_BINDING_TESTS}"
RESULT=$?
echo "RESULT: $RESULT"

# save log
ibmcloud login --apikey "${API_KEY}" -r "${IBM_REGION}"
sshpass -p $WINDOWS_PASSWORD scp -o StrictHostKeyChecking=no Administrator@$WINDOWS_IP:~/AppData/Local/Temp/${LOGFILE} /tmp/${LOGFILE}
save_logs "${LOGFILE}" "${TEST_NAME}" $RESULT

# save results
sshpass -p $WINDOWS_PASSWORD scp -o StrictHostKeyChecking=no Administrator@$WINDOWS_IP:~/AppData/Local/Temp/test-integration-nc.xml /tmp/
sshpass -p $WINDOWS_PASSWORD scp -o StrictHostKeyChecking=no Administrator@$WINDOWS_IP:~/AppData/Local/Temp/test-integration.xml /tmp/
sshpass -p $WINDOWS_PASSWORD scp -o StrictHostKeyChecking=no Administrator@$WINDOWS_IP:~/AppData/Local/Temp/test-e2e.xml /tmp/
save_results "/tmp/test-integration-nc.xml" "${LOGFILE}" "${TEST_NAME}" "${BUILD_NUMBER}"
save_results "/tmp/test-integration.xml" "${LOGFILE}" "${TEST_NAME}" "${BUILD_NUMBER}"
save_results "/tmp/test-e2e.xml" "${LOGFILE}" "${TEST_NAME}" "${BUILD_NUMBER}"

# cleanup
sshpass -p $WINDOWS_PASSWORD ssh Administrator@$WINDOWS_IP -o StrictHostKeyChecking=no rm -rf /tmp/windows-test-script.ps1

exit ${RESULT}
