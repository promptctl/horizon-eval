// Package cli implements mackup's command-line boundary: the invocation
// grammar, the global options table, the dispatch order, and the exit codes
// specified in appspec/02-invocation.md.
package cli

import (
	"fmt"
	"io"

	"github.com/promptctl/macklebox/internal/version"
)

// Exit codes. appspec/02-invocation.md, "Exit codes": 0 is "the requested
// action completed"; 1 is a fatal, cleanly-handled error.
const (
	ExitOK    = 0
	ExitFatal = 1
)

// ForceConflictMessage is a literal contract token: supplying both --force and
// --force-no writes exactly this line to stderr and exits 1
// (appspec/07-output-safety-lifecycle.md, "Error behavior summary").
const ForceConflictMessage = "Options --force and --force-no are mutually exclusive."

// Streams are the process streams the program reads and writes, injected so the
// boundary is observable from a test without a subprocess.
type Streams struct {
	Out io.Writer
	Err io.Writer
	In  io.Reader
}

// config is the resolved startup state: the storage location, the synced
// sub-directory, and the application set in scope. Resolving it is the second
// step of the dispatch order and belongs to the resolvers epic; until that
// lands, loading is a no-op that records nothing.
type config struct{}

// loadConfig is step 2 of the startup pipeline: load and validate the user
// config, resolving the storage location, for every command except --help and
// --version (appspec/02-invocation.md, "Command dispatch order").
//
// Stub: config discovery, validation, and storage-engine resolution land with
// the resolvers epic. The call site is real so the order — and the fact that
// --help/--version and the force-flag conflict precede it — is already
// observable.
func loadConfig(Options) (config, error) { return config{}, nil }

// assembleDatabase is step 3: build the key -> (display name, home-relative file
// set) lookup table from the layered definition directories.
//
// Stub: lands with the resolvers epic. It is named here because appspec/01 §4
// counts it among the steps *every* command passes identically.
func assembleDatabase(config) error { return nil }

// checkEnvironment is step 4, the universal environment gate: the root guard
// (effective UID 0 without --root aborts any command) and the storage-root
// directory's existence.
//
// Stub: lands with the resolvers epic, which is when a storage root exists to
// check. It belongs here and nowhere else — appspec/01 §4 calls this "one gate
// every command passes identically" and "the sole place root is refused", so a
// command that repeats the root check has misread the architecture. Options.Root
// is parsed today and consumed here when this lands.
func checkEnvironment(Options, config) error { return nil }

// Run executes one invocation and returns the process exit code.
func Run(argv []string, s Streams) int {
	return run(argv, s, loadConfig)
}

// run is Run with the config loader injected, so a test can observe whether a
// given argv reached the config-load step.
//
// The steps below are appspec/01 §4's one startup pipeline, which every command
// flows through identically; only the command-specific storage gate (step 5,
// which lands with the sync commands) varies.
func run(argv []string, s Streams, load func(Options) (config, error)) int {
	// Step 1: parse argv. --help/--version print and exit here; conflicting
	// force flags are rejected here, before any config is read.
	opts, err := Parse(argv)
	if err != nil {
		// Every Parse error is a non-matching argv, so the response is always the
		// offending argument followed by the usage block, on stderr
		// (appspec/02-invocation.md, "Argument-parser behavior";
		// appspec/07-output-safety-lifecycle.md, "Output streams").
		fmt.Fprintf(s.Err, "mackup: %s\n", err)
		fmt.Fprint(s.Err, Help)
		return ExitFatal
	}

	switch {
	case opts.Help:
		fmt.Fprint(s.Out, Help)
		return ExitOK
	case opts.Version:
		fmt.Fprintf(s.Out, "Mackup %s\n", version.String())
		return ExitOK
	}

	if opts.Force && opts.ForceNo {
		fmt.Fprintln(s.Err, ForceConflictMessage)
		return ExitFatal
	}

	// A bare invocation is a usage display, not a usage error: print the usage
	// block and exit 0 (appspec/02-invocation.md, "Argument-parser behavior").
	if opts.Command == CmdNone {
		fmt.Fprint(s.Out, Help)
		return ExitOK
	}

	// Steps 2-4 abort every command uniformly on failure, including list and
	// show. Their diagnostics print verbatim: appspec/07-output-safety-lifecycle.md
	// specifies some as a single "Error: ... Aborting." line and others as
	// multi-line messages carrying no such prefix, so the user-facing shape is
	// the error value's to own, not this call site's to impose.

	// Step 2: load config, resolving the storage location.
	cfg, err := load(opts)
	if err != nil {
		fmt.Fprintln(s.Err, err)
		return ExitFatal
	}

	// Step 3: assemble the application database.
	if err := assembleDatabase(cfg); err != nil {
		fmt.Fprintln(s.Err, err)
		return ExitFatal
	}

	// Step 4: the universal environment gate — root guard, storage root exists.
	if err := checkEnvironment(opts, cfg); err != nil {
		fmt.Fprintln(s.Err, err)
		return ExitFatal
	}

	// Dispatch. The command-specific storage gate (appspec/01 §4 step 5) belongs
	// to each command, and lands with them.
	return dispatch(opts, cfg, s)
}

// dispatch routes a parsed invocation to its command.
//
// The commands themselves land in the later epics; each is wired here as it
// arrives. Until then a recognized command reports that it is not implemented
// on stderr and exits non-zero, so no run can report success for work that did
// not happen.
func dispatch(opts Options, _ config, s Streams) int {
	fmt.Fprintf(s.Err, "Error: '%s' is not implemented yet.\n", opts.Command)
	return ExitFatal
}
