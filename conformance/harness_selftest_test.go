package conformance

import (
	"os"
	"path/filepath"
	"reflect"
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
		"real":    "contents",
		"pointer": "link -> " + target,
		"sub":     "dir",
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
		".vimrc":           "set nocompatible\n",
		".config":          "dir",
		"nested":           "dir",
		"nested/deep":      "dir",
		"nested/deep/file": "x",
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
