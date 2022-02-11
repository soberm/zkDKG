#!/bin/bash

cd "$(dirname "$0")" || exit 1

build=../build/${1%.*}

zokrates verify -j $build/proof.json -v $build/verification.key