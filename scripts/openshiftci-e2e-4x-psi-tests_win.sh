#!/bin/sh

export SENDQUEUE=amqp.ci.queue.win.e2e.send
export SENDTOPIC=amqp.ci.topic.win.e2e.send

. scripts/openshift-e2e-4x-psi-tests.sh