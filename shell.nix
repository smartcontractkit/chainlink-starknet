{ stdenv, pkgs, lib }:

# juno requires building with clang, not gcc
(pkgs.mkShell.override { stdenv = pkgs.clangStdenv; }) {
  buildInputs = with pkgs; [
    python39
    python39Packages.pip
    python39Packages.venvShellHook
    python39Packages.fastecdsa # so libgmp is correctly sourced
    gmp
    nodejs-18_x
    (yarn.override { nodejs = nodejs-18_x; })
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

  LD_LIBRARY_PATH="${stdenv.cc.cc.lib}/lib64:$LD_LIBRARY_PATH";
  HELM_REPOSITORY_CONFIG=./.helm-repositories.yaml;

  venvDir = "./.venv";

  postShellHook = ''
    pip install -r ${./contracts/requirements.txt} -c ${./contracts/constraints.txt}
    helm repo update
  '';
}
