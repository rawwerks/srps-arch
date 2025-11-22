{
  description = "System Resource Protection Script (SRPS): ananicy-cpp + earlyoom + helpers";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        srpsPkg = pkgs.stdenv.mkDerivation {
          pname = "srps";
          version =
            if self ? shortRev then self.shortRev
            else if self ? rev then builtins.substring 0 7 self.rev
            else "dev";
          src = ./.;
          dontBuild = true;
          installPhase = ''
            mkdir -p $out/share/srps
            cp install.sh verify.sh README.md $out/share/srps/
            mkdir -p $out/bin
            ln -s $out/share/srps/install.sh $out/bin/srps-install
            ln -s $out/share/srps/verify.sh $out/bin/srps-verify
          '';
          nativeBuildInputs = [ pkgs.makeWrapper ];
        };
      in {
        packages.default = srpsPkg;
        apps.default = {
          type = "app";
          program = "${srpsPkg}/bin/srps-install";
        };
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ bash git shellcheck cmake pkg-config ];
        };
      }
    );
}
