#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
# The minimum Go version is a decision, not a side effect of `go get`
# (a dependency upgrade can raise it silently). To raise it, change
# EXPECTED_GO here and go.mod's go line together, on purpose.
EXPECTED_GO="1.25.0"
actual="$(sed -n 's/^go //p' go.mod)"
if [ "$actual" != "$EXPECTED_GO" ]; then
  echo "gomod: go.mod's go line is $actual, expected $EXPECTED_GO." >&2
  echo "gomod: a dependency upgrade probably raised it. Pin an older version, or change EXPECTED_GO in scripts/gomod.sh on purpose." >&2
  exit 1
fi
if ! go mod tidy -diff; then
  echo "gomod: go.mod/go.sum are not tidy; run go mod tidy and commit the result." >&2
  exit 1
fi
echo "gomod: ok"
