{ stdenv, pkgs, lib }:

# juno requires building with clang, not gcc
(pkgs.mkShell.override { stdenv = pkgs.clangStdenv; }) {
  buildInputs = with pkgs; [
    python39
    python39Packages.venvShellHook
    gmp
    nodejs-16_x
    (yarn.override { nodejs = nodejs-16_x; })
    nodePackages.typescript
    nodePackages.typescript-language-server
    nodePackages.npm

    go_1_18
    gopls
    delve
    golangci-lint
    gotools
  ];

  LD_LIBRARY_PATH="${stdenv.cc.cc.lib}/lib64:$LD_LIBRARY_PATH";

  venvDir = "./.venv";

  postShellHook = ''
    pip install -r contracts/requirements.txt -c contracts/constraints.txt
  '';
}
