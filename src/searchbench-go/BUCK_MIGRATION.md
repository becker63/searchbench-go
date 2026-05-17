# Go BUCK migration ledger (#93)

Per-package `BUCK` files under `src/searchbench-go/` are generated from `go list` via `tools/generate_go_buck.py`. The module-root `BUCK` (check suite, Pkl targets) is hand-maintained.

## Regenerating

```bash
cd src/searchbench-go && go mod vendor
python3 tools/generate_vendor_buck.py
python3 tools/generate_go_buck.py
nix develop -c python3 tools/generate_go_check_tests.py
```

## Label conventions

| Kind | Label pattern |
| --- | --- |
| First-party library | `//src/searchbench-go/internal/...:<pkg>` |
| Vendor | `//src/searchbench-go/vendor/<import/path>:<last_segment>` |
| CLI binary | `//src/searchbench-go/cmd/searchbench:searchbench` |
| Package tests | `//src/searchbench-go/...:<pkg>_test` (`go_external_package_test`) |

## Special cases

- **External test packages** (`package foo_test`): `go_external_package_test` in `build_defs/go_external_test.bzl` (runs `go test` from module root).
- **CGO / treesitter**: tests use `//go:build cgo`; default check runs with `CGO_ENABLED=0` (no cgo tests).
- **Embeds**: `embed_srcs` on `go_library` (expanded paths, not globs).
- **Module root**: not overwritten by the generator.
