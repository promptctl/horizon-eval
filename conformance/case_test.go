package conformance

import (
	"strings"
	"testing"
)

// A Check asserts something about one captured stream. stream is the stream's
// name, for the failure message.
type Check func(t *testing.T, stream, text string)

// Empty requires the stream to carry nothing at all.
func Empty() Check {
	return func(t *testing.T, stream, text string) {
		t.Helper()
		if text != "" {
			t.Errorf("%s = %q, want empty", stream, text)
		}
	}
}

// Exactly requires the stream to be the given text, byte for byte. Use it for
// the spec's literal contract tokens.
func Exactly(want string) Check {
	return func(t *testing.T, stream, text string) {
		t.Helper()
		if text != want {
			t.Errorf("%s = %q, want exactly %q", stream, text, want)
		}
	}
}

// Contains requires every substring to appear somewhere in the stream. Use it
// for human-facing wording, which appspec/02 states is not a machine-read
// contract.
func Contains(subs ...string) Check {
	return func(t *testing.T, stream, text string) {
		t.Helper()
		for _, sub := range subs {
			if !strings.Contains(text, sub) {
				t.Errorf("%s = %q, want it to contain %q", stream, text, sub)
			}
		}
	}
}

// InOrder requires every substring to appear, each after the one before it.
func InOrder(subs ...string) Check {
	return func(t *testing.T, stream, text string) {
		t.Helper()
		rest := text
		for _, sub := range subs {
			i := strings.Index(rest, sub)
			if i < 0 {
				t.Errorf("%s = %q, want it to contain %q after the substrings before it", stream, text, sub)
				return
			}
			rest = rest[i+len(sub):]
		}
	}
}

// Case is one conformance case: an invocation plus what a caller must observe.
//
// Stdout and Stderr are checked against the colour-stripped streams, since
// appspec/07 makes colour unconditional but only the text and the stream it
// lands on are contract. A case about colour itself reads Result.Stdout.
type Case struct {
	Name  string
	Args  []string
	Stdin string
	Env   map[string]string
	Home  map[string]string

	Code   int
	Stdout Check
	Stderr Check

	// HomeUnchanged requires the run to have left the home directory exactly as
	// it was seeded — the post-condition appspec/07 states for error paths and
	// appspec/01 §3 states for --dry-run.
	HomeUnchanged bool
}

// RunCases runs every case as its own subtest.
func RunCases(t *testing.T, cases []Case) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			r := Run(t, Invocation{
				Args:  c.Args,
				Env:   c.Env,
				Stdin: c.Stdin,
				Home:  c.Home,
			})
			if r.Code != c.Code {
				t.Errorf("exit = %d, want %d (stdout %q, stderr %q)",
					r.Code, c.Code, r.PlainStdout(), r.PlainStderr())
			}
			if c.Stdout != nil {
				c.Stdout(t, "stdout", r.PlainStdout())
			}
			if c.Stderr != nil {
				c.Stderr(t, "stderr", r.PlainStderr())
			}
			if c.HomeUnchanged {
				AssertUnchanged(t, r.Home, r.HomeBefore)
			}
		})
	}
}
