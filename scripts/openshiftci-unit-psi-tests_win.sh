#!/bin/sh

export SENDQUEUE=amqp.ci.queue.win.unit.send
export SENDTOPIC=amqp.ci.topic.win.unit.send
export SETUPSCRIPT=${SETUPSCRIPT:-"scripts/setup_script_unit.sh"}
export RUNSCRIPT=${RUNSCRIPT:-"scripts/run_script_unit.sh"}

. scripts/openshift-e2e-4x-psi-tests.sh