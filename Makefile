BIN_DIR = $(abspath bin)
export GOPATH ?= $(shell go env GOPATH)
export GO111MODULE ?= on

LINUX=LINUX
OSX=OSX
WINDOWS=WIN32
OSFLAG :=
ifeq ($(OS),Windows_NT)
	OSFLAG = $(WINDOWS)
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		OSFLAG = $(LINUX)
	endif
	ifeq ($(UNAME_S),Darwin)
		OSFLAG = $(OSX)
	endif
endif

.PHONY: install
install:
ifeq ($(OSFLAG),$(WINDOWS))
	@echo "Windows system detected - no automated setup available."
	@echo "Please install your developer enviroment manually (@see .tool-versions)."
	@echo
	exit 1
endif
ifeq ($(OSFLAG),$(OSX))
	@echo "MacOS system detected - installing the required toolchain via asdf (@see .tool-versions)."
	@echo
	brew install asdf
	asdf plugin add golang || true
	asdf plugin add nodejs || true
	asdf plugin add python || true
	asdf plugin add mockery || true
	asdf plugin add golangci-lint || true
	asdf plugin add actionlint || true
	asdf plugin add shellcheck || true
	asdf plugin add k3d || true
	asdf plugin add kubectl || true
	asdf plugin add k9s || true
	asdf plugin add helm || true
	asdf plugin add helmenv https://github.com/smartcontractkit/asdf-helmenv.git || true
	@echo
	asdf install
endif
ifeq ($(OSFLAG),$(LINUX))
	@echo "Linux system detected - please install and use NIX (@see shell.nix)."
	@echo
