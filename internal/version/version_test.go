package version

import "testing"

func TestStringFallsBackWhenNoPackageVersionIsAvailable(t *testing.T) {
	// A test binary is built from the tree, so no module version is available —
	// the spec's uninstalled-tree case.
	if got := String(); got != Fallback {
		t.Errorf("String() = %q, want %q", got, Fallback)
	}
}

func TestStringPrefersTheStampedVersion(t *testing.T) {
	t.Cleanup(func() { override = "" })
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
