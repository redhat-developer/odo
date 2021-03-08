#!/bin/sh

export SENDQUEUE=amqp.ci.queue.mac.unit.send
export SENDTOPIC=amqp.ci.topic.mac.unit.send
export SETUPSCRIPT=${SETUPSCRIPT:-"scripts/setup_script_unit.sh"}
export RUNSCRIPT=${RUNSCRIPT:-"scripts/run_script_unit.sh"}

. scripts/openshift-e2e-4x-psi-tests.sh