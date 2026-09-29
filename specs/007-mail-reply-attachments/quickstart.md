# Validation
From this worktree, with existing .tools/bin link and pinned Go toolchain:

```sh
go test ./internal/adapters/graph ./internal/adapters/cli
make acceptance
make verify
make acceptance-mutation
```

First observe missing attachment payload RED before production edits, then lock a test-only commit. After minimal GREEN all commands must pass. Validate names/exact decoded bytes, comment/no-body parity, zero POSTs on local file read failure, one POST on success, error propagation and dry-run no sends. No live data or sends. Controlled recipient verification is a later explicitly authorized manual check. Repository-local Git identity was confirmed before RED commit 3e9f580. Automated GREEN validation passed; independent review and live recipient checks remain pending.

Observed results are recorded in tasks.md. Test temporary files are removed by Go testing cleanup; validation uses the authorized TMPDIR. Generated acceptance artifacts live under ignored acceptance/generated and build/acceptance; regenerate with `make acceptance`.
