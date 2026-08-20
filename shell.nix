{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  packages = with pkgs; [
    go
    gopls
    gotools
    delve
    golangci-lint

    git
    gnumake
    pkg-config
    gcc
  ];

  shellHook = ''
    export GOPATH="''${GOPATH:-$HOME/go}"
    export PATH="$GOPATH/bin:$PATH"
  '';
}
