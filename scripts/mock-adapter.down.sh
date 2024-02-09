#!/usr/bin/env bash

set -euo pipefail

echo "Killing any running mock-adapter processes.."
killall "dummy-external-adapter" &>/dev/null || true
echo "Done."
