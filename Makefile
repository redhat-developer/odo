BUILD_FLAGS := -ldflags="-w -X github.com/redhat-developer/ocdev/cmd.GITCOMMIT=$(GITCOMMIT)"

default: bin

.PHONY: bin
bin:
	go build ${BUILD_FLAGS} -o ocdev main.go

.PHONY: install
install:
	go install ${BUILD_FLAGS}

# run all validation tests
.PHONY: validate
validate: gofmt vet lint

.PHONY: gofmt
gofmt:
	./scripts/check-gofmt.sh

# golint errors are only recommendations
.PHONY: lint
lint:
	golint $(PKGS)

.PHONY: vet
vet:
	go vet $(PKGS)
