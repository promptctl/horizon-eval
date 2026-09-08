package version

import (
	"strings"
	"testing"
)

// The version string is never empty: it is the package's own version when the
// build carries one, and the stable fallback token otherwise (appspec/00
// "Provenance").
func TestVersionIsAlwaysResolved(t *testing.T) {
	if got := Version(); got == "" {
		t.Fatal("Version() = \"\", want a version or the fallback token")
	}
}

// --version prints "Mackup <version>": the product token is the spec's, not
// the project's package name.
func TestStringUsesTheSpecProductToken(t *testing.T) {
	got := String()
	if !strings.HasPrefix(got, "Mackup ") {
		t.Errorf("String() = %q, want a %q prefix", got, "Mackup ")
	}
	if strings.TrimPrefix(got, "Mackup ") != Version() {
		t.Errorf("String() = %q, want %q", got, "Mackup "+Version())
	}
}

func TestNormalizeRejectsToolchainPlaceholders(t *testing.T) {
	tests := map[string]string{
		"v0.11.1": "0.11.1",
		"0.11.1":  "0.11.1",
		" v1.2.3": "1.2.3",
		"(devel)": "",
		"devel":   "",
		"":        "",
	}
	for in, want := range tests {
		if got := normalize(in); got != want {
			t.Errorf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

// A build with no metadata reports the fallback, not an empty or invented
// version.
func TestFallbackTokenIsStable(t *testing.T) {
	if Fallback != "unknown" {
		t.Errorf("Fallback = %q, want %q", Fallback, "unknown")
	}
}
