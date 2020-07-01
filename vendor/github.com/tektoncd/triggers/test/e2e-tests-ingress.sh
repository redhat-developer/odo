#!/usr/bin/env bash

# Copyright 2019 The Tekton Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

source $(dirname $0)/e2e-common.sh

trap "cleanup" SIGINT EXIT

function cleanup() {
  echo -e "\n\nCleaning up test resources"
  kubectl delete eventlistener ${EVENTLISTENER_NAME} || true
  kubectl delete secret ${CERTIFICATE_SECRET_NAME} || true
  kubectl delete taskrun ${INGRESS_TASKRUN_NAME} || true
  kubectl delete -f ${REPO_ROOT_DIR}/test/ingress || true
}

# Parameters: $1 - TaskRun name
function wait_until_taskrun_completed() {
  echo "Waiting until TaskRun $1 is completed"
  for i in {1..150}; do  # timeout after 10 minutes
    reason=$(kubectl get taskrun $1 -o=jsonpath='{.status.conditions[0].reason}')
    case ${reason} in
    "TaskRunValidationFailed"|"TaskRunTimeout")
      echo -e "\n\nERROR: ${reason}"
      kubectl get taskrun $1 -o=jsonpath='{.status.conditions[0].message}'
      exit 1
      ;;
    "Failed")
      echo -e "\n\nERROR: TaskRun Failed"
      echo "Grabbing container logs:"
      kubectl_debug="$(kubectl get taskrun $1 -o=jsonpath='{.status.conditions[0].message}' | sed 's/^.*\(kubectl.*\)/\1/')"
      eval ${kubectl_debug}
      exit 1
      ;;
    "Succeeded")
      return
      ;;
    esac
    sleep 2
  done
  echo -e "\n\nERROR: timeout waiting for taskrun successful completion\n"
  exit 1
}

# Parameters: $1 - pod name prefix
function wait_until_pod_started() {
  echo "Waiting until Pod $1 is running"
  for i in {1..150}; do  # timeout after 5 minutes
    kubectl get pod | grep $1 | grep "Running" && return || sleep 2
  done
  echo -e "\n\nERROR: timeout waiting for pod successful start\n"
  exit 1
}

# Parameters: $1 - EventListener name
function get_eventlistener_service() {
  echo "Getting ServiceName for EventListener $1"
  for i in {1..150}; do  # timeout after 5 minutes
    SERVICE_NAME=$(kubectl get eventlistener $1 -o=jsonpath='{.status.configuration.generatedName}')
    if [[ -z "$SERVICE_NAME" ]]
    then
      sleep 2
    else
      break
    fi
  done
}

# Parameters: $1 - Service name
function get_service_uid() {
  echo "Getting UID for Service $1"
  for i in {1..150}; do  # timeout after 5 minutes
    SERVICE_UID=$(kubectl get svc $1 -o=jsonpath='{.metadata.uid}')
    if [ -z "${SERVICE_UID}" ];
    then
        sleep 2
    else
      break
    fi
  done
}

# Parameters:
# $1 - Debug message before match
# $2, $3 - Expected, Actual value
function matchOrFail() {
  echo $1
  if [ $2 != $3 ];then
    echo "Match fail: expected: $2, actual: $3"
    exit 1
  fi
}

set -o errexit
set -o pipefail
set -x

# Apply ClusterRole/ClusterRoleBinding for default SA to run create Ingress Task
kubectl apply -f ${REPO_ROOT_DIR}/test/ingress
# Apply Ingress Task
kubectl apply -f ${REPO_ROOT_DIR}/docs/create-ingress.yaml
kubectl apply -f ${REPO_ROOT_DIR}/examples/triggerbindings/triggerbinding.yaml
kubectl apply -f ${REPO_ROOT_DIR}/examples/triggertemplates/triggertemplate.yaml

EVENTLISTENER_NAME="ingress-test-eventlistener"

