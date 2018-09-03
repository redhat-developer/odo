PROJECT := github.com/redhat-developer/odo
GITCOMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
PKGS := $(shell go list  ./... | grep -v $(PROJECT)/vendor)
BUILD_FLAGS := -ldflags="-w -X $(PROJECT)/cmd.GITCOMMIT=$(GITCOMMIT)"

default: bin

.PHONY: bin
bin:
	go build ${BUILD_FLAGS} -o odo main.go

.PHONY: install
install:
	go install ${BUILD_FLAGS}

# run all validation tests
.PHONY: validate
validate: gofmt check-vendor vet #lint

.PHONY: gofmt
gofmt:
	./scripts/check-gofmt.sh

.PHONY: check-vendor
check-vendor:
	./scripts/check-vendor.sh

# golint errors are only recommendations
.PHONY: lint
lint:
	golint $(PKGS)

.PHONY: vet
vet:
	go vet $(PKGS)

# install tools used for building, tests and  validations
.PHONY: goget-tools
goget-tools:
	go get -u github.com/Masterminds/glide
	# go get -u golang.org/x/lint/golint
	go get -u github.com/mitchellh/gox

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

# Run e2e tests
.PHONY: test-e2e
test-e2e:
ifdef TIMEOUT
	go test -v github.com/redhat-developer/odo/tests/e2e --ginkgo.focus="odoe2e" -ginkgo.v -timeout $(TIMEOUT)
else
	go test -v github.com/redhat-developer/odo/tests/e2e --ginkgo.focus="odoe2e" -ginkgo.v
endif

.PHONY: test-demo
test-demo:
	go test -v github.com/redhat-developer/odo/tests/e2e --ginkgo.focus="katacodaDemo" -ginkgo.v

.PHONY: test-update-e2e
test-update-e2e:
	go test -v github.com/redhat-developer/odo/tests/e2e --ginkgo.focus="updateE2e" -ginkgo.v

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