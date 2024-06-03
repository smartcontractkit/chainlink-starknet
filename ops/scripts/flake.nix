{
  inputs = {
    nixpkgs.url = "https://github.com/NixOS/nixpkgs/archive/e1ee359d16a1886f0771cc433a00827da98d861c.tar.gz";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, utils }:
    utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };

        devnet-hardhat-down-script-name = "devnet-hardhat-down";
        devnet-hardhat-down-script-path = (pkgs.writeScriptBin devnet-hardhat-down-script-name (builtins.readFile ./devnet-hardhat-down.sh)).overrideAttrs(old: {
          buildCommand = "${old.buildCommand}\n patchShebangs $out";
        });

        devnet-hardhat-script-name = "devnet-hardhat";
        devnet-hardhat-script-path = (pkgs.writeScriptBin devnet-hardhat-script-name (builtins.readFile ./devnet-hardhat.sh)).overrideAttrs(old: {
          buildCommand = "${old.buildCommand}\n patchShebangs $out";
        });

        build-deps = with pkgs; [
          docker
          bash
          devnet-hardhat-down-script-path
        ];
      in rec {
        defaultPackage = pkgs.symlinkJoin {
          name = devnet-hardhat-script-name;
          paths = [ devnet-hardhat-script-path ] ++ build-deps;
          buildInputs = [ pkgs.makeWrapper ];
          postBuild = "wrapProgram $out/bin/${devnet-hardhat-down-script-name} --prefix PATH : $out/bin";
        };
      }
    );
}

