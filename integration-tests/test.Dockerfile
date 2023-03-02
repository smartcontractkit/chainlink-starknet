ARG BASE_IMAGE
ARG IMAGE_VERSION=latest
FROM ${BASE_IMAGE}:${IMAGE_VERSION}
# FROM nixos/nix:latest

ARG SUITES=smoke soak

COPY . /go/testdir/
WORKDIR /go/testdir
RUN pwd && \
    ls -la && \
    # curl -o ./install-nix.sh -L https://raw.githubusercontent.com/cachix/install-nix-action/master/install-nix.sh && \
    # chmod +x ./install-nix.sh && \
    # export INPUT_EXTRA_NIX_CONFIG="" && \
    # export INPUT_INSTALL_OPTIONS="" && \
    # export INPUT_NIX_PATH=nixpkgs=channel:nixos-unstable && \
    # export NIXPATH=$INPUT_NIX_PATH && \
    # export GITHUB_PATH=/dev/null && \
    # export GITHUB_ENV=/dev/null && \
    # PATH="/nix/var/nix/profiles/default/bin:$PATH" && \
    # PATH="$HOME/.nix-profile/bin:$PATH" && \
    apt-get update && \
    apt-get -y install sudo xz-utils libgmp-dev make python3 python3-pip python3-venv python-is-python3 && \
    pip3 install fastecdsa --no-binary :all: && \
    # echo "alias python=python3" >> ~/.bash_profile && \
    # echo "alias pip=pip3" >> ~/.bash_profile && \
    # ln -s /usr/local/bin/pip3 /usr/local/bin/pip && \
    # ./install-nix.sh && \
    # nix --extra-experimental-features flakes --extra-experimental-features nix-command develop -c /go/testdir/integration-tests/scripts/buildTests "${SUITES}"
    /go/testdir/integration-tests/scripts/buildTests "${SUITES}"
ENTRYPOINT ["/go/testdir/integration-tests/scripts/entrypoint"]
