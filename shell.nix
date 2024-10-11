{
  stdenv,
  pkgs,
  lib,
  scriptDir,
}:
with pkgs; let
  go = pkgs.go_1_23;

  mkShell' = mkShell.override {
    # juno requires building with clang, not gcc
    stdenv = pkgs.clangStdenv;
  };
in
  mkShell' {
    nativeBuildInputs =
      [
        stdenv.cc.cc.lib
        (rust-bin.stable.latest.default.override {extensions = ["rust-src"];})
        nodejs_20
        (yarn.override {nodejs = nodejs_20;})
        nodePackages.typescript
        nodePackages.typescript-language-server
        nodePackages.npm
        python3

        python311Packages.ledgerwallet
        go

        gopls
        delve
        golangci-lint
        gotools

        kubectl
        kubernetes-helm

        postgresql_15 # psql
      ]
      ++ lib.optionals stdenv.isLinux [
        # ledger specific packages
        libudev-zero
        libusb1
      ];

    LD_LIBRARY_PATH = lib.makeLibraryPath [pkgs.zlib stdenv.cc.cc.lib]; # lib64

    GOROOT = "${go}/share/go";
    CGO_ENABLED = 1;
    HELM_REPOSITORY_CONFIG = "${scriptDir}/.helm-repositories.yaml";

    shellHook = ''
      # Update helm repositories
      helm repo update > /dev/null
      # setup go bin for nix
      export GOBIN=$HOME/.nix-go/bin
      mkdir -p $GOBIN
      export PATH=$GOBIN:$PATH
      # install gotestloghelper
      go install github.com/smartcontractkit/chainlink-testing-framework/tools/gotestloghelper@latest
    '';
  }
