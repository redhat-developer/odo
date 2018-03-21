PROJECT := github.com/redhat-developer/ocdev
GITCOMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
PKGS := $(shell go list  ./... | grep -v $(PROJECT)/vendor)
BUILD_FLAGS := -ldflags="-w -X $(PROJECT)/cmd.GITCOMMIT=$(GITCOMMIT)"

default: bin

.PHONY: bin
bin:
	go build ${BUILD_FLAGS} -o ocdev main.go

.PHONY: install
install:
	go install ${BUILD_FLAGS}

# run all validation tests
.PHONY: validate
validate: gofmt check-vendor vet lint

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
	go get -u github.com/golang/lint/golint
	go get -u github.com/mitchellh/gox

# Run unit tests and collect coverage
.PHONY: test-coverage
test-coverage:
	./scripts/generate-coverage.sh

# compile for multiple platforms
.PHONY: cross
cross:
	gox -osarch="darwin/amd64 linux/amd64 linux/arm windows/amd64" -output="dist/bin/{{.OS}}-{{.Arch}}/ocdev" $(BUILD_FLAGS)

.PHONY: generate-cli-docs
generate-cli-docs:
	go run scripts/generate-cli-documentation.go

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
	go test -v github.com/redhat-developer/ocdev/tests/e2e

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
