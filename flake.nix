{
    description = "nth development environment";

    inputs = {
        nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
        flake-utils.url = "github:numtide/flake-utils";
        treefmt-nix = {
            url = "github:numtide/treefmt-nix";
            inputs.nixpkgs.follows = "nixpkgs";
        };
    };

    outputs =
        {
            self,
            nixpkgs,
            flake-utils,
            treefmt-nix,
        }:
        flake-utils.lib.eachDefaultSystem (
            system:
            let
                pkgs = nixpkgs.legacyPackages.${system};
                treefmtEval = treefmt-nix.lib.evalModule pkgs ./treefmt.nix;
                buildGoModule = pkgs.buildGoModule.override { go = pkgs.go_1_26; };
            in
            {
                formatter = treefmtEval.config.build.wrapper;

                checks = {
                    formatting = treefmtEval.config.build.check self;
                    go = buildGoModule {
                        pname = "nth-check";
                        version = "0";
                        src = self;
                        vendorHash = "sha256-3D5k6GSV5zSURmjr0G6S+LkRposYelaSQWVgT0E5RwI=";
                        env = {
                            CGO_ENABLED = "1";
                            GOWORK = "off";
                        };
                        nativeBuildInputs = [ pkgs.go-tools ];
                        preCheck = ''
                            go vet ./...
                            XDG_CACHE_HOME="$TMPDIR" staticcheck ./...
                        '';
                        checkFlags = [ "-race" ];
                    };
                    markdown = pkgs.runCommand "nth-markdownlint" { nativeBuildInputs = [ pkgs.markdownlint-cli2 ]; } ''
                        cd ${self}
                        markdownlint-cli2 "**/*.md"
                        touch $out
                    '';
                    workflows = pkgs.runCommand "nth-actionlint" { nativeBuildInputs = [ pkgs.actionlint ]; } ''
                        cd ${self}
                        actionlint .github/workflows/*.yml
                        touch $out
                    '';
                    installers =
                        pkgs.runCommand "nth-installers-check"
                            {
                                nativeBuildInputs = [
                                    pkgs.powershell
                                    pkgs.shellcheck
                                ];
                            }
                            ''
                                cd ${self}
                                shellcheck install.sh
                                pwsh -NoLogo -NoProfile -NonInteractive -Command '
                                    $tokens = $null
                                    $parseErrors = $null
                                    [void] [System.Management.Automation.Language.Parser]::ParseFile(
                                        (Resolve-Path "install.ps1"),
                                        [ref] $tokens,
                                        [ref] $parseErrors
                                    )
                                    if ($parseErrors.Count -gt 0) {
                                        throw ($parseErrors.Message -join [Environment]::NewLine)
                                    }
                                '
                                touch $out
                            '';
                    release =
                        pkgs.runCommand "nth-goreleaser-check"
                            {
                                nativeBuildInputs = [
                                    pkgs.gitMinimal
                                    pkgs.goreleaser
                                ];
                            }
                            ''
                                        cp -r ${self} source
                                chmod -R u+w source
                                cd source
                                git init --quiet
                                if git remote get-url origin >/dev/null 2>&1; then
                                    git remote set-url origin https://github.com/Hayao0819/nth.git
                                else
                                    git remote add origin https://github.com/Hayao0819/nth.git
                                fi
                                goreleaser check
                                        touch $out
                            '';
                };

                devShells.default = pkgs.mkShell {
                    packages = [
                        pkgs.go_1_26
                        pkgs.gopls
                        pkgs.go-tools
                        pkgs.delve
                        pkgs.goreleaser
                        pkgs.actionlint
                        pkgs.markdownlint-cli2
                        pkgs.shellcheck
                        pkgs.gnumake
                        treefmtEval.config.build.wrapper
                    ];
                };
            }
        );
}
