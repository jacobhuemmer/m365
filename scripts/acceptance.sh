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
# One feature list for both stages: the generator's names (cli/help ->
# cli_help) and its walk, so every feature gets both IR and a test.
list=build/acceptance/features.tsv
go run ./cmd/acceptance-entrypoint-generator -list > "$list"
tab=$(printf '\t')
findings=0
flagged=0
while IFS="$tab" read -r name f; do
  gherkin-parser "$f" "build/acceptance/ir/${name}.json"
  if command -v gherkin-ir-dry-checker >/dev/null 2>&1; then
    # The checker exits 0 with findings; a crash (bad IR, I/O) fails here.
    dry="build/acceptance/dry/${name}.json"
    gherkin-ir-dry-checker "build/acceptance/ir/${name}.json" "$dry"
    # Findings are advisory wording hints (the constitution's optional
    # dry-check), so they are shown, not failed.
    n=$(sed -n 's/.*"findings": \([0-9][0-9]*\).*/\1/p' "$dry" | head -n 1)
    if [ "${n:-0}" -gt 0 ]; then
      findings=$((findings + n))
      flagged=$((flagged + 1))
    fi
  fi
done < "$list"
if [ "$findings" -gt 0 ]; then
  echo "dry-check: $findings advisory finding(s) in $flagged feature(s); reports in build/acceptance/dry/" >&2
fi
go run ./cmd/acceptance-entrypoint-generator
go test ./acceptance/generated
