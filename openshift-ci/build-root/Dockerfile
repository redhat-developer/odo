# Dockerfile to bootstrap build and test in openshift-ci

FROM registry.svc.ci.openshift.org/openshift/release:golang-1.12

RUN yum -y install make wget gcc git httpd-tools

RUN mkdir -p /tmp/secret
