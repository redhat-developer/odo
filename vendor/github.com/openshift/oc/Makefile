all: build
.PHONY: all

# Include the library makefile
include $(addprefix ./vendor/github.com/openshift/build-machinery-go/make/, \
	golang.mk \
	targets/openshift/images.mk \
	targets/openshift/rpm.mk \
	targets/openshift/deps-gomod.mk \
)

GO_LD_EXTRAFLAGS :=-X k8s.io/component-base/version.gitMajor="1" \
                   -X k8s.io/component-base/version.gitMinor="17" \
                   -X k8s.io/component-base/version.gitVersion="v1.17.0-4-g38212b5" \
                   -X k8s.io/component-base/version.gitCommit="$(SOURCE_GIT_COMMIT)" \
                   -X k8s.io/component-base/version.buildDate="$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')" \
                   -X k8s.io/component-base/version.gitTreeState="clean" \
                   -X k8s.io/client-go/pkg/version.gitVersion="$(SOURCE_GIT_TAG)" \
                   -X k8s.io/client-go/pkg/version.gitCommit="$(SOURCE_GIT_COMMIT)" \
                   -X k8s.io/client-go/pkg/version.buildDate="$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')" \
                   -X k8s.io/client-go/pkg/version.gitTreeState="$(SOURCE_GIT_TREE_STATE)"

GO_BUILD_PACKAGES :=$(strip \
	./cmd/... \
	$(wildcard ./tools/*) \
)
# These tags make sure we can statically link and avoid shared dependencies
GO_BUILD_FLAGS :=-tags 'include_gcs include_oss containers_image_openpgp gssapi'
GO_BUILD_FLAGS_DARWIN :=-tags 'include_gcs include_oss containers_image_openpgp'
GO_BUILD_FLAGS_WINDOWS :=-tags 'include_gcs include_oss containers_image_openpgp'
GO_BUILD_FLAGS_LINUX_CROSS :=-tags 'include_gcs include_oss containers_image_openpgp'

OUTPUT_DIR :=_output
CROSS_BUILD_BINDIR :=$(OUTPUT_DIR)/bin
RPM_VERSION :=$(shell set -o pipefail && echo '$(SOURCE_GIT_TAG)' | sed -E 's/v([0-9]+\.[0-9]+\.[0-9]+)-.*/\1/')
RPM_EXTRAFLAGS := \
	--define 'local_build true' \
	--define 'os_git_vars ignore' \
	--define 'version $(RPM_VERSION)' \
	--define 'dist .el7' \
	--define 'release 1'

IMAGE_REGISTRY :=registry.svc.ci.openshift.org

# This will call a macro called "build-image" which will generate image specific targets based on the parameters:
# $1 - target name
# $2 - image ref
# $3 - Dockerfile path
# $4 - context
$(call build-image,ocp-cli,$(IMAGE_REGISTRY)/ocp/4.2:cli,./images/cli/Dockerfile.rhel,.)

$(call build-image,ocp-cli-artifacts,$(IMAGE_REGISTRY)/ocp/4.2:cli-artifacts,./images/cli-artifacts/Dockerfile.rhel,.)
image-ocp-cli-artifacts: image-ocp-cli

$(call build-image,ocp-deployer,$(IMAGE_REGISTRY)/ocp/4.2:deployer,./images/deployer/Dockerfile.rhel,.)
image-ocp-deployer: image-ocp-cli

$(call build-image,ocp-recycler,$(IMAGE_REGISTRY)/ocp/4.2:recycler,./images/recycler/Dockerfile.rhel,.)
image-ocp-recycler: image-ocp-cli

oc: GO_BUILD_PACKAGES :=./cmd/oc
oc: build
.PHONY: oc

update: update-generated-completions
.PHONY: update

verify: verify-cli-conventions verify-generated-completions verify-imports
.PHONY: verify

verify-cli-conventions:
	go run ./tools/clicheck
.PHONY: verify-cli-conventions

update-generated-completions: build
	hack/update-generated-completions.sh
.PHONY: update-generated-completions

verify-generated-completions: build
	hack/verify-generated-completions.sh
.PHONY: verify-generated-completions

verify-imports:
	hack/verify-imports.sh
.PHONY: verify-imports

cross-build-darwin-amd64:
	+@GOOS=darwin GOARCH=amd64 $(MAKE) --no-print-directory build GO_BUILD_PACKAGES:=./cmd/oc GO_BUILD_FLAGS:="$(GO_BUILD_FLAGS_DARWIN)" GO_BUILD_BINDIR:=$(CROSS_BUILD_BINDIR)/darwin_amd64
.PHONY: cross-build-darwin-amd64

cross-build-windows-amd64:
	+@GOOS=windows GOARCH=amd64 $(MAKE) --no-print-directory build GO_BUILD_PACKAGES:=./cmd/oc GO_BUILD_FLAGS:="$(GO_BUILD_FLAGS_WINDOWS)" GO_BUILD_BINDIR:=$(CROSS_BUILD_BINDIR)/windows_amd64
.PHONY: cross-build-windows-amd64

cross-build-linux-amd64:
	+@GOOS=linux GOARCH=amd64 $(MAKE) --no-print-directory build GO_BUILD_PACKAGES:=./cmd/oc GO_BUILD_FLAGS:="$(GO_BUILD_FLAGS_LINUX_CROSS)" GO_BUILD_BINDIR:=$(CROSS_BUILD_BINDIR)/linux_amd64
.PHONY: cross-build-linux-amd64

cross-build-linux-arm64:
	+@GOOS=linux GOARCH=arm64 $(MAKE) --no-print-directory build GO_BUILD_PACKAGES:=./cmd/oc GO_BUILD_FLAGS:="$(GO_BUILD_FLAGS_LINUX_CROSS)" GO_BUILD_BINDIR:=$(CROSS_BUILD_BINDIR)/linux_arm64
.PHONY: cross-build-linux-arm64

cross-build-linux-ppc64le:
	+@GOOS=linux GOARCH=ppc64le $(MAKE) --no-print-directory build GO_BUILD_PACKAGES:=./cmd/oc GO_BUILD_FLAGS:="$(GO_BUILD_FLAGS_LINUX_CROSS)" GO_BUILD_BINDIR:=$(CROSS_BUILD_BINDIR)/linux_ppc64le
.PHONY: cross-build-linux-ppc64le

cross-build-linux-s390x:
	+@GOOS=linux GOARCH=s390x $(MAKE) --no-print-directory build GO_BUILD_PACKAGES:=./cmd/oc GO_BUILD_FLAGS:="$(GO_BUILD_FLAGS_LINUX_CROSS)" GO_BUILD_BINDIR:=$(CROSS_BUILD_BINDIR)/linux_s390x
.PHONY: cross-build-linux-s390x

cross-build: cross-build-darwin-amd64 cross-build-windows-amd64 cross-build-linux-amd64 cross-build-linux-arm64 cross-build-linux-ppc64le cross-build-linux-s390x
.PHONY: cross-build

clean-cross-build:
	$(RM) -r '$(CROSS_BUILD_BINDIR)'
	if [ -d '$(OUTPUT_DIR)' ]; then rmdir --ignore-fail-on-non-empty '$(OUTPUT_DIR)'; fi
.PHONY: clean-cross-build

clean: clean-cross-build
