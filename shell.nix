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
    pip install -r requirements.txt -c constraints.txt
  '';
}
