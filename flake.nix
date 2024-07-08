{
  description = "Starknet integration";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    rust-overlay.url = "github:oxalica/rust-overlay";
  };

  outputs = inputs@{ self, nixpkgs, flake-utils, rust-overlay, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; overlays = [ rust-overlay.overlays.default ]; };
      in
      {
        devShell = pkgs.callPackage ./shell.nix {
          inherit pkgs;
          scriptDir = toString ./.; # This converts the flake's root directory to a string
        };

        packages = {
          starknet-devnet = pkgs.stdenv.mkDerivation rec {
            name = "starknet-devnet";
            src = ./ops/scripts;
            installPhase = ''
              mkdir -p $out/bin
              cp $src/devnet-hardhat.sh $out/bin/${name}
              cp $src/devnet-hardhat-down.sh $out/bin/
              chmod +x $out/bin/${name}
            '';
          };

          starknet-devnet-down = pkgs.stdenv.mkDerivation rec {
            name = "starknet-devnet-down";
            src = ./ops/scripts;
            installPhase = ''
              mkdir -p $out/bin
              cp $src/devnet-hardhat-down.sh $out/bin/${name}
              chmod +x $out/bin/${name}
            '';
          };
        };

        formatter = pkgs.nixpkgs-fmt;
      });
}
