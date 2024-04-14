{
  description = "Development Flake with Node, TS, Go, Postgre and Docker";

  inputs = {
    #latest pkgs
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    systems.url = "github:nix-systems/default";
  };

  outputs = {
    systems,
    nixpkgs,
    ...
  } @ inputs: let 
    pkgs = import nixpkgs { 
      system = "x86_64-linux"; 
      config.allowUnfree = true; 
    };
  in {
    devShells.${pkgs.system} = {
      default = pkgs.mkShell {
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

          # Postgre
          postgresql
          # Deployments
          docker
          terraform
        ];
      };
    };
  };
}