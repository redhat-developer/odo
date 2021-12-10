#!/bin/bash 
set -x
apt-get update -y
apt-get install -y sshpass

LOGFILE="pr-${GIT_PR_NUMBER}-windows-tests-${BUILD_NUMBER}"
export REPO=${REPO:-"https://github.com/anandrkskd/odo"}
#copy test script inside /tmp/
#sshpass -p $PASS scp -o StrictHostKeyChecking=no ./.ibm/pipelines/openshift-windows-tests.sh    Administrator@161.156.170.160:/tmp/openshift-windows-tests-${BUILD_NUMBER}.sh
sshpass -p $PASS scp -o StrictHostKeyChecking=no ./.ibm/pipelines/windows-test-script.ps1    Administrator@161.156.170.160:/tmp/windows-test-script.ps1

#(
#execute test from the test script
#sshpass -p $PASS ssh Administrator@161.156.170.160 -o StrictHostKeyChecking=no /tmp/openshift-windows-tests-${BUILD_NUMBER}.sh "${GIT_PR_NUMBER}" "${BUILD_NUMBER}" "${API_KEY}" "${IBM_OPENSHIFT_ENDPOINT}" "${LOGFILE}" "${REPO}"
sshpass -p $PASS ssh Administrator@161.156.170.160 -o StrictHostKeyChecking=no powershell /tmp/windows-test-script.ps1 "${GIT_PR_NUMBER}" "${BUILD_NUMBER}" "${API_KEY}" "${IBM_OPENSHIFT_ENDPOINT}" "${LOGFILE}" "${REPO}"
RESULT=$?
echo "RESULT: $RESULT"

# save log
source .ibm/pipelines/functions.sh
ibmcloud login --apikey "${API_KEY}" -r "${IBM_REGION}"
sshpass -p $PASS scp  -o StrictHostKeyChecking=no Administrator@161.156.170.160:/tmp/${LOGFILE}   /tmp/${LOGFILE}
save_logs "${LOGFILE}" "OpenShift Windows Tests"  $RESULT

# # cleanup 
sshpass -p $PASS ssh Administrator@161.156.170.160    -o StrictHostKeyChecking=no rm -rf /tmp/openshift-windows-tests-${BUILD_NUMBER}.sh

exit ${RESULT}