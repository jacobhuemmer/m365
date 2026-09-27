#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
# The minimum Go version is a decision, not a side effect of `go get`
# (a dependency upgrade can raise it silently). To raise it, change
# EXPECTED_GO here and go.mod's go line together, on purpose.
EXPECTED_GO="1.25.0"
# Ask Go for the parsed version: comments, CRLF and spacing don't matter,
# and a malformed go.mod fails here instead of passing.
actual="$(go list -m -f '{{.GoVersion}}')"
if [ "$actual" != "$EXPECTED_GO" ]; then
  echo "gomod: go.mod's go line is $actual, expected $EXPECTED_GO." >&2
  echo "gomod: a dependency upgrade probably raised it. Pin an older version, or change EXPECTED_GO in scripts/gomod.sh on purpose." >&2
  exit 1
fi
# -diff prints the changes tidy would make and exits nonzero; it never
# edits files. A nonzero exit with no diff means tidy itself failed
# (cache, network, toolchain), which is not the same as "not tidy".
if diff="$(go mod tidy -diff)"; then
  echo "gomod: ok"
  exit 0
fi
if [ -n "$diff" ]; then
  printf '%s\n' "$diff"
  echo "gomod: go.mod/go.sum are not tidy; run go mod tidy and commit the result." >&2
else
  echo "gomod: go mod tidy -diff could not run; see the Go error above." >&2
fi
exit 1
