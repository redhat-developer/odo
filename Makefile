PROJECT := github.com/redhat-developer/odo
GITCOMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
PKGS := $(shell go list  ./... | grep -v $(PROJECT)/vendor)
BUILD_FLAGS := -ldflags="-w -X $(PROJECT)/cmd.GITCOMMIT=$(GITCOMMIT)"
FILES := odo dist

default: bin

.PHONY: bin
bin:
	go build ${BUILD_FLAGS} -o odo main.go

.PHONY: install
install:
	go install ${BUILD_FLAGS}

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
	wwhrd check
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
	gox -osarch="darwin/amd64 linux/amd64 linux/arm windows/amd64" -output="dist/bin/{{.OS}}-{{.Arch}}/odo" $(BUILD_FLAGS)

.PHONY: generate-cli-structure
generate-cli-structure:
	go run scripts/cli-structure/generate-cli-structure.go

.PHONY: generate-cli-reference
generate-cli-reference:
	go run scripts/cli-reference/generate-cli-reference.go > docs/cli-reference.md

# create gzipped binaries in ./dist/release/
# for uploading to GitHub release page
# run make cross before this!
.PHONY: prepare-release
prepare-release: cross
	./scripts/prepare-release.sh

.PHONY: test
test:
	go test -race $(PKGS)

# Run main e2e tests
.PHONY: test-main-e2e
test-main-e2e:
ifdef TIMEOUT
	go test -v github.com/redhat-developer/odo/tests/e2e --ginkgo.focus="odoe2e" -ginkgo.v -timeout $(TIMEOUT)
else
	go test -v github.com/redhat-developer/odo/tests/e2e --ginkgo.focus="odoe2e" -ginkgo.v
endif

# Run component e2e tests
.PHONY: test-cmp-e2e
test-cmp-e2e:
ifdef TIMEOUT
	go test -v github.com/redhat-developer/odo/tests/e2e --ginkgo.focus="odoCmpE2e" -ginkgo.v -timeout $(TIMEOUT)
else
	go test -v github.com/redhat-developer/odo/tests/e2e --ginkgo.focus="odoCmpE2e" -ginkgo.v
endif

# Run java e2e tests
.PHONY: test-java-e2e
test-java-e2e:
ifdef TIMEOUT
	go test -v github.com/redhat-developer/odo/tests/e2e --ginkgo.focus="odoJavaE2e" -ginkgo.v -timeout $(TIMEOUT)
else
	go test -v github.com/redhat-developer/odo/tests/e2e --ginkgo.focus="odoJavaE2e" -ginkgo.v
endif

# Run service catalog e2e tests
.PHONY: test-service-e2e
test-service-e2e:
ifdef TIMEOUT
	go test -v github.com/redhat-developer/odo/tests/e2e --ginkgo.focus="odoServiceE2e" -ginkgo.v -timeout $(TIMEOUT)
else
	go test -v github.com/redhat-developer/odo/tests/e2e --ginkgo.focus="odoServiceE2e" -ginkgo.v
endif

# Run all e2e tests
.PHONY: test-e2e
test-e2e:
ifdef TIMEOUT
	go test -v github.com/redhat-developer/odo/tests/e2e -ginkgo.v -timeout $(TIMEOUT)
else
	go test -v github.com/redhat-developer/odo/tests/e2e -ginkgo.v
endif

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
