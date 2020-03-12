# Dockerfile to bootstrap build and test in openshift-ci

FROM registry.svc.ci.openshift.org/openshift/release:golang-1.11

RUN mkdir -p ~/tmp/secret

RUN yum -y install make wget gcc git httpd-tools

RUN mkdir -p /tmp/secret