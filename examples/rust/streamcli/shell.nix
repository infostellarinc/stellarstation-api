{
  sources ? import ./.nix/sources.nix,
  unstable ? import sources.unstable {}
}:

unstable.mkShell {
  buildInputs = [
    unstable.protobuf
  ];
}

