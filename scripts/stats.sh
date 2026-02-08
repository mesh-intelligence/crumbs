#!/usr/bin/env bash
#
# Count Go source lines and documentation words in this repository.
#

cd "${1:-$(dirname "$0")/..}" || exit 1

# Count non-blank lines in Go files, excluding tests and vendor
go_prod=$(find . -type f -name "*.go" ! -name "*_test.go" ! -path "*/vendor/*" -print0 2>/dev/null |
  xargs -0 wc -l 2>/dev/null | tail -1 | awk '{print $1}')

# Count non-blank lines in Go test files
go_test=$(find . -type f -name "*_test.go" ! -path "*/vendor/*" -print0 2>/dev/null |
  xargs -0 wc -l 2>/dev/null | tail -1 | awk '{print $1}')

# Count words in markdown documentation
doc_wc=$(cat README.md docs/*.md docs/**/*.md 2>/dev/null | wc -w | awk '{print $1}')

printf '{"go_production_loc":%s,"go_test_loc":%s,"go_total_loc":%s,"documentation_words":%s}\n' \
  "${go_prod:-0}" "${go_test:-0}" "$((${go_prod:-0} + ${go_test:-0}))" "${doc_wc:-0}"
