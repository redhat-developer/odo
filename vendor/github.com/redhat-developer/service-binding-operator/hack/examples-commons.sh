#!/bin/bash
export EXAMPLE_NAMESPACE="service-binding-demo"

function pathname() {
  DIR="${1%/*}"
  (cd "$DIR" && echo "$(pwd -P)")
}
export HACK_YAMLS=${HACK_YAMLS:-$(pathname $0)/yamls}

## OpenShift Project/Namespace
function create_project {
    oc new-project $EXAMPLE_NAMESPACE
}

function delete_project {
    oc delete project $EXAMPLE_NAMESPACE --ignore-not-found=true
}

## Generic OperatorSources
function print_operator_source {
    REGISTRY_NAMESPACE=$1
    NAME=$2
    sed -e 's,REPLACE_OPSRC_NAME,'$NAME',g' $HACK_YAMLS/operator-source.template.yaml \
    | sed -e 's,REPLACE_REGISTRY_NAMESPACE,'$REGISTRY_NAMESPACE',g'
}

function install_operator_source {
    print_operator_source $@ | oc apply --wait -f -
}

function uninstall_operator_source {
    print_operator_source $@ | oc delete --wait --ignore-not-found=true -f -
}

## Generic operator Subscriptions

function get_current_csv {
    PACKAGE_NAME=$1
    CATALOG=$2
    CHANNEL=$3

    #oc get packagemanifest $PACKAGE_NAME -o jsonpath='{.status.channels[?(@.name == "'$CHANNEL'")].currentCSV}'
    oc get packagemanifests -o json | jq -r '.items[] | select(.metadata.name=="'$PACKAGE_NAME'") | select(.status.catalogSource=="'$CATALOG'").status.channels[] | select(.name=="'$CHANNEL'").currentCSV'
}

function print_operator_subscription {
    PACKAGE_NAME=$1
    OPSRC_NAME=$2
    CHANNEL=$3

    CSV_VERSION=$(get_current_csv $PACKAGE_NAME $OPSRC_NAME $CHANNEL)
    sed -e 's,REPLACE_CSV_VERSION,'$CSV_VERSION',g' $HACK_YAMLS/subscription.template.yaml \
    | sed -e 's,REPLACE_CHANNEL,'$CHANNEL',g' \
    | sed -e 's,REPLACE_OPSRC_NAME,'$OPSRC_NAME',g' \
    | sed -e 's,REPLACE_NAME,'$PACKAGE_NAME',g';
}

function install_operator_subscription {
    if [[ ! -z $(get_current_csv $1 $2 $3) ]]; then
        print_operator_subscription $1 $2 $3 | oc apply --wait -f -
    else
        echo "ERROR: packagemanifest $1 not found";
        exit 1;
    fi
}

function uninstall_operator_subscription {
    print_operator_subscription $1 $2 $3 | oc delete --ignore-not-found=true --wait -f -
}

function wait_for_packagemanifest {
    PACKAGE_NAME=$1
    OPSRC_NAME=$2
    CHANNEL=$3
    i=1
    while [[ -z "$(get_current_csv $1 $2 $3)" ]] && [ $i -le 10 ]; do
        echo "Waiting for package install to complete..."
        sleep 5
        i=$(($i+1))
    done
}

function uninstall_current_csv {
    PACKAGE_NAME=$1
    OPSRC_NAME=$2
    CHANNEL=$3

    oc delete csv $(get_current_csv $PACKAGE_NAME $OPSRC_NAME $CHANNEL) -n openshift-operators --ignore-not-found=true
}

## Backing DB (PostgreSQL) Operator

function install_postgresql_operator_source {
    OPSRC_NAMESPACE=pmacik
    OPSRC_NAME=db-operators
    PACKAGE_NAME=db-operators
    CHANNEL=stable

    install_operator_source $OPSRC_NAMESPACE $OPSRC_NAME
    wait_for_packagemanifest $PACKAGE_NAME $OPSRC_NAME $CHANNEL
}

function uninstall_postgresql_operator_source {
    OPSRC_NAMESPACE=pmacik
    OPSRC_NAME=db-operators

    uninstall_operator_source $OPSRC_NAMESPACE $OPSRC_NAME
}

function install_postgresql_operator_subscription {
    NAME=db-operators
    OPSRC_NAME=db-operators
    CHANNEL=stable

    install_operator_subscription $NAME $OPSRC_NAME $CHANNEL
}

function uninstall_postgresql_operator_subscription {
    NAME=db-operators
    OPSRC_NAME=db-operators
    CHANNEL=stable

    uninstall_operator_subscription $NAME $OPSRC_NAME $CHANNEL
    uninstall_current_csv $NAME $OPSRC_NAME $CHANNEL
}

function install_postgresql_db_instance {
    oc apply -f $HACK_YAMLS/postgresql-database.yaml
}

function uninstall_postgresql_db_instance {
    oc delete -f $HACK_YAMLS/postgresql-database.yaml --ignore-not-found=true
}

## Service Binding Operator

### Community operators
function install_service_binding_operator_subscription_community {
    NAME=service-binding-operator
    OPSRC_NAME=community-operators
    CHANNEL=alpha

    install_operator_subscription $NAME $OPSRC_NAME $CHANNEL
}

