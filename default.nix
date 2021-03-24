{ pkgs ? import <nixpkgs> {},
  version ? "git",
  src ? pkgs.lib.cleanSource ./. }:
with pkgs;
buildGoModule {
  inherit version src;
  pname = "hey";
  vendorSha256 = null;
  meta = with lib; {
    description = "hey is a tiny program that sends some load to a web application";
    homepage = "https://github.com/rakyll/hey";
    license = licenses.asl20;
  };
}
