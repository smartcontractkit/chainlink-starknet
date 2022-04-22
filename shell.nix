{ stdenv, pkgs, lib }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    pkgs.python3
    pkgs.python3Packages.venvShellHook
    gmp
  ];

  venvDir = "./.venv";

  postShellHook = ''
    pip install -r requirements.txt
  '';
}
