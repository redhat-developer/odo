#!/usr/bin/env sh

if [[ $CLUSTER_TYPE == "PSI" ]]; then
    ￼ #PSI cluster login
    ￼ oc login -u kubeadmin -p ${OCP4X_KUBEADMIN_PASSWORD} --insecure-skip-tls-verify ${OCP4X_API_URL}
else
    ￼ # Login to IBM Cloud using service account API Key
    ￼ ##ibmcloud login --apikey $IBMC_OCP47_APIKEY -a cloud.ibm.com -r eu-de -g "Developer CI and QE"
    ￼
    ￼ # Login to cluster in IBM Cloud using cluster API key
    ￼ ##oc login --token=$IBMC_OCLOGIN_APIKEY --server=$IBMC_OCP47_SERVER
fi

PROJECT_AND_TIME=$(kubectl get projects -o jsonpath='{range .items[*]}{.metadata.name}{"|"}{.metadata.creationTimestamp} {"\n"}{end}')

for PROJECT in ${PROJECT_AND_TIME}; do
    IFS='|'
    read -ra ADDR <<<"$PROJECT"
    echo "INFO: Project="${ADDR[0]}, "Date=" ${ADDR[1]}
    if [[ ${ADDR[0]} == "openshift"* || ${ADDR[0]} == "kube"* || ${ADDR[0]} == "default" ]]; then
        echo "Skipped"
    else
        datetime=${ADDR[1]}
        timeago='2 days ago'

        dtSec=$(date --date "$datetime" +'%s') # For "now", use $(date +'%s')
        taSec=$(date --date "$timeago" +'%s')

        echo "INFO: dtSec=$dtSec, taSec=$taSec" >&2

        if [ $dtSec -lt $taSec ]; then 
        echo too old project : ${ADDR[0]}
        oc delete project ${ADDR[0]}
        fi
        [ $dtSec -gt $taSec ] && echo new
        IFS=' '
    fi
done
