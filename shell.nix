{ stdenv, pkgs, lib }:

# juno requires building with clang, not gcc
(pkgs.mkShell.override { stdenv = pkgs.clangStdenv; }) {
  buildInputs = with pkgs; [
    stdenv.cc.cc.lib
    (rust-bin.stable.latest.default.override { extensions = ["rust-src"]; })
    nodejs-18_x
    (yarn.override { nodejs = nodejs-18_x; })
    nodePackages.typescript
    nodePackages.typescript-language-server
    nodePackages.npm
    python3

    go_1_21
    gopls
    delve
    (golangci-lint.override { buildGoModule = buildGo121Module; })
    gotools

    kubectl
    kubernetes-helm

    postgresql_15 # psql

  ] ++ lib.optionals stdenv.isLinux [
    # ledger specific packages
    libudev-zero
    libusb1
  ];

  LD_LIBRARY_PATH = lib.makeLibraryPath [pkgs.zlib stdenv.cc.cc.lib]; # lib64
  HELM_REPOSITORY_CONFIG = "./.helm-repositories.yaml";
}
