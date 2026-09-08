package cli

import (
	"errors"

	"github.com/promptctl/macklebox/internal/ui"
	"github.com/promptctl/macklebox/internal/version"
)

// Exit codes, per the appspec/02 "Exit codes" table.
const (
	ExitOK = 0
	// ExitFatal is a fatal, cleanly-handled error: a single diagnostic on
	// stderr and no partial effect.
	ExitFatal = 1
)

// forceConflictMessage is a literal contract token (appspec/07 "Error behavior
// summary"): scripts and tests match this line exactly.
const forceConflictMessage = "Options --force and --force-no are mutually exclusive."

// FatalError is a cleanly-handled failure: one diagnostic on stderr, exit 1,
// no filesystem effect. appspec/01 §6 calls this the guarded regime.
type FatalError struct {
	Message string
}

func (e *FatalError) Error() string { return e.Message }

// Main runs one invocation and returns the process exit code.
//
// The order below is the startup pipeline fixed by appspec/02 "Command
// dispatch order" and appspec/01 §4: parse argv (with --help/--version
// short-circuiting and the force-flag conflict rejected here), then load the
// user config, then dispatch. Only --help and --version bypass the config-load
// gate.
func Main(argv []string, io ui.IO) int {
	// 1. Parse argv.
	inv, err := Parse(argv)
	if err != nil {
		var usageErr *UsageError
		if errors.As(err, &usageErr) {
			return reportUsageError(io, usageErr)
		}
		io.Error("Error: " + err.Error())
		return ExitFatal
	}

	// --help and --version print to stdout and exit 0 before anything else,
	// reading no config.
	if inv.Help {
		io.Print(usageText)
		return ExitOK
	}
	if inv.Version {
		io.Print(version.String())
		return ExitOK
	}

	// The force flags are mutually exclusive, and the conflict is rejected
	// here — before config is loaded and before any action is taken.
	if inv.Force && inv.ForceNo {
		io.Error(forceConflictMessage)
		return ExitFatal
	}

	// A bare invocation is a usage display, not an error: usage on stdout,
	// exit 0 (appspec/02 "Argument-parser behavior", "No subcommand").
	if inv.Command == CmdNone {
		io.Print(usageText)
		return ExitOK
	}

	// 2. Load and validate the user config, resolving the storage location and
	//    assembling the application database. Stubbed until the resolvers land;
	//    every command except --help/--version passes through this gate.
	cfg, err := loadConfig(inv)
	if err != nil {
		return reportFatal(io, err)
	}

	// 3. Dispatch to the requested subcommand.
	if err := dispatch(inv, cfg, io); err != nil {
		return reportFatal(io, err)
	}
	return ExitOK
}

// reportUsageError writes the parser warning and the usage block to stderr.
// appspec/07 "Output streams" routes argument-parser usage and warning text on
// a usage error to stderr.
func reportUsageError(io ui.IO, err *UsageError) int {
	io.Errorf("mackup: %s", err.Warning)
	io.Error(usageText)
	return ExitFatal
}

// reportFatal writes a cleanly-handled failure to stderr and returns its code.
func reportFatal(io ui.IO, err error) int {
	var fatal *FatalError
	if errors.As(err, &fatal) {
		io.Error("Error: " + fatal.Message)
		return ExitFatal
	}
	io.Error("Error: " + err.Error())
	return ExitFatal
}
