#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
export PATH="$PWD/.tools/bin:$PATH"
# CRAP = complexity^2 x (1 - coverage)^3 + complexity, per non-test
# function, with coverage measured across the module. A score over 15
# fails unless the function is in scripts/crap-baseline.txt (a waiver;
# see its header), a baselined score must not rise, and stale entries
# must be removed. Run it through make (pinned toolchain): Go versions
# measure coverage slightly differently. To list current offenders in
# baseline format: GOTOOLCHAIN=$(sed -n 's/^toolchain //p' go.mod) sh scripts/crap.sh -print-offenders
if ! command -v gocyclo >/dev/null 2>&1; then
  echo "crap: gocyclo missing; run scripts/install-tools.sh" >&2
  exit 1
fi
mkdir -p build/crap
if ! go test -coverpkg=./internal/...,./cmd/... -coverprofile=build/crap/cover.out ./internal/... ./cmd/... > build/crap/test.log 2>&1; then
  cat build/crap/test.log >&2
  exit 1
fi
go tool cover -func=build/crap/cover.out > build/crap/func.txt
gocyclo -ignore '_test' cmd internal > build/crap/cyclo.txt
go build -o build/crap/crap-gate ./cmd/crap-gate
build/crap/crap-gate -module "$(go list -m)" -cover build/crap/func.txt -cyclo build/crap/cyclo.txt -baseline scripts/crap-baseline.txt "$@"
