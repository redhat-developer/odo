PROJECT := github.com/openshift/odo
ifdef GITCOMMIT
        GITCOMMIT := $(GITCOMMIT)
else
        GITCOMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
endif
PKGS := $(shell go list  ./... | grep -v $(PROJECT)/vendor | grep -v $(PROJECT)/tests )
COMMON_FLAGS := -X $(PROJECT)/pkg/odo/cli/version.GITCOMMIT=$(GITCOMMIT)
BUILD_FLAGS := -ldflags="-w $(COMMON_FLAGS)"
DEBUG_BUILD_FLAGS := -ldflags="$(COMMON_FLAGS)"
FILES := odo dist
TIMEOUT ?= 7200s

# Env variable TEST_EXEC_NODES is used to pass spec execution type
# (parallel or sequential) for ginkgo tests. To run the specs sequentially use
# TEST_EXEC_NODES=1, otherwise by default the specs are run in parallel on 2 ginkgo test node.
# NOTE: Any TEST_EXEC_NODES value greater than one runs the spec in parallel
# on the same number of ginkgo test nodes.
TEST_EXEC_NODES ?= 2

# Slow spec threshold for ginkgo tests. After this time (in second), ginkgo marks test as slow
SLOW_SPEC_THRESHOLD := 120

default: bin

.PHONY: debug
debug:
	go build ${DEBUG_BUILD_FLAGS} cmd/odo/odo.go

.PHONY: bin
bin:
	go build ${BUILD_FLAGS} cmd/odo/odo.go

.PHONY: install
install:
	go install ${BUILD_FLAGS} ./cmd/odo/

# run all validation tests
.PHONY: validate
validate: gofmt check-vendor vet validate-vendor-licenses sec #lint

.PHONY: gofmt
gofmt:
	./scripts/check-gofmt.sh

.PHONY: check-vendor
check-vendor:
	./scripts/check-vendor.sh

.PHONY: validate-vendor-licenses
validate-vendor-licenses:
	wwhrd check -q
# golint errors are only recommendations
.PHONY: lint
lint:
	golint $(PKGS)

.PHONY: vet
vet:
	go vet $(PKGS)

.PHONY: sec
sec:
	@which gosec 2> /dev/null >&1 || { echo "gosec must be installed to lint code";  exit 1; }
	gosec -severity medium -confidence medium -exclude G304,G204 -quiet  ./...

.PHONY: clean
clean:
	@rm -rf $(FILES)

# install tools used for building, tests and  validations
.PHONY: goget-tools
goget-tools:
	go get -u github.com/Masterminds/glide
	# go get -u golang.org/x/lint/golint
	go get github.com/frapposelli/wwhrd
	go get -u github.com/onsi/ginkgo/ginkgo
	go get -u github.com/securego/gosec/cmd/gosec

# Run unit tests and collect coverage
.PHONY: test-coverage
test-coverage:
	./scripts/generate-coverage.sh

# compile for multiple platforms
.PHONY: cross
cross:
	./scripts/cross-compile.sh '$(COMMON_FLAGS)'

.PHONY: generate-cli-structure
generate-cli-structure:
	go run cmd/cli-doc/cli-doc.go structure

.PHONY: generate-cli-reference
generate-cli-reference:
	go run cmd/cli-doc/cli-doc.go reference > docs/cli-reference.adoc

# create gzipped binaries in ./dist/release/
# for uploading to GitHub release page
# run make cross before this!
.PHONY: prepare-release
prepare-release: cross
	./scripts/prepare-release.sh

.PHONY: configure-installer-tests-cluster
configure-installer-tests-cluster:
	. ./scripts/configure-installer-tests-cluster.sh

.PHONY: test
test:
	go test -race $(PKGS)

# -randomizeAllSpecs - If set, ginkgo will randomize all specs together.
# By default, ginkgo only randomizes the top level Describe, Context and When groups.

# Run generic integration tests
.PHONY: test-generic
test-generic:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo generic" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run odo login and logout tests
.PHONY: test-cmd-login-logout
test-cmd-login-logout:
	ginkgo -v -nodes=1 -focus="odo login and logout command tests" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/loginlogout/ -timeout $(TIMEOUT)

# Run link and unlink command tests
.PHONY: test-cmd-link-unlink
test-cmd-link-unlink:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo link and unlink command tests" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run odo service command tests
.PHONY: test-cmd-service
test-cmd-service:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo service command tests" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run odo project command tests
.PHONY: test-cmd-project
test-cmd-project:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo project command tests" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run odo app command tests
.PHONY: test-cmd-app
test-cmd-app:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo app command tests" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run odo component command tests
.PHONY: test-cmd-cmp
test-cmd-cmp:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo component command tests" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run odo component subcommands tests
.PHONY: test-cmd-cmp-sub
test-cmd-cmp-sub:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo sub component command tests" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run odo preference and config command tests
.PHONY: test-cmd-pref-config
test-cmd-pref-config:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo preference and config command tests" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run odo push command tests
.PHONY: test-cmd-push
test-cmd-push:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo push command tests" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run odo storage command tests
.PHONY: test-cmd-storage
test-cmd-storage:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo storage command tests" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run odo url command tests
.PHONY: test-cmd-url
test-cmd-url:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo url command tests" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run odo watch command tests
.PHONY: test-cmd-watch
test-cmd-watch:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo watch command tests" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run command's integration tests irrespective of service catalog status in the cluster.
# Service, link and login/logout command tests are not the part of this test run
.PHONY: test-integration
test-integration:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run command's integration tests which are depend on service catalog enabled cluster.
# Only service and link command tests are the part of this test run
.PHONY: test-integration-service-catalog
test-integration-service-catalog:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/servicecatalog/ -timeout $(TIMEOUT)

# Run core beta flow e2e tests
.PHONY: test-e2e-beta
test-e2e-beta:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo core beta flow" \
	-randomizeAllSpecs slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) tests/e2escenarios/ -timeout $(TIMEOUT)

# Run java e2e tests
.PHONY: test-e2e-java
test-e2e-java:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo java e2e tests" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/e2escenarios/ -timeout $(TIMEOUT)

# Run source e2e tests
.PHONY: test-e2e-source
test-e2e-source:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo source e2e tests" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/e2escenarios/ -timeout $(TIMEOUT)

# Run all e2e test scenarios
.PHONY: test-e2e-all
test-e2e-all:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -randomizeAllSpecs \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) tests/e2escenarios/ -timeout $(TIMEOUT)

# this test shouldn't be in paralel -  it will effect the results
.PHONY: test-benchmark
test-benchmark:
	ginkgo -v -nodes=1 slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) \
	tests/benchmark -timeout $(TIMEOUT)

# create deb and rpm packages using fpm in ./dist/pkgs/
# run make cross before this!
.PHONY: packages
packages:
	./scripts/create-packages.sh

# upload packages greated by 'make packages' to bintray repositories
# run 'make cross' and 'make packages' before this!
.PHONY: upload-packages
upload-packages:
	./scripts/upload-packages.sh

# Update vendoring
.PHONY: vendor-update
vendor-update:
	glide update --strip-vendor



.PHONY: openshiftci-presubmit-unittests
openshiftci-presubmit-unittests:
	./scripts/openshiftci-presubmit-unittests.sh
