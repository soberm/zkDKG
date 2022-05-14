#!/bin/bash

cd "$(dirname "$0")"/../ || exit 1

participants=$1

buildRoot=./build/$participants
contracts=../contracts/contracts

mkdir -p $buildRoot

prefix="// This file is MIT Licensed.
//
// Copyright 2017 Christian Reitwiessner
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the \"Software\"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED \"AS IS\", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
pragma solidity ^0.8.0;\n"
prefixWithImport="${prefix}\nimport \"./Pairing.sol\";\n\n"

declare -A inputs
inputs["poly_eval"]="ShareVerifier"
inputs["poly_eval_input"]="ShareInputVerifier"
inputs["key_deriv"]="KeyVerifier"

for name in ${!inputs[@]}; do
    source=./$name.zok
    buildDir=$buildRoot/$name
    generated=./$name.gen
    contractName=${inputs[$name]}
    checksumFile=$buildDir/$name.sha1

    mkdir -p $buildDir

    # Zokrates has problems with accepting input directly from stdin via /dev/stdin and piping, so temporarily store "generated" file
    sed -E "s/(const u32 N =) \?/\1 $participants/" $source > $generated
    trap "rm $generated" EXIT

    # A matching checksum indicates that all files were already built with the same input file, continue
    if [[ -f $checksumFile ]] && $(sha1sum -c --status $checksumFile); then
        echo "Build files for $name are up-to-date, skipping"
        continue
    fi

    zokrates compile -i $generated -s $buildDir/abi.json -o $buildDir/out

    zokrates setup -i $buildDir/out --proving-key-path $buildDir/proving.key --verification-key-path $buildDir/verification.key

    zokrates export-verifier -i $buildDir/verification.key -o $buildDir/verifier.sol

    # The pairing contract is only required once
    if [[ $name == "poly_eval" ]]; then
        pairing=$(sed -n "/^library Pairing/,/^}/p" $buildDir/verifier.sol)
        printf "${prefix}${pairing}" > $contracts/Pairing.sol
    fi

    verifier=$(sed -ne "s/Verifier/$contractName/" -e "/^contract $contractName/,/^}/p" $buildDir/verifier.sol)
    printf "${prefixWithImport}${verifier}" > $contracts/$contractName.sol

    # Compute checksum to indicate a succesful build using the current source file
    sha1sum $generated > $checksumFile
done
