# feat(json/v2): add json/v2 marshaler/unmarshaler hooks

## Summary

Adds `encoding/json/v2` method implementations behind `//go:build goexperiment.jsonv2 && go1.27` so that Go 1.27+ builds — where json/v2 is the default — get the v2 fast path on the types that have custom v1 JSON behavior. Builds that opt out via `GOEXPERIMENT=nojsonv2`, and all pre-1.27 builds, are unchanged and use the v1 path. See [Build tags](#build-tags) for why both constraints are required.

Scope is intentionally narrow: a repo-wide audit found exactly two types with custom `MarshalJSON`/`UnmarshalJSON` methods, and both are addressed:

- `protocol.URLEncodedBase64` — adds `MarshalJSONTo` (zero-alloc fast path using `encoder.AvailableBuffer()` + `RawURLEncoding.AppendEncode`). Mirrors v1's `MarshalJSON` byte-for-byte: `nil → null`, `[]byte{} → ""`, otherwise raw URL-encoded base64.
- `webauthn.Credential` — adds `UnmarshalJSONFrom` that decodes via a `credentialAlias` shim (to avoid method recursion) and then runs the existing `AttestationType → AttestationFormat` legacy-record migration.

### Why no `URLEncodedBase64.UnmarshalJSONFrom`?

Deliberate, and measured. `encoding/json/v2/arshal_methods.go` wires up the v1 `json.Unmarshaler` in its unmarshal dispatch chain and only overrides it with `UnmarshalerFrom` when the latter is present. With `URLEncodedBase64` defining only the v1 `UnmarshalJSON`, json/v2 cleanly delegates to it — preserving v1's exact padding-trim and `null` semantics with no duplication.

A native method was benchmarked rather than assumed: an `UnmarshalJSONFrom` built on `decoder.ReadValue` is byte-for-byte parity with the delegation (both **113 ns / 2 allocs** on the go1.27 line), and a `decoder.ReadToken` variant is *slower* (**151 ns / 3 allocs** — the extra alloc is `Token.String`). Since a hand-written method cannot beat delegation here, none is added. A note in `protocol/base64_jsonv2.go` documents this for future maintainers.

### Build tags

The v2 files are gated on `//go:build goexperiment.jsonv2 && go1.27`. Both terms are load-bearing:

- `goexperiment.jsonv2` — true exactly when `encoding/json/v2` and `encoding/json/jsontext` are importable. In Go 1.27 they are default-on but still removed under `GOEXPERIMENT=nojsonv2` (the std packages carry this same constraint). A bare `//go:build go1.27` stays true under `nojsonv2`, so it would pull these files into a build where the imports no longer exist — breaking `go build` on the documented opt-out.
- `go1.27` — raises the files' effective language version to 1.27 so the `stdversion` vet check (run by `go test` by default as of 1.27) accepts the 1.27-only stdlib symbols. Without it, `stdversion` reports `requires go1.27 or later (file is go1.25)` against the module's `go 1.25.0` directive.

Consequence: v2 is used only on Go 1.27+ with the experiment active. The module stays at `go 1.25.0`, so the minimum supported version is unchanged; pre-1.27 `GOEXPERIMENT=jsonv2` builds fall back to v1.

## Changes

- `protocol/base64_jsonv2.go` (new, build-tagged) — `URLEncodedBase64.MarshalJSONTo`
- `webauthn/credential_jsonv2.go` (new, build-tagged) — `Credential.UnmarshalJSONFrom`
- Test files mirroring the existing `*_msgp_gen_test.go` structure (Marshal/Unmarshal + Encode/Decode, `Zero`/`Populated` variants):
  - `protocol/options_json_test.go`
  - `protocol/base64_test.go` (additions)
  - `protocol/base64_jsonv2_test.go` (build-tagged)
  - `webauthn/authenticator_json_test.go`
  - `webauthn/credential_json_test.go`
  - `webauthn/credential_jsonv2_test.go` (build-tagged)
  - `webauthn/types_session_json_test.go`
- `protocol/client_bench_test.go`, `protocol/decoder_bench_test.go` — benchmarks for `CollectedClientData.Unmarshal` and the `decodeBody`/`decodeBytes` request-path entry points. Both transparently route through `MarshalJSONTo` under `GOEXPERIMENT=jsonv2` and quantify the speedup.

`webauthn/credential_json_test.go` defines `assertJSONRoundTripEqual` for two cases where the Go value does not round-trip byte-for-byte: `CredentialFlags` carries an unexported raw byte dropped on unmarshal, and `SessionData.Extensions` is `map[string]any`, so numeric values come back as `float64`. The helper compares JSON output instead of Go values for these cases; the rationale is documented inline.

## Compatibility

- No public API change. New methods are additive and only compile under `//go:build goexperiment.jsonv2 && go1.27`.
- No new runtime dependencies.
- No changes to v1 behavior; existing `MarshalJSON`/`UnmarshalJSON` are untouched.

## Benchmarks

Median of 3 runs, `-benchtime=1s -benchmem`, Apple M4 / darwin/arm64. Go versions: `go1.26.4`, `gotip go1.27-devel_f2f369db`. All numbers in ns/op (lower is better); allocations per op shown in parentheses where they changed.

| Benchmark | go 1.26.4 (v1) | go 1.26.4 +`GOEXPERIMENT=jsonv2` | gotip 1.27 (default) | gotip 1.27 +`GOEXPERIMENT=nojsonv2` |
| --- | ---: | ---: | ---: | ---: |
| `URLEncodedBase64Marshal`             | 1608 (11 alloc) |  547 (4 alloc) |  535 (4 alloc) | 1672 (11 alloc) |
| `URLEncodedBase64Unmarshal`           | 1520 (8 alloc)  |  750 (4 alloc) |  756 (4 alloc) | 1583 (8 alloc)  |
| `CollectedClientData_Unmarshal`       |  742 (8 alloc)  |  330 (1 alloc) |  316 (1 alloc) |  731 (8 alloc)  |
| `DecodeBody_AssertionResponse`        | 4008 (27 alloc) | 2734 (21 alloc)| 2648 (21 alloc)| 4072 (27 alloc) |
| `JSONMarshalCredential` (webauthn)    |  498            |  762           |  787           |  513            |
| `JSONUnmarshalCredential` (webauthn)  | 8127 (27 alloc) | 2517 (12 alloc)| 2455 (12 alloc)| 7692 (27 alloc) |

### Independent re-verification (go1.27rc2)

Reproduced on `go1.27rc2` darwin/arm64, `-benchtime=1s -benchmem`. Columns are the two shipped paths: v1 via `GOEXPERIMENT=nojsonv2` and v2 default-on.

| Benchmark | v1 (`nojsonv2`) | v2 (default) |
| --- | ---: | ---: |
| `URLEncodedBase64Marshal`       | 1764 (11 alloc) |  527 (4 alloc)  |
| `URLEncodedBase64Unmarshal`     | 1622 (8 alloc)  |  741 (4 alloc)  |
| `CollectedClientData_Unmarshal` |  747 (8 alloc)  |  315 (1 alloc)  |
| `JSONMarshalCredential`         |  516 (1 alloc)  |  827 (1 alloc)  |
| `JSONUnmarshalCredential`       | 7694 (27 alloc) | 2495 (12 alloc) |

### Takeaways

- **`gotip default` ≈ `go 1.26.4 + GOEXPERIMENT=jsonv2`**, and **`gotip + nojsonv2` ≈ `go 1.26.4` default** — confirms json/v2 is the default in Go 1.27, and that opting back out via `GOEXPERIMENT=nojsonv2` restores the legacy path cleanly. Note: the shipped build tags enable v2 only on the `gotip 1.27 default` column; the `go 1.26.4 + GOEXPERIMENT=jsonv2` column exercises the same v2 implementation and is retained as corroborating evidence of the speedup, but pre-1.27 builds of this branch now use v1.
- **Significant speedups on the hot paths.** `URLEncodedBase64.Marshal` is ~3× faster with 64% fewer allocations. `CollectedClientData.Unmarshal` (called once per assertion / attestation request) is ~2.3× faster and drops from 8 allocs to 1. `Credential.Unmarshal` is ~3.2× faster with 56% fewer allocations.
- **One regression: `JSONMarshalCredential` is ~1.5–1.6× slower under v2** (487 → 788 ns/op on the re-verification run), with the same single allocation. Root cause confirmed by isolation: this is v2's reflection *marshaler* cost on the Credential struct shape, **not** the new `MarshalJSONTo`. A controlled benchmark of a struct with `URLEncodedBase64` sub-fields shows the v2 `MarshalJSONTo` field path is *faster* than a v1-only `MarshalJSON` field (438 ns / 3 allocs vs 531 ns / 12 allocs) — so this PR's method mitigates struct-marshal cost; the residual is upstream. Accepted and not worked around: a hand-written `Credential.MarshalJSONTo` would be substantial hand-maintained code for a marginal gain on the *rarer* emit path (servers receive credentials far more often than they emit them), so the net remains strongly positive.

## Test plan

- [x] `go test -race ./...` on Go 1.26.x — v1 path (v2 files excluded below 1.27)
- [x] `go test -race ./...` on Go 1.27+ — v2 path, incl. the default `stdversion` vet check
- [x] `GOEXPERIMENT=nojsonv2 go build ./...` on Go 1.27+ — v2 files excluded, v1 restored, builds cleanly
- [x] `go test -run=^$ -bench=. ./protocol ./webauthn` on Go 1.27rc2 — confirms benchmarks run and capture v2 speedup

Checked items verified locally with a Go 1.27-line toolchain (`gotip`). This PR **adds** a `jsonv2` job to `.github/workflows/go.yml` (Go 1.27 default + `GOEXPERIMENT=nojsonv2`) — previously CI ran only Go 1.25/1.26 with the default experiment, so the v2-tagged files were never compiled or tested. The new job runs the first three checks automatically and also asserts (via `go list | grep jsonv2`) that the v2 files are actually built by default and excluded under `nojsonv2`.

## Future: json/v2 graduation

The v2 files are gated on `goexperiment.jsonv2`. Whenever `encoding/json/v2` eventually *graduates* (the experiment is removed and the package becomes unconditionally available), the `goexperiment.jsonv2` build tag stops matching and these files would **silently drop out of the build**, reverting the package to the slower v1 path with no error. The new CI assertion (`go list -f '{{.GoFiles}}' ./protocol ./webauthn | grep -q jsonv2` on the latest Go) is the tripwire: it turns that silent fallback into a red build so the tags can be updated deliberately.
