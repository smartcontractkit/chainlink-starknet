FROM nixos/nix:latest

ARG SUITES=smoke soak
ENV NIX_USER_CONF_FILES=/repo/nix.conf

COPY . /repo/
WORKDIR /repo
RUN nix develop -c /repo/integration-tests/scripts/buildTests "${SUITES}"
ENTRYPOINT ["/repo/integration-tests/scripts/entrypoint"]
