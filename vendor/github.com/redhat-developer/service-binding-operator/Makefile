.DEFAULT_GOAL := help

# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /bin/bash

#-----------------------------------------------------------------------------
# VERBOSE target
#-----------------------------------------------------------------------------

# When you run make VERBOSE=1 (the default), executed commands will be printed
# before executed. If you run make VERBOSE=2 verbose flags are turned on and
# quiet flags are turned off for various commands. Use V_FLAG in places where
# you can toggle on/off verbosity using -v. Use Q_FLAG in places where you can
# toggle on/off quiet mode using -q. Use S_FLAG where you want to toggle on/off
# silence mode using -s...
VERBOSE ?= 1
Q = @
Q_FLAG = -q
QUIET_FLAG = --quiet
V_FLAG =
S_FLAG = -s
X_FLAG =
ZAP_ENCODER_FLAG = --zap-level=debug --zap-encoder=console
ZAP_LEVEL_FLAG =
ifeq ($(VERBOSE),1)
	Q =
endif
ifeq ($(VERBOSE),2)
	Q =
	Q_FLAG =
	QUIET_FLAG =
	S_FLAG =
	V_FLAG = -v
	X_FLAG = -x
	ZAP_LEVEL_FLAG = --zap-level 1
endif
ifeq ($(VERBOSE),3)
	Q_FLAG =
	QUIET_FLAG =
	S_FLAG =
	V_FLAG = -v
	X_FLAG = -x
	ZAP_LEVEL_FLAG = --zap-level 2
endif

ZAP_FLAGS = $(ZAP_ENCODER_FLAG) $(ZAP_LEVEL_FLAG)

# Create output directory for artifacts and test results. ./out is supposed to
# be a safe place for all targets to write to while knowing that all content
# inside of ./out is wiped once "make clean" is run.
$(shell mkdir -p ./out);

## -- Utility targets --

## Print help message for all Makefile targets
## Run `make` or `make help` to see the help
.PHONY: help
help: ## Credit: https://gist.github.com/prwhite/8168133#gistcomment-2749866

	@printf "Usage:\n  make <target>";

	@awk '{ \
			if ($$0 ~ /^.PHONY: [a-zA-Z\-\_0-9]+$$/) { \
				helpCommand = substr($$0, index($$0, ":") + 2); \
				if (helpMessage) { \
					printf "\033[36m%-20s\033[0m %s\n", \
						helpCommand, helpMessage; \
					helpMessage = ""; \
				} \
			} else if ($$0 ~ /^[a-zA-Z\-\_0-9.]+:/) { \
				helpCommand = substr($$0, 0, index($$0, ":")); \
				if (helpMessage) { \
					printf "\033[36m%-20s\033[0m %s\n", \
						helpCommand, helpMessage; \
					helpMessage = ""; \
				} \
			} else if ($$0 ~ /^##/) { \
				if (helpMessage) { \
					helpMessage = helpMessage"\n                     "substr($$0, 3); \
				} else { \
					helpMessage = substr($$0, 3); \
				} \
			} else { \
				if (helpMessage) { \
					print "\n                     "helpMessage"\n" \
				} \
				helpMessage = ""; \
			} \
		}' \
		$(MAKEFILE_LIST)


#-----------------------------------------------------------------------------
# Global Variables
#-----------------------------------------------------------------------------

# By default the project should be build under GOPATH/src/github.com/<orgname>/<reponame>
GO_PACKAGE_ORG_NAME ?= $(shell basename $$(dirname $$PWD))
GO_PACKAGE_REPO_NAME ?= $(shell basename $$PWD)
GO_PACKAGE_PATH ?= github.com/${GO_PACKAGE_ORG_NAME}/${GO_PACKAGE_REPO_NAME}

PROJECT_DIR ?= $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
SBR_MANIFESTS ?= ${PROJECT_DIR}service-binding-operator-manifests

CGO_ENABLED ?= 0
GO111MODULE ?= on
GOCACHE ?= "$(shell echo ${PWD})/out/gocache"
OS ?= linux
ARCH ?= amd64

# This variable is for artifacts to be archived by Prow jobs at OpenShift CI
# The actual value will be set by the OpenShift CI accordingly
ARTIFACT_DIR ?= "$(shell echo ${PWD})/out"

GOCOV_DIR ?= $(ARTIFACT_DIR)/test-coverage
GOCOV_FILE_TEMPL ?= $(GOCOV_DIR)/REPLACE_TEST.txt
GOCOV ?= "-covermode=atomic -coverprofile REPLACE_FILE"

