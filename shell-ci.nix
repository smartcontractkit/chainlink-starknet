{
  stdenv,
  pkgs,
  scriptDir,
}: let
  go = pkgs.go_1_26;

  helm = pkgs.kubernetes-helm.overrideAttrs (_old: {
    doCheck = false;
  });

  mkShell' = pkgs.mkShell.override {
    stdenv = pkgs.clangStdenv;
  };
in
  mkShell' {
    nativeBuildInputs =
      [
        stdenv.cc.cc.lib
        go
        helm
        pkgs.kubectl
      ]
      ++ pkgs.lib.optionals pkgs.stdenv.isLinux [
        pkgs.libudev-zero
        pkgs.libusb1
      ];

    LD_LIBRARY_PATH = pkgs.lib.makeLibraryPath [pkgs.zlib stdenv.cc.cc.lib];

    GOROOT = "${go}/share/go";
    CGO_ENABLED = 1;
    HELM_REPOSITORY_CONFIG = "${scriptDir}/.helm-repositories.yaml";

    shellHook = ''
      helm repo update > /dev/null
    '';
  }
