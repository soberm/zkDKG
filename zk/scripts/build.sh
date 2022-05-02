#!/bin/bash

cd "$(dirname "$0")" || exit 1

sourceName=${1%.*}
build=../build/$sourceName

mkdir -p $build

compileFlags=""
if [[ $sourceName == "poly_eval" ]]; then
    # Required for checking if the commits are valid points
    compileFlags="--isolate-branches"
fi

zokrates compile -i ../"$@" -s $build/abi.json -o $build/out $compileFlags
zokrates setup -i $build/out --proving-key-path $build/proving.key --verification-key-path $build/verification.key
zokrates export-verifier -i $build/verification.key -o $build/verifier.sol