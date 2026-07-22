## What changed and why

<!-- The product reasoning matters as much as the diff. -->

## Checklist

- [ ] `go test ./...` passes; new behavior is pinned by a test
- [ ] `go run ./cmd/orven validate plugins/demo-activity examples/radarr-queue` still reports 0 errors / 0 warnings
- [ ] Conforms to [docs/CONSTRAINTS.md](../blob/main/docs/CONSTRAINTS.md) (read-only, facts only, calm voice, additive-only contract)
- [ ] User-visible text reads like the rest of the product
- [ ] Docs updated where behavior changed
