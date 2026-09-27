package conformance

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
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
		// Names the home itself rather than an entry in it.
		".",
		"nested/..",
	} {
		if got, err := resolveInHome(home, path); err == nil {
			t.Errorf("resolveInHome(%q) = %q, want it rejected", path, got)
		}
	}
}

func TestResolveInHomeAcceptsPathsInsideTheHome(t *testing.T) {
	home := t.TempDir()
	for path, want := range map[string]string{
		".vimrc":            filepath.Join(home, ".vimrc"),
		"nested/deep/file":  filepath.Join(home, "nested/deep/file"),
		"nested/../.gitcfg": filepath.Join(home, ".gitcfg"),
		".config/":          filepath.Join(home, ".config"),
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

// The rig's whole isolation promise lives in how Run wires the child process, and
// the command under test cannot show it: it does not read its environment yet, so
// deleting `cmd.Env = envv` would leave every conformance case green while every
// run inherited the developer's real HOME, MACKUP_CONFIG, and PATH — and every
// HomeUnchanged assertion became vacuous. This runs a probe through the same
// wiring and reads back what actually arrived.
func TestRunDeliversTheScrubbedEnvironmentWorkingDirectoryAndSeededHome(t *testing.T) {
	// A variable set in this process is exactly what an inherited environment
	// would leak into the child.
	t.Setenv("MACKLEBOX_SELFTEST_SENTINEL", "leaked")

	probe := `printf 'pwd=%s\n' "$(pwd)"; printf 'home=%s\n' "$HOME"; ` +
		`printf 'seeded=%s' "$(cat .vimrc)"; printf '\n--env--\n'; env | sort`

	r := runProgram(t, "/bin/sh", Invocation{
		Args: []string{"-c", probe},
		Env:  map[string]string{"MACKUP_CONFIG": "sentinel.cfg"},
		Home: map[string]string{".vimrc": "seeded contents"},
	})
	if r.Code != 0 {
		t.Fatalf("probe exited %d: %s", r.Code, r.Stderr)
	}

	head, envBlock, found := strings.Cut(r.Stdout, "\n--env--\n")
	if !found {
		t.Fatalf("probe output has no env block: %q", r.Stdout)
	}

	// The working directory is the throwaway home. Compared through
	// EvalSymlinks because a temp directory reaches it via a symlink on macOS.
	wantHome, err := filepath.EvalSymlinks(r.Home)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"pwd=" + wantHome,
		"home=" + r.Home,
		"seeded=seeded contents",
	} {
		if !strings.Contains(head, want) {
			t.Errorf("probe reported %q, want it to contain %q", head, want)
		}
	}

	// The environment is childEnv's and nothing else. This is the assertion that
	// fails if Run stops setting cmd.Env: an inherited environment would both
	// lose HOME=<throwaway> and carry the host's own variables.
	wantEnv, err := childEnv(r.Home, map[string]string{"MACKUP_CONFIG": "sentinel.cfg"})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(wantEnv)
	gotEnv := strings.Split(strings.TrimSuffix(envBlock, "\n"), "\n")

	for _, want := range wantEnv {
		if !slices.Contains(gotEnv, want) {
			t.Errorf("child environment = %q, want it to contain %q", gotEnv, want)
		}
	}
	// The probe shell sets these three itself; anything else is inherited.
	shellInjected := map[string]bool{"PWD": true, "SHLVL": true, "_": true}
	wanted := map[string]bool{}
	for _, line := range wantEnv {
		name, _, _ := strings.Cut(line, "=")
		wanted[name] = true
	}
	for _, line := range gotEnv {
		name, _, _ := strings.Cut(line, "=")
		if !wanted[name] && !shellInjected[name] {
			t.Errorf("%s reached the child; the environment is being inherited, not built", line)
		}
	}
}

// HomeBefore is taken after seeding and before the run, which is what makes
// HomeUnchanged mean anything. A probe that writes proves the ordering.
func TestHomeBeforeIsSnapshottedAfterSeedingAndBeforeTheRun(t *testing.T) {
	r := runProgram(t, "/bin/sh", Invocation{
		Args: []string{"-c", "printf written > .written"},
		Home: map[string]string{".vimrc": "seeded contents"},
	})
	if r.Code != 0 {
		t.Fatalf("probe exited %d: %s", r.Code, r.Stderr)
	}
	if got, want := r.HomeBefore[".vimrc"], "file 0600 seeded contents"; got != want {
		t.Errorf("HomeBefore[.vimrc] = %q, want %q — the snapshot must follow seeding", got, want)
	}
	if _, ok := r.HomeBefore[".written"]; ok {
		t.Error("HomeBefore contains a file the run created; the snapshot must precede the run")
	}
	diffs := diffTree(r.HomeBefore, Snapshot(t, r.Home))
	if len(diffs) != 1 || !strings.Contains(diffs[0], ".written was created") {
		t.Errorf("diffTree = %q, want it to report only the created file", diffs)
	}
}
