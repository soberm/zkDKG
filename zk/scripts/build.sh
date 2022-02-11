#!/bin/bash

cd "$(dirname "$0")" || exit 1

build=../build/${1%.*}

mkdir -p $build

zokrates compile -i ../"$@" -s $build/abi.json -o $build/out 
zokrates setup -i $build/out --proving-key-path $build/proving.key --verification-key-path $build/verification.key
zokrates export-verifier -i $build/verification.key -o $build/verifier.sol