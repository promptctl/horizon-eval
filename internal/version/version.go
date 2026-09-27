// Package version resolves the version string the program reports.
package version

import "runtime/debug"

// Fallback is the version string used when no package version is available —
// the spec's "uninstalled tree" case (appspec/00-overview.md, Provenance).
const Fallback = "unknown"

// override is set at link time (-ldflags "-X .../internal/version.override=X")
// for builds that carry a stamped version.
var override string

// String returns the program's version: its own package version when the
// binary was installed as a module, and the stable Fallback token otherwise.
func String() string {
	if override != "" {
		return override
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" && v != "devel" {
			return v
		}
	}
	return Fallback
}
