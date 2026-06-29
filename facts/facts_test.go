package facts_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/gomatic/go-git/facts"
)

// fakeRunner drives the facts package without invoking the real git binary.
type fakeRunner struct {
	fn func(args ...facts.Arg) (facts.CommandOutput, error)
}

func (f fakeRunner) Run(args ...facts.Arg) (facts.CommandOutput, error) { return f.fn(args...) }

var errFake = errors.New("fake failure")

func TestBranch(t *testing.T) {
	r := fakeRunner{fn: func(...facts.Arg) (facts.CommandOutput, error) { return "master\n", nil }}
	got, err := facts.Branch(r)
	if err != nil {
		t.Fatalf("Branch returned error: %v", err)
	}
	if got != "master" {
		t.Fatalf("Branch = %q, want %q", got, "master")
	}
}

func TestBranchDetached(t *testing.T) {
	r := fakeRunner{fn: func(...facts.Arg) (facts.CommandOutput, error) { return "", errFake }}
	got, err := facts.Branch(r)
	if !errors.Is(err, facts.ErrDetachedHead) {
		t.Fatalf("Branch error = %v, want ErrDetachedHead", err)
	}
	if !errors.Is(err, errFake) {
		t.Fatalf("Branch error = %v, want wrapped cause errFake", err)
	}
	if got != "" {
		t.Fatalf("Branch = %q, want empty on error", got)
	}
}

func TestOrigin(t *testing.T) {
	want := "git@github.com:gomatic/go-git.git"
	r := fakeRunner{fn: func(...facts.Arg) (facts.CommandOutput, error) {
		return facts.CommandOutput(want + "\n"), nil
	}}
	got, err := facts.Origin(r)
	if err != nil {
		t.Fatalf("Origin returned error: %v", err)
	}
	if string(got) != want {
		t.Fatalf("Origin = %q, want %q", got, want)
	}
}

func TestOriginMissing(t *testing.T) {
	r := fakeRunner{fn: func(...facts.Arg) (facts.CommandOutput, error) { return "", errFake }}
	got, err := facts.Origin(r)
	if !errors.Is(err, facts.ErrNoOrigin) {
		t.Fatalf("Origin error = %v, want ErrNoOrigin", err)
	}
	if !errors.Is(err, errFake) {
		t.Fatalf("Origin error = %v, want wrapped cause errFake", err)
	}
	if got != "" {
		t.Fatalf("Origin = %q, want empty on error", got)
	}
}

func TestOwnerOfUpstream(t *testing.T) {
	r := fakeRunner{fn: func(a ...facts.Arg) (facts.CommandOutput, error) {
		if a[2] == "remote.upstream.url" {
			return "https://github.com/up-owner/repo.git\n", nil
		}
		return "", errFake
	}}
	got, err := facts.OwnerOf(r)
	if err != nil || got != "up-owner" {
		t.Fatalf("OwnerOf = %q, %v", got, err)
	}
}

func TestOwnerOfOriginFallback(t *testing.T) {
	r := fakeRunner{fn: func(a ...facts.Arg) (facts.CommandOutput, error) {
		if a[2] == "remote.origin.url" {
			return "git@github.com:org-owner/repo.git\n", nil
		}
		return "", errFake
	}}
	got, err := facts.OwnerOf(r)
	if err != nil || got != "org-owner" {
		t.Fatalf("OwnerOf = %q, %v", got, err)
	}
}

func TestOwnerOfSSHPort(t *testing.T) {
	// A scheme-qualified ssh URL with an explicit port must not let the port
	// colon shift the owner index.
	r := fakeRunner{fn: func(a ...facts.Arg) (facts.CommandOutput, error) {
		if a[2] == "remote.upstream.url" {
			return "ssh://git@github.com:22/port-owner/repo.git\n", nil
		}
		return "", errFake
	}}
	got, err := facts.OwnerOf(r)
	if err != nil || got != "port-owner" {
		t.Fatalf("OwnerOf = %q, %v, want port-owner", got, err)
	}
}

func TestOwnerOfNone(t *testing.T) {
	r := fakeRunner{fn: func(...facts.Arg) (facts.CommandOutput, error) { return "", errFake }}
	if _, err := facts.OwnerOf(r); !errors.Is(err, facts.ErrNoOrigin) {
		t.Fatalf("OwnerOf error = %v, want ErrNoOrigin", err)
	}
}

func TestOwnerOfUnparseable(t *testing.T) {
	// A configured upstream that parses to an empty owner must not shadow a
	// valid origin: OwnerOf skips the empty owner and falls through to origin.
	r := fakeRunner{fn: func(a ...facts.Arg) (facts.CommandOutput, error) {
		if a[2] == "remote.upstream.url" {
			return "nonsense\n", nil
		}
		return "git@github.com:org-owner/repo.git\n", nil
	}}
	got, err := facts.OwnerOf(r)
	if err != nil || got != "org-owner" {
		t.Fatalf("OwnerOf = %q, %v, want org-owner via fall-through", got, err)
	}
}

func TestOwnerOfEmptyOnly(t *testing.T) {
	// When the only configured remote parses to an empty owner, OwnerOf must
	// report ErrNoOrigin rather than a silent empty owner.
	r := fakeRunner{fn: func(a ...facts.Arg) (facts.CommandOutput, error) {
		if a[2] == "remote.upstream.url" {
			return "nonsense\n", nil
		}
		return "", errFake
	}}
	if _, err := facts.OwnerOf(r); !errors.Is(err, facts.ErrNoOrigin) {
		t.Fatalf("OwnerOf error = %v, want ErrNoOrigin", err)
	}
}

func TestEnsureForwardOnly(t *testing.T) {
	onBranch := fakeRunner{fn: func(...facts.Arg) (facts.CommandOutput, error) { return "master\n", nil }}
	if err := facts.EnsureForwardOnly(onBranch); err != nil {
		t.Fatalf("EnsureForwardOnly on a branch returned error: %v", err)
	}

	detached := fakeRunner{fn: func(...facts.Arg) (facts.CommandOutput, error) { return "", errFake }}
	if err := facts.EnsureForwardOnly(detached); !errors.Is(err, facts.ErrDetachedHead) {
		t.Fatalf("EnsureForwardOnly detached error = %v, want ErrDetachedHead", err)
	}
}

func TestExecRunner(t *testing.T) {
	r := facts.NewExecRunner()

	out, err := r.Run("--version")
	if err != nil {
		t.Fatalf("ExecRunner git --version returned error: %v", err)
	}
	if !strings.Contains(string(out), "git version") {
		t.Fatalf("ExecRunner git --version = %q, want it to contain %q", out, "git version")
	}

	if _, err := r.Run("definitely-not-a-real-subcommand"); err == nil {
		t.Fatal("ExecRunner expected error for a bogus git subcommand, got nil")
	}
}

func TestErrorString(t *testing.T) {
	if facts.ErrNoOrigin.Error() != string(facts.ErrNoOrigin) {
		t.Fatal("Error() does not return the underlying string")
	}
}
