BIN_DIR = bin
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
	asdf plugin add ginkgo || true
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
	go install github.com/onsi/ginkgo/v2/ginkgo@v$(shell cat ./.tool-versions | grep ginkgo | sed -En "s/ginkgo.(.*)/\1/p")
endif

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

.PHONY: build-ts
build-ts: build-ts-workspace build-ts-contracts build-ts-examples

.PHONY: build-ts-workspace
build-ts-workspace:
	yarn install --frozen-lockfile
	yarn build

# TODO: use yarn workspaces features instead of managing separately like this
# https://yarnpkg.com/cli/workspaces/foreach
.PHONY: build-ts-contracts
build-ts-contracts:
	cd contracts/ && \
		yarn install --frozen-lockfile && \
		yarn compile

.PHONY: build-ts-examples
build-ts-examples:
	cd examples/contracts/aggregator-consumer && \
		yarn install --frozen-lockfile && \
		yarn compile

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
format-go: format-go-fmt format-go-mod-tidy

.PHONY: format-go-fmt
format-go-fmt:
	cd ./relayer && go fmt ./...
	cd ./ops && go fmt ./...
	cd ./integration-tests && go fmt ./...

.PHONY: format-go-mod-tidy
format-go-mod-tidy:
	cd ./relayer && go mod tidy
	cd ./ops && go mod tidy
	cd ./integration-tests && go mod tidy

.PHONY: format-cairo
format-cairo:
	find ./contracts/src -name "*.cairo" -type f \
		-exec cairo-format -i --one_item_per_line {} +
	find ./examples -name "*.cairo" -type f \
		-exec cairo-format -i --one_item_per_line {} +

.PHONY: format-cairo-check
format-cairo-check:
	find ./contracts/src -name "*.cairo" -type f \
		-exec cairo-format -c --one_item_per_line {} +
	find ./examples -name "*.cairo" -type f \
		-exec cairo-format -c --one_item_per_line {} +

.PHONY: format-ts
format-ts:
	yarn format

.PHONY: format-ts-check
format-ts-check:
	yarn format:check

.PHONY: test-go
test-go: test-unit-go test-integration-go

.PHONY: test-unit
test-unit: test-unit-go

.PHONY: test-unit-go
test-unit-go:
	cd ./relayer && go test -v ./...
	cd ./relayer && go test -v ./... -race -count=10

.PHONY: test-integration-go
# only runs tests with TestIntegration_* + //go:build integration
test-integration-go:
	cd ./relayer && go test -v ./... -run TestIntegration -tags integration

.PHONY: test-integration
test-integration: test-integration-smoke test-integration-contracts test-integration-gauntlet

.PHONY: test-integration-smoke
test-integration-smoke: build-ts-contracts
	ginkgo -v -r --junit-report=tests-smoke-report.xml --keep-going --trace integration-tests/smoke

.PHONY: test-integration-contracts
# TODO: better network lifecycle setup - requires external network (L1 + L2)
test-integration-contracts: build-ts env-devnet-hardhat
	cd examples/contracts/aggregator-consumer/ && \
		yarn test
	cd packages-ts/integration-eqlabs-multisig/ && \
		yarn test
	cd packages-ts/integration-starkgate/ && \
		yarn test

.PHONY: test-integration-gauntlet
# TODO: better network lifecycle setup - tests setup/run their own network (L1 + conflict w/ above if not cleaned up)
test-integration-gauntlet: build-ts env-devnet-hardhat-down
	cd packages-ts/starknet-gauntlet/ && \
		yarn test
	cd packages-ts/starknet-gauntlet-argent/ && \
		yarn test
	cd packages-ts/starknet-gauntlet-cli/ && \
		yarn test
	cd packages-ts/starknet-gauntlet-example/ && \
		yarn test
	cd packages-ts/starknet-gauntlet-multisig/ && \
		yarn test
	cd packages-ts/starknet-gauntlet-ocr2/ && \
		yarn test
	cd packages-ts/starknet-gauntlet-oz/ && \
		yarn test
	cd packages-ts/starknet-gauntlet-starkgate/ && \
		yarn test

.PHONY: test-ts
test-ts: test-ts-contracts test-integration-contracts test-integration-gauntlet

.PHONY: test-ts-contracts
test-ts-contracts: build-ts-contracts build-ts-workspace env-devnet-hardhat
	cd contracts/ && \
		yarn test

# TODO: this script needs to be replaced with a predefined K8s enviroment
.PHONY: env-devnet-hardhat
env-devnet-hardhat:
	./ops/scripts/devnet-hardhat.sh

.PHONY: env-devnet-hardhat-down
env-devnet-hardhat-down:
	./ops/scripts/devnet-hardhat-down.sh