GIT_COMMIT_ID ?= $(shell git rev-parse --short HEAD)

OPERATOR_VERSION ?= 0.1.1
OPERATOR_GROUP ?= ${GO_PACKAGE_ORG_NAME}
OPERATOR_IMAGE ?= quay.io/${OPERATOR_GROUP}/${GO_PACKAGE_REPO_NAME}
OPERATOR_IMAGE_REL ?= quay.io/${OPERATOR_GROUP}/app-binding-operator
OPERATOR_TAG_SHORT ?= $(OPERATOR_VERSION)
OPERATOR_TAG_LONG ?= $(OPERATOR_VERSION)-$(GIT_COMMIT_ID)
OPERATOR_IMAGE_BUILDER ?= buildah
OPERATOR_SDK_EXTRA_ARGS ?= "--debug"
COMMIT_COUNT := $(shell git rev-list --count HEAD)
BASE_BUNDLE_VERSION ?= $(OPERATOR_VERSION)
BUNDLE_VERSION ?= $(BASE_BUNDLE_VERSION)-$(COMMIT_COUNT)
OPERATOR_IMAGE_REF ?= $(OPERATOR_IMAGE_REL):$(GIT_COMMIT_ID)
CSV_PACKAGE_NAME ?= $(GO_PACKAGE_REPO_NAME)
CSV_CREATION_TIMESTAMP ?= $(shell TZ=GMT date '+%FT%TZ')

QUAY_TOKEN ?= ""
QUAY_BUNDLE_TOKEN ?= ""

MANIFESTS_DIR ?= $(shell echo ${PWD})/manifests
MANIFESTS_TMP ?= $(shell echo ${PWD})/tmp/manifests
HACK_DIR ?= $(shell echo ${PWD})/hack
OUTPUT_DIR ?= $(shell echo ${PWD})/out
OLM_CATALOG_DIR ?= $(shell echo ${PWD})/deploy/olm-catalog
CRDS_DIR ?= $(shell echo ${PWD})/deploy/crds
LOGS_DIR ?= $(OUTPUT_DIR)/logs
GOLANGCI_LINT_BIN=$(OUTPUT_DIR)/golangci-lint

# -- Variables for uploading code coverage reports to Codecov.io --
# This default path is set by the OpenShift CI
CODECOV_TOKEN_PATH ?= "/usr/local/redhat-developer-service-binding-operator-codecov-token/token"
CODECOV_TOKEN ?= @$(CODECOV_TOKEN_PATH)
REPO_OWNER := $(shell echo $$CLONEREFS_OPTIONS | jq '.refs[0].org')
REPO_NAME := $(shell echo $$CLONEREFS_OPTIONS | jq '.refs[0].repo')
BASE_COMMIT := $(shell echo $$CLONEREFS_OPTIONS | jq '.refs[0].base_sha')
PR_COMMIT := $(shell echo $$CLONEREFS_OPTIONS | jq '.refs[0].pulls[0].sha')
PULL_NUMBER := $(shell echo $$CLONEREFS_OPTIONS | jq '.refs[0].pulls[0].number')

## -- Static code analysis (lint) targets --

.PHONY: lint
## Runs linters on Go code files and YAML files - DISABLED TEMPORARILY
lint: setup-venv lint-go-code lint-yaml

YAML_FILES := $(shell find . -path ./vendor -prune -o -type f -regex ".*y[a]ml" -print)
.PHONY: lint-yaml
## runs yamllint on all yaml files
lint-yaml: ${YAML_FILES}
	$(Q)$(OUTPUT_DIR)/venv3/bin/pip install yamllint
	$(Q)$(OUTPUT_DIR)/venv3/bin/yamllint -c .yamllint $(YAML_FILES)

.PHONY: lint-go-code
## Checks the code with golangci-lint
lint-go-code: $(GOLANGCI_LINT_BIN)
	# This is required for OpenShift CI enviroment
	# Ref: https://github.com/openshift/release/pull/3438#issuecomment-482053250
	$(Q)GOCACHE=$(GOCACHE) $(OUTPUT_DIR)/golangci-lint ${V_FLAG} run --deadline=30m

$(GOLANGCI_LINT_BIN):
	$(Q)curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./out v1.18.0

.PHONY: setup-venv
## Setup virtual environment
setup-venv:
	$(Q)python3 -m venv $(OUTPUT_DIR)/venv3
	$(Q)$(OUTPUT_DIR)/venv3/bin/pip install --upgrade setuptools
	$(Q)$(OUTPUT_DIR)/venv3/bin/pip install --upgrade pip

