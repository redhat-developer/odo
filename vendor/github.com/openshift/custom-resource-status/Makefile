CODEGEN_PKG ?= ./vendor/k8s.io/code-generator

all: test verify-deepcopy

update-deepcopy: ## Update the deepcopy generated code
	./tools/update-deepcopy.sh

verify-deepcopy: ## Verify deepcopy generated code
	VERIFY=--verify-only ./tools/update-deepcopy.sh

test: ## Run unit tests
	go test -count=1 -short ./conditions/...
	go test -count=1 -short ./objectreferences/...

help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''

.PHONY: update-deepcopy verify-deepcopy
