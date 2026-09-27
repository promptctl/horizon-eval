package conformance

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// The rig's own change detection has to work, or every HomeUnchanged assertion
// in the suite is silently vacuous.
func TestDiffTreeDetectsEveryKindOfChange(t *testing.T) {
	before := map[string]string{
		".vimrc":   "old",
		".config":  "dir",
		".gitcfg":  "kept",
		".removed": "gone soon",
	}
	after := map[string]string{
		".vimrc":  "new",
		".config": "dir",
		".gitcfg": "kept",
		".added":  "brand new",
	}
	// Changes to pre-existing paths come first in sorted path order, then
	// creations.
	want := []string{
		".removed was removed",
		`.vimrc changed to "new", was "old"`,
		".added was created",
	}
	got := diffTree(before, after)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("diffTree = %q, want %q", got, want)
	}
}

func TestDiffTreeFindsNothingInAnUnchangedTree(t *testing.T) {
	tree := map[string]string{".vimrc": "x", ".config": "dir"}
	if got := diffTree(tree, tree); got != nil {
		t.Errorf("diffTree = %q, want no differences", got)
	}
}

// Snapshot must record what is at a path, not what it resolves to: whether a
// home path is a symlink is itself contract (appspec/01 §2).
func TestSnapshotRecordsSymlinksWithoutFollowingThem(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "real")
	if err := os.WriteFile(target, []byte("contents"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, "pointer")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "sub"), 0o700); err != nil {
		t.Fatal(err)
	}

	got := Snapshot(t, root)
	want := map[string]string{
		"real":    "file 0600 contents",
		"pointer": "link -> " + target,
		"sub":     "dir 0700",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Snapshot = %v, want %v", got, want)
	}
}

// Seeding is what makes a case's starting state explicit; a trailing slash means
// a directory.
func TestSeedHomeCreatesFilesAndDirectories(t *testing.T) {
	home := t.TempDir()
	seedHome(t, home, map[string]string{
		".vimrc":           "set nocompatible\n",
		".config/":         "",
		"nested/deep/file": "x",
	})
	got := Snapshot(t, home)
	want := map[string]string{
		".vimrc":           "file 0600 set nocompatible\n",
		".config":          "dir 0700",
		"nested":           "dir 0700",
		"nested/deep":      "dir 0700",
		"nested/deep/file": "file 0600 x",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("seeded tree = %v, want %v", got, want)
	}
}

// PlainStdout/PlainStderr strip colour so a case can assert on text, since
// appspec/07 makes colour unconditional but only the text and stream are
// contract.
func TestPlainStreamsStripColour(t *testing.T) {
	r := Result{
		Stdout: "\x1b[33mBacking up .vimrc\x1b[0m\n",
		Stderr: "\x1b[91mError: nope\x1b[0m\n",
	}
	if got, want := r.PlainStdout(), "Backing up .vimrc\n"; got != want {
		t.Errorf("PlainStdout = %q, want %q", got, want)
	}
	if got, want := r.PlainStderr(), "Error: nope\n"; got != want {
		t.Errorf("PlainStderr = %q, want %q", got, want)
	}
}

// appspec/06 "Attribute cleanup" has the program chmod files and shell out to
// strip ACLs and immutable flags, so a mode change with identical bytes is a real
// filesystem change and HomeUnchanged has to catch it.
func TestSnapshotSeesAModeChange(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".vimrc")
	if err := os.WriteFile(path, []byte("unchanged"), 0o600); err != nil {
		t.Fatal(err)
	}
	before := Snapshot(t, root)

	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	diffs := diffTree(before, Snapshot(t, root))
	if len(diffs) != 1 {
		t.Fatalf("diffTree = %q, want one difference for the mode change", diffs)
	}
	if !strings.Contains(diffs[0], "0644") || !strings.Contains(diffs[0], "0600") {
		t.Errorf("diffTree = %q, want it to report both modes", diffs)
	}
}

// A seed path must not be able to write outside the throwaway home: the parent
// temp directory holds sibling cases' homes, and Snapshot walks only this one, so
// an escaped write would be invisible to HomeUnchanged as well.
func TestResolveInHomeRejectsEscapingPaths(t *testing.T) {
	home := t.TempDir()
	for _, path := range []string{
		"../outside.cfg",
		"..",
		"nested/../../outside.cfg",
	} {
		if got, err := resolveInHome(home, path); err == nil {
			t.Errorf("resolveInHome(%q) = %q, want it rejected", path, got)
		}
	}
}

func TestResolveInHomeAcceptsPathsInsideTheHome(t *testing.T) {
	home := t.TempDir()
	for path, want := range map[string]string{
		".vimrc":           filepath.Join(home, ".vimrc"),
		"nested/deep/file": filepath.Join(home, "nested/deep/file"),
		"nested/..":        home,
		".config/":         filepath.Join(home, ".config"),
	} {
		got, err := resolveInHome(home, path)
		if err != nil {
			t.Errorf("resolveInHome(%q): %v", path, err)
			continue
		}
		if got != want {
			t.Errorf("resolveInHome(%q) = %q, want %q", path, got, want)
		}
	}
}
