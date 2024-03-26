#!/usr/bin/make -f

DOCKER := $(shell which docker)

export GO111MODULE = on

####################
###   Building  ####
####################

include simapp/Makefile

install:
	$(MAKE) -C simapp/ install

####################
###   Testing   ####
####################

include e2e/Makefile

test:
	@echo "--> Running tests"
	go test -v ./...

ictest-poa:
	$(MAKE) -C e2e/ ictest-poa

ictest-jail:
	$(MAKE) -C e2e/ ictest-jail

coverage: ## Run coverage report
	@echo "--> Running coverage"
	@go test -race -cpu=$$(nproc) -covermode=atomic -coverprofile=coverage.out $$(go list ./...) ./e2e/... ./simapp/... -coverpkg=github.com/strangelove-ventures/poa/... > /dev/null 2>&1
	@echo "--> Running coverage filter"
	@./scripts/filter-coverage.sh
	@echo "--> Running coverage report"
	@go tool cover -func=coverage-filtered.out
	@echo "--> Running coverage html"
	@go tool cover -html=coverage-filtered.out -o coverage.html
	@echo "--> Coverage report available at coverage.html"
	@echo "--> Cleaning up coverage files"
	@rm coverage.out
	@echo "--> Running coverage complete"

.PHONY: coverage

###############################################################################
###                                  Docker                                 ###
###############################################################################

local-image:
	docker build . -t poa:local

###################
###  Protobuf  ####
###################

protoVer=0.13.2
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating protobuf files..."
	@$(protoImage) sh ./scripts/protocgen.sh
	@go mod tidy

proto-format:
	@$(protoImage) find ./ -name "*.proto" -exec clang-format -i {} \;

proto-lint:
	@$(protoImage) buf lint

.PHONY: proto-all proto-gen proto-format proto-lint

##################
###  Linting  ####
##################

golangci_lint_cmd=golangci-lint
golangci_version=v1.55.2

lint:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run ./... --timeout 15m

lint-fix:
	@echo "--> Running linter and fixing issues"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run ./... --fix --timeout 15m

.PHONY: lint lint-fix