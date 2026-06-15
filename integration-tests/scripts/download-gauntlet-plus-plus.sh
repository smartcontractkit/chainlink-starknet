#!/usr/bin/env bash
# Downloads and extracts a gauntlet-plus-plus nops release tarball.
# Prints: export GAUNTLET_PLUS_PLUS_DIR=<install-dir>
#
# Nops tarballs bundle pre-built Starknet ops plugins (gauntlet-plus-plus #1708).
set -euo pipefail

VERSION="${GAUNTLET_PLUS_PLUS_VERSION:-2.6.6}"
TAG="@chainlink/gauntlet-bundle/v${VERSION}"
REPO="smartcontractkit/gauntlet-plus-plus"
CACHE_ROOT="${GAUNTLET_PLUS_PLUS_CACHE:-${PWD}/.cache/gauntlet-plus-plus/v${VERSION}}"

case "$(uname -s)-$(uname -m)" in
  Linux-x86_64 | Linux-amd64) OS=linux; ARCH=x64 ;;
  Linux-aarch64 | Linux-arm64) OS=linux; ARCH=arm64 ;;
  Darwin-arm64) OS=macos; ARCH=arm64 ;;
  Darwin-x86_64) OS=macos; ARCH=x64 ;;
  *)
    echo "unsupported platform: $(uname -s)-$(uname -m)" >&2
    exit 1
    ;;
esac

TARBALL="gauntlet-nops-v${VERSION}-${OS}-${ARCH}.tar.xz"
PREFIX="${TARBALL%.tar.xz}"
INSTALL_DIR="${CACHE_ROOT}/${PREFIX}"

if [[ -x "${INSTALL_DIR}/bin/gauntlet" ]]; then
  echo "export GAUNTLET_PLUS_PLUS_DIR=${INSTALL_DIR}"
  exit 0
fi

TOKEN="${GITHUB_TOKEN:-${GH_TOKEN:-${GATI_TOKEN:-}}}"
if [[ -z "${TOKEN}" ]]; then
  echo "GITHUB_TOKEN, GH_TOKEN, or GATI_TOKEN is required to download gauntlet++ releases" >&2
  exit 1
fi

mkdir -p "${CACHE_ROOT}"
TAG_ENCODED=$(python3 -c 'import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=""))' "${TAG}")
RELEASE_URL="https://api.github.com/repos/${REPO}/releases/tags/${TAG_ENCODED}"

ASSET_URL=$(
  curl -fsSL \
    -H "Authorization: Bearer ${TOKEN}" \
    -H "Accept: application/vnd.github+json" \
    "${RELEASE_URL}" \
    | python3 -c '
import json, sys
tarball = sys.argv[1]
for asset in json.load(sys.stdin).get("assets", []):
    if asset.get("name") == tarball:
        print(asset["url"])
        break
else:
    raise SystemExit(f"release asset {tarball!r} not found")
' "${TARBALL}"
)

curl -fsSL \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Accept: application/octet-stream" \
  -o "${CACHE_ROOT}/${TARBALL}" \
  "${ASSET_URL}"

tar xf "${CACHE_ROOT}/${TARBALL}" -C "${CACHE_ROOT}"

if [[ ! -x "${INSTALL_DIR}/bin/gauntlet" ]]; then
  echo "gauntlet binary missing after extract at ${INSTALL_DIR}/bin/gauntlet" >&2
  exit 1
fi

echo "export GAUNTLET_PLUS_PLUS_DIR=${INSTALL_DIR}"
