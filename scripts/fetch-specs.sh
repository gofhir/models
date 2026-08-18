#!/usr/bin/env bash
#
# fetch-specs.sh — download the FHIR specification bundles the generator needs.
#
# The specs are gitignored (~143 MB), so a fresh clone cannot regenerate code
# until this runs. Every file is verified against the sha256 pinned in
# specs.lock; a mismatch is a hard failure, because a silently changed spec would
# turn the CI drift check into noise.
#
# Usage:
#   scripts/fetch-specs.sh              # all versions
#   scripts/fetch-specs.sh r4 r5        # only the ones named
#   scripts/fetch-specs.sh --verify     # verify what is on disk, download nothing
#
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOCKFILE="$REPO_ROOT/specs.lock"
SPECS_DIR="$REPO_ROOT/specs"

VERIFY_ONLY=0
VERSIONS=()

for arg in "$@"; do
  case "$arg" in
    --verify) VERIFY_ONLY=1 ;;
    -h|--help) sed -n '2,18p' "${BASH_SOURCE[0]}"; exit 0 ;;
    -*) echo "unknown flag: $arg" >&2; exit 2 ;;
    *) VERSIONS+=("$arg") ;;
  esac
done

if [[ ! -f "$LOCKFILE" ]]; then
  echo "error: $LOCKFILE not found" >&2
  exit 1
fi

# sha256 of a file, portable between macOS (shasum) and Linux (sha256sum).
sha256_of() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -d' ' -f1
  else
    shasum -a 256 "$1" | cut -d' ' -f1
  fi
}

# All versions mentioned in the lockfile, in order, deduplicated.
lock_versions() {
  awk '!/^#/ && NF >= 5 {print $1}' "$LOCKFILE" | awk '!seen[$0]++'
}

if [[ ${#VERSIONS[@]} -eq 0 ]]; then
  while IFS= read -r v; do VERSIONS+=("$v"); done < <(lock_versions)
fi

# download_version <version> <url>
# Fetches the release zip once and extracts only the files the lockfile names.
download_version() {
  local version="$1" url="$2"
  local dest="$SPECS_DIR/$version"
  local tmp
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' RETURN

  echo "  fetching $url"
  if ! curl -fsSL --retry 3 -o "$tmp/definitions.zip" "$url"; then
    echo "  error: download failed for $version" >&2
    return 1
  fi

  mkdir -p "$dest"
  local wanted
  wanted="$(awk -v v="$version" '!/^#/ && NF >= 5 && $1 == v {print $3}' "$LOCKFILE")"
  while IFS= read -r file; do
    [[ -n "$file" ]] || continue
    if ! unzip -o -q -j "$tmp/definitions.zip" "$file" -d "$dest"; then
      echo "  error: $file not present in the archive for $version" >&2
      return 1
    fi
  done <<< "$wanted"
}

# verify_version <version> — returns non-zero if anything is missing or altered.
verify_version() {
  local version="$1" failed=0
  while read -r v release file want _url; do
    [[ "$v" == "$version" ]] || continue
    local path="$SPECS_DIR/$version/$file"
    if [[ ! -f "$path" ]]; then
      echo "  MISSING  $version/$file"
      failed=1
      continue
    fi
    local got
    got="$(sha256_of "$path")"
    if [[ "$got" != "$want" ]]; then
      echo "  MISMATCH $version/$file"
      echo "           expected $want"
      echo "           got      $got"
      failed=1
    else
      echo "  ok       $version/$file  (FHIR $release)"
    fi
  done < <(awk '!/^#/ && NF >= 5' "$LOCKFILE")
  return $failed
}

exit_code=0
for version in "${VERSIONS[@]}"; do
  url="$(awk -v v="$version" '!/^#/ && NF >= 5 && $1 == v {print $5; exit}' "$LOCKFILE")"
  if [[ -z "$url" ]]; then
    echo "error: version '$version' is not in specs.lock" >&2
    exit 1
  fi

  echo "$version:"
  if [[ $VERIFY_ONLY -eq 0 ]]; then
    if verify_version "$version" >/dev/null 2>&1; then
      echo "  already present and verified, skipping download"
    else
      download_version "$version" "$url" || { exit_code=1; continue; }
    fi
  fi

  verify_version "$version" || exit_code=1
done

if [[ $exit_code -ne 0 ]]; then
  echo
  echo "specs are not in the expected state." >&2
  if [[ $VERIFY_ONLY -eq 1 ]]; then
    echo "run scripts/fetch-specs.sh to download them." >&2
  else
    echo "a hash mismatch means the published spec changed: update specs.lock deliberately." >&2
  fi
  exit 1
fi

echo
echo "specs verified against specs.lock"
