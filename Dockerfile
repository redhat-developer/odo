# This Dockerfile is used by CI to publish odo binary artifacts.
# It builds an image containing the Mac, Win and Linux version of odo binary on the
# OpenShift golang image.

FROM registry.svc.ci.openshift.org/openshift/release:golang-1.11 AS builder
COPY . /go/src/github.com/openshift/odo
WORKDIR /go/src/github.com/openshift/odo
RUN go get github.com/mitchellh/gox &&\
    make cross

FROM registry.access.redhat.com/ubi7/ubi
COPY --from=builder /go/src/github.com/openshift/odo/dist/bin/darwin-amd64/odo /usr/share/openshift/odo/mac/odo
COPY --from=builder /go/src/github.com/openshift/odo/dist/bin/windows-amd64/odo.exe /usr/share/openshift/odo/windows/odo.exe
COPY --from=builder /go/src/github.com/openshift/odo/dist/bin/linux-amd64/odo /usr/share/openshift/odo/linux/odo
