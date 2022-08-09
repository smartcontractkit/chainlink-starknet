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

.PHONY: e2e_test
e2e_test:
	ginkgo -v -r --junit-report=tests-smoke-report.xml --keep-going --trace integration-tests/smoke
	
.PHONY: gowork
gowork:
	go work init
	go work use ./ops
	go work use ./relayer
	go work use ./integration-tests

.PHONY: gowork_rm
gowork_rm:
	rm go.work*
