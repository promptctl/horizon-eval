// Command mackup keeps application settings in sync by keeping the one real
// copy of each config file in a folder that already syncs between machines.
//
// The observable surface — command grammar, contract tokens, config filenames,
// the version line — is mackup's, as specified in appspec/. macklebox is the
// project and module name only.
package main

import (
	"os"

	"github.com/promptctl/macklebox/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], cli.Streams{
		Out: os.Stdout,
		Err: os.Stderr,
		In:  os.Stdin,
	}))
}
