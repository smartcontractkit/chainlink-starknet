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
        devnet-hardhat = import ./ops/scripts { inherit pkgs; };
      in rec {
        devShell = pkgs.callPackage ./shell.nix {
          inherit pkgs;
          scriptDir = toString ./.;  # This converts the flake's root directory to a string
        };

        apps.starknet-devnet = {
          type = "app";
          program = "${devnet-hardhat}/ops/scripts/devnet-hardhat.sh";
        };

        formatter = pkgs.nixpkgs-fmt;
      });
}