## -- Test targets --

# Generate namespace name for test
out/test-namespace:
	@echo -n "test-namespace-$(shell uuidgen | tr '[:upper:]' '[:lower:]')" > $(OUTPUT_DIR)/test-namespace

.PHONY: get-test-namespace
get-test-namespace: out/test-namespace
	$(eval TEST_NAMESPACE := $(shell cat $(OUTPUT_DIR)/test-namespace))

# E2E test
.PHONY: e2e-deploy-3rd-party-crds
e2e-deploy-3rd-party-crds: get-test-namespace
	$(Q)kubectl --namespace $(TEST_NAMESPACE) apply -f ./test/third-party-crds/

.PHONY: e2e-create-namespace
e2e-create-namespace:
	$(Q)kubectl create namespace $(TEST_NAMESPACE)

.PHONY: e2e-setup
e2e-setup: e2e-cleanup e2e-create-namespace e2e-deploy-3rd-party-crds
	$(Q)mkdir -p ${LOGS_DIR}/e2e

.PHONY: e2e-cleanup
e2e-cleanup: get-test-namespace
	$(Q)-kubectl delete namespace $(TEST_NAMESPACE) --timeout=45s --wait

.PHONY: test-e2e
## Runs the e2e tests locally from test/e2e dir
test-e2e: e2e-setup deploy-crds
	$(info Running E2E test: $@)
	$(Q)set -o pipefail; GO111MODULE=$(GO111MODULE) GOCACHE=$(GOCACHE) SERVICE_BINDING_OPERATOR_DISABLE_ELECTION=true \
		operator-sdk --verbose test local ./test/e2e \
			--namespace $(TEST_NAMESPACE) \
			--up-local \
			--go-test-flags "-timeout=15m" \
			--local-operator-flags "$(ZAP_FLAGS)" \
			$(OPERATOR_SDK_EXTRA_ARGS) \
			| tee $(LOGS_DIR)/e2e/test-e2e.log

.PHONY: parse-test-e2e-operator-log
## Extract the local operator log from the logs of the last e2e tests run
parse-test-e2e-operator-log:
	${HACK_DIR}/e2e-log-parser.sh ${LOGS_DIR}/e2e/test-e2e.log > ${LOGS_DIR}/e2e/local-operator.log

.PHONY: test-unit
## Runs the unit tests without code coverage
test-unit:
	$(info Running unit test: $@)
	$(Q)GO111MODULE=$(GO111MODULE) GOCACHE=$(GOCACHE) \
		go test $(shell GOCACHE="$(GOCACHE)" go list ./...|grep -v e2e) -v -mod vendor $(TEST_EXTRA_ARGS)

.PHONY: test-e2e-image
## Run e2e tests on operator image
test-e2e-image: push-image
	$(info Running e2e test on operator image: $@)
	$(eval NAMESPACE := test-image-$(shell </dev/urandom tr -dc 'a-z0-9' | head -c 7  ; echo))
	echo "$(NAMESPACE)"
	$(Q)kubectl create namespace $(NAMESPACE)
	$(Q)kubectl --namespace $(NAMESPACE) apply -f ./test/third-party-crds/postgresql_v1alpha1_database_crd.yaml
	$(Q)GO111MODULE=$(GO111MODULE) GOCACHE=$(GOCACHE) SERVICE_BINDING_OPERATOR_DISABLE_ELECTION=true \
		operator-sdk --verbose test local ./test/e2e \
			--namespace "$(NAMESPACE)" \
			--image "$(OPERATOR_IMAGE):$(OPERATOR_TAG_LONG)" \
			--go-test-flags "-timeout=15m" \
			--local-operator-flags "$(ZAP_FLAGS)" \
			$(OPERATOR_SDK_EXTRA_ARGS)

.PHONY: test-unit-with-coverage
## Runs the unit tests with code coverage
test-unit-with-coverage:
	$(info Running unit test: $@)
	$(eval GOCOV_FILE := $(shell echo $(GOCOV_FILE_TEMPL) | sed -e 's,REPLACE_TEST,$(@),'))
	$(eval GOCOV_FLAGS := $(shell echo $(GOCOV) | sed -e 's,REPLACE_FILE,$(GOCOV_FILE),'))
	$(Q)mkdir -p $(GOCOV_DIR)
	$(Q)rm -vf '$(GOCOV_DIR)/*.txt'
	$(Q)GO111MODULE=$(GO111MODULE) GOCACHE=$(GOCACHE) \
		go test $(shell GOCACHE="$(GOCACHE)" go list ./...|grep -v e2e) $(GOCOV_FLAGS) -v -mod vendor $(TEST_EXTRA_ARGS)
	$(Q)GOCACHE=$(GOCACHE) go tool cover -func=$(GOCOV_FILE)

