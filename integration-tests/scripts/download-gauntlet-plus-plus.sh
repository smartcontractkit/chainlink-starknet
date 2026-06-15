#!/usr/bin/env bash
# Downloads and extracts a gauntlet-plus-plus release tarball and installs plugins.
# Prints: export GAUNTLET_PLUS_PLUS_DIR=<install-dir>
#
# Uses the full release tarball but installs only Starknet (+ core) plugins needed for
# smoke/soak OCR tests. Switch to a nops tarball once G++ ships Starknet ops in nops.
set -euo pipefail

VERSION="${GAUNTLET_PLUS_PLUS_VERSION:-2.6.6}"
TAG="@chainlink/gauntlet-bundle/v${VERSION}"
REPO="smartcontractkit/gauntlet-plus-plus"
CACHE_ROOT="${GAUNTLET_PLUS_PLUS_CACHE:-${PWD}/.cache/gauntlet-plus-plus/v${VERSION}}"

# Plugins required for TestOCRBasic / DeployGauntletPP (see ops/gauntlet/gauntlet_plus_plus_starknet.go).
# Order here does not matter; install order comes from dependencies.txt in the tarball.
gpp_plugin_allowed() {
  case "$1" in
    chainlink-gauntlet-core-sequences-v*) return 0 ;;
    chainlink-gauntlet-environment-v*) return 0 ;;
    chainlink-gauntlet-core-v*) return 0 ;;
    chainlink-gauntlet-starknet-data-feeds-v*) return 0 ;;
    chainlink-gauntlet-starknet-token-v*) return 0 ;;
    chainlink-gauntlet-starknet-ownable-v*) return 0 ;;
    chainlink-gauntlet-starknet-storage-v*) return 0 ;;
    chainlink-gauntlet-starknet-v*) return 0 ;;
    *) return 1 ;;
  esac
}

filter_dependencies_txt() {
  local deps_file=$1
  if [[ ! -f "${deps_file}" ]]; then
    echo "dependencies.txt missing in gauntlet++ tarball" >&2
    exit 1
  fi

  local dep total=0 kept=0 filtered=""
  for dep in $(cat "${deps_file}"); do
    total=$((total + 1))
    if gpp_plugin_allowed "${dep}"; then
      filtered+="${dep} "
      kept=$((kept + 1))
    fi
  done

  if [[ "${kept}" -eq 0 ]]; then
    echo "no allowed plugins matched dependencies.txt (total ${total})" >&2
    exit 1
  fi

  printf '%s\n' "${filtered}" > "${deps_file}"
  echo "Filtered gauntlet++ plugins: ${kept}/${total} (Starknet + core only)" >&2
}

case "$(uname -s)-$(uname -m)" in
  Linux-x86_64 | Linux-amd64) TARBALL="gauntlet-v${VERSION}-ubuntu-24.04.tar.gz" ;;
  Linux-aarch64 | Linux-arm64) TARBALL="gauntlet-v${VERSION}-ubuntu-24.04-4cores-16GB-ARM.tar.gz" ;;
  Darwin-arm64) TARBALL="gauntlet-v${VERSION}-macos-latest.tar.gz" ;;
  Darwin-x86_64) TARBALL="gauntlet-v${VERSION}-macos-latest.tar.gz" ;;
  *)
    echo "unsupported platform: $(uname -s)-$(uname -m)" >&2
    exit 1
    ;;
esac

PREFIX="${TARBALL%.tar.gz}"
INSTALL_DIR="${CACHE_ROOT}/${PREFIX}"
PLUGINS_INSTALLED_MARKER="${INSTALL_DIR}/.plugins-installed-starknet-filtered"

if [[ -x "${INSTALL_DIR}/bin/gauntlet" && -f "${PLUGINS_INSTALLED_MARKER}" ]]; then
  echo "export GAUNTLET_PLUS_PLUS_DIR=${INSTALL_DIR}"
  exit 0
fi

TOKEN="${GITHUB_TOKEN:-${GH_TOKEN:-${GATI_TOKEN:-}}}"
if [[ -z "${TOKEN}" ]]; then
  echo "GITHUB_TOKEN, GH_TOKEN, or GATI_TOKEN is required to download gauntlet++ releases" >&2
  exit 1
fi

if ! command -v node >/dev/null 2>&1; then
  echo "node is required to install gauntlet++ plugins from the full release tarball" >&2
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

mkdir -p "${INSTALL_DIR}"
curl -fsSL \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Accept: application/octet-stream" \
  -o "${CACHE_ROOT}/${TARBALL}" \
  "${ASSET_URL}"

tar xzf "${CACHE_ROOT}/${TARBALL}" -C "${INSTALL_DIR}"

if [[ ! -x "${INSTALL_DIR}/bin/gauntlet" ]]; then
  echo "gauntlet binary missing after extract at ${INSTALL_DIR}/bin/gauntlet" >&2
  exit 1
fi

if [[ ! -f "${PLUGINS_INSTALLED_MARKER}" ]]; then
  echo "Installing Starknet gauntlet++ plugins from release tarball..." >&2
  (
    cd "${INSTALL_DIR}"
    export GAUNTLET_DATA_DIR="${INSTALL_DIR}/data"
    export GAUNTLET_CONFIG_DIR="${INSTALL_DIR}/config"
    export GAUNTLET_CACHE_DIR="${INSTALL_DIR}/cache"
    filter_dependencies_txt ./dependencies.txt
    bash ./install-plugins.sh
  ) >&2
  touch "${PLUGINS_INSTALLED_MARKER}"
fi

echo "export GAUNTLET_PLUS_PLUS_DIR=${INSTALL_DIR}"
