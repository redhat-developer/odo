#!/bin/sh
# fail if some commands fails
set -e
# Hide command that deals with secrets
set +x
if [[ -f $ODO_RABBITMQ_AMQP_URL ]]; then
    export AMQP_URI=$(cat $ODO_RABBITMQ_AMQP_URL)
fi

##### These are varialbes used by ci-firewall as one of the ways to get its parameters
# If AMQP_URI is not set by the time we reach here, show error message and exit.
export AMQP_URI=${AMQP_URI:?"Please set AMQP_URI env with amqp uri or provide path of file containing it as ODO_RABBITMQ_AMQP_URL env"}
case $1 in
    minikube)
        export JOB_NAME="odo-minikube-pr-build"
        ;;
    minishift)
        export JOB_NAME="odo-minishift-pr-tests"
        ;;
    *)
        echo "Must pass minikube or minishift as paramater"
        exit 1
        ;;
esac

export REPO_URL="https://github.com/openshift/odo"
# Extract PR NUMBER from prow job spec, which is injected by prow.
export TARGET="$(jq .refs.pulls[0].number <<< $(echo $JOB_SPEC))"
##### ci-firewall parameters end

# The version of CI_FIREWALL TO USE
export CI_FIREWALL_VERSION="valpha"

echo "Getting ci-firewall, see https://github.com,/mohammedzee1000/ci-firewall"
# show commands
set -x

curl -kLO https://github.com/mohammedzee1000/ci-firewall/releases/download/valpha/ci-firewall-linux-amd64.tar.gz
tar -xzf ci-firewall-linux-amd64.tar.gz

case $1 in
    minikube)
        ./ci-firewall request --sendqueue amqp.ci.queue.minikube.send --sendtopic amqp.ci.topic.minikube.send --runscript scripts/kubernetes-all-test.sh  --timeout 2h15m
        ;;
    minishift)
        ./ci-firewall request --sendqueue amqp.ci.queue.minishift.send --sendtopic amqp.ci.topic.minishift.send --setupscript minishift-setup-env.sh --runscript scripts/minishift-execute-test.sh  --timeout 4h00m
        ;;
esac
