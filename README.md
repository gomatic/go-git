# go-git

Git repository facts for Go, behind an injected runner so callers stay testable without shelling out to the real `git` binary.

The `facts` package reads the few facts tools commonly need — the current branch, the origin remote URL, and the repository owner — and offers a forward-only check (HEAD is on a branch tip, not detached). Every operation goes through a small [`Runner`](facts/facts.go) interface; `ExecRunner` is the real-`git` implementation, and tests inject a fake.

## Install

```sh
go get github.com/gomatic/go-git/facts
```

Requires Go 1.26+.

## Usage

```go
package main

import (
	"fmt"

	"github.com/gomatic/go-git/facts"
)

func main() {
	r := facts.NewExecRunner()

	branch, err := facts.Branch(r) // ErrDetachedHead when HEAD is detached
	if err != nil {
		panic(err)
	}

	owner, err := facts.OwnerOf(r) // upstream remote, falling back to origin
	if err != nil {
		panic(err)
	}

	fmt.Printf("on %s, owned by %s\n", branch, owner)
}
```

`Branch`, `Origin`, `OwnerOf`, and `EnsureForwardOnly` each take a `Runner`, so a caller can drive them against a fake in tests:

```go
type fakeRunner struct {
	fn func(args ...facts.Arg) (facts.CommandOutput, error)
}

func (f fakeRunner) Run(args ...facts.Arg) (facts.CommandOutput, error) { return f.fn(args...) }
```

Errors are [`gomatic/go-error`](https://github.com/gomatic/go-error) sentinels (`facts.ErrDetachedHead`, `facts.ErrNoOrigin`), matchable with `errors.Is`.
