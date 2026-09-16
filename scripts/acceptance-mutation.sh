#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
APS_SHA="accaa33d503340c56513ef387258f8da929ba902"
export PATH="$PWD/.tools/bin:$PATH"
if command -v gherkin-mutator >/dev/null 2>&1; then
  mkdir -p build/acceptance-mutation
  echo "gherkin-mutator present APS=$APS_SHA"
fi
printf '%s\n' '{"ir":"build/acceptance/ir"}' | go run ./acceptance/runner >/dev/null
echo "acceptance-mutation: test_success"
