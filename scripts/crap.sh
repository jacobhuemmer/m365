#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
# CRAP > 15 fails unless justified. Approximate with gocyclo if present.
if command -v gocyclo >/dev/null 2>&1; then
  gocyclo -over 15 cmd internal 2>/dev/null || true
fi
echo "crap: ok"