.PHONY: test
## Test: Runs unit and integration (e2e) tests
test: test-unit test-e2e

.PHONY: test-e2e-olm-ci
## OLM-E2E: Adds the operator as a subscription, and run e2e tests without any setup.
test-e2e-olm-ci:
	$(Q)sed -e "s,REPLACE_IMAGE,registry.svc.ci.openshift.org/${OPENSHIFT_BUILD_NAMESPACE}/stable:service-binding-operator-registry," ./test/operator-hub/catalog_source.yaml | kubectl apply -f -
	$(Q)kubectl apply -f ./test/operator-hub/subscription.yaml
	$(eval DEPLOYED_NAMESPACE := openshift-operators)
	$(Q)$(HACK_DIR)/check-crds.sh
	$(Q)operator-sdk --verbose test local ./test/e2e \
			--no-setup \
			--go-test-flags "-timeout=15m" \
			--local-operator-flags "$(ZAP_FLAGS)" \
			$(OPERATOR_SDK_EXTRA_ARGS)

## -- Build Go binary and OCI image targets --

.PHONY: build
## Build: compile the operator for Linux/AMD64.
build: out/operator

out/operator:
	$(Q)GOARCH=$(ARCH) GOOS=$(OS) go build ${V_FLAG} -o $(OUTPUT_DIR)/operator cmd/manager/main.go

## Build-Image: using operator-sdk to build a new image
build-image:
	$(Q)operator-sdk build \
		--image-builder=$(OPERATOR_IMAGE_BUILDER) \
		"$(OPERATOR_IMAGE):$(OPERATOR_TAG_LONG)"

## Generate-K8S: after modifying _types, generate Kubernetes scaffolding.
generate-k8s:
	$(Q)GOCACHE=$(GOCACHE) operator-sdk generate k8s

build-openapi-gen-cli: $(OUTPUT_DIR)/openapi-gen
	$(Q)GOCACHE=$(GOCACHE) go build -o $(OUTPUT_DIR)/openapi-gen k8s.io/kube-openapi/cmd/openapi-gen

## Generate-OpenAPI: after modifying _types, generate OpenAPI scaffolding.
generate-openapi: build-openapi-gen-cli
	# Build the latest openapi-gen from source
	$(Q)GOCACHE=$(GOCACHE) $(OUTPUT_DIR)/openapi-gen --logtostderr=true -o "" -i $(GO_PACKAGE_PATH)/pkg/apis/apps/v1alpha1 -O zz_generated.openapi -p ./pkg/apis/apps/v1alpha1 -h ./hack/boilerplate.go.txt -r "-"

## Vendor: 'go mod vendor' resets the vendor folder to what is defined in go.mod.
vendor: go.mod go.sum
	$(Q)GOCACHE=$(GOCACHE) go mod vendor ${V_FLAG}

## Generate CSV: using oeprator-sdk generate cluster-service-version for current operator version
generate-csv:
	operator-sdk generate csv --csv-version=$(OPERATOR_VERSION) --verbose

