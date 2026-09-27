// Package conformance is the project's black-box test rig.
//
// The specification in appspec/ defines only observable behavior, so every
// behavioral claim is checked here the way appspec/00-overview.md Provenance
// describes the spec being written: run the real command under a throwaway home
// directory and observe stdout, stderr, the exit code, and what happened on
// disk. Nothing in this package reaches inside the program.
//
// Run the suite with `make conformance` (or `go test ./conformance/`).
package conformance

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// binary is the compiled command under test, built once for the whole suite.
var binary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "mackup-conformance")
	if err != nil {
		panic(err)
	}

	binary = filepath.Join(dir, "mackup")
	build := exec.Command("go", "build", "-o", binary, "../cmd/mackup")
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		os.RemoveAll(dir)
		panic(err)
	}

	// os.Exit skips deferred calls, so the cleanup is explicit.
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// Invocation is one run of the command: everything the program is allowed to
// read, per appspec/00-overview.md "The boundary".
type Invocation struct {
	// Args is argv without the program name.
	Args []string
	// Env adds to or overrides the scrubbed base environment. Use it for the
	// three variables the spec names: HOME, XDG_CONFIG_HOME, MACKUP_CONFIG.
	Env map[string]string
	// Stdin is what the program reads when it prompts.
	Stdin string
	// Home seeds the throwaway home directory: each key is a home-relative
	// path, and a key ending in "/" makes a directory instead of a file.
	Home map[string]string
}

// Result is what a caller observes from one run.
type Result struct {
	// Stdout and Stderr are raw, including any SGR colour sequences. appspec/07
	// specifies colour as emitted unconditionally, even when piped, so a case
	// asserting on text should use PlainStdout/PlainStderr and one asserting on
	// colour should read these.
	Stdout string
	Stderr string
	Code   int
	// Home is the throwaway home directory this run used.
	Home string
	// HomeBefore is a Snapshot of Home taken after seeding and before the run,
	// so a case can assert the run changed nothing.
	HomeBefore map[string]string
}

// sgr matches an ANSI SGR (colour) escape sequence.
var sgr = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// PlainStdout is Stdout with colour sequences removed.
func (r Result) PlainStdout() string { return sgr.ReplaceAllString(r.Stdout, "") }

// PlainStderr is Stderr with colour sequences removed.
func (r Result) PlainStderr() string { return sgr.ReplaceAllString(r.Stderr, "") }

// Run executes the command for one invocation and returns what it observed.
//
// The environment is built from scratch rather than inherited: a developer's own
// MACKUP_CONFIG, XDG_CONFIG_HOME, or HOME must not be able to change what a
// conformance run observes.
func Run(t *testing.T, inv Invocation) Result {
	t.Helper()

	home := t.TempDir()
	seedHome(t, home, inv.Home)

	env := map[string]string{
		"HOME": home,
		// Kept because appspec/06 and 07 have the program invoke external
		// commands for filesystem attribute cleanup.
		"PATH": os.Getenv("PATH"),
	}
	for k, v := range inv.Env {
		env[k] = v
	}
	envv := make([]string, 0, len(env))
	for k, v := range env {
		envv = append(envv, k+"="+v)
	}

	before := Snapshot(t, home)

	cmd := exec.Command(binary, inv.Args...)
	cmd.Env = envv
	cmd.Dir = home
	cmd.Stdin = strings.NewReader(inv.Stdin)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb

	code := 0
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("running %v: %v", inv.Args, err)
		}
		code = exitErr.ExitCode()
	}

	return Result{
		Stdout:     out.String(),
		Stderr:     errb.String(),
		Code:       code,
		Home:       home,
		HomeBefore: before,
	}
}

// seedHome writes the files and directories a case wants in place before the
// run. A path ending in "/" is a directory; any parent directories are created.
func seedHome(t *testing.T, home string, tree map[string]string) {
	t.Helper()
	for path, contents := range tree {
		full := filepath.Join(home, path)
		if strings.HasSuffix(path, "/") {
			if err := os.MkdirAll(full, 0o700); err != nil {
				t.Fatalf("seeding %s: %v", path, err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
			t.Fatalf("seeding %s: %v", path, err)
		}
		if err := os.WriteFile(full, []byte(contents), 0o600); err != nil {
			t.Fatalf("seeding %s: %v", path, err)
		}
	}
}

// Snapshot records a directory tree's observable content, so a case can assert
// that a run changed nothing — the post-condition appspec/07 states for every
// error path and appspec/01 §3 states for --dry-run.
//
// Each key is a path relative to root. The value describes what is there: a
// symlink's target, "dir", or a regular file's contents. Symlinks are recorded,
// never followed, because whether a home path is a link is itself contract
// (appspec/01 §2).
func Snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	tree := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		switch {
		case d.Type()&os.ModeSymlink != 0:
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			tree[rel] = "link -> " + target
		case d.IsDir():
			tree[rel] = "dir"
		default:
			contents, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			tree[rel] = string(contents)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("snapshotting %s: %v", root, err)
	}
	return tree
}

// AssertUnchanged reports every difference between a tree and an earlier
// snapshot of it.
func AssertUnchanged(t *testing.T, root string, before map[string]string) {
	t.Helper()
	for _, diff := range diffTree(before, Snapshot(t, root)) {
		t.Errorf("%s, want the tree unchanged", diff)
	}
}

// diffTree describes every difference between two snapshots, in sorted path
// order. It is separate from AssertUnchanged so the rig's own change detection is
// testable: an assertion helper that cannot fail asserts nothing.
func diffTree(before, after map[string]string) []string {
	var diffs []string
	for _, path := range sortedKeys(before) {
		want := before[path]
		got, ok := after[path]
		switch {
		case !ok:
			diffs = append(diffs, path+" was removed")
		case got != want:
			diffs = append(diffs, fmt.Sprintf("%s changed to %q, was %q", path, got, want))
		}
	}
	for _, path := range sortedKeys(after) {
		if _, ok := before[path]; !ok {
			diffs = append(diffs, path+" was created")
		}
	}
	return diffs
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
