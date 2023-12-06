PROJECT := github.com/redhat-developer/odo
ifdef GITCOMMIT
        GITCOMMIT := $(GITCOMMIT)
else
        GITCOMMIT := $(shell git describe --no-match --always --abbrev=9 --dirty --broken 2>/dev/null || git rev-parse --short HEAD 2>/dev/null)
endif

COMMON_GOFLAGS := -mod=vendor
COMMON_LDFLAGS := -X $(PROJECT)/pkg/version.GITCOMMIT=$(GITCOMMIT)
BUILD_FLAGS := $(COMMON_GOFLAGS) -ldflags="$(COMMON_LDFLAGS)"
RELEASE_BUILD_FLAGS := $(COMMON_GOFLAGS) -ldflags="-s -w -X $(PROJECT)/pkg/segment.writeKey=R1Z79HadJIrphLoeONZy5uqOjusljSwN $(COMMON_LDFLAGS)"
PKGS := $(shell go list $(COMMON_GOFLAGS)  ./... | grep -v $(PROJECT)/vendor | grep -v $(PROJECT)/tests)
FILES := odo dist
TIMEOUT ?= 14400s

# We should NOT output any color when running interactive tests
# or else we may have issues with regards to comparing coloured output strings
NO_COLOR = true

# Env variable TEST_EXEC_NODES is used to pass spec execution type
# (parallel or sequential) for ginkgo tests. To run the specs sequentially use
# TEST_EXEC_NODES=1, otherwise by default the specs are run in parallel on 4 ginkgo test node if running on PSI cluster or 24 nodes if running on IBM Cloud cluster.

# NOTE: Any TEST_EXEC_NODES value greater than one runs the spec in parallel
# on the same number of ginkgo test nodes.
ifdef TEST_EXEC_NODES
   TEST_EXEC_NODES := $(TEST_EXEC_NODES)
else
   TEST_EXEC_NODES := 4
endif

# After this time, Ginkgo will emit progress reports, so we can get visibility into long-running tests.
POLL_PROGRESS_INTERVAL := 120s

# Env variable GINKGO_TEST_ARGS is used to get control over enabling ginkgo test flags against each test target run.
# For example:
# To enable verbosity export or set env GINKGO_TEST_ARGS like "GINKGO_TEST_ARGS=-v"
GINKGO_TEST_ARGS ?=

# ODO_LOG_LEVEL sets the verbose log level for the make tests
export ODO_LOG_LEVEL ?= 4

# Env variable UNIT_TEST_ARGS is used to get control over enabling test flags along with go test.
# For example:
# To enable verbosity export or set env GINKGO_TEST_ARGS like "GINKGO_TEST_ARGS=-v"
UNIT_TEST_ARGS ?=

export ARTIFACT_DIR ?= .

ifdef PODMAN_EXEC_NODES
   PODMAN_EXEC_NODES := $(PODMAN_EXEC_NODES)
else
   PODMAN_EXEC_NODES := 1
endif

GINKGO_FLAGS_ALL = $(GINKGO_TEST_ARGS) --randomize-all --poll-progress-after=$(POLL_PROGRESS_INTERVAL) --poll-progress-interval=$(POLL_PROGRESS_INTERVAL) -timeout $(TIMEOUT) --no-color

# Flags to run one test per core.
GINKGO_FLAGS_AUTO = $(GINKGO_FLAGS_ALL) -p
# Flags for tests that may be run in parallel
GINKGO_FLAGS=$(GINKGO_FLAGS_ALL) -nodes=$(TEST_EXEC_NODES)
# Flags for Podman tests that may be run in parallel, and not waiting for terminating containers
GINKGO_FLAGS_PODMAN=$(GINKGO_FLAGS_ALL) -nodes=$(PODMAN_EXEC_NODES) --output-interceptor-mode=none 
# Flags for tests that must not be run in parallel
GINKGO_FLAGS_ONE=$(GINKGO_FLAGS_ALL) -nodes=1
# GolangCi version for unit-validate test
GOLANGCI_LINT_VERSION=1.49.0

RUN_GINKGO = go run -mod=vendor github.com/onsi/ginkgo/v2/ginkgo

# making sure that mockgen binary version is the same as github.com/golang/mock library version
GO_MOCK_LIBRARY_VERSION = $(shell go list -mod=readonly -m -f '{{.Version}}' github.com/golang/mock)

default: bin

.PHONY: help
help: ## Show this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: ui-static
ui-static: ## build static files for UI to be served by embedded API server
	podman run --rm \
		-v ${PWD}:/local \
		-t docker.io/library/node:18 \
		/bin/sh -c "cd /local && (cd ui && npm install && npm run build) && rm -rf pkg/apiserver-impl/ui/* && mv ui/dist/devfile-builder/* pkg/apiserver-impl/ui/"

