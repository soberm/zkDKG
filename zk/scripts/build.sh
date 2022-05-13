#!/bin/bash

cd "$(dirname "$0")" || exit 1

build=../build/${1%.*}
source=../$1
generated=../$1.gen

mkdir -p $build

# Zokrates has problems with accepting input directly from stdin via /dev/stdin and piping, so temporarily store "generated" file
sed -E "s/(const u32 N =) \?/\1 $2/" $source > $generated
zokrates compile -i $generated -s $build/abi.json -o $build/out
rm $generated

zokrates setup -i $build/out --proving-key-path $build/proving.key --verification-key-path $build/verification.key

prefix="// This file is MIT Licensed.
//
// Copyright 2017 Christian Reitwiessner
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the \"Software\"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED \"AS IS\", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
pragma solidity ^0.8.0;"
zokrates export-verifier -i $build/verification.key -o $build/verifier.sol

pairing=$(sed -n "/^library Pairing/,/^}/p" $build/verifier.sol)
printf "$prefix\n$pairing" > ../../contracts/contracts/Pairing.sol

shareVerifier=$(sed -ne "s/Verifier/ShareVerifier/" -e "/^contract ShareVerifier/,/^}/p" $build/verifier.sol)
printf "$prefix\nimport \"./Pairing.sol\";\n$shareVerifier" > ../../contracts/contracts/ShareVerifier.sol
