# Save the logs from the file passed as parameter #1
# and send a message to GitHub PR using parameter #2 as name of test
# Env vars:
#  IBM_RESOURCE_GROUP: Resource group of the Cloud ObjectStorage
#  IBM_COS: Cloud Object Storage containing the bucket on which to save logs
#  IBM_BUCKET: Bucket name on which to save logs
save_logs() {
    LOGFILE="$1"
    NAME="$2"
    RESULT="$3"

    ansi2html <"/tmp/${LOGFILE}" >"/tmp/${LOGFILE}.html"
    ansi2txt <"/tmp/${LOGFILE}" >"/tmp/${LOGFILE}.txt"

    ibmcloud login --apikey "${API_KEY}"
    ibmcloud target -g "${IBM_RESOURCE_GROUP}"  -r "${IBM_REGION}"
    CRN=$(ibmcloud resource service-instance ${IBM_COS} --output json | jq -r .[0].guid)
    ibmcloud cos config crn --crn "${CRN}"

    ibmcloud cos upload --bucket "${IBM_BUCKET}" --key "${LOGFILE}.html" --file "/tmp/${LOGFILE}.html"
    ibmcloud cos upload --bucket "${IBM_BUCKET}" --key "${LOGFILE}.txt" --file "/tmp/${LOGFILE}.txt"

    BASE_URL="https://s3.${IBM_REGION}.cloud-object-storage.appdomain.cloud/${IBM_BUCKET}"
    if [[ $RESULT == "0" ]]; then
        STATUS="successfully"
    else
        STATUS="with errors"
    fi
    cat <<EOF | odo-robot -key-from-env-var ROBOT_KEY -pr-comment ${GIT_PR_NUMBER} -pipeline "${NAME}"
${NAME} on commit ${GIT_COMMIT} finished ${STATUS}.
View logs: [TXT](${BASE_URL}/${LOGFILE}.txt) [HTML](${BASE_URL}/${LOGFILE}.html)
EOF
}

# Delete namespaces from cluster containing a configmap named config-map-for-cleanup
# with values: "team: odo" and "type: testing"
cleanup_namespaces() {
    PROJECTS=$(kubectl get cm -A | grep config-map-for-cleanup | awk '{ print $1 }')
    for PROJECT in ${PROJECTS}; do
        TEAM=$(kubectl get configmaps config-map-for-cleanup -n ${PROJECT} -o jsonpath='{.data.team}')
        TYPE=$(kubectl get configmaps config-map-for-cleanup -n ${PROJECT} -o jsonpath='{.data.type}')
        if [[ "${TYPE}" -eq "testing" ]] && [[ "${TEAM}" -eq "odo" ]]; then
            kubectl delete namespace ${PROJECT} --wait=false
        fi
    done
}

skip_if_only() {
    echo "Checking if tests need to be executed..."
    NAMES=$(git diff --merge-base --name-only main)
    for change in ${NAMES}; do
        skip $change
        if [[ $? == 0 ]]; then
            return
        fi
    done
    echo   "  => Skipping the tests."
    exit 0
}

skip() {
    SKIP_IF_ONLY="docs/ CONTRIBUTING.md OWNERS README.md USAGE_DATA.md scripts/ .github/ .threatmodel/ ui/"
    change=$1
    for skip in ${SKIP_IF_ONLY}; do
        if [[ "${change}" == "${skip}"* ]]; then
            echo "  - ${change} is ${skip}*"
            return 1
        fi
    done
    echo "  - $change not in $SKIP_IF_ONLY"
    return 0
}

save_results() {
    (        
        cd .ibm/tools/tests-results
        echo -n $RESULTS_SERVICE_ACCOUNT > sa.json
        JUNIT="$1"
        LOGFILE="$2"
        NAME="$3"
        JOB="$4"
        BASE_URL="https://s3.${IBM_REGION}.cloud-object-storage.appdomain.cloud/${IBM_BUCKET}"

        GOOGLE_APPLICATION_CREDENTIALS=$PWD/sa.json go run main.go \
            --sheetId "${RESULTS_SHEET_ID}" \
            --junit "${JUNIT}" \
            --pr "${GIT_PR_NUMBER}" \
            --job "${JOB}" \
            --test "${NAME}" \
            --logfile "${BASE_URL}/${LOGFILE}.txt"
    )
}
