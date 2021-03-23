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
export CI_FIREWALL_VERSION="valpha"
export AMQP_URI=${AMQP_URI:?"Please set AMQP_URI env with amqp uri or provide path of file containing it as ODO_RABBITMQ_AMQP_URL env"}
export JOB_NAME="odo-minishift-pr-tests"
export SENDQUEUE="amqp.ci.queue.minishift.send"
export SENDTOPIC="amqp.ci.topic.minishift.send"
export SETUP_SCRIPT="scripts/minikube-minishift-setup-env.sh"
export RUN_SCRIPT="scripts/minikube-minishift-all-tests.sh"
export TIMEOUT="4h00m"
export REPO_URL="https://github.com/openshift/odo"
# Extract PR NUMBER from prow job spec, which is injected by prow.
export TARGET="$(jq .refs.pulls[0].number <<< $(echo $JOB_SPEC))"
##### ci-firewall parameters end

echo "Getting ci-firewall, see https://github.com,/mohammedzee1000/ci-firewall"
# show commands
set -x

curl -kLO https://github.com/mohammedzee1000/ci-firewall/releases/download/$CI_FIREWALL_VERSION/ci-firewall-linux-amd64.tar.gz
tar -xzf ci-firewall-linux-amd64.tar.gz

./ci-firewall request --sendqueue $SENDQUEUE --sendtopic $SENDTOPIC --setupscript $SETUP_SCRIPT --runscript $RUN_SCRIPT --timeout $TIMEOUT
