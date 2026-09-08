// Command mackup keeps application settings in sync by keeping the one real
// copy of each config file in an already-replicated folder.
//
// The observable surface — the command name, the grammar, the literal contract
// tokens — is the one specified in appspec/; "macklebox" is the project and
// package name only.
package main

import (
	"os"

	"github.com/promptctl/macklebox/internal/cli"
	"github.com/promptctl/macklebox/internal/ui"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], ui.Std()))
}
