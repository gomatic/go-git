# go-git

Git repository facts for Go, behind an injected runner. The `facts` package reads the current branch, the origin remote URL, and the repository owner, and offers a forward-only check (HEAD on a branch tip, not detached) — all through a small `Runner` interface so callers are testable without invoking the real `git` binary. Extracted from `skykernel/skym`'s `internal/git`; generic — lives in `gomatic`.

- Package `facts` (import `github.com/gomatic/go-git/facts`): `Branch`, `Origin`, `OwnerOf`, `EnsureForwardOnly` over a `Runner`; `ExecRunner` is the real-`git` implementation; `ownerFromURL` parses https and scp-style remotes.
- Errors are [`gomatic/go-error`](https://github.com/gomatic/go-error) sentinels (`ErrDetachedHead`, `ErrNoOrigin`), matched with `errors.Is`. Never `errors.New`; `fmt.Errorf` only to wrap with `%w` (prefer `errs.Const.With`).
- Value receivers, immutable types, private by default. Gate: gofumpt, vet, staticcheck, govulncheck, gocognit ≤ 7, 100% coverage.
- `Makefile`, `.golangci.yaml`, `.editorconfig`, `.gitignore`, `.github/` are owned and pushed by `nicerobot/tools.repository` — do not edit in-tree; per-repo divergence goes in a `Makefile.local`.
