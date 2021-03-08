#!/bin/sh

case $1 in

  win)
    export SENDQUEUE=amqp.ci.queue.win.e2e.send
    export SENDTOPIC=amqp.ci.topic.win.e2e.send
    export JOB_NAME=odo-windows-e2e-pr-build
    ;;

  mac)
    export SENDQUEUE=amqp.ci.queue.mac.e2e.send
    export SENDTOPIC=amqp.ci.topic.mac.e2e.send
    export JOB_NAME=odo-mac-e2e-pr-build
    ;;
esac

. scripts/openshift-e2e-4x-psi-tests.sh