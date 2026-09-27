package cli

import "fmt"

// Command identifies one of the invocation forms in appspec/02-invocation.md.
type Command int

// The commands. CmdNone is a bare invocation, which the spec treats as
// "show usage" rather than as an error.
const (
	CmdNone Command = iota
	CmdList
	CmdShow
	CmdBackup
	CmdRestore
	CmdLink
	CmdLinkInstall
	CmdLinkUninstall
)

// String returns the command as it is spelled on the command line.
func (c Command) String() string {
	switch c {
	case CmdList:
		return "list"
	case CmdShow:
		return "show"
	case CmdBackup:
		return "backup"
	case CmdRestore:
		return "restore"
	case CmdLink:
		return "link"
	case CmdLinkInstall:
		return "link install"
	case CmdLinkUninstall:
		return "link uninstall"
	}
	return ""
}

// Options is the decided content of one argv: the global options table from
// appspec/02-invocation.md plus the selected command and its optional
// application key.
type Options struct {
	Help       bool
	Version    bool
	Force      bool
	ForceNo    bool
	Root       bool
	DryRun     bool
	Verbose    bool
	ConfigFile string

	Command Command
	// Application is the application key given on the command line, or "" when
	// none was given. Only show requires it.
	Application string
}

// usageError is a non-matching argv: the program prints the offending argument
// and the usage block (appspec/02-invocation.md, "Argument-parser behavior").
type usageError struct{ msg string }

func (e *usageError) Error() string { return e.msg }

func usagef(format string, args ...any) error {
	return &usageError{msg: fmt.Sprintf(format, args...)}
}

// valued names the long options that take an argument.
var valued = map[string]bool{"--config-file": true}

// longFlags maps each valueless long option to its field.
func (o *Options) setLong(name string) bool {
	switch name {
	case "--help":
		o.Help = true
	case "--version":
		o.Version = true
	case "--force":
		o.Force = true
	case "--force-no":
		o.ForceNo = true
	case "--root":
		o.Root = true
	case "--dry-run":
		o.DryRun = true
	case "--verbose":
		o.Verbose = true
	default:
		return false
	}
	return true
}

// shortToLong maps each short option to its long form. Short and long forms are
// interchangeable and produce identical behavior (appspec/02-invocation.md).
var shortToLong = map[byte]string{
	'h': "--help",
	'f': "--force",
	'r': "--root",
	'n': "--dry-run",
	'v': "--verbose",
	'c': "--config-file",
}

// Parse turns argv (without the program name) into Options. It reports a usage
// error for any argv matching none of the spec's invocation forms.
//
// Options may appear before or after the subcommand, a lone "--" ends option
// parsing, and --help/--version are honored even alongside an otherwise
// non-matching argv, because they are specified to print and exit taking "no
// other action".
func Parse(argv []string) (Options, error) {
	var opts Options
	var positional []string
	var optErr error

	fail := func(err error) {
		if optErr == nil {
			optErr = err
		}
	}

	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		switch {
		case arg == "--":
			positional = append(positional, argv[i+1:]...)
			i = len(argv)
		case len(arg) > 2 && arg[:2] == "--":
			name, value, hasValue := splitLongOption(arg)
			switch {
			case valued[name]:
				if !hasValue {
					if i+1 >= len(argv) {
						fail(usagef("%s requires an argument", name))
						continue
					}
					i++
					value = argv[i]
				}
				opts.ConfigFile = value
			case opts.setLong(name):
				if hasValue {
					fail(usagef("%s does not take an argument", name))
				}
			default:
				fail(usagef("unrecognized option: %s", name))
			}
		case len(arg) > 1 && arg[0] == '-':
			// A short-option cluster: valueless flags may be stacked, and the
			// last one may take the rest of the token or the next argv entry as
			// its value (-c path, -cpath).
			for j := 1; j < len(arg); j++ {
				name, ok := shortToLong[arg[j]]
				if !ok {
					fail(usagef("unrecognized option: -%c", arg[j]))
					break
				}
				if !valued[name] {
					opts.setLong(name)
					continue
				}
				if rest := arg[j+1:]; rest != "" {
					opts.ConfigFile = rest
				} else if i+1 < len(argv) {
					i++
					opts.ConfigFile = argv[i]
				} else {
					fail(usagef("-%c requires an argument", arg[j]))
				}
				break
			}
		default:
			positional = append(positional, arg)
		}
	}

	// --help and --version print and exit without taking any other action, so
	// they outrank both option and grammar errors.
	if opts.Help || opts.Version {
		return opts, nil
	}
	if optErr != nil {
		return opts, optErr
	}
	if err := opts.resolveCommand(positional); err != nil {
		return opts, err
	}
	return opts, nil
}

// splitLongOption splits "--name=value" into its parts.
func splitLongOption(arg string) (name, value string, hasValue bool) {
	for i := 2; i < len(arg); i++ {
		if arg[i] == '=' {
			return arg[:i], arg[i+1:], true
		}
	}
	return arg, "", false
}

// resolveCommand binds the positional arguments to one invocation form.
func (o *Options) resolveCommand(positional []string) error {
	if len(positional) == 0 {
		o.Command = CmdNone
		return nil
	}

	rest := positional[1:]
	switch positional[0] {
	case "list":
		o.Command = CmdList
		return o.takeApplication(rest, 0)
	case "show":
		o.Command = CmdShow
		if len(rest) == 0 {
			return usagef("show requires an <application>")
		}
		return o.takeApplication(rest, 1)
	case "backup":
		o.Command = CmdBackup
		return o.takeApplication(rest, 1)
	case "restore":
		o.Command = CmdRestore
		return o.takeApplication(rest, 1)
	case "link":
		o.Command = CmdLink
		if len(rest) > 0 {
			switch rest[0] {
			case "install":
				o.Command = CmdLinkInstall
				rest = rest[1:]
			case "uninstall":
				o.Command = CmdLinkUninstall
				rest = rest[1:]
			}
		}
		return o.takeApplication(rest, 1)
	default:
		return usagef("unrecognized command: %s", positional[0])
	}
}

// takeApplication accepts up to max trailing application keys, rejecting any
// argument beyond the grammar's allowance.
func (o *Options) takeApplication(rest []string, max int) error {
	if len(rest) > max {
		return usagef("unrecognized argument: %s", rest[max])
	}
	if len(rest) == 1 && max == 1 {
		o.Application = rest[0]
	}
	return nil
}
