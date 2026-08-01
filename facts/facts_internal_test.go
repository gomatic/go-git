package facts

// White-box test for remote-URL parsing. Git remotes come in two unrelated
// syntaxes in which the colon means opposite things, so the parser is the one
// place a wrong answer looks exactly like a right one.

import "testing"

// TestOwnerFromURLParsesBothRemoteSyntaxesAndRefusesTheRest names
// ownerFromURL's claim. Git remotes come in two unrelated shapes and the colon
// means opposite things in each: in scp-style `host:owner/repo` it separates
// host from path, and in a URL `https://host:443/owner/repo` it introduces a
// port. Treating one as the other silently yields the wrong owner — not an
// error, an answer — and every downstream decision keyed on ownership then
// applies to the wrong org.
//
// Returning "" for anything unparseable is the other half: a guess would be
// worse than a refusal here, because the caller has no way to tell one from the
// other.
func TestOwnerFromURLParsesBothRemoteSyntaxesAndRefusesTheRest(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		url  OriginURL
		want Owner
		why  string
	}{
		{url: "git@github.com:gomatic/go-git.git", want: "gomatic", why: "scp-style with userinfo"},
		{url: "github.com:gomatic/go-git", want: "gomatic", why: "scp-style without userinfo or suffix"},
		{url: "https://github.com/gomatic/go-git.git", want: "gomatic", why: "https URL"},
		{url: "https://user@github.com/gomatic/go-git", want: "gomatic", why: "URL with userinfo"},
		{url: "ssh://git@github.com/gomatic/go-git.git", want: "gomatic", why: "ssh URL"},
		{
			url: "https://github.com:443/gomatic/go-git", want: "gomatic",
			why: "a colon in a URL is a PORT and must not be read as an scp separator",
		},

		{url: "", want: "", why: "an empty remote"},
		{url: "not-a-remote", want: "", why: "no owner/repo path at all"},
		{url: "https://github.com/", want: "", why: "a URL with no path"},
	} {
		if got := ownerFromURL(tc.url); got != tc.want {
			t.Errorf("ownerFromURL(%q) = %q, want %q: %s", tc.url, got, tc.want, tc.why)
		}
	}
}
