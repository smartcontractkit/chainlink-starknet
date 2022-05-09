{ stdenv, pkgs, lib }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    pkgs.python39
    pkgs.python39Packages.venvShellHook
    gmp
    nodejs
    nodePackages.npm
  ];

  venvDir = "./.venv";

  postShellHook = ''
    pip install -r contracts/requirements.txt -c contracts/constraints.txt
  '';
}
