{
  description = "required tools";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs-terraform.url = "github:stackbuilders/nixpkgs-terraform";
  };

  nixConfig = {
    extra-substituters = "https://nixpkgs-terraform.cachix.org";
    extra-trusted-public-keys = "nixpkgs-terraform.cachix.org-1:8Sit092rIdAVENA3ZVeH9hzSiqI/jng6JiCrQ1Dmusw=";
  };

  outputs =
    { ... }@inputs:
    inputs.flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import inputs.nixpkgs {
          inherit system;
          config.allowUnfree = true;
        };

        terraform-version = pkgs.lib.strings.removeSuffix "\n" (
          builtins.readFile ./hack/terraform/.terraform-version
        );
        terraform = inputs.nixpkgs-terraform.packages.${system}.${terraform-version};
      in
      {
        formatter = pkgs.alejandra;
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            pre-commit
            shellcheck
            just
            graphviz
            gotools
            golangci-lint
            upx
            kubernetes-helm
            openssl
            terraform
            ratchet
          ];
        };
      }
    );
}
