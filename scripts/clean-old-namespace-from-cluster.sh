#!/usr/bin/env sh
# this scripts logins as kubeadmin and removes the namespaces that are more than oneday old.
# this script is used in a kubernetes cronjob for PSI/IBM cloud openshift cluster.  

# login as kubeadmin
if [[ $CLUSTER_TYPE == "PSI" ]]; then
    ￼ #PSI cluster login
    ￼ oc login -u kubeadmin -p ${OCP4X_KUBEADMIN_PASSWORD} --insecure-skip-tls-verify ${OCP4X_API_URL}
else
    ￼ # Login to IBM Cloud using service account API Key
    ￼ ibmcloud login --apikey $IBMC_OCP47_APIKEY -a cloud.ibm.com -r eu-de -g "Developer CI and QE"
    ￼
    ￼ # Login to cluster in IBM Cloud using cluster API key
    ￼ oc login --token=$IBMC_OCLOGIN_APIKEY --server=$IBMC_OCP47_SERVER
fi

# PROJECT_AND_TIME var will contain namespace and date with time seperated with `|`
# PROJECT_AND_TIME doesn't contain openshift/ibm/kube namespace
# eg. cmd-push-test157kgb|2021-08-17T12:40:20Z
PROJECT_AND_TIME=$(kubectl get projects -o jsonpath='{range .items[*]}{.metadata.name}{"|"}{.metadata.creationTimestamp} {"\n"}{end}' | grep -v '^openshift\|^kube\|^default\|^ibm\|^calico\|^tigera')

for PROJECT in ${PROJECT_AND_TIME}; do
    IFS='|'
    read -ra ADDR <<<"$PROJECT"
    echo "INFO: Project="${ADDR[0]}, "Date=" ${ADDR[1]}

    datetime=${ADDR[1]}
    timeago='1 days ago'

    dtSec=$(date --date "$datetime" +'%s')
    taSec=$(date --date "$timeago" +'%s')

    echo "INFO: dtSec=$dtSec, taSec=$taSec" >&2

    if [ $dtSec -lt $taSec ]; then
        echo too old project : ${ADDR[0]}

    # delete namespace 
    oc delete project ${ADDR[0]}

    fi
    IFS=' '
done
