#!/bin/bash

###################################################
# This script is used to run the test on windows
# using the IBM DevOps Services.
#

set -x

LOGFILE="pr-${GIT_PR_NUMBER}-windows-tests-${BUILD_NUMBER}"
export REPO=${REPO:-"https://github.com/redhat-developer/odo"}
#copy test script inside /tmp/
sshpass -p $WINDOWS_PASSWORD scp -o StrictHostKeyChecking=no ./.ibm/pipelines/windows-test-script.ps1 Administrator@$WINDOWS_IP:/tmp/windows-test-script.ps1
sshpass -p $WINDOWS_PASSWORD scp -ro StrictHostKeyChecking=no  /go/odo_1  Administrator@$WINDOWS_IP:$BUILD_NUMBER

#execute test from the test script
export TEST_EXEC_NODES=${TEST_EXEC_NODES:-"16"}
sshpass -p $WINDOWS_PASSWORD ssh Administrator@$WINDOWS_IP -o StrictHostKeyChecking=no powershell /tmp/windows-test-script.ps1 "${GIT_PR_NUMBER}" "${BUILD_NUMBER}" "${API_KEY_QE}" "${IBM_OPENSHIFT_ENDPOINT}" "${LOGFILE}" "${REPO}" "${CLUSTER_ID}" "${TEST_EXEC_NODES}"
RESULT=$?
echo "RESULT: $RESULT"

# save log
source .ibm/pipelines/functions.sh
ibmcloud login --apikey "${API_KEY}" -r "${IBM_REGION}"
sshpass -p $WINDOWS_PASSWORD scp -o StrictHostKeyChecking=no Administrator@$WINDOWS_IP:~/AppData/Local/Temp/${LOGFILE} /tmp/${LOGFILE}
save_logs "${LOGFILE}" "Windows Tests (OCP)" $RESULT

# cleanup
sshpass -p $WINDOWS_PASSWORD ssh Administrator@$WINDOWS_IP -o StrictHostKeyChecking=no rm -rf /tmp/windows-test-script.ps1

exit ${RESULT}
