// Package facts exposes git repository facts (current branch, origin remote,
// owner) and a forward-only check, all through an injected [Runner] so the logic
// is testable without invoking the real git binary.
package facts

import (
	"os/exec"
	"strings"

	errs "github.com/gomatic/go-error"
)

const (
	// ErrDetachedHead is returned when HEAD is not on a branch (a branch tip is required).
	ErrDetachedHead errs.Const = "git: HEAD is detached; not on a branch tip"
	// ErrNoOrigin is returned when remote.origin.url is not configured.
	ErrNoOrigin errs.Const = "git: no origin remote configured"
)

// Arg is a single git command-line argument.
type Arg string

// CommandOutput is the raw stdout of a git invocation.
type CommandOutput string

// BranchName is a git branch name.
type BranchName string

// OriginURL is the URL of the origin remote.
type OriginURL string

// Runner runs a git subcommand and returns its standard output.
type Runner interface {
	Run(args ...Arg) (CommandOutput, error)
}

// Branch returns the current branch name, or ErrDetachedHead when HEAD is detached.
func Branch(r Runner) (BranchName, error) {
	out, err := r.Run("symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", ErrDetachedHead.With(err)
	}
	return BranchName(strings.TrimSpace(string(out))), nil
}

// Origin returns the origin remote URL, or ErrNoOrigin when it is not configured.
func Origin(r Runner) (OriginURL, error) {
	out, err := r.Run("config", "--get", "remote.origin.url")
	if err != nil {
		return "", ErrNoOrigin.With(err)
	}
	return OriginURL(strings.TrimSpace(string(out))), nil
}

// Owner is the account/org that owns a repository (e.g. "nicerobot").
type Owner string

// OwnerOf returns the owner parsed from the upstream remote, falling back to
// origin. ErrNoOrigin is returned when neither remote is configured.
func OwnerOf(r Runner) (Owner, error) {
	for _, name := range []string{"upstream", "origin"} {
		if url, err := remoteURL(r, name); err == nil {
			return ownerFromURL(url), nil
		}
	}
	return "", ErrNoOrigin
}

func remoteURL(r Runner, name string) (OriginURL, error) {
	out, err := r.Run("config", "--get", Arg("remote."+name+".url"))
	if err != nil {
		return "", err
	}
	return OriginURL(strings.TrimSpace(string(out))), nil
}

// ownerFromURL extracts the owner from a git remote URL (https or scp-style),
// returning "" when it cannot be parsed.
func ownerFromURL(u OriginURL) Owner {
	s := strings.TrimSuffix(string(u), ".git")
	if i := strings.Index(s, "://"); i >= 0 {
		s = s[i+3:]
	}
	if i := strings.LastIndex(s, "@"); i >= 0 {
		s = s[i+1:]
	}
	s = strings.Replace(s, ":", "/", 1)
	if parts := strings.Split(s, "/"); len(parts) >= 2 {
		return Owner(parts[1])
	}
	return ""
}

// EnsureForwardOnly verifies the working tree satisfies the forward-only
// invariant: HEAD is on a branch tip rather than a detached commit.
func EnsureForwardOnly(r Runner) error {
	_, err := Branch(r)
	return err
}

// ExecRunner is a Runner backed by the real git binary.
type ExecRunner struct{}

// NewExecRunner returns a Runner that invokes git.
func NewExecRunner() ExecRunner { return ExecRunner{} }

// Run invokes git with the given arguments and returns its standard output.
func (ExecRunner) Run(args ...Arg) (CommandOutput, error) {
	strArgs := make([]string, len(args))
	for i, a := range args {
		strArgs[i] = string(a)
	}
	out, err := exec.Command("git", strArgs...).Output()
	if err != nil {
		return "", err
	}
	return CommandOutput(out), nil
}
