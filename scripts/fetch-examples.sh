#!/usr/bin/env bash
#
# fetch-examples.sh — download the official FHIR example corpora.
#
# These feed the conformance suite in conformance/, which round-trips every
# published example through the library. About 200 MB unpacked per version, so
# they are gitignored and fetched on demand. Archives are verified against the
# sha256 pinned in examples.lock.
#
# The conformance suite skips itself when the corpus is missing, so this is only
# needed to actually run it.
#
# Usage:
#   scripts/fetch-examples.sh                # every version, json and xml
#   scripts/fetch-examples.sh r4             # one version
#   scripts/fetch-examples.sh r4 --json      # one version, one format
#   scripts/fetch-examples.sh --verify       # verify archives on disk
#
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOCKFILE="$REPO_ROOT/examples.lock"
DEST_ROOT="$REPO_ROOT/conformance/testdata/examples"
CACHE_DIR="$REPO_ROOT/conformance/testdata/.archives"

VERIFY_ONLY=0
KINDS=()
VERSIONS=()

for arg in "$@"; do
  case "$arg" in
    --verify) VERIFY_ONLY=1 ;;
    --json) KINDS+=("json") ;;
    --xml) KINDS+=("xml") ;;
    -h|--help)
      awk 'NR>1 && /^#/ {sub(/^# ?/, ""); print; next} NR>1 {exit}' "${BASH_SOURCE[0]}"
      exit 0
      ;;
    -*) echo "unknown flag: $arg" >&2; exit 2 ;;
    *) VERSIONS+=("$arg") ;;
  esac
done

[[ ${#KINDS[@]} -eq 0 ]] && KINDS=("json" "xml")

if [[ ! -f "$LOCKFILE" ]]; then
  echo "error: $LOCKFILE not found" >&2
  exit 1
fi

sha256_of() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -d' ' -f1
  else
    shasum -a 256 "$1" | cut -d' ' -f1
  fi
}

if [[ ${#VERSIONS[@]} -eq 0 ]]; then
  while IFS= read -r v; do VERSIONS+=("$v"); done \
    < <(awk '!/^#/ && NF >= 4 {print $1}' "$LOCKFILE" | awk '!seen[$0]++')
fi

# in_list <needle> <haystack...>
in_list() {
  local needle="$1"; shift
  local item
  for item in "$@"; do
    [[ "$item" == "$needle" ]] && return 0
  done
  return 1
}

exit_code=0
mkdir -p "$CACHE_DIR"

while read -r version kind want url; do
  in_list "$version" "${VERSIONS[@]}" || continue
  in_list "$kind" "${KINDS[@]}" || continue

  archive="$CACHE_DIR/$version-$kind.zip"
  dest="$DEST_ROOT/$version/$kind"

  if [[ ! -f "$archive" ]]; then
    if [[ $VERIFY_ONLY -eq 1 ]]; then
      echo "  MISSING  $version/$kind (run without --verify to download)"
      exit_code=1
      continue
    fi
    echo "$version/$kind: fetching $url"
    if ! curl -fsSL --retry 3 -o "$archive" "$url"; then
      echo "  error: download failed" >&2
      rm -f "$archive"
      exit_code=1
      continue
    fi
  fi

  got="$(sha256_of "$archive")"
  if [[ "$got" != "$want" ]]; then
    echo "  MISMATCH $version/$kind"
    echo "           expected $want"
    echo "           got      $got"
    echo "           the published corpus changed; update examples.lock deliberately"
    exit_code=1
    continue
  fi

  # Unpack only when the destination is empty or the archive is newer, so
  # re-running is cheap.
  if [[ $VERIFY_ONLY -eq 0 ]]; then
    if [[ ! -d "$dest" ]] || [[ -z "$(ls -A "$dest" 2>/dev/null)" ]] || [[ "$archive" -nt "$dest" ]]; then
      mkdir -p "$dest"
      unzip -o -q -j "$archive" -d "$dest"
      touch "$dest"
    fi
  fi

  n=0
  [[ -d "$dest" ]] && n="$(find "$dest" -type f -name "*.$kind" | wc -l | tr -d ' ')"
  echo "  ok       $version/$kind  ($n files)"
done < <(awk '!/^#/ && NF >= 4' "$LOCKFILE")

if [[ $exit_code -ne 0 ]]; then
  echo
  echo "example corpus is not in the expected state." >&2
  exit 1
fi

echo
echo "examples verified against examples.lock"
