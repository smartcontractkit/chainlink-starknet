{ stdenv, pkgs, lib }:

# juno requires building with clang, not gcc
(pkgs.mkShell.override { stdenv = pkgs.clangStdenv; }) {
  buildInputs = with pkgs; [
    stdenv.cc.cc.lib
    (rust-bin.stable.latest.default.override { extensions = ["rust-src"]; })
    python39
    python39Packages.pip
    python39Packages.venvShellHook
    python39Packages.fastecdsa # so libgmp is correctly sourced
    zlib # for numpy
    gmp
    # use nodejs 16.x due to https://github.com/NomicFoundation/hardhat/issues/3877
    nodejs-16_x
    (yarn.override { nodejs = nodejs-16_x; })
    nodePackages.typescript
    nodePackages.typescript-language-server
    nodePackages.npm

    go_1_20
    gopls
    delve
    golangci-lint
    gotools

    kube3d
    kubectl
    k9s
    kubernetes-helm

  ] ++ lib.optionals stdenv.isLinux [
    # ledger specific packages
    libudev-zero
    libusb1
  ];

  LD_LIBRARY_PATH = lib.makeLibraryPath [pkgs.zlib stdenv.cc.cc.lib]; # lib64
  HELM_REPOSITORY_CONFIG=./.helm-repositories.yaml;

  venvDir = "./.venv";

  postShellHook = ''
    helm repo update
  '';
}
