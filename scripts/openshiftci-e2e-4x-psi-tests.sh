#!/bin/sh

# fail if some commands fails
set -e

# Hide command that deals with secrets
set +x
if [[ -f $ODO_RABBITMQ_AMQP_URL ]]; then
    export AMQP_URI=$(cat $ODO_RABBITMQ_AMQP_URL)
fi
export AMQP_URI=${AMQP_URI:?"Please set AMQP_URI env with amqp uri or provide path of file containing it as ODO_RABBITMQ_AMQP_URL env"}

# show commands
set -x

export EXCHANGE="amqp.ci.exchange.send"
export JOB_NAME="odo-pr-build"
export REPO_URL="https://github.com/openshift/odo"
# Extract PR NUMBER from prow job spec, which is injected by prow.
export TARGET="$(jq .refs.pulls[0].number <<< $(echo $JOB_SPEC))"
export CUSTOM_HOMEDIR=$ARTIFACT_DIR

##### ci-firewall parameters end
# The version of CI_FIREWALL TO USE
export CI_FIREWALL_VERSION="v0.1.1"

echo "Getting ci-firewall, see https://github.com,/mohammedzee1000/ci-firewall"
curl -kLO https://github.com/mohammedzee1000/ci-firewall/releases/download/valpha/ci-firewall-linux-amd64.tar.gz
tar -xzf ci-firewall-linux-amd64.tar.gz
./ci-firewall request --sendexchange $EXCHANGE --runscript scripts/run_all_tests.sh --timeout 4h00m --mainbranch main
