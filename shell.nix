{ stdenv, pkgs, lib }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    pkgs.python39
    pkgs.python39Packages.venvShellHook
    gmp
  ];

  venvDir = "./.venv";

  postShellHook = ''
    pip install -r requirements.txt -c constraints.txt
  '';
}
