package cli

import (
	"errors"
	"testing"
)

// Every usage line in appspec/02 "Invocation forms" must parse, and the
// resulting command/app must be the one the grammar names.
func TestParseAcceptsEveryInvocationForm(t *testing.T) {
	tests := []struct {
		name string
		argv []string
		want Invocation
	}{
		{"list", []string{"list"}, Invocation{Command: CmdList}},
		{"show app", []string{"show", "vim"}, Invocation{Command: CmdShow, App: "vim"}},
		{"backup", []string{"backup"}, Invocation{Command: CmdBackup}},
		{"backup app", []string{"backup", "vim"}, Invocation{Command: CmdBackup, App: "vim"}},
		{"restore", []string{"restore"}, Invocation{Command: CmdRestore}},
		{"restore app", []string{"restore", "git"}, Invocation{Command: CmdRestore, App: "git"}},
		{"link", []string{"link"}, Invocation{Command: CmdLink}},
		{"link app", []string{"link", "vim"}, Invocation{Command: CmdLink, App: "vim"}},
		{"link install", []string{"link", "install"}, Invocation{Command: CmdLinkInstall}},
		{"link install app", []string{"link", "install", "vim"}, Invocation{Command: CmdLinkInstall, App: "vim"}},
		{"link uninstall", []string{"link", "uninstall"}, Invocation{Command: CmdLinkUninstall}},
		{"link uninstall app", []string{"link", "uninstall", "vim"}, Invocation{Command: CmdLinkUninstall, App: "vim"}},
		{"help short", []string{"-h"}, Invocation{Help: true}},
		{"help long", []string{"--help"}, Invocation{Help: true}},
		{"version", []string{"--version"}, Invocation{Version: true}},
		{"bare", nil, Invocation{Command: CmdNone}},
		// An application key that happens to collide with a reserved word is
		// still reachable in the <application> slot.
		{"app named install", []string{"link", "install", "install"}, Invocation{Command: CmdLinkInstall, App: "install"}},
		{"app after terminator", []string{"backup", "--", "-weird-app"}, Invocation{Command: CmdBackup, App: "-weird-app"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.argv)
			if err != nil {
				t.Fatalf("Parse(%q) returned error %v, want success", tt.argv, err)
			}
			if *got != tt.want {
				t.Errorf("Parse(%q) = %+v, want %+v", tt.argv, *got, tt.want)
			}
		})
	}
}

// The appspec/02 "Global options" table: short and long forms are
// interchangeable and produce identical behavior.
func TestParseGlobalOptions(t *testing.T) {
	tests := []struct {
		name  string
		argv  []string
		check func(*Invocation) bool
	}{
		{"force long", []string{"--force", "backup"}, func(i *Invocation) bool { return i.Force }},
		{"force short", []string{"-f", "backup"}, func(i *Invocation) bool { return i.Force }},
		{"force-no", []string{"--force-no", "backup"}, func(i *Invocation) bool { return i.ForceNo }},
		{"root long", []string{"--root", "list"}, func(i *Invocation) bool { return i.Root }},
		{"root short", []string{"-r", "list"}, func(i *Invocation) bool { return i.Root }},
		{"dry-run long", []string{"--dry-run", "backup"}, func(i *Invocation) bool { return i.DryRun }},
		{"dry-run short", []string{"-n", "backup"}, func(i *Invocation) bool { return i.DryRun }},
		{"verbose long", []string{"--verbose", "backup"}, func(i *Invocation) bool { return i.Verbose }},
		{"verbose short", []string{"-v", "backup"}, func(i *Invocation) bool { return i.Verbose }},
		{"clustered shorts", []string{"-fnv", "backup"}, func(i *Invocation) bool { return i.Force && i.DryRun && i.Verbose }},
		{"config long separate", []string{"--config-file", "/h/c.cfg", "list"}, func(i *Invocation) bool { return i.ConfigFile == "/h/c.cfg" }},
		{"config long equals", []string{"--config-file=/h/c.cfg", "list"}, func(i *Invocation) bool { return i.ConfigFile == "/h/c.cfg" }},
		{"config short separate", []string{"-c", "/h/c.cfg", "list"}, func(i *Invocation) bool { return i.ConfigFile == "/h/c.cfg" }},
		{"config short attached", []string{"-c/h/c.cfg", "list"}, func(i *Invocation) bool { return i.ConfigFile == "/h/c.cfg" }},
		{"config at end of cluster", []string{"-nc", "/h/c.cfg", "list"}, func(i *Invocation) bool { return i.DryRun && i.ConfigFile == "/h/c.cfg" }},
		// Options may appear before the subcommand; the reference parser also
		// accepts them after it.
		{"option after subcommand", []string{"backup", "-v", "vim"}, func(i *Invocation) bool {
			return i.Verbose && i.Command == CmdBackup && i.App == "vim"
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.argv)
			if err != nil {
				t.Fatalf("Parse(%q) returned error %v, want success", tt.argv, err)
			}
			if !tt.check(got) {
				t.Errorf("Parse(%q) = %+v, option not applied as expected", tt.argv, *got)
			}
		})
	}
}

// Anything that does not match one of the listed usage lines is a usage error.
func TestParseRejectsNonMatchingForms(t *testing.T) {
	tests := []struct {
		name string
		argv []string
	}{
		{"unrecognized subcommand", []string{"frobnicate"}},
		{"show without application", []string{"show"}},
		{"extra positional after list", []string{"list", "vim"}},
		{"extra positional after show", []string{"show", "vim", "git"}},
		{"extra positional after backup", []string{"backup", "vim", "git"}},
		{"extra positional after link install", []string{"link", "install", "vim", "git"}},
		{"unknown long option", []string{"--nope", "list"}},
		{"unknown short option", []string{"-z", "list"}},
		{"config file without value", []string{"--config-file"}},
		{"short config without value", []string{"list", "-c"}},
		{"value on a valueless option", []string{"--verbose=yes", "list"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.argv)
			if err == nil {
				t.Fatalf("Parse(%q) succeeded, want a usage error", tt.argv)
			}
			var usageErr *UsageError
			if !errors.As(err, &usageErr) {
				t.Fatalf("Parse(%q) returned %T, want *UsageError", tt.argv, err)
			}
			if usageErr.Warning == "" {
				t.Errorf("Parse(%q) usage error has no warning line", tt.argv)
			}
		})
	}
}

// --help and --version short-circuit the grammar: they are terminal, so an
// otherwise-invalid argv still prints help rather than a usage error.
func TestHelpAndVersionShortCircuitTheGrammar(t *testing.T) {
	for _, argv := range [][]string{
		{"--help", "frobnicate"},
		{"--version", "frobnicate"},
		{"show", "--help"},
	} {
		got, err := Parse(argv)
		if err != nil {
			t.Fatalf("Parse(%q) returned error %v, want success", argv, err)
		}
		if !got.Help && !got.Version {
			t.Errorf("Parse(%q) = %+v, want Help or Version set", argv, *got)
		}
	}
}
