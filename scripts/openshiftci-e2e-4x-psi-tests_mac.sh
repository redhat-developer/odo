#!/bin/sh

export SENDQUEUE=amqp.ci.queue.mac.e2e.send
export SENDTOPIC=amqp.ci.topic.mac.e2e.send

. scripts/openshift-e2e-4x-psi-tests.sh