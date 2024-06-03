{
  pkgs ? import <nixpkgs> {}
}:
  let
    inherit (pkgs) stdenv;
  in
    stdenv.mkDerivation {
      name = "starknet-devnet-hardhat-scripts";
      src = ./.;

      buildInputs = with pkgs; [
        docker
        bash
      ];

      installPhase = ''
        mkdir -p $out/ops/scripts
        cp -rv $src/* $out/ops/scripts
      '';
    }

