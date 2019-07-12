PROJECT := github.com/openshift/odo
GITCOMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
PKGS := $(shell go list  ./... | grep -v $(PROJECT)/vendor | grep -v $(PROJECT)/tests )
COMMON_FLAGS := -X $(PROJECT)/pkg/odo/cli/version.GITCOMMIT=$(GITCOMMIT)
BUILD_FLAGS := -ldflags="-w $(COMMON_FLAGS)"
DEBUG_BUILD_FLAGS := -ldflags="$(COMMON_FLAGS)"
FILES := odo dist
TIMEOUT ?= 1800s

# Env variable TEST_EXEC_NODES is used to pass spec execution type
# (parallel or sequential) for ginkgo tests. To run the specs sequentially use
# TEST_EXEC_NODES=1, otherwise by default the specs are run in parallel on 4 ginkgo test node.
# NOTE: Any TEST_EXEC_NODES value greater than one runs the spec in parallel
# on the same number of ginkgo test nodes.
TEST_EXEC_NODES ?= 4

# Slow spec threshold for ginkgo tests. After this time (in second), ginkgo marks test as slow
SLOW_SPEC_THRESHOLD := 120

CLUSTER_LOGIN_URL ?=
CLUSTER_USER_NAME ?= developer
CLUSTER_PASSWORD ?= developer

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
validate: gofmt check-vendor vet validate-vendor-licenses #lint

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

# Run unit tests and collect coverage
.PHONY: test-coverage
test-coverage:
	./scripts/generate-coverage.sh

# compile for multiple platforms
.PHONY: cross
cross:
	gox -osarch="darwin/amd64 linux/amd64 windows/amd64" -output="dist/bin/{{.OS}}-{{.Arch}}/odo" $(BUILD_FLAGS) ./cmd/odo/

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
	odo login -u $(CLUSTER_USER_NAME) -p $(CLUSTER_PASSWORD) $(CLUSTER_LOGIN_URL) --insecure-skip-tls-verify
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo generic" slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)
	odo logout

# Run json outout tests
.PHONY: test-json-format-output
test-json-format-output:
	odo login -u $(CLUSTER_USER_NAME) -p $(CLUSTER_PASSWORD) $(CLUSTER_LOGIN_URL) --insecure-skip-tls-verify
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odojsonoutput" slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)
	odo logout

# Run component e2e tests
.PHONY: test-cmp-e2e
test-cmp-e2e:
	odo login -u $(CLUSTER_USER_NAME) -p $(CLUSTER_PASSWORD) $(CLUSTER_LOGIN_URL) --insecure-skip-tls-verify
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoCmpE2e" slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)
	odo logout

# Run component subcommands e2e tests
.PHONY: test-cmp-sub-e2e
test-cmp-sub-e2e:
	odo login -u $(CLUSTER_USER_NAME) -p $(CLUSTER_PASSWORD) $(CLUSTER_LOGIN_URL) --insecure-skip-tls-verify
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoCmpE2e" slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)
	odo logout

# Run java e2e tests
.PHONY: test-java-e2e
test-java-e2e:
	odo login -u $(CLUSTER_USER_NAME) -p $(CLUSTER_PASSWORD) $(CLUSTER_LOGIN_URL) --insecure-skip-tls-verify
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoJavaE2e" slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)
	odo logout

# Run source e2e tests
.PHONY: test-source-e2e
test-source-e2e:
	odo login -u $(CLUSTER_USER_NAME) -p $(CLUSTER_PASSWORD) $(CLUSTER_LOGIN_URL) --insecure-skip-tls-verify
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoSourceE2e" slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)
	odo logout

# Run service catalog e2e tests
.PHONY: test-service-e2e
test-service-e2e:
	odo login -u $(CLUSTER_USER_NAME) -p $(CLUSTER_PASSWORD) $(CLUSTER_LOGIN_URL) --insecure-skip-tls-verify
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoServiceE2e" slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)
	odo logout

# Run link e2e tests
.PHONY: test-link-e2e
test-link-e2e:
	odo login -u $(CLUSTER_USER_NAME) -p $(CLUSTER_PASSWORD) $(CLUSTER_LOGIN_URL) --insecure-skip-tls-verify
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoLinkE2e" slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)
	odo logout

# Run watch e2e tests
.PHONY: test-watch-e2e
test-watch-e2e:
	odo login -u $(CLUSTER_USER_NAME) -p $(CLUSTER_PASSWORD) $(CLUSTER_LOGIN_URL) --insecure-skip-tls-verify
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoWatchE2e" slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)
	odo logout

# Run storage command integration tests
.PHONY: test-cmd-storage
test-cmd-storage:
	odo login -u $(CLUSTER_USER_NAME) -p $(CLUSTER_PASSWORD) $(CLUSTER_LOGIN_URL) --insecure-skip-tls-verify
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo storage command" slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)
	odo logout

# Run odo app cmd tests
.PHONY: test-cmd-app
test-cmd-app:
	odo login -u $(CLUSTER_USER_NAME) -p $(CLUSTER_PASSWORD) $(CLUSTER_LOGIN_URL) --insecure-skip-tls-verify
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoCmdApp" slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)
	odo logout

# Run login e2e tests
.PHONY: test-odo-login-e2e
test-odo-login-e2e:
	go test -v github.com/openshift/odo/tests/integration --ginkgo.focus="odoLoginE2e" -ginkgo.slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -ginkgo.v -timeout $(TIMEOUT)

# Run config tests
.PHONY: test-odo-config
test-odo-config:
	odo login -u $(CLUSTER_USER_NAME) -p $(CLUSTER_PASSWORD) $(CLUSTER_LOGIN_URL) --insecure-skip-tls-verify
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo config test" slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)
	odo logout

# Run all integration tests
.PHONY: test-integration
test-integration:
	odo login -u $(CLUSTER_USER_NAME) -p $(CLUSTER_PASSWORD) $(CLUSTER_LOGIN_URL) --insecure-skip-tls-verify
	go test -v github.com/openshift/odo/tests/integration -ginkgo.slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -ginkgo.v -timeout $(TIMEOUT)
	odo logout

# Run url integreation tests
.PHONY: test-odo-url-int
test-odo-url-int:
	odo login -u $(CLUSTER_USER_NAME) -p $(CLUSTER_PASSWORD) $(CLUSTER_LOGIN_URL) --insecure-skip-tls-verify
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odoURLIntegration" slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)
	odo logout

# Run push command e2e
.PHONY: test-cmd-push
test-cmd-push:
	ginkgo -v -nodes=$(TEST_EXEC_NODES) -focus="odo push command tests" slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/integration/ -timeout $(TIMEOUT)

# Run e2e test scenarios
.PHONY: test-e2e-scenarios
test-e2e-scenarios:
	odo login -u $(CLUSTER_USER_NAME) -p $(CLUSTER_PASSWORD) $(CLUSTER_LOGIN_URL) --insecure-skip-tls-verify
	ginkgo -v -nodes=$(TEST_EXEC_NODES) slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -randomizeAllSpecs  tests/e2escenarios/ -timeout $(TIMEOUT)
	odo logout

# this test shouldn't be in paralel -  it will effect the results
.PHONY: test-benchmark
test-benchmark:
	odo login -u $(CLUSTER_USER_NAME) -p $(CLUSTER_PASSWORD) $(CLUSTER_LOGIN_URL) --insecure-skip-tls-verify
	go test -v github.com/openshift/odo/tests/benchmark -timeout $(TIMEOUT)
	odo logout

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
