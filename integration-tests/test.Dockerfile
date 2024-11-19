FROM nixos/nix:latest

ARG SUITES=smoke soak
ARG GITHUB_TOKEN=""
RUN 
ENV NIX_USER_CONF_FILES=/repo/nix.conf
ENV PATH="/repo/cairo-build/bin:/repo/scarb-build/bin:${PATH}"

COPY . /repo/
WORKDIR /repo
RUN nix develop -c helm repo update
RUN nix develop -c /repo/integration-tests/scripts/buildTests "${SUITES}" "${GITHUB_TOKEN}"
ENTRYPOINT ["/repo/integration-tests/scripts/entrypoint"]
