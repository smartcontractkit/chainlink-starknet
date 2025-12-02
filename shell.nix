{
  stdenv,
  pkgs,
  lib,
  scriptDir,
}: let
  go = pkgs.go_1_25;

  mkShell' = pkgs.mkShell.override {
    stdenv = pkgs.clangStdenv;
  };

  custom-golangci-lint = pkgs.buildGoModule rec {
    pname = "golangci-lint";
    version = "2.6.2";

    src = pkgs.fetchFromGitHub {
      owner = "golangci";
      repo = "golangci-lint";
      rev = "v${version}";
      sha256 = "sha256-E/qubcF5AFgWicszZcGNBlRLYqk0ZpGilR1LNw7bO/c=";
    };

    vendorHash = "sha256-KuTACpyMqq3h2MXfJ3wuD+9Z803sXROLfxtZXCbX1RA=";
    subPackages = ["cmd/golangci-lint"];
  };
in
  mkShell' {
    buildInputs = [
      pkgs.zizmor
    ];

    nativeBuildInputs =
      [
        stdenv.cc.cc.lib
        (pkgs.rust-bin.stable.latest.default.override {extensions = ["rust-src"];})
        pkgs.nodejs_20
        (pkgs.yarn.override {nodejs = pkgs.nodejs_20;})
        pkgs.nodePackages.typescript
        pkgs.nodePackages.typescript-language-server
        pkgs.nodePackages.npm
        pkgs.python3
        pkgs.python311Packages.ledgerwallet
        go
        pkgs.gopls
        pkgs.delve
        custom-golangci-lint
        pkgs.gotools
        pkgs.kubectl
        pkgs.kubernetes-helm
        pkgs.postgresql_15
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
      export GOBIN=$HOME/.nix-go/bin
      mkdir -p $GOBIN
      export PATH=$GOBIN:$PATH
      go install github.com/smartcontractkit/chainlink-testing-framework/tools/gotestloghelper@latest
    '';
  }
