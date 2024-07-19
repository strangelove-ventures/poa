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

ictest-val-add:
	$(MAKE) -C e2e/ ictest-val-add

ictest-val-remove:
	$(MAKE) -C e2e/ ictest-val-remove

ictest-gov:
	$(MAKE) -C e2e/ ictest-gov

.PHONY: test ictest-poa ictest-jail ictest-val-add ictest-val-remove

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

##################
### Simulation ###
##################

SIM_PARAMS ?= $(shell pwd)/simulation/sim_params.json
SIM_NUM_BLOCKS ?= 100
SIM_PERIOD ?= 5
SIM_COMMIT ?= true
SIM_ENABLED ?= true
SIM_VERBOSE ?= false
SIM_TIMEOUT ?= 24h
SIM_SEED ?= 42
SIM_COMMON_ARGS = -NumBlocks=${SIM_NUM_BLOCKS} -Enabled=${SIM_ENABLED} -Commit=${SIM_COMMIT} -Period=${SIM_PERIOD} -Params=${SIM_PARAMS} -Verbose=${SIM_VERBOSE} -Seed=${SIM_SEED} -v -timeout ${SIM_TIMEOUT}

sim-full-app:
	@echo "--> Running full app simulation (blocks: ${SIM_NUM_BLOCKS}, commit: ${SIM_COMMIT}, period: ${SIM_PERIOD}, seed: ${SIM_SEED}, params: ${SIM_PARAMS}"
	@go test ./simapp -run TestFullAppSimulation ${SIM_COMMON_ARGS}

sim-full-app-random:
	$(MAKE) sim-full-app SIM_SEED=$$RANDOM

# Note: known to fail when using app wiring v1
sim-import-export:
	@echo "--> Running app import/export simulation (blocks: ${SIM_NUM_BLOCKS}, commit: ${SIM_COMMIT}, period: ${SIM_PERIOD}, seed: ${SIM_SEED}, params: ${SIM_PARAMS}"
	@go test ./simapp -run TestAppImportExport ${SIM_COMMON_ARGS}

# Note: known to fail when using app wiring v1
sim-import-export-random:
	$(MAKE) sim-import-export SIM_SEED=$$RANDOM

sim-after-import:
	@echo "--> Running app after import simulation (blocks: ${SIM_NUM_BLOCKS}, commit: ${SIM_COMMIT}, period: ${SIM_PERIOD}, seed: ${SIM_SEED}, params: ${SIM_PARAMS}"
	@go test ./simapp -run TestAppSimulationAfterImport ${SIM_COMMON_ARGS}

sim-after-import-random:
	$(MAKE) sim-after-import SIM_SEED=$$RANDOM

sim-app-determinism:
	@echo "--> Running app determinism simulation (blocks: ${SIM_NUM_BLOCKS}, commit: ${SIM_COMMIT}, period: ${SIM_PERIOD}, seed: ${SIM_SEED}, params: ${SIM_PARAMS}"
	@go test ./simapp -run TestAppStateDeterminism ${SIM_COMMON_ARGS}

sim-app-determinism-random:
	$(MAKE) sim-app-determinism SIM_SEED=$$RANDOM

.PHONY: sim-full-app sim-full-app-random sim-import-export sim-after-import sim-app-determinism sim-import-export-random sim-after-import-random sim-app-determinism-random