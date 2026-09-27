// Package version resolves the version string the program reports.
package version

import (
	"runtime/debug"
	"strings"
)

// Fallback is the version string used when no package version is available —
// the spec's "uninstalled tree" case (appspec/00-overview.md, Provenance).
const Fallback = "unknown"

// override is set at link time (-ldflags "-X .../internal/version.override=X")
// for builds that carry a stamped version.
var override string

// String returns the program's version: its own package version when the
// binary was installed as a module, and the stable Fallback token otherwise.
func String() string {
	var build string
	if info, ok := debug.ReadBuildInfo(); ok {
		build = info.Main.Version
	}
	return resolve(override, build)
}

// resolve picks the reported version from a link-time stamp and the module
// version recorded in the build, so the choice is testable without a build.
func resolve(stamp, build string) string {
	// Neither source is v-prefixed in the reported version
	// (appspec/00-overview.md: "Mackup 0.11.1", never "Mackup v0.11.1"). Go
	// module versions always carry the prefix, and the natural way to stamp a
	// release — -X ...override=$(git describe --tags) — carries it too.
	if stamp != "" {
		return strings.TrimPrefix(stamp, "v")
	}
	switch build {
	case "", "(devel)", "devel":
		// Built from a tree, not installed as a module version.
		return Fallback
	}
	return strings.TrimPrefix(build, "v")
}
