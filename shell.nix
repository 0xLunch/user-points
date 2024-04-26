{ pkgs ? import <nixpkgs> {
  config.allowUnfree = true;
}, ... }:
pkgs.mkShell {
  buildInputs = with pkgs; [
    # Node
    nodejs
    nodePackages.npm

    nodePackages.typescript
    nodePackages.typescript-language-server
    nodePackages.prettier

    # Go
    go
    gopls
    gotools
    go-tools
    golines

    # Deployments
    terraform
  ];
}
