package version

import "testing"

func TestResolve(t *testing.T) {
	tests := []struct {
		name  string
		stamp string
		build string
		want  string
	}{
		{"uninstalled tree", "", "(devel)", Fallback},
		{"no build info", "", "", Fallback},
		{"devel without parens", "", "devel", Fallback},
		{"installed module version loses its v prefix", "", "v0.11.1", "0.11.1"},
		{"prerelease module version", "", "v1.2.3-rc.1", "1.2.3-rc.1"},
		{"unprefixed build version is left alone", "", "0.11.1", "0.11.1"},
		{"a link-time stamp wins", "0.11.1", "v9.9.9", "0.11.1"},
		{"a v-prefixed stamp loses its prefix too", "v0.11.1", "", "0.11.1"},
		{"a link-time stamp wins over no build info", "0.11.1", "", "0.11.1"},
	}
	for _, tt := range tests {
		if got := resolve(tt.stamp, tt.build); got != tt.want {
			t.Errorf("%s: resolve(%q, %q) = %q, want %q", tt.name, tt.stamp, tt.build, got, tt.want)
		}
	}
}

// A test binary is built from the tree, so with no link-time stamp no module
// version is available — the spec's uninstalled-tree case.
func TestStringFallsBackWhenNoPackageVersionIsAvailable(t *testing.T) {
	prior := override
	t.Cleanup(func() { override = prior })
	override = ""
	if got := String(); got != Fallback {
		t.Errorf("String() = %q, want %q", got, Fallback)
	}
}

func TestStringPrefersTheStampedVersion(t *testing.T) {
	prior := override
	t.Cleanup(func() { override = prior })
	override = "0.11.1"
	if got := String(); got != "0.11.1" {
		t.Errorf("String() = %q, want %q", got, "0.11.1")
	}
}

func TestStringIsNeverEmpty(t *testing.T) {
	if String() == "" {
		t.Error("String() is empty; a version string must always be reported")
	}
}
