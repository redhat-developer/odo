#!/bin/sh
# yaml file must call this script with parameter values minikube or minishift
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
        export SENDQUEUE="amqp.ci.queue.minikube.send"
        export SENDTOPIC="amqp.ci.topic.minikube.send"
        export EXCHANGE="amqp.ci.exchange.minikube.send"
        export SETUP_SCRIPT="scripts/minikube-minishift-setup-env.sh minikube"
        export RUN_SCRIPT="scripts/minikube-minishift-all-tests.sh minikube"
        export TIMEOUT="4h00m"
        ;;
    *)
        echo "Must pass minikube as paramater"
        exit 1
        ;;
esac

export REPO_URL="https://github.com/redhat-developer/odo"
# Extract PR NUMBER from prow job spec, which is injected by prow.
export TARGET="$(jq .refs.pulls[0].number <<< $(echo $JOB_SPEC))"
##### ci-firewall parameters end

# The version of CI_FIREWALL TO USE
export CI_FIREWALL_VERSION="v0.1.2"

echo "Getting ci-firewall, see https://github.com,/mohammedzee1000/ci-firewall"
# show commands
set -x

curl -kLO https://github.com/mohammedzee1000/ci-firewall/releases/download/$CI_FIREWALL_VERSION/ci-firewall-linux-amd64.tar.gz
tar -xzf ci-firewall-linux-amd64.tar.gz

./ci-firewall request --sendqueue $SENDQUEUE --sendtopic $SENDTOPIC --sendexchange $EXCHANGE --setupscript "$SETUP_SCRIPT" --runscript "$RUN_SCRIPT" --jenkinsproject $JOB_NAME --timeout $TIMEOUT --mainbranch main
