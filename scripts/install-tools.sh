#!/bin/sh
set -eu
# APS SHA recorded: accaa33d503340c56513ef387258f8da929ba902
# unclebob/Acceptance-Pipeline-Specification 2026-06-12
ROOT="$(CDPATH="" cd "$(dirname "$0")/.." && pwd)"
APS_SHA="accaa33d503340c56513ef387258f8da929ba902"
mkdir -p "$ROOT/.tools/bin"
GOBIN="$ROOT/.tools/bin"
export GOBIN
go install github.com/securego/gosec/v2/cmd/gosec@v2.25.0
go install golang.org/x/vuln/cmd/govulncheck@v1.6.0
go install github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0
if [ -d "$HOME/.local/bin" ]; then
  ln -sf "$GOBIN/gosec" "$HOME/.local/bin/gosec"
  ln -sf "$GOBIN/govulncheck" "$HOME/.local/bin/govulncheck"
  ln -sf "$GOBIN/gocyclo" "$HOME/.local/bin/gocyclo"
fi
if [ ! -d "$ROOT/.tools/aps/.git" ]; then
  git clone --filter=blob:none https://github.com/unclebob/Acceptance-Pipeline-Specification.git "$ROOT/.tools/aps"
fi
git -C "$ROOT/.tools/aps" fetch --depth 1 origin "$APS_SHA"
git -C "$ROOT/.tools/aps" checkout "$APS_SHA"
GOBIN="$ROOT/.tools/bin" go install -C "$ROOT/.tools/aps" ./cmd/gherkin-parser ./cmd/gherkin-ir-dry-checker ./cmd/gherkin-mutator
echo "tools installed APS=$APS_SHA"
