// Package version resolves the version string the program reports.
package version

import (
	"runtime/debug"
	"strings"
)

// Fallback is the stable token reported when no package metadata is available
// (e.g. a binary built straight out of an uninstalled tree). appspec/00
// "Provenance" requires the version to be the package's own version when
// installed and a stable fallback token otherwise.
const Fallback = "unknown"

// buildVersion is overridable at link time for release builds:
//
//	go build -ldflags "-X github.com/promptctl/macklebox/internal/version.buildVersion=0.11.1"
var buildVersion string

// Version returns the bare version string, without any product prefix.
func Version() string {
	if v := normalize(buildVersion); v != "" {
		return v
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := normalize(info.Main.Version); v != "" {
			return v
		}
	}
	return Fallback
}

// String returns the line printed by --version: "Mackup <version>".
//
// The product token is "Mackup", not "macklebox": appspec/02 fixes the literal
// observable surface; macklebox is the project/package name only.
func String() string {
	return "Mackup " + Version()
}

// normalize strips Go's "v" prefix and rejects the placeholders the toolchain
// substitutes when a module has no released version attached.
func normalize(v string) string {
	v = strings.TrimSpace(v)
	switch v {
	case "", "(devel)", "devel":
		return ""
	}
	return strings.TrimPrefix(v, "v")
}
