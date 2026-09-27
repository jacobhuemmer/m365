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
# Exactly the non-test files Go compiles on this platform (GoFiles), so a
# file named like foo_testhelper.go is measured and build-tagged files
# outside this build are not scored at 0% coverage.
go list -f '{{$d := .Dir}}{{range .GoFiles}}{{$d}}/{{.}}{{"\n"}}{{end}}' ./cmd/... ./internal/... |
  sed "s|^$PWD/||" > build/crap/files.txt
# Null-separated so a filename with a space stays one argument.
tr '\n' '\0' < build/crap/files.txt | xargs -0 gocyclo > build/crap/cyclo.txt
go build -o build/crap/crap-gate ./cmd/crap-gate
build/crap/crap-gate -module "$(go list -m)" -cover build/crap/func.txt -cyclo build/crap/cyclo.txt -baseline scripts/crap-baseline.txt "$@"
