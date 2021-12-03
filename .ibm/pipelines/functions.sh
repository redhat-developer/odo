# Save the logs from the file passed as parameter #1
# and send a message to GitHub PR using parameter #2 as name of test
# Env vars:
#  IBM_RESOURCE_GROUP: Resource group of the Cloud ObjectStorage
#  IBM_COS: Cloud Object Storage containing the bucket on which to save logs
#  IBM_BUCKET: Bucket name on which to save logs
save_logs() {
    LOGFILE="$1"
    NAME="$2"
    apt update
    apt install jq colorized-logs --yes

    ansi2html < "/tmp/${LOGFILE}" > "/tmp/${LOGFILE}.html"
    ansi2txt < "/tmp/${LOGFILE}" > "/tmp/${LOGFILE}.txt"

    ibmcloud target -g "${IBM_RESOURCE_GROUP}"
    CRN=$(ibmcloud resource service-instance ${IBM_COS} --output json | jq -r .[0].guid)
    ibmcloud cos config crn --crn "${CRN}"

    ibmcloud cos upload --bucket "${IBM_BUCKET}" --key "${LOGFILE}.html" --file "/tmp/${LOGFILE}.html"
    ibmcloud cos upload --bucket "${IBM_BUCKET}" --key "${LOGFILE}.txt" --file "/tmp/${LOGFILE}.txt"

    echo -n ${GITHUB_TOKEN} | gh auth login --with-token
    BASE_URL="https://s3.${IBM_REGION}.cloud-object-storage.appdomain.cloud/${IBM_BUCKET}"
    cat <<EOF | gh pr comment ${GIT_PR_NUMBER} --body-file -
${NAME} finished.
View logs: [TXT](${BASE_URL}/${LOGFILE}.txt) [HTML](${BASE_URL}/${LOGFILE}.html)
EOF
}
