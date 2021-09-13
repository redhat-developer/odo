#!/usr/bin/env sh
# this scripts logins as kubeadmin and removes the namespaces that are more than oneday old.
# this script is used in a kubernetes cronjob for PSI/IBM cloud openshift cluster.

# TEAM_PROVIDED is used to check if the namespace belongs to that team or not 
TEAM_PROVIDED=${TEAM_PROVIDED:-"odo"}

# login as kubeadmin
if [[ $CLUSTER_TYPE == "PSI" ]]; then
    ￼ #PSI cluster login
    ￼ oc login -u kubeadmin -p ${OCP4X_KUBEADMIN_PASSWORD} --insecure-skip-tls-verify ${OCP4X_API_URL}
else
    ￼ # Login to cluster in IBM Cloud using cluster API key
    ￼ oc login -u ${IBM_OC_LOGIN_USER} -p ${IBMC_ADMIN_OCLOGIN_APIKEY} --server=${IBMC_OCP47_SERVER}
fi

# PROJECT_AND_TIME var will contain namespace and date with time seperated with `|`
# PROJECT_AND_TIME is selected using labels app and team. e.g: app=test and team=odo
# eg. cmd-push-test157kgb|2021-08-17T12:40:20Z
PROJECT_AND_TIME=$(kubectl get namespace -o jsonpath='{range .items[*]}{.metadata.name}{"|"}{.metadata.creationTimestamp} {"\n"}{end}' | grep -v '^openshift\|^kube\|^default\|^ibm\|^calico\|^tigera\|^odo-operator-test')

for PROJECT in ${PROJECT_AND_TIME}; do
    IFS='|' read -r PRJ TIME <<<"$PROJECT" # seperate the Namespace and time of creation using IFS(Input Field Seperators) value `|`

    CONFIGMAP=$(kubectl get configmaps -n $PRJ -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}' | grep config-map-for-cleanup) # check if PROJECT contains configmap named 'config-map-for-cleanup'
                                                                                                                                    # if not then the value of CONFIGMAP will be empty
    if [[ ! -z $CONFIGMAP ]]; then

        TEAMANDTYPE=$(kubectl get configmaps $CONFIGMAP -n $PRJ -o jsonpath='{.data.team}{"|"}{.data.type}') # fetch team and tyoe(testing) data from the configmap
        IFS='|' read -r TEAM TYPE <<<"$TEAMANDTYPE"                                                          # seperate the TEAM and TYPE of creation using IFS(Input Field Seperators) value `|`

        if [[ $TYPE -eq "testing" ]] &&  [[ $TEAM -eq $TEAM_PROVIDED ]]; then # check if type if testing

            dtSec=$(date --date "${TIME}" +'%s')    # convert time in sec for namespace age
            taSec=$(date --date "1 days ago" +'%s') # convert time allowed for the namespace to be in the cluster

            if [ $dtSec -lt $taSec ]; then
                echo "INFO: Project="${PRJ}, "Date=" ${TIME}
                echo too old project : ${TIME}
                # delete namespace
                echo Delete project : ${PRJ}
                kubectl delete namespace ${PRJ} # delete namespace
                echo ---------
            fi
        fi
    fi
done