.PHONY: prebuild
prebuild: ## Step to place go embedded files into Go sources before to build the go executable
	cp ododevapispec.yaml pkg/apiserver-impl/swagger-ui/swagger.yaml

.PHONY: bin
bin: prebuild ## build the odo binary
	go build ${BUILD_FLAGS} cmd/odo/odo.go

.PHONY: release-bin
release-bin: prebuild ## build the odo binary
	go build ${RELEASE_BUILD_FLAGS} cmd/odo/odo.go

.PHONY: install
install: prebuild
	go install ${BUILD_FLAGS} ./cmd/odo/

.PHONY: validate
validate: gofmt check-fit check-vendor vet validate-vendor-licenses sec golint ## run all validation tests

.PHONY: gofmt
gofmt:
	./scripts/check-gofmt.sh

.PHONY: check-vendor
check-vendor:
	go mod verify

.PHONY: check-fit
check-fit:
	./scripts/check-fit.sh

.PHONY: validate-vendor-licenses
validate-vendor-licenses:
	go run $(COMMON_GOFLAGS) github.com/frapposelli/wwhrd check -q

.PHONY: golint
golint:
	golangci-lint run ./... --timeout 15m

.PHONY: lint
lint: ## golint errors are only recommendations
	golint $(PKGS)

.PHONY: vet
vet:
	go vet $(PKGS)

.PHONY: sec
sec:
	go run $(COMMON_GOFLAGS) github.com/securego/gosec/v2/cmd/gosec -severity medium -confidence medium -exclude G304,G204,G107 -quiet  ./tests/integration/... ./tests/helper... ./tests/e2escenarios/... ./tests/documentation/...
	go run $(COMMON_GOFLAGS) github.com/securego/gosec/v2/cmd/gosec -severity medium -confidence medium -exclude G304,G204 -quiet  ./cmd/... ./pkg/...

.PHONY: clean
clean:
	@rm -rf $(FILES)

.PHONY: goget-tools
goget-tools:  ## Install binaries of the tools used by odo (current list: golangci-lint, mockgen)
	cd / && go install -mod=mod github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION)
	cd / && go install -mod=mod github.com/golang/mock/mockgen@$(GO_MOCK_LIBRARY_VERSION)

.PHONY: goget-ginkgo
goget-ginkgo:
	@echo "This is no longer used."
	@echo "Ginkgo can be executed directly from this repository using command '$(RUN_GINKGO)'"

.PHONY: test-coverage
test-coverage: ## Run unit tests and collect coverage
	./scripts/generate-coverage.sh

.PHONY: cross
cross: ## compile for multiple platforms
	./scripts/cross-compile.sh $(RELEASE_BUILD_FLAGS)

.PHONY: generate-cli-structure
generate-cli-structure:
	go run cmd/cli-doc/cli-doc.go structure

# run make cross before this!
.PHONY: prepare-release
prepare-release: cross ## create gzipped binaries in ./dist/release/ for uploading to GitHub release page
	./scripts/prepare-release.sh

.PHONY: configure-installer-tests-cluster
configure-installer-tests-cluster:
	. ./scripts/configure-installer-tests-cluster.sh

.PHONY: configure-installer-tests-cluster-s390x
configure-installer-tests-cluster-s390x: ## configure cluster to run tests on s390x arch
	. ./scripts/configure-installer-tests-cluster-s390x.sh

.PHONY: configure-installer-tests-cluster-ppc64le
configure-installer-tests-cluster-ppc64le: ## configure cluster to run tests on ppc64le arch
	. ./scripts/configure-installer-tests-cluster-ppc64le.sh

.PHONY: configure-supported-311-is
configure-supported-311-is:
	. ./scripts/supported-311-is.sh

.PHONY: test
test:
	go test $(UNIT_TEST_ARGS) -race $(PKGS)

.PHONY: test-windows
test-windows:
	go test $(UNIT_TEST_ARGS)  $(PKGS)

# run make cross before this!
.PHONY: packages
packages: ## create deb and rpm packages using fpm in ./dist/pkgs/
	./scripts/create-packages.sh

# run 'make cross' and 'make packages' before this!
.PHONY: upload-packages
upload-packages: ## upload packages created by 'make packages' to bintray repositories
	./scripts/upload-packages.sh

.PHONY: vendor-update
vendor-update: ## Update vendoring
	go mod vendor

.PHONY: openshiftci-presubmit-unittests
openshiftci-presubmit-unittests:
	./scripts/openshiftci-presubmit-unittests.sh

.PHONY: test-integration-cluster
test-integration-cluster:
	$(RUN_GINKGO) $(GINKGO_FLAGS) --junit-report="test-integration.xml" --label-filter="!unauth && !nocluster && !podman" tests/integration

