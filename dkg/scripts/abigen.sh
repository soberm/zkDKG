#!/bin/sh

cd "$(dirname "$0")" || exit 1

if ! which abigen >/dev/null; then
  echo "error: abigen not installed" >&2
  exit 1
fi

abigen --abi ../../contracts/artifacts/contracts/ZKDKG.sol/ZKDKG.abi --pkg dkg --type ZKDKGContract --out ../pkg/dkg/contract.go