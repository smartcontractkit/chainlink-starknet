FROM nixos/nix:latest

ARG SUITES=smoke soak
ENV NIX_USER_CONF_FILES=/repo/nix.conf

COPY . /repo/
WORKDIR /repo
RUN nix develop -c helm repo add chainlink-qa https://raw.githubusercontent.com/smartcontractkit/qa-charts/gh-pages/ && \
    nix develop -c helm repo add bitnami https://charts.bitnami.com/bitnami && \
    nix develop -c helm repo update && \
    nix develop -c /repo/integration-tests/scripts/buildTests "${SUITES}"
ENTRYPOINT ["/repo/integration-tests/scripts/entrypoint"]
