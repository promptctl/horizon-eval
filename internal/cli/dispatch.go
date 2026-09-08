package cli

import (
	"fmt"

	"github.com/promptctl/macklebox/internal/ui"
)

// Config is the decided-facts record the configuration and application-database
// resolvers produce (appspec/01 §1). It is a placeholder until those resolvers
// land; what matters at this altitude is that the load step sits between argv
// parsing and dispatch, and that every command except --help/--version passes
// through it.
type Config struct{}

// loadConfig is step 2 of the startup pipeline: load and validate the user
// config (resolving the storage engine's folder) and assemble the application
// database. A fatal error here terminates the run for any command, including
// list and show.
//
// Stub: the resolvers epic replaces this body. It must keep returning a
// *FatalError for guarded config failures so the exit-code table holds.
func loadConfig(inv *Invocation) (*Config, error) {
	_ = inv
	return &Config{}, nil
}

// dispatch routes to the requested subcommand. Every entry is a leaf on one
// tree (appspec/01 §1), not an independent program.
func dispatch(inv *Invocation, cfg *Config, io ui.IO) error {
	switch inv.Command {
	case CmdList, CmdShow, CmdBackup, CmdRestore, CmdLink, CmdLinkInstall, CmdLinkUninstall:
		return notImplemented(inv.Command)
	default:
		// Unreachable: Parse only produces the commands above.
		return &FatalError{Message: fmt.Sprintf("Unknown command: %s", inv.Command)}
	}
}

func notImplemented(cmd Command) error {
	return &FatalError{Message: fmt.Sprintf("The '%s' command is not implemented yet.", cmd)}
}
