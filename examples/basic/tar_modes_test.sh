#!/usr/bin/env bash
# Asserts the site tar contains only mode 0644 regular files and 0755 directories.
# Cloudflare and other strict consumers omit Content-Type when an asset has the
# executable bit set, which breaks social previews and content negotiation.

set -euo pipefail

TAR="${1:?usage: tar_modes_test.sh <tar>}"

modes=$(tar tvf "$TAR" | awk '{print $1}' | sort -u)

expected=$'-rw-r--r--\ndrwxr-xr-x'
expected_sorted=$(printf '%s\n' "$expected" | sort -u)

if [[ "$modes" != "$expected_sorted" ]]; then
  echo "FAIL: unexpected file modes in $TAR" >&2
  echo "got:" >&2
  printf '%s\n' "$modes" >&2
  echo "want:" >&2
  printf '%s\n' "$expected_sorted" >&2
  echo "" >&2
  echo "Offending entries:" >&2
  tar tvf "$TAR" | grep -vE '^(-rw-r--r--|drwxr-xr-x) ' >&2 || true
  exit 1
fi
