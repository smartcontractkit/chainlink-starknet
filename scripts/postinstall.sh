#!/usr/bin/env bash
# Patches @chainlink/gauntlet-core to handle starknet.js RpcError which has
# a getter-only .code property. Without this, the catch block in cli.js throws
# "TypeError: Cannot set property code of Error which has only a getter"
# and masks the real underlying error.

TARGET="node_modules/@chainlink/gauntlet-core/dist/cli.js"

if [ -f "$TARGET" ]; then
  if grep -q 'e.code = Error_1.GauntletError.GauntletErrorCode.UNKNOWN;' "$TARGET"; then
    sed -i.bak 's/e\.code = Error_1\.GauntletError\.GauntletErrorCode\.UNKNOWN;/try { e.code = Error_1.GauntletError.GauntletErrorCode.UNKNOWN; } catch (_) {}/' "$TARGET"
    rm -f "${TARGET}.bak"
    echo "postinstall: patched gauntlet-core cli.js (.code assignment)"
  else
    echo "postinstall: gauntlet-core cli.js already patched or changed"
  fi
else
  echo "postinstall: gauntlet-core cli.js not found, skipping patch"
fi
