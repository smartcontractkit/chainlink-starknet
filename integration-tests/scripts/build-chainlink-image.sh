#!/usr/bin/env bash
# Build a local Chainlink image matching CI: core + plugins, with the current
# chainlink-starknet relayer compiled from this repo (not Hub 2.50.0).
#
# Usage:
#   ./integration-tests/scripts/build-chainlink-image.sh
#
# Then run smoke:
#   export CHAINLINK_IMAGE=chainlink
#   export CHAINLINK_VERSION=starknet.$(git rev-parse HEAD)
#   export GAUNTLET_PLUS_PLUS_DIR=...
#   cd integration-tests && go test -v -count=1 -timeout 45m ./smoke/ -run TestOCRBasic
#
# Env:
#   CHAINLINK_DIR     path to chainlink repo (default: ../../chainlink from repo root)
#   CHAINLINK_REF     git ref to build core from (default: develop)
#   DOCKER_PLATFORM   default linux/arm64 on arm64 Mac, else linux/amd64
#   SKIP_BASE_BUILD   set to 1 to reuse existing chainlink-plugins:local-base

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SN_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CHAINLINK_DIR="${CHAINLINK_DIR:-$(cd "${SN_ROOT}/../chainlink" && pwd)}"
CHAINLINK_REF="${CHAINLINK_REF:-develop}"
SN_SHA="$(git -C "${SN_ROOT}" rev-parse HEAD)"
IMAGE_NAME="${IMAGE_NAME:-chainlink}"
IMAGE_TAG="starknet.${SN_SHA}"
BASE_TAG="chainlink-plugins:local-base"
FINAL_TAG="${IMAGE_NAME}:${IMAGE_TAG}"

if [[ "$(uname -m)" == "arm64" ]]; then
  DOCKER_PLATFORM="${DOCKER_PLATFORM:-linux/arm64}"
else
  DOCKER_PLATFORM="${DOCKER_PLATFORM:-linux/amd64}"
fi

if ! command -v docker >/dev/null; then
  echo "docker is required" >&2
  exit 1
fi
if ! command -v gh >/dev/null; then
  echo "gh CLI is required for private module access during docker build" >&2
  exit 1
fi

export GITHUB_TOKEN="${GITHUB_TOKEN:-$(gh auth token)}"

echo "=== chainlink-starknet plugin SHA: ${SN_SHA}"
echo "=== chainlink core repo: ${CHAINLINK_DIR} @ ${CHAINLINK_REF}"
echo "=== docker platform: ${DOCKER_PLATFORM}"
echo "=== output image: ${FINAL_TAG}"

if [[ ! -d "${CHAINLINK_DIR}/plugins" ]]; then
  echo "chainlink repo not found at ${CHAINLINK_DIR}; set CHAINLINK_DIR" >&2
  exit 1
fi

if [[ "${SKIP_BASE_BUILD:-0}" != "1" ]]; then
  echo "=== [1/2] Building chainlink-plugins base (this takes several minutes)..."
  git -C "${CHAINLINK_DIR}" fetch origin "${CHAINLINK_REF}" --quiet || true
  CORE_TREE="${CHAINLINK_DIR}"
  if ! git -C "${CHAINLINK_DIR}" rev-parse --verify "${CHAINLINK_REF}^{commit}" >/dev/null 2>&1; then
    echo "ref ${CHAINLINK_REF} not found in ${CHAINLINK_DIR}" >&2
    exit 1
  fi
  if [[ "$(git -C "${CHAINLINK_DIR}" rev-parse HEAD)" != "$(git -C "${CHAINLINK_DIR}" rev-parse "${CHAINLINK_REF}^{commit}")" ]]; then
    WORKTREE="$(mktemp -d)"
    trap 'rm -rf "${WORKTREE}"' EXIT
    git -C "${CHAINLINK_DIR}" worktree add --detach "${WORKTREE}" "${CHAINLINK_REF}^{commit}" >/dev/null
    CORE_TREE="${WORKTREE}"
    echo "    using detached worktree at ${CORE_TREE}"
  fi

  docker buildx build \
    --load \
    --platform "${DOCKER_PLATFORM}" \
    --secret id=GIT_AUTH_TOKEN,env=GITHUB_TOKEN \
    --build-arg CL_AUTO_DOCKER_TAG=local-base \
    --build-arg CL_IS_PROD_BUILD=true \
    -f "${CORE_TREE}/plugins/chainlink.Dockerfile" \
    -t "${BASE_TAG}" \
    "${CORE_TREE}"
else
  echo "=== [1/2] Skipping base build (SKIP_BASE_BUILD=1)"
fi

echo "=== [2/2] Building overlay with local chainlink-starknet plugin..."
docker buildx build \
  --load \
  --platform "${DOCKER_PLATFORM}" \
  --secret id=GIT_AUTH_TOKEN,env=GITHUB_TOKEN \
  --build-arg CHAINLINK_PLUGINS_BASE="${BASE_TAG}" \
  -f "${SN_ROOT}/integration-tests/docker/chainlink-starknet-plugin.Dockerfile" \
  -t "${FINAL_TAG}" \
  "${SN_ROOT}"

echo ""
echo "Done. Use for smoke:"
echo "  export CHAINLINK_IMAGE=${IMAGE_NAME}"
echo "  export CHAINLINK_VERSION=${IMAGE_TAG}"
