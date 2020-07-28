#!/bin/bash

set -e
set -u

SUBSCRIPTION="./test/operator-hub/subscription2.yaml"
BUNDLE_VERSION=`curl -s https://raw.githubusercontent.com/operator-framework/community-operators/master/community-operators/service-binding-operator/service-binding-operator.package.yaml | ./out/venv3/bin/yq -r '.channels[] | select (.name == '\"$CHANNEL\"') | .currentCSV | sub("service-binding-operator.v"; "") '`
INSTALL_PLAN_PRIOR=service-binding-operator.v${BUNDLE_VERSION}

# Subscribing to the operator
kubectl apply -f ${SUBSCRIPTION}
RUNNING_STATUS=`kubectl get pods -n openshift-marketplace | grep "community-operators" | awk '{print $3}'`
if [[ ${RUNNING_STATUS} = "Running" ]] ; then
	echo "Operator marketplace pod is running"
fi
./hack/check-crds.sh
./hack/check-csvs.sh
VERSION_NUMBER=`kubectl get csvs service-binding-operator.v${BUNDLE_VERSION} -n=default -o jsonpath='{.spec.version}'`
if [ "${VERSION_NUMBER}" = "${BUNDLE_VERSION}" ] ; then
    echo -e "OLM Bundle Version validation succeeded \n ";
	rm -f ./output.json;
	kubectl get csvs service-binding-operator.v${BUNDLE_VERSION} -n=default -o jsonpath='{.metadata.annotations.alm-examples}' | cut -d "[" -f 2 | cut -d "]" -f 1 > output.json;
	kubectl apply -n=default -f ./output.json;
	rm -f ./output.json;
	if [ $? = 0 ] ; then
		echo "CSV alm example validation succeeded"
	fi
    install_plan_name=`kubectl get subscriptions service-binding-operator -n openshift-operators -o jsonpath='{.status.installPlanRef.name}'`
	if [ `kubectl get installplans ${install_plan_name} -n=openshift-operators -o jsonpath='{.status.phase}'` == "Complete" ] ; then
		INSTALL_PLAN=`kubectl get installplans ${install_plan_name} -n=openshift-operators -o jsonpath='{.spec.clusterServiceVersionNames[0]}'`
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