.PHONY: test-integration-cluster-no-service-binding
test-integration-cluster-no-service-binding:
	$(RUN_GINKGO) $(GINKGO_FLAGS) --junit-report="test-integration.xml" --label-filter="!unauth && !nocluster && !podman &&!servicebinding" tests/integration

.PHONY: test-integration-service-binding
test-integration-cluster-service-binding:
	$(RUN_GINKGO) $(GINKGO_FLAGS) --junit-report="test-integration.xml" --label-filter="servicebinding" tests/integration

.PHONY: test-integration-openshift-unauth
test-integration-openshift-unauth:
	$(RUN_GINKGO) $(GINKGO_FLAGS) --junit-report="test-integration-unauth.xml" --label-filter="unauth" tests/integration

.PHONY: test-integration-no-cluster
test-integration-no-cluster:
	$(RUN_GINKGO) $(GINKGO_FLAGS_AUTO)  --junit-report="test-integration-nc.xml" --label-filter=nocluster tests/integration

# Running by default on 1 node because of issues when running too many Podman containers locally,
# but you can override this behavior by setting the PODMAN_EXEC_NODES env var if needed.
# For example: "PODMAN_EXEC_NODES=5".
.PHONY: test-integration-podman
test-integration-podman:
	$(RUN_GINKGO) $(GINKGO_FLAGS_PODMAN) --junit-report="test-integration-podman.xml" --label-filter=podman tests/integration

.PHONY: test-integration
test-integration: test-integration-no-cluster test-integration-cluster

.PHONY: test-e2e-no-service-binding
test-e2e-no-service-binding:
	$(RUN_GINKGO) $(GINKGO_FLAGS) --junit-report="test-e2e.xml" --label-filter="!servicebinding" tests/e2escenarios

.PHONY: test-e2e-service-binding
test-e2e-service-binding:
	$(RUN_GINKGO) $(GINKGO_FLAGS) --junit-report="test-e2e.xml" --label-filter="servicebinding" tests/e2escenarios

.PHONY: test-e2e
test-e2e:
	$(RUN_GINKGO) $(GINKGO_FLAGS) --junit-report="test-e2e.xml"  tests/e2escenarios


.PHONY: test-doc-automation
test-doc-automation:
	$(RUN_GINKGO) $(GINKGO_FLAGS) --junit-report="test-doc-automation.xml"  tests/documentation/...


# Generate OpenAPISpec library based on ododevapispec.yaml inside pkg/apiserver-gen; this will only generate interfaces
# Actual implementation must be done inside pkg/apiserver-impl
# Apart from generating the files, this target also formats the generated files
# and removes openapi.yaml to avoid any confusion regarding ododevapispec.yaml file and which file to use.
.PHONY: generate-apiserver
generate-apiserver: ## Generate OpenAPISpec library based on ododevapispec.yaml inside pkg/apiserver-gen
	podman run --rm \
    		-v ${PWD}:/local \
    		docker.io/openapitools/openapi-generator-cli:v6.6.0 \
    		generate \
    		-i /local/ododevapispec.yaml \
    		-g go-server \
    		-o /local/pkg/apiserver-gen \
    		--additional-properties=outputAsLibrary=true,onlyInterfaces=true,hideGenerationTimestamp=true && \
    		echo "Formatting generated files:" && go fmt ./pkg/apiserver-gen/... && \
    		echo "Removing pkg/apiserver-gen/api/openapi.yaml" && rm ./pkg/apiserver-gen/api/openapi.yaml

.PHONY: generate-apifront
generate-apifront: ## Generate OpenAPISpec library based on ododevapispec.yaml inside ui/src/app
	podman run --rm \
    		-v ${PWD}:/local \
    		docker.io/openapitools/openapi-generator-cli:v6.6.0 \
    		generate \
    		-i /local/ododevapispec.yaml \
    		-g typescript-angular \
    		-o /local/ui/src/app/api-gen

.PHONY: generate-api
generate-api: generate-apiserver generate-apifront ## Generate code based on ododevapispec.yaml

.PHONY: copy-swagger-ui
copy-swagger-ui:
	./scripts/copy-swagger-ui.sh

generate-test-registry-build: ## Rebuild the local registry artifacts. Only for testing
	mkdir -p "${PWD}"/tests/helper/registry_server/testdata/registry-build
	rm -rf "${PWD}"/tests/helper/registry_server/testdata/registry-build/*
	podman container run --rm \
		-v "${PWD}"/tests/helper/registry_server/testdata:/code \
		-t docker.io/golang:1.19 \
		bash /code/build-registry.sh /code/registry/ /code/registry-build/
