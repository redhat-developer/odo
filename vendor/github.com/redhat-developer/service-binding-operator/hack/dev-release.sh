#!/bin/bash

set -e
set -u

OPERATOR_SOURCE="./test/operator-hub/operator_source.yaml"
SUBSCRIPTION="./test/operator-hub/subscription.yaml"
INSTALL_PLAN_PRIOR=service-binding-operator.v${BUNDLE_VERSION}

sed -i -e "s,REPLACE_CSV_VERSION,service-binding-operator.v${BUNDLE_VERSION},g" ${SUBSCRIPTION}

kubectl apply -f ${OPERATOR_SOURCE}
./hack/check-operatorsource-pods.sh
RUNNING_STATUS=`kubectl get pods -n openshift-marketplace | grep "example-operators" | awk '{print $3}'`
if [ ${RUNNING_STATUS} != "Running" ] ; then
	echo "Operator marketplace pod is not running"
fi
# Subscribing to the operator
kubectl apply -f ${SUBSCRIPTION}
./hack/check-installplan.sh
INSTALL_PLAN=`kubectl get subscriptions service-binding-operator -n openshift-operators -o jsonpath='{.status.installPlanRef.name}'`
kubectl patch installplan ${INSTALL_PLAN} -n openshift-operators --type merge --patch '{"spec": {"approved": true}}'

./hack/check-crds.sh
./hack/check-csvs.sh
INSTALL_PLAN_PRIOR=service-binding-operator.v${BUNDLE_VERSION}
VERSION_NUMBER=`kubectl get csvs  -n=default -o jsonpath='{.items[*].spec.version}'`
if [ "${VERSION_NUMBER}" = "${BUNDLE_VERSION}" ] ; then
    echo -e "OLM Bundle Version validation succeeded \n ";
	kubectl get csvs -n=default -o jsonpath='{.items[*].metadata.annotations.alm-examples}' | cut -d "[" -f 2 | cut -d "]" -f 1 > output.json;
	kubectl apply -n=default -f ./output.json;
	if [ $? = 0 ] ; then
		echo "CSV alm example validation succeeded"
	fi
	if [ `kubectl get installplans -n=openshift-operators -o jsonpath='{.items[*].status.phase}'` = "Complete" ] ; then
		INSTALL_PLAN=`kubectl get installplans -n=openshift-operators -o jsonpath='{.items[*].spec.clusterServiceVersionNames[0]}'`
		if [ "${INSTALL_PLAN_PRIOR}" = "${INSTALL_PLAN}" ] ; then
			echo "Install Plan validation succeeded. OLM Bundle Validation succeeded"
		fi
	fi
	exit 0
else
	echo -e "OLM Bundle validation failed \n"
	echo "Version number: ${VERSION_NUMBER} \nBuild version: ${BUNDLE_VERSION}"
	exit 1
fi
