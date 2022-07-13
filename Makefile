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
	echo "If you are running windows and know how to install what is needed, please contribute by adding it here!"
	exit 1
endif
ifeq ($(OSFLAG),$(OSX))
	curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
	brew install asdf
	asdf plugin-add golang || true
	asdf plugin-add nodejs || true
	asdf plugin-add python || true
	asdf plugin-add golangci-lint || true
	asdf plugin-add ginkgo || true
	asdf plugin add actionlint || true
	asdf plugin add shellcheck || true
	asdf install
endif
ifeq ($(OSFLAG),$(LINUX))
	curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
ifneq ($(CI),true)
	# install nix
	sh <(curl -L https://nixos-nix-install-tests.cachix.org/serve/vij683ly7sl95nnhb67bdjjfabclr85m/install) --daemon --tarball-url-prefix https://nixos-nix-install-tests.cachix.org/serve --nix-extra-conf-file ./nix.conf
endif
	go install github.com/onsi/ginkgo/v2/ginkgo@v$(shell cat ./.tool-versions | grep ginkgo | sed -En "s/ginkgo.(.*)/\1/p")
endif

.PHONY: e2e_test
e2e_test:
	ginkgo -r integration-tests/smoke