ifneq ($(CI),true)
	# install nix
	sh <(curl -L https://nixos-nix-install-tests.cachix.org/serve/vij683ly7sl95nnhb67bdjjfabclr85m/install) --daemon --tarball-url-prefix https://nixos-nix-install-tests.cachix.org/serve --nix-extra-conf-file ./nix.conf
endif
endif

.PHONY: nix-container
nix-container:
	docker run -it --rm -v $(shell pwd):/repo -e NIX_USER_CONF_FILES=/repo/nix.conf --workdir /repo nixos/nix:latest /bin/sh

.PHONY: nix-flake-update
nix-flake-update:
	docker run -it --rm -v $(shell pwd):/repo -e NIX_USER_CONF_FILES=/repo/nix.conf --workdir /repo nixos/nix:latest /bin/sh -c "nix flake update"

.PHONY: build
build: build-go build-ts

.PHONY: build-go
build-go: build-go-relayer build-go-ops build-go-integration-tests

.PHONY: build-go-relayer
build-go-relayer:
	cd relayer/ && go build ./...

.PHONY: build-go-ops
build-go-ops:
	cd ops/ && go build ./...

.PHONY: build-go-integration-tests
build-go-integration-tests:
	cd integration-tests/ && go build ./...

# TODO: fix and readd build-ts-examples
.PHONY: build-ts
build-ts: build-ts-workspace build-cairo-contracts build-sol-contracts

.PHONY: build-ts-workspace
build-ts-workspace:
	yarn install --frozen-lockfile
	yarn build

# TODO: use yarn workspaces features instead of managing separately like this
# https://yarnpkg.com/cli/workspaces/foreach
.PHONY: build-sol-contracts
build-sol-contracts:
	cd contracts/ && \
		yarn install --frozen-lockfile && \
		yarn compile:solidity

# TODO: this should build cairo contracts when they are rewritten
.PHONY: build-ts-examples
build-ts-examples:
	cd examples/contracts/aggregator-consumer && \
		yarn install --frozen-lockfile && \
		yarn compile:solidity

.PHONY: gowork
gowork:
	go work init
	go work use ./ops
	go work use ./relayer
	go work use ./integration-tests

.PHONY: gowork_rm
gowork_rm:
	rm go.work*

.PHONY: format
format: format-go format-cairo format-ts

.PHONY: format-check
format-check: format-cairo-check format-ts-check

.PHONY: format-go
format-go: format-go-fmt gomodtidy

.PHONY: format-go-fmt
format-go-fmt:
	cd ./relayer && go fmt ./...
	cd ./ops && go fmt ./...
	cd ./integration-tests && go fmt ./...

.PHONY: gomodtidy
gomodtidy:
	cd ./relayer && go mod tidy
	cd ./monitoring && go mod tidy
	cd ./ops && go mod tidy
	cd ./integration-tests && go mod tidy

.PHONY: format-cairo
format-cairo:
	cairo-format -i ./contracts/src/**/*.cairo
	cairo-format -i ./examples/**/*.cairo

.PHONY: format-cairo-check
format-cairo-check:
	cairo-format -c ./contracts/src/**/*.cairo
	cairo-format -c ./examples/**/*.cairo

.PHONY: format-ts
format-ts:
	yarn format

.PHONY: format-ts-check
format-ts-check:
	yarn format:check

.PHONY: lint-go-ops
lint-go-ops:
	cd ./ops && golangci-lint --color=always --out-format checkstyle:golangci-lint-ops-report.xml run

.PHONY: lint-go-relayer
lint-go-relayer:
	cd ./relayer && golangci-lint --color=always --out-format checkstyle:golangci-lint-relayer-report.xml run

.PHONY: lint-go-test
lint-go-test:
	cd ./integration-tests && golangci-lint --color=always --exclude=dot-imports --out-format checkstyle:golangci-lint-integration-tests-report.xml run

.PHONY: test-go
test-go: test-unit-go test-unit-go-race test-integration-go

.PHONY: test-unit
test-unit: test-unit-go test-unit-go-race

LOG_PATH ?= ./gotest.log
.PHONY: test-unit-go
test-unit-go:
	cd ./relayer && go test -json ./... -covermode=atomic -coverpkg=./... -coverprofile=coverage.txt 2>&1 | tee $(LOG_PATH) | gotestloghelper -ci

.PHONY: test-unit-go-race
test-unit-go-race:
	cd ./relayer && CGO_ENABLED=1 go test -v ./... -race -count=10 -coverpkg=./... -coverprofile=race_coverage.txt

.PHONY: test-integration-go
# only runs tests with TestIntegration_* + //go:build integration
test-integration-go: env-devnet-hardhat
	cd ./relayer && go test -json ./... -run TestIntegration -tags integration 2>&1 | tee $(LOG_PATH) | gotestloghelper -ci

.PHONY: test-integration-prep
test-integration-prep:
	cd ./contracts
	make build

.PHONY: test-integration
test-integration: test-integration-smoke test-integration-contracts test-integration-gauntlet

.PHONY: test-integration-smoke
test-integration-smoke: test-integration-prep
	cd integration-tests/ && \
		go test --timeout=2h -v ./smoke

# CI Already has already ran test-integration-prep
.PHONY: test-integration-smoke-ci
test-integration-smoke-ci:
	cd integration-tests/ && \
		go test --timeout=2h -v -count=1 -run TestOCRBasic/$(test) -json ./smoke | tee /tmp/gotest.log | gotestloghelper -ci -singlepackage

.PHONY: test-integration-soak
test-integration-soak: test-integration-prep
	cd integration-tests/ && \
		go test --timeout=1h -v -json ./soak

# CI Already has already ran test-integration-prep
.PHONY: test-integration-soak-ci
test-integration-soak-ci:
	cd integration-tests/ && \
		go test --timeout=1h -v -count=1 -json ./soak

.PHONY: test-examples
test-examples:
	cd ./examples/contracts/aggregator_consumer && \
		snforge test

.PHONY: test-integration-gauntlet
# TODO: fix example
# cd packages-ts/starknet-gauntlet-example/ && \
#   yarn test
test-integration-gauntlet: build-ts env-devnet-hardhat
	cd packages-ts/starknet-gauntlet/ && \
		yarn test
	cd packages-ts/starknet-gauntlet-argent/ && \
		yarn test
	cd packages-ts/starknet-gauntlet-cli/ && \
		yarn test
	cd packages-ts/starknet-gauntlet-multisig/ && \
		yarn test
	cd packages-ts/starknet-gauntlet-ocr2/ && \
		yarn test
	cd packages-ts/starknet-gauntlet-oz/ && \
		yarn test
	cd packages-ts/starknet-gauntlet-token/ && \
		yarn test
	cd packages-ts/starknet-gauntlet-emergency-protocol/ && \
		yarn test

.PHONY: test-ts
test-ts: test-ts-contracts test-integration-contracts test-integration-gauntlet

.PHONY: test-ts-contracts
test-ts-contracts: build-ts env-devnet-hardhat
	cd contracts/ && \
		yarn test

.PHONY: build-cairo-contracts
build-cairo-contracts:
	cd contracts && scarb --profile release build

.PHONY: test-cairo-contracts
test-cairo-contracts:
	cd contracts && scarb test

# TODO: this script needs to be replaced with a predefined K8s enviroment
.PHONY: env-devnet-hardhat
env-devnet-hardhat:
	./ops/scripts/devnet-hardhat.sh

.PHONY: env-devnet-hardhat-down
env-devnet-hardhat-down:
	./ops/scripts/devnet-hardhat-down.sh