generate-olm:
	operator-courier --verbose nest $(MANIFESTS_DIR) $(MANIFESTS_TMP)
	cp -vf deploy/crds/*_crd.yaml $(MANIFESTS_TMP)

## -- Publish image and manifests targets --

## Prepare-CSV: using a temporary location copy all operator CRDs and metadata to generate a CSV.
prepare-csv: build-image
	$(eval ICON_BASE64_DATA := $(shell cat ./assets/icon/red-hat-logo.png | base64))
	@rm -rf $(MANIFESTS_TMP) || true
	@mkdir -p ${MANIFESTS_TMP}
	operator-courier --verbose nest $(MANIFESTS_DIR) $(MANIFESTS_TMP)
	cp -vf deploy/crds/*_crd.yaml $(MANIFESTS_TMP)
	sed -i -e 's,REPLACE_IMAGE,"$(OPERATOR_IMAGE):latest",g' $(MANIFESTS_TMP)/*.yaml
	sed -i -e 's,REPLACE_ICON_BASE64_DATA,$(ICON_BASE64_DATA),' $(MANIFESTS_TMP)/*.yaml
	operator-courier --verbose verify $(MANIFESTS_TMP)

.PHONY: push-operator
## Push-Operator: Uplaod operator to Quay.io application repository
push-operator: prepare-csv
	operator-courier push $(MANIFESTS_TMP) $(OPERATOR_GROUP) $(GO_PACKAGE_REPO_NAME) $(OPERATOR_VERSION) "$(QUAY_TOKEN)"

## Push-Image: push container image to upstream, including latest tag.
push-image: build-image
	podman tag "$(OPERATOR_IMAGE):$(OPERATOR_TAG_LONG)" "$(OPERATOR_IMAGE):latest"
	-podman push "$(OPERATOR_IMAGE):$(OPERATOR_TAG_LONG)"
	-podman push "$(OPERATOR_IMAGE):latest"

## -- Local deployment targets --

.PHONY: local
## Local: Run operator locally
local: deploy-clean deploy-rbac deploy-crds
	$(Q)operator-sdk --verbose run --local --operator-flags "$(ZAP_FLAGS)"

.PHONY: deploy-rbac
## Deploy-RBAC: Setup service account and deploy RBAC
deploy-rbac:
	$(Q)kubectl create -f deploy/service_account.yaml
	$(Q)kubectl create -f deploy/role.yaml
	$(Q)kubectl create -f deploy/role_binding.yaml

.PHONY: deploy-crds
## Deploy-CRD: Deploy CRD
deploy-crds:
	$(Q)kubectl apply -f deploy/crds/apps.openshift.io_servicebindingrequests_crd.yaml

.PHONY: deploy-clean
## Deploy-Clean: Removing CRDs and CRs
deploy-clean:
	$(Q)-kubectl delete -f deploy/crds/apps_v1alpha1_servicebindingrequest_cr.yaml
	$(Q)-kubectl delete -f deploy/crds/apps.openshift.io_servicebindingrequests_crd.yaml
	$(Q)-kubectl delete -f deploy/operator.yaml
	$(Q)-kubectl delete -f deploy/role_binding.yaml
	$(Q)-kubectl delete -f deploy/role.yaml
	$(Q)-kubectl delete -f deploy/service_account.yaml

.PHONY: deploy
## Deploy:
deploy: deploy-rbac deploy-crds


## -- Cleanup targets --

.PHONY: clean
## Removes temp directories
clean:
	$(Q)-rm -rf ${V_FLAG} $(OUTPUT_DIR)


## -- Targets for uploading code coverage reports to Codecov.io --

.PHONY: upload-codecov-report
## Uploads the test coverage reports to codecov.io.
## DO NOT USE LOCALLY: must only be called by OpenShift CI when processing new PR and when a PR is merged!
upload-codecov-report:
ifneq ($(PR_COMMIT), null)
	@echo "uploading test coverage report for pull-request #$(PULL_NUMBER)..."
	@/bin/bash <(curl -s https://codecov.io/bash) \
		-t $(CODECOV_TOKEN) \
		-f $(GOCOV_DIR)/*.txt \
		-C $(PR_COMMIT) \
		-r $(REPO_OWNER)/$(REPO_NAME) \
		-P $(PULL_NUMBER) \
		-Z > codecov-upload.log
else
	@echo "uploading test coverage report after PR was merged..."
	@/bin/bash <(curl -s https://codecov.io/bash) \
		-t $(CODECOV_TOKEN) \
		-f $(GOCOV_DIR)/*.txt \
		-C $(BASE_COMMIT) \
		-r $(REPO_OWNER)/$(REPO_NAME) \
		-Z > codecov-upload.log
endif

## -- Target for maintaining consistent crd copies --

.PHONY: consistent-crds-manifests-upstream
## Copy crd from deploy/crds to manifests-upstream/
consistent-crds-manifests-upstream:
	$(Q)cd ./manifests-upstream/${OPERATOR_VERSION}/ && ln -srf ../../deploy/crds/apps_v1alpha1_servicebindingrequest_crd.yaml \
	servicebindingrequests.apps.openshift.io.crd.yaml

## -- Target for merge to master dev release --
.PHONY: merge-to-master-release
## Make a dev release on every merge to master
merge-to-master-release:
	echo "${QUAY_TOKEN}" | docker login -u "redhat-developer+travis" --password-stdin quay.io
	$(eval COMMIT_COUNT := $(shell git rev-list --count HEAD))
	$(Q)docker build -f Dockerfile.rhel -t $(OPERATOR_IMAGE_REF) .
	docker push "$(OPERATOR_IMAGE_REF)"


## -- Target for pushing manifest bundle to service-binding-operator-manifest repo --
.PHONY: push-to-manifest-repo
## Push manifest bundle to service-binding-operator-manifest repo
push-to-manifest-repo:
	@rm -rf $(MANIFESTS_TMP) || true
	@mkdir -p ${MANIFESTS_TMP}/${BUNDLE_VERSION}
	operator-sdk generate csv --csv-version $(BUNDLE_VERSION) --from-version=0.0.23
	cp -vrf $(OLM_CATALOG_DIR)/$(GO_PACKAGE_REPO_NAME)/$(BUNDLE_VERSION)/* $(MANIFESTS_TMP)/$(BUNDLE_VERSION)/
	cp -vrf $(OLM_CATALOG_DIR)/$(GO_PACKAGE_REPO_NAME)/*package.yaml $(MANIFESTS_TMP)/
	cp -vrf $(CRDS_DIR)/*_crd.yaml $(MANIFESTS_TMP)/${BUNDLE_VERSION}/
	sed -i -e 's,REPLACE_IMAGE,$(OPERATOR_IMAGE_REF),g' $(MANIFESTS_TMP)/${BUNDLE_VERSION}/*.clusterserviceversion.yaml
	sed -i -e 's,CSV_CREATION_TIMESTAMP,$(CSV_CREATION_TIMESTAMP),g' $(MANIFESTS_TMP)/${BUNDLE_VERSION}/*.clusterserviceversion.yaml
	awk -i inplace '!/^[[:space:]]+replaces:[[:space:]]+[[:graph:]]+/ { print $0 }' $(MANIFESTS_TMP)/${BUNDLE_VERSION}/*.clusterserviceversion.yaml
	sed -i -e 's,BUNDLE_VERSION,$(BUNDLE_VERSION),g' $(MANIFESTS_TMP)/*.yaml
	sed -i -e 's,PACKAGE_NAME,$(CSV_PACKAGE_NAME),g' $(MANIFESTS_TMP)/*.yaml
	sed -i -e 's,ICON_BASE64_DATA,$(shell base64 --wrap=0 ./assets/icon/red-hat-logo.svg),g' $(MANIFESTS_TMP)/${BUNDLE_VERSION}/*.clusterserviceversion.yaml
	sed -i -e 's,ICON_MEDIA_TYPE,image/svg+xml,g' $(MANIFESTS_TMP)/${BUNDLE_VERSION}/*.clusterserviceversion.yaml

## -- Target for preparing manifest bundle to quay application --
.PHONY: prepare-bundle-to-quay
## Prepare manifest bundle to quay application
prepare-bundle-to-quay:
	$(Q)python3.7 -m venv $(OUTPUT_DIR)/venv3
	$(Q)$(OUTPUT_DIR)/venv3/bin/pip install --upgrade setuptools
	$(Q)$(OUTPUT_DIR)/venv3/bin/pip install --upgrade pip
	$(Q)$(OUTPUT_DIR)/venv3/bin/pip install operator-courier==2.1.2
	$(Q)$(OUTPUT_DIR)/venv3/bin/operator-courier --version
	$(Q)$(OUTPUT_DIR)/venv3/bin/operator-courier verify $(MANIFESTS_TMP)
	rm -rf deploy/olm-catalog/$(GO_PACKAGE_REPO_NAME)/$(BUNDLE_VERSION)

## -- Target to push bundle to quay
.PHONY: push-bundle-to-quay
## Push manifest bundle to quay application
push-bundle-to-quay:
	$(Q)$(OUTPUT_DIR)/venv3/bin/operator-courier verify $(SBR_MANIFESTS)
	$(Q)$(OUTPUT_DIR)/venv3/bin/operator-courier push $(SBR_MANIFESTS) redhat-developer service-binding-operator $(BUNDLE_VERSION) "$(QUAY_BUNDLE_TOKEN)"

## -- Target for validating the operator --
.PHONY: dev-release
## validating the operator by installing new quay releases
dev-release:
	BUNDLE_VERSION=$(BUNDLE_VERSION) ./hack/dev-release.sh

## -- Target to validate release --
.PHONY: validate-release
## validate the operator by installing the releases
validate-release: setup-venv
	$(Q)$(OUTPUT_DIR)/venv3/bin/pip install yq==2.10.0
	BUNDLE_VERSION=$(BASE_BUNDLE_VERSION) CHANNEL="alpha" ./hack/validate-release.sh
