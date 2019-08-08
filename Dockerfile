# This Dockerfile is used by CI to publish openshift/odo:binary-artifacts
# It builds an image containing the Mac, Win and Linux version of odo binary on the
# OpenShift golang image.

FROM registry.svc.ci.openshift.org/openshift/release:golang-1.11
RUN mkdir -p /go/src/github.com/openshift/odo
WORKDIR /go/src/github.com/openshift/odo
COPY . .
RUN go get github.com/mitchellh/gox &&\
    make cross
