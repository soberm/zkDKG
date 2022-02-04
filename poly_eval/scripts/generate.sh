#!/bin/bash

cd "$(dirname "$0")" || exit 1

zokrates compute-witness -i ../build/out -s ../build/abi.json -a "$@" -o ../build/witness
zokrates generate-proof -i ../build/out --proof-path ../build/proof.json -p ../build/proving.key -w ../build/witness