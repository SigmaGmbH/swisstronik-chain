{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    rust-overlay = {
      url = "github:oxalica/rust-overlay";
      inputs = {
        nixpkgs.follows = "nixpkgs";
        flake-utils.follows = "flake-utils";
      };
    };
  };
  outputs = { self, nixpkgs, flake-utils, rust-overlay }:
    flake-utils.lib.eachSystem
      (with flake-utils.lib.system; [ "x86_64-linux" ])
      (system:
        let
          overlays = [ (import rust-overlay) ];
          pkgs = import nixpkgs {
            inherit system overlays;
          };
          rustToolchain = pkgs.pkgsBuildHost.rust-bin.fromRustupToolchainFile ./go-sgxvm/rust-toolchain;
        in
        with pkgs;
        {
          devShells.default = mkShell {
            shellHook = ''
            export SHELL=$(which bash)
            source ${sgx-sdk}/environment
            '';
            buildInputs = [ 
                rustToolchain
                go
                curl
                gnumake
                automake
                autoconf
                cmake
                git
                openssl
                libtool
                bash
                stdenv
                protobuf3_20
                protoc-gen-rust
                protoc-gen-go
                protoc-gen-go-grpc
                sgx-sdk
                sgx-psw
                libunwind
                libclang
                clang_multi
                glibc
            ];
          };
        }
      );
}