# Create EventListener
cat << DONE | kubectl apply -f -
apiVersion: triggers.tekton.dev/v1alpha1
kind: EventListener
metadata:
  name: ${EVENTLISTENER_NAME}
spec:
  serviceAccountName: default
  triggers:
  - bindings:
    - name: pipeline-binding
    template:
      name: pipeline-template
DONE

INGRESS_TASKRUN_NAME="create-ingress-taskrun"
CERTIFICATE_KEY_PASSPHRASE="pass1"
CERTIFICATE_SECRET_NAME="secret1"
get_eventlistener_service ${EVENTLISTENER_NAME}
get_service_uid ${SERVICE_NAME}
EXTERNAL_DOMAIN="${SERVICE_NAME}.192.168.0.1.nip.io"

# Create Ingress using Ingress Task
cat << DONE | kubectl apply -f -
apiVersion: tekton.dev/v1beta1
kind: TaskRun
metadata:
  name: ${INGRESS_TASKRUN_NAME}
spec:
  taskRef:
    name: create-ingress
  params:
  - name: CertificateKeyPassphrase
    value: ${CERTIFICATE_KEY_PASSPHRASE}
  - name: CertificateSecretName
    value: ${CERTIFICATE_SECRET_NAME}
  - name: ExternalDomain
    value: ${EXTERNAL_DOMAIN}
  - name: Service
    value: ${SERVICE_NAME}
  - name: ServicePort
    value: "8080"
  - name: ServiceUID
    value: ${SERVICE_UID}
  timeout: 1000s
  serviceAccountName: default
DONE
wait_until_taskrun_completed ${INGRESS_TASKRUN_NAME}

# Check certificate
echo -e "Testing certificate"
crt=$(kubectl get secret ${CERTIFICATE_SECRET_NAME} -o=jsonpath='{.data.tls\.crt}')
echo $crt | base64 --decode | grep "\-\-\-\-\-BEGIN CERTIFICATE\-\-\-\-\-"
echo $crt | base64 --decode | grep "\-\-\-\-\-END CERTIFICATE\-\-\-\-\-"

key=$(kubectl get secret ${CERTIFICATE_SECRET_NAME} -o=jsonpath='{.data.tls\.key}')
echo $key | base64 --decode | grep "\-\-\-\-\-BEGIN RSA PRIVATE KEY\-\-\-\-\-"
echo $key | base64 --decode | grep "\-\-\-\-\-END RSA PRIVATE KEY\-\-\-\-\-"
echo -e "Certificate is OK"

# Check ingress
ingress_svc=$(kubectl get ingress ${SERVICE_NAME} -o=jsonpath='{.spec.rules[0].http.paths[0].backend.serviceName}')
ingress_rules_host=$(kubectl get ingress ${SERVICE_NAME} -o=jsonpath='{.spec.rules[0].host}')
ingress_tls_host=$(kubectl get ingress ${SERVICE_NAME} -o=jsonpath='{.spec.tls[0].hosts[0]}')
ingress_tls_secret=$(kubectl get ingress ${SERVICE_NAME} -o=jsonpath='{.spec.tls[0].secretName}')
ingress_owner_reference_uid=$(kubectl get ingress ${SERVICE_NAME} -o=jsonpath='{.metadata.ownerReferences[0].uid}')
matchOrFail "Checking the Ingress Service" ${SERVICE_NAME} ${ingress_svc}
matchOrFail "Checking the Ingress Rules Host" ${EXTERNAL_DOMAIN} ${ingress_rules_host}
matchOrFail "Checking the Ingress TLS Host" ${EXTERNAL_DOMAIN} ${ingress_tls_host}
matchOrFail "Checking the Ingress TLS Secret" ${CERTIFICATE_SECRET_NAME} ${ingress_tls_secret}
matchOrFail "Checking the Ingress OwnerReference" ${SERVICE_UID} ${ingress_owner_reference_uid}
