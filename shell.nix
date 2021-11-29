{ pkgs ? import <nixpkgs> {} }:
with pkgs;
mkShell {
  buildInputs = [
    go
    gnumake # for make
    google-cloud-sdk # for gsutil
  ];
}
