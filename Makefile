PROJECT := github.com/openshift/odo
GITCOMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
PKGS := $(shell go list  ./... | grep -v $(PROJECT)/vendor | grep -v $(PROJECT)/tests )
COMMON_FLAGS := -X $(PROJECT)/pkg/odo/cli/version.GITCOMMIT=$(GITCOMMIT)
BUILD_FLAGS := -ldflags="-w $(COMMON_FLAGS)"
DEBUG_BUILD_FLAGS := -ldflags="$(COMMON_FLAGS)"
FILES := odo dist
TIMEOUT ?= 7200s

# Env variable TEST_EXEC_NODES is used to pass spec execution type
# (parallel or sequential) for ginkgo tests. To run the specs sequentially use
# TEST_EXEC_NODES=1, otherwise by default the specs are run in parallel on 4 ginkgo test node.
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
	go get -u github.com/mitchellh/gox
	go get github.com/frapposelli/wwhrd
	go get -u github.com/onsi/ginkgo/ginkgo
	go get -u github.com/securego/gosec/cmd/gosec

# Run unit tests and collect coverage
.PHONY: test-coverage
test-coverage:
	./scripts/generate-coverage.sh

# compile for linux platform
.PHONY: linux-amd64
linux-amd64:
	go build $(BUILD_FLAGS) -o dist/bin/linux-amd64/odo ./cmd/odo/

# compile for darwin platform
.PHONY: darwin-amd64
darwin-amd64:
	go build $(BUILD_FLAGS) -o dist/bin/darwin-amd64/odo ./cmd/odo/

# compile for windows platform
.PHONY: windows-amd64
windows-amd64:
	go build $(BUILD_FLAGS) -o "dist/bin/windows-amd64/odo.exe" ./cmd/odo/

# compile for multiple platforms
.PHONY: cross
cross:
	make linux-amd64
	make darwin-amd64
	make windows-amd64

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

# Run json outout tests
.PHONY: test-json-format-output
test-json-format-output:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odojsonoutput" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run component e2e tests
.PHONY: test-cmp-e2e
test-cmp-e2e:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoCmpE2e" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run component subcommands e2e tests
.PHONY: test-cmp-sub-e2e
test-cmp-sub-e2e:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoCmpSubE2e" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run java e2e tests
.PHONY: test-java-e2e
test-java-e2e:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoJavaE2e" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run source e2e tests
.PHONY: test-source-e2e
test-source-e2e:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoSourceE2e" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run service catalog e2e tests
.PHONY: test-service-e2e
test-service-e2e:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoServiceE2e" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/servicecatalog/ -timeout $(TIMEOUT)

# Run link e2e tests
.PHONY: test-link-e2e
test-link-e2e:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoLinkE2e" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/servicecatalog/ -timeout $(TIMEOUT)

# Run watch e2e tests
.PHONY: test-watch-e2e
test-watch-e2e:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoWatchE2e" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run storage command integration tests
.PHONY: test-cmd-storage
test-cmd-storage:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo storage command" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run odo app cmd tests
.PHONY: test-cmd-app
test-cmd-app:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoCmdApp" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run login e2e tests
# This test shouldn't run spec in paralel because it will break the test behaviour
# due to race condition in parallel run.
.PHONY: test-odo-login-e2e
test-odo-login-e2e:
	ginkgo -v -nodes=1 -focus="odoLoginE2e" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/loginlogout/ -timeout $(TIMEOUT)

# Run config tests
.PHONY: test-odo-config
test-odo-config:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo config test" \
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
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/servicecatalog -timeout $(TIMEOUT)

# Run url integreation tests
.PHONY: test-odo-url-int
test-odo-url-int:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoURLIntegration" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run push command e2e
.PHONY: test-cmd-push
test-cmd-push:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo push command tests" \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run e2e test scenarios
.PHONY: test-e2e-scenarios
test-e2e-scenarios:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -randomizeAllSpecs \
	slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) tests/e2escenarios/ -timeout $(TIMEOUT)

# this test shouldn't be in paralel -  it will effect the results
.PHONY: test-benchmark
test-benchmark:
	go test -v github.com/openshift/odo/tests/benchmark -timeout $(TIMEOUT)

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

.PHONY: openshiftci-presubmit-e2e
openshiftci-presubmit-e2e:
	./scripts/openshiftci-presubmit-e2e.sh

.PHONY: openshiftci-presubmit-unittests
openshiftci-presubmit-unittests:
	./scripts/openshiftci-presubmit-unittests.sh
