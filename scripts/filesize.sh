#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
# Constitution II: a Go file (source or test) over 250 lines needs a refactoring note in
# the plan; over 500 it must be split. Notes are the "- `path`: ..." lines
# under "### File-size notes" in specs/001-m365-cli/plan.md. A note for a
# file that is now 250 lines or fewer, or gone, is stale and must go.
notes_file=specs/001-m365-cli/plan.md
mkdir -p build/filesize
# go list writes to a file first: in a pipeline its failure would be lost.
# Source and test files alike (owner's decision); generated tests excluded.
go list -f '{{$d := .Dir}}{{range .GoFiles}}{{$d}}/{{.}}{{"\n"}}{{end}}{{range .TestGoFiles}}{{$d}}/{{.}}{{"\n"}}{{end}}{{range .XTestGoFiles}}{{$d}}/{{.}}{{"\n"}}{{end}}' ./... > build/filesize/list.txt
sed "s|^$PWD/||" build/filesize/list.txt | grep -v '^acceptance/generated/' | sort > build/filesize/files.txt
if [ ! -s build/filesize/files.txt ]; then
  echo "filesize: go list found no Go files" >&2
  exit 1
fi
fail=0
# A note is "- `path`: reason"; the reason must not be empty.
awk '/^### File-size notes/{on=1; next} on && /^#/{on=0} on && /^- `[^`]*\.go`/{
  line=$0; sub(/^- `/, "", line); path=line; sub(/`.*/, "", path)
  reason=line; sub(/^[^`]*`[: \t]*/, "", reason); print path "\t" reason }' "$notes_file" > build/filesize/entries.tsv
tab=$(printf '\t')
while IFS="$tab" read -r f reason; do
  if ! printf '%s' "$reason" | grep -q '[[:alnum:]]'; then
    echo "filesize: note for $f in $notes_file has no reason; say how to split it" >&2
    fail=1
  fi
done < build/filesize/entries.tsv
cut -f1 build/filesize/entries.tsv | sort > build/filesize/notes.txt
while IFS= read -r f; do
  n=$(wc -l < "$f" | tr -d ' ')
  if [ "$n" -gt 500 ]; then
    echo "filesize: $f has $n lines (> 500); split it" >&2
    fail=1
  elif [ "$n" -gt 250 ] && ! grep -qxF "$f" build/filesize/notes.txt; then
    echo "filesize: $f has $n lines (> 250); split it or add a note under '### File-size notes' in $notes_file" >&2
    fail=1
  fi
done < build/filesize/files.txt
while IFS= read -r f; do
  if [ ! -f "$f" ] || [ "$(wc -l < "$f" | tr -d ' ')" -le 250 ]; then
    echo "filesize: note for $f is stale (file gone or 250 lines or fewer); remove it from $notes_file" >&2
    fail=1
  fi
done < build/filesize/notes.txt
[ "$fail" -eq 0 ] || exit 1
echo "filesize: ok ($(wc -l < build/filesize/files.txt | tr -d ' ') files, $(wc -l < build/filesize/notes.txt | tr -d ' ') noted)"
