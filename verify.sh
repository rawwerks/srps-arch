#!/usr/bin/env bash
set -euo pipefail

REPO="Dicklesworthstone/system_resource_protection_script"
TAG="${1:-latest}"

die() { echo "[verify] $*" >&2; exit 1; }

require() { command -v "$1" >/dev/null 2>&1 || die "Missing required tool: $1"; }

require curl

sha_cmd=""
if command -v sha256sum >/dev/null 2>&1; then
  sha_cmd="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  sha_cmd="shasum -a 256"
else
  die "Missing required tool: sha256sum or shasum"
fi

if [ "$TAG" = "latest" ]; then
  require jq
  TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | jq -r .tag_name)
fi

[ -z "$TAG" ] && die "Could not determine tag"

BASE="https://github.com/${REPO}/releases/download/${TAG}"
echo "[verify] Using release tag: $TAG"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

curl -fsSL "$BASE/install.sh" -o "$tmp/install.sh" || die "Failed to download install.sh"
curl -fsSL "$BASE/install.sh.sha256" -o "$tmp/install.sh.sha256" || die "Failed to download install.sh.sha256"

echo "[verify] Verifying checksum..."
cd "$tmp"
$sha_cmd -c install.sh.sha256
echo "[verify] OK"

echo "[verify] To run the installer:"
echo "bash install.sh [--plan|--install|--uninstall]"
