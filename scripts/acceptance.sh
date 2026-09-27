#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
APS_SHA="accaa33d503340c56513ef387258f8da929ba902"
export PATH="$PWD/.tools/bin:$PATH"
if ! command -v gherkin-parser >/dev/null 2>&1; then
  echo "gherkin-parser missing; run scripts/install-tools.sh (APS $APS_SHA)" >&2
  exit 1
fi
rm -rf build/acceptance/ir build/acceptance/dry
mkdir -p build/acceptance/ir build/acceptance/dry
for f in features/*/*.feature; do
  # Same path-based name as the generated test (cli/help -> cli_help), so
  # features with the same file name in different folders don't collide.
  name=$(printf '%s' "${f#features/}" | sed 's/\.feature$//; s/[^A-Za-z0-9_]/_/g')
  gherkin-parser "$f" "build/acceptance/ir/${name}.json"
  if command -v gherkin-ir-dry-checker >/dev/null 2>&1; then
    gherkin-ir-dry-checker "build/acceptance/ir/${name}.json" "build/acceptance/dry/${name}.json" || true
  fi
done
go run ./cmd/acceptance-entrypoint-generator
go test ./acceptance/generated
