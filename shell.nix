{
  sources ? import ./nix/sources.nix,
  pkgs ? import sources.nixpkgs {},
}:

let
  protobuf = pkgs.protobuf_25;
  buf = pkgs.buf;
  just = pkgs.just;
in
  pkgs.mkShell {
    buildInputs = [
      pkgs.git
      buf
      just
      protobuf

      # shell tools
      pkgs.gnused
      pkgs.gnugrep
      pkgs.just
      pkgs.jq
    ];

    shellHook = ''
      export TEMP_DIR="$PWD/.temp"
    '';
  }