function uninstall_service_binding_operator_subscription_community {
    NAME=service-binding-operator
    OPSRC_NAME=community-operators
    CHANNEL=alpha

    uninstall_operator_subscription $NAME $OPSRC_NAME $CHANNEL
    uninstall_current_csv $NAME $OPSRC_NAME $CHANNEL
}

### Latest master
function install_service_binding_operator_source_master {
    OPSRC_NAMESPACE=redhat-developer
    OPSRC_NAME=redhat-developer-operators
    PACKAGE_NAME=service-binding-operator
    CHANNEL=alpha

    install_operator_source $OPSRC_NAMESPACE $OPSRC_NAME
    wait_for_packagemanifest $PACKAGE_NAME $OPSRC_NAME $CHANNEL
}

function uninstall_service_binding_operator_source_master {
    OPSRC_NAMESPACE=redhat-developer
    OPSRC_NAME=redhat-developer-operators

    uninstall_operator_source $OPSRC_NAMESPACE $OPSRC_NAME
}

function install_service_binding_operator_subscription_master {
    NAME=service-binding-operator
    OPSRC_NAME=redhat-developer-operators
    CHANNEL=alpha

    install_operator_subscription $NAME $OPSRC_NAME $CHANNEL
}

function uninstall_service_binding_operator_subscription_master {
    NAME=service-binding-operator
    OPSRC_NAME=redhat-developer-operators
    CHANNEL=alpha

    uninstall_operator_subscription $NAME $OPSRC_NAME $CHANNEL
    uninstall_current_csv $NAME $OPSRC_NAME $CHANNEL
}



## Serverless Operator

function install_serverless_operator_subscription {
    NAME=serverless-operator
    OPSRC_NAME=redhat-operators
    CHANNEL=techpreview

    install_operator_subscription $NAME $OPSRC_NAME $CHANNEL
}

function uninstall_serverless_operator_subscription {
    NAME=serverless-operator
    OPSRC_NAME=redhat-operators
    CHANNEL=techpreview

    uninstall_operator_subscription $NAME $OPSRC_NAME $CHANNEL
    uninstall_current_csv $NAME $OPSRC_NAME $CHANNEL
}

## Service Mesh Operator

function install_service_mesh_operator_subscription {
    NAME=servicemeshoperator
    OPSRC_NAME=redhat-operators
    CHANNEL='1.0'

    install_operator_subscription $NAME $OPSRC_NAME $CHANNEL
}

function uninstall_service_mesh_operator_subscription {
    NAME=servicemeshoperator
    OPSRC_NAME=redhat-operators
    CHANNEL=1.0

    uninstall_operator_subscription $NAME $OPSRC_NAME $CHANNEL
    uninstall_current_csv $NAME $OPSRC_NAME $CHANNEL
}

## Knative Serving
function install_knative_serving {
    echo " ==   -  STEP 1/5 -   ==  "
    echo "  -  Cleanup process  -  "
    oc delete -f $HACK_YAMLS/service-mesh-control-plane.yaml --ignore-not-found=true
    oc delete -f $HACK_YAMLS/service-mesh-member-roll.yaml --ignore-not-found=true
    oc delete -f $HACK_YAMLS/knative-serving.yaml --ignore-not-found=true
    sleep 5
    echo " ==   -  STEP 2/5 -   == "
    echo "  - SETTING ENVIRONMENT NAMESPACE CONTROLLERS  - "
    oc new-project serverless-test
    sleep 5
    oc new-project istio-system
    echo " - In the installation process!! THIS SHOULD TAKE 4-5 MINUTES  - "
    oc apply -f $HACK_YAMLS/service-mesh-control-plane.yaml
    sleep 300
    echo " - watch the progress of the pods during the installation process!! - "
    oc get pods -n istio-system
    echo " ==   -  STEP 3/5 -   ==  "
    echo " - Installing a ServiceMeshMemberRoll  - "
    oc apply -f $HACK_YAMLS/service-mesh-member-roll.yaml
    echo " ==   -  STEP 4/5 -   ==  "
    echo "  -  Installing a ServiceMeshMemberRoll  -  "
    sleep 15
    oc apply -f $HACK_YAMLS/knative-serving.yaml
    echo " ==   -  STEP 5/5 -   ==  "
    echo " -  Installing Knative Serving!! THIS SHOULD TAKE 1-2 MINUTES - "
    sleep 120
    oc get knativeserving/knative-serving -n knative-serving --template='{{range .status.conditions}}{{printf "%s=%s\n" .type .status}}{{end}}'
}

function uninstall_knative_serving {
    oc delete -f $HACK_YAMLS/service-mesh-control-plane.yaml
    oc delete -f $HACK_YAMLS/service-mesh-member-roll.yaml
    oc delete -f $HACK_YAMLS/knative-serving.yaml
}

## UBI Quarkus Native S2I Builder Image
function install_ubi_quarkus_native_s2i_builder_image {
    oc import-image quay.io/quarkus/ubi-quarkus-native-s2i:19.1.1 -n openshift --confirm
	oc patch is ubi-quarkus-native-s2i -n openshift -p '{"spec": {"tags": [{"name" : "19.1.1", "annotations": {"tags": "builder"}}]}}'
}
