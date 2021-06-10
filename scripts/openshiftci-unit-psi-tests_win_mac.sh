#!/bin/sh


case $1 in

  win)
    export SENDQUEUE=amqp.ci.queue.win.unit.send
    export SENDTOPIC=amqp.ci.topic.win.unit.send
    export SETUPSCRIPT="scripts/setup_script_unit.sh"
    export RUNSCRIPT="scripts/run_script_unit.sh"
    export JOB_NAME=odo-windows-unit-pr-build
    export SENDEXCHANGE=amqp.ci.exchange.win.unit.send 
    export TIMEOUT="20m"
    ;;

  mac)
    export SENDQUEUE=amqp.ci.queue.mac.unit.send
    export SENDTOPIC=amqp.ci.topic.mac.unit.send
    export SETUPSCRIPT="scripts/setup_script_unit.sh"
    export RUNSCRIPT="scripts/run_script_unit.sh"
    export JOB_NAME=odo-mac-unit-pr-build
    export SENDEXCHANGE=amqp.ci.exchange.mac.unit.send 
    export TIMEOUT="20m"
    ;;
esac

. scripts/openshiftci-e2e-4x-psi-tests.sh