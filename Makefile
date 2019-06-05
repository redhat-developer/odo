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

# Run generic integration tests
.PHONY: test-generic
test-generic:
	go test -v github.com/openshift/odo/tests/integration --ginkgo.focus="odo generic" -ginkgo.slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -ginkgo.v -timeout $(TIMEOUT)

# Run json outout tests
.PHONY: test-json-format-output
test-json-format-output:
	go test -v github.com/openshift/odo/tests/integration --ginkgo.focus="odojsonoutput" -ginkgo.slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -ginkgo.v -timeout $(TIMEOUT)

# Run component e2e tests
.PHONY: test-cmp-e2e
test-cmp-e2e:
	go test -v github.com/openshift/odo/tests/integration --ginkgo.focus="odoCmpE2e" -ginkgo.slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -ginkgo.v -timeout $(TIMEOUT)

# Run component subcommands e2e tests
.PHONY: test-cmp-sub-e2e
test-cmp-sub-e2e:
	go test -v github.com/openshift/odo/tests/integration --ginkgo.focus="odoCmpSubE2e" -ginkgo.slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -ginkgo.v -timeout $(TIMEOUT)

# Run java e2e tests
.PHONY: test-java-e2e
test-java-e2e:
	go test -v github.com/openshift/odo/tests/integration --ginkgo.focus="odoJavaE2e" -ginkgo.slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -ginkgo.v -timeout $(TIMEOUT)

# Run source e2e tests
.PHONY: test-source-e2e
test-source-e2e:
	go test -v github.com/openshift/odo/tests/integration --ginkgo.focus="odoSourceE2e" -ginkgo.slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -ginkgo.v -timeout $(TIMEOUT)

# Run service catalog e2e tests
.PHONY: test-service-e2e
test-service-e2e:
	go test -v github.com/openshift/odo/tests/integration --ginkgo.focus="odoServiceE2e" -ginkgo.slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -ginkgo.v -timeout $(TIMEOUT)

# Run link e2e tests
.PHONY: test-link-e2e
test-link-e2e:
	go test -v github.com/openshift/odo/tests/integration --ginkgo.focus="odoLinkE2e" -ginkgo.slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -ginkgo.v -timeout $(TIMEOUT)

# Run watch e2e tests
.PHONY: test-watch-e2e
test-watch-e2e:
	go test -v github.com/openshift/odo/tests/integration --ginkgo.focus="odoWatchE2e" -ginkgo.slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -ginkgo.v -timeout $(TIMEOUT)

# Run storage e2e tests
.PHONY: test-storage-e2e
test-storage-e2e:
	go test -v github.com/openshift/odo/tests/integration --ginkgo.focus="odoStorageE2e" -ginkgo.slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -ginkgo.v -timeout $(TIMEOUT)

# Run cmd app e2e tests
.PHONY: test-cmd-app-e2e
test-cmd-app-e2e:
	go test -v github.com/openshift/odo/tests/integration --ginkgo.focus="odoCmdAppE2e" -ginkgo.slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -ginkgo.v -timeout $(TIMEOUT)

# Run login e2e tests
.PHONY: test-odo-login-e2e
test-odo-login-e2e:
	go test -v github.com/openshift/odo/tests/integration --ginkgo.focus="odoLoginE2e" -ginkgo.slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -ginkgo.v -timeout $(TIMEOUT)

# Run config tests
.PHONY: test-odo-config
test-odo-config:
	go test -v github.com/openshift/odo/tests/integration --ginkgo.focus="odo config test" -ginkgo.slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -ginkgo.v -timeout $(TIMEOUT)

# Run all integration tests
.PHONY: test-integration
test-integration:
	go test -v github.com/openshift/odo/tests/integration -ginkgo.slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -ginkgo.v -timeout $(TIMEOUT)

# Run e2e test scenarios
.PHONY: test-e2e-scenarios
test-e2e-scenarios:
	go test -v github.com/openshift/odo/tests/e2escenarios -ginkgo.slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -ginkgo.v -timeout $(TIMEOUT)

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
