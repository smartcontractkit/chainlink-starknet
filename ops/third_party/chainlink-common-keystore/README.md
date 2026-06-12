# Vendored `chainlink-common/keystore` (temporary)

## Why this exists

`integration-tests` pulls `github.com/smartcontractkit/chainlink-common/keystore` transitively. Upstream **v1.0.2** still uses starknet.go v0.9 APIs (`curve.Curve`), which were removed in **starknet.go v0.17.1**. The `go.mod` replace in `integration-tests/go.mod` points here so integration tests compile.

This fork is **not used by the relayer or production Chainlink images** — only integration-test builds.

## Upstream baseline

| Field | Value |
|-------|-------|
| Module | `github.com/smartcontractkit/chainlink-common/keystore` |
| Version | **v1.0.2** |
| Introduced | commit `9bb588e4` on `feature/bump-starknet-go` |

## Actual patch (starknet.go v0.17)

Only **`corekeys/starkkey/`** differs from upstream for v0.17 compatibility:

| File | Change |
|------|--------|
| `corekeys/starkkey/curve_compat.go` | Local `curveOrder` (replaces removed `curve.Curve.Order()`) |
| `corekeys/starkkey/utils.go` | `GenerateKey` via package-level `curve.Sign` / `curve.PrivateKeyToPoint` |
| `corekeys/starkkey/key.go` | Package-level curve funcs instead of `curve.Curve.*` |
| `corekeys/starkkey/ocr2key.go` | Malleability check uses `curveOrder`; `curve.GetYCoordinate`, `curve.Verify` |

The rest of this tree is an unmodified copy of upstream keystore v1.0.2 (with `go:generate`/protoc removed for CI).

A machine-readable diff of the starkkey patch lives in [`patches/starkkey-starknet-go-v0.17.patch`](patches/starkkey-starknet-go-v0.17.patch).

## Removal criteria

Delete this fork when **all** of the following are true:

1. `chainlink-common/keystore` releases with starknet.go **v0.17+** support (upstream issue: track in smartcontractkit/chainlink-common).
2. `integration-tests/go.mod` drops the `replace` directive.
3. CI passes without this directory.

## Regenerating from upstream

```bash
# 1. Copy keystore v1.0.2 from chainlink-common
# 2. Apply patches/starkkey-starknet-go-v0.17.patch under corekeys/starkkey/
# 3. Remove protoc go:generate directives (CI has no protoc)
# 4. go mod tidy in this module and integration-tests
```
