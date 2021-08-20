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
PROJECT_AND_TIME=$(kubectl get projects -o jsonpath='{range .items[*]}{.metadata.name}{"|"}{.metadata.creationTimestamp} {"\n"}{end}' | grep -v '^openshift\|^kube\|^default\|^ibm\|^calico\|^tigera\|^odo-operator-test')

for PROJECT in ${PROJECT_AND_TIME}; do
    IFS='|' read -r PRJ TIME <<<"$PROJECT"                          # seperate the Namespace and time of creation using IFS(Input Field Seperators) value `|`
    echo "INFO: Project="${PRJ}, "Date=" ${TIME}

    dtSec=$(date --date "${TIME}" +'%s')                            # convert time in sec for namespace age
    taSec=$(date --date "1 days ago" +'%s')                         # convert time allowed for the namespace to be in the cluster

    if [ $dtSec -lt $taSec ]; then
        echo too old project : ${TIME}
        # delete namespace
        oc delete project ${PRJ}
        echo ---------
    fi
done
