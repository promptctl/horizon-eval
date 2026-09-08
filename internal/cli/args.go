// Package cli implements the command-line boundary specified in
// appspec/02-invocation.md: the invocation grammar, the global-option table,
// the dispatch order, and the exit codes.
package cli

import (
	"fmt"
	"strings"
)

// Command is one of the subcommand shapes in appspec/02 "Invocation forms".
type Command string

const (
	CmdNone          Command = ""
	CmdList          Command = "list"
	CmdShow          Command = "show"
	CmdBackup        Command = "backup"
	CmdRestore       Command = "restore"
	CmdLink          Command = "link"
	CmdLinkInstall   Command = "link install"
	CmdLinkUninstall Command = "link uninstall"
)

// Invocation is the parsed argv: which command was asked for, which
// application it was scoped to, and the global options in effect.
type Invocation struct {
	Command Command
	// App is the application key the command was scoped to, or "" for the
	// configured set. appspec/02 "Selecting which applications a command acts on".
	App string

	Help    bool
	Version bool

	Force      bool
	ForceNo    bool
	Root       bool
	DryRun     bool
	Verbose    bool
	ConfigFile string
}

// UsageError is a form that matches none of the usage lines. appspec/02
// "Argument-parser behavior": a warning line identifying the problem, then the
// usage block, both on stderr.
type UsageError struct {
	Warning string
}

func (e *UsageError) Error() string { return e.Warning }

// option describes one entry of the appspec/02 "Global options" table.
type option struct {
	long     string
	short    string // without the leading dash; "" when there is no short form
	takesArg bool
	apply    func(inv *Invocation, value string)
}

var options = []option{
	{long: "help", short: "h", apply: func(i *Invocation, _ string) { i.Help = true }},
	{long: "version", apply: func(i *Invocation, _ string) { i.Version = true }},
	{long: "force", short: "f", apply: func(i *Invocation, _ string) { i.Force = true }},
	{long: "force-no", apply: func(i *Invocation, _ string) { i.ForceNo = true }},
	{long: "root", short: "r", apply: func(i *Invocation, _ string) { i.Root = true }},
	{long: "dry-run", short: "n", apply: func(i *Invocation, _ string) { i.DryRun = true }},
	{long: "verbose", short: "v", apply: func(i *Invocation, _ string) { i.Verbose = true }},
	{long: "config-file", short: "c", takesArg: true, apply: func(i *Invocation, v string) { i.ConfigFile = v }},
}

func lookupLong(name string) *option {
	for i := range options {
		if options[i].long == name {
			return &options[i]
		}
	}
	return nil
}

func lookupShort(c byte) *option {
	for i := range options {
		if options[i].short != "" && options[i].short[0] == c {
			return &options[i]
		}
	}
	return nil
}

// Parse turns argv (without the program name) into an Invocation.
//
// Options may appear before the subcommand (appspec/02 "Invocation forms");
// they are also accepted after and between positionals, which is what the
// reference parser does and is a superset of the documented grammar. "--"
// terminates option parsing so an application key may look like a flag.
func Parse(argv []string) (*Invocation, error) {
	inv := &Invocation{}
	var positional []string
	endOfOptions := false

	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		switch {
		case endOfOptions:
			positional = append(positional, arg)

		case arg == "--":
			endOfOptions = true

		case strings.HasPrefix(arg, "--"):
			name, value, hasValue := strings.Cut(arg[2:], "=")
			opt := lookupLong(name)
			if opt == nil {
				return nil, &UsageError{Warning: fmt.Sprintf("unrecognized option: --%s", name)}
			}
			switch {
			case opt.takesArg && hasValue:
				// --config-file=<path>
			case opt.takesArg:
				if i+1 >= len(argv) {
					return nil, &UsageError{Warning: fmt.Sprintf("option --%s requires an argument", name)}
				}
				i++
				value = argv[i]
			case hasValue:
				return nil, &UsageError{Warning: fmt.Sprintf("option --%s does not take an argument", name)}
			}
			opt.apply(inv, value)

		case len(arg) > 1 && arg[0] == '-':
			// A short-option cluster: -v, -fn, -c <path>, -c<path>, -nc <path>.
			for j := 1; j < len(arg); j++ {
				opt := lookupShort(arg[j])
				if opt == nil {
					return nil, &UsageError{Warning: fmt.Sprintf("unrecognized option: -%c", arg[j])}
				}
				if !opt.takesArg {
					opt.apply(inv, "")
					continue
				}
				value := arg[j+1:]
				if value == "" {
					if i+1 >= len(argv) {
						return nil, &UsageError{Warning: fmt.Sprintf("option -%c requires an argument", arg[j])}
					}
					i++
					value = argv[i]
				}
				opt.apply(inv, strings.TrimPrefix(value, "="))
				break // the remainder of the cluster was consumed as the value
			}

		default:
			positional = append(positional, arg)
		}
	}

	// --help and --version are terminal: they short-circuit before the grammar
	// is enforced, so `mackup --help frobnicate` still prints help.
	if inv.Help || inv.Version {
		return inv, nil
	}

	if err := applyPositionals(inv, positional); err != nil {
		return nil, err
	}
	return inv, nil
}

// applyPositionals matches the positional arguments against the usage lines in
// appspec/02 "Invocation forms". Anything matching none of them is a usage error.
func applyPositionals(inv *Invocation, pos []string) error {
	if len(pos) == 0 {
		inv.Command = CmdNone
		return nil
	}

	switch pos[0] {
	case "list":
		inv.Command = CmdList
		return rejectExtra(pos[1:])

	case "show":
		inv.Command = CmdShow
		if len(pos) < 2 {
			return &UsageError{Warning: "the following argument is required: <application>"}
		}
		inv.App = pos[1]
		return rejectExtra(pos[2:])

	case "backup", "restore":
		inv.Command = Command(pos[0])
		rest := pos[1:]
		if len(rest) > 0 {
			inv.App = rest[0]
			rest = rest[1:]
		}
		return rejectExtra(rest)

	case "link":
		// `link install` / `link uninstall` are matched before `link <application>`,
		// so "install" and "uninstall" are reserved words in this position.
		rest := pos[1:]
		if len(rest) > 0 && (rest[0] == "install" || rest[0] == "uninstall") {
			inv.Command = Command("link " + rest[0])
			rest = rest[1:]
		} else {
			inv.Command = CmdLink
		}
		if len(rest) > 0 {
			inv.App = rest[0]
			rest = rest[1:]
		}
		return rejectExtra(rest)

	default:
		return &UsageError{Warning: fmt.Sprintf("unrecognized arguments: %s", strings.Join(pos, " "))}
	}
}

func rejectExtra(extra []string) error {
	if len(extra) == 0 {
		return nil
	}
	return &UsageError{Warning: fmt.Sprintf("unrecognized arguments: %s", strings.Join(extra, " "))}
}
