package cli

import (
	"fmt"
	"strings"
)

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

// usagef builds a usage error: a non-matching argv, for which the program prints
// the offending argument and the usage block (appspec/02-invocation.md,
// "Argument-parser behavior"). Every error Parse reports is one of these, which
// is what makes the usage block the right response to all of them.
func usagef(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// valued names the long options that take an argument.
var valued = map[string]bool{"--config-file": true}

// longFlags maps each valueless long option to the field it sets. Membership is
// the definition of "is a known valueless option", so no caller has to keep a
// second list in step with this one.
var longFlags = map[string]func(*Options){
	"--help":     func(o *Options) { o.Help = true },
	"--version":  func(o *Options) { o.Version = true },
	"--force":    func(o *Options) { o.Force = true },
	"--force-no": func(o *Options) { o.ForceNo = true },
	"--root":     func(o *Options) { o.Root = true },
	"--dry-run":  func(o *Options) { o.DryRun = true },
	"--verbose":  func(o *Options) { o.Verbose = true },
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

// longToShort is shortToLong inverted, for naming an option in a diagnostic. When
// a long option ever gains a second short alias, the smaller letter wins, so
// spell's output cannot vary between runs with map iteration order.
var longToShort = func() map[string]byte {
	inverted := make(map[string]byte, len(shortToLong))
	for short, long := range shortToLong {
		if existing, ok := inverted[long]; ok && existing < short {
			continue
		}
		inverted[long] = short
	}
	return inverted
}()

// spell names an option the way a diagnostic should: both spellings when it has
// two, so one mistake reads identically however it was typed
// (appspec/02-invocation.md, "Invocation forms") without pointing the user at a
// form they did not use.
func spell(long string) string {
	if short, ok := longToShort[long]; ok {
		return fmt.Sprintf("-%c/%s", short, long)
	}
	return long
}

// Parse turns argv (without the program name) into Options. It reports a usage
// error for any argv matching none of the spec's invocation forms.
//
// Options may appear before or after the subcommand, and a lone "--" ends option
// parsing — POSIX's meaning, so a later argument may begin with a dash; it does
// not change how positionals bind to the grammar.
//
// --help/--version short-circuit past a positional the grammar does not accept,
// because they are specified to print and exit taking "no other action". They do
// not rescue a malformed option: --help=1 and --frobnicate --help match no usage
// line, so they are usage errors like any other.
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
						fail(usagef("%s requires an argument", spell(name)))
						continue
					}
					i++
					value = argv[i]
				}
				if value == "" {
					fail(usagef("%s requires a non-empty argument", spell(name)))
					continue
				}
				opts.ConfigFile = value
			case longFlags[name] != nil:
				// Reject before setting the flag: an accepted --help=1 would
				// otherwise short-circuit the run and discard this error.
				if hasValue {
					fail(usagef("%s does not take an argument", spell(name)))
					continue
				}
				longFlags[name](&opts)
			default:
				fail(usagef("unrecognized option: %s", name))
			}

		case len(arg) > 1 && arg[0] == '-':
			consumed, err := opts.applyShortCluster(arg, argv[i+1:])
			if err != nil {
				fail(err)
				continue
			}
			i += consumed

		default:
			positional = append(positional, arg)
		}
	}

	// An unrecognized or malformed option is reported even alongside
	// --help/--version: it matches no usage line, and silently accepting a
	// typo'd option would be worse than the help text is useful.
	if optErr != nil {
		return opts, optErr
	}
	// Past that, --help and --version print and exit without taking any other
	// action, so they short-circuit the grammar.
	if opts.Help || opts.Version {
		return opts, nil
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

// applyShortCluster parses one short-option token: -f, a stack like -fnv, or a
// valued option as -c path, -cpath, or -c=path. next is the remaining argv, and
// consumed says how many of its entries the token took as its value.
//
// The whole token is validated before anything is set. A token that half-applies
// and then fails is the bug -h=1 was: Help would be set, and Parse's
// help/version short-circuit would discard the error and exit 0, while the long
// form --help=1 correctly reported a usage error.
func (o *Options) applyShortCluster(token string, next []string) (consumed int, err error) {
	var setters []func(*Options)
	var configFile string
	var hasConfigFile bool

	last := ""
	for j := 1; j < len(token); j++ {
		if token[j] == '=' && last != "" {
			// -v=yes: the flag is real, the value is not allowed. Report it the
			// way the long form does rather than as an unknown option "-=".
			return 0, usagef("%s does not take an argument", spell(last))
		}
		// Every diagnostic below names the option through spell, so that short and
		// long spellings of one mistake are observably identical. Only an unknown
		// option letter is reported as typed: it names no known option, so there
		// is no long form to pair it with.
		name, ok := shortToLong[token[j]]
		if !ok {
			return 0, usagef("unrecognized option: -%c", token[j])
		}
		last = name
		if !valued[name] {
			// Every shortToLong value is a key of longFlags or of valued, which
			// TestEveryShortOptionHasALongForm holds to.
			setters = append(setters, longFlags[name])
			continue
		}
		// A valued option takes the rest of the token, or the next argv entry.
		// The "=" spelling is stripped so -c=path means what --config-file=path
		// means: short and long forms are interchangeable
		// (appspec/02-invocation.md, "Invocation forms").
		rest := token[j+1:]
		value := strings.TrimPrefix(rest, "=")
		if rest == "" {
			if len(next) == 0 {
				return 0, usagef("%s requires an argument", spell(name))
			}
			value = next[0]
			consumed = 1
		}
		if value == "" {
			return 0, usagef("%s requires a non-empty argument", spell(name))
		}
		configFile, hasConfigFile = value, true
		break
	}

	for _, set := range setters {
		set(o)
	}
	if hasConfigFile {
		o.ConfigFile = configFile
	}
	return consumed, nil
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
		// An empty key must not read as "no application given", which would
		// silently widen the run from one app to the whole configured set
		// (appspec/02-invocation.md, "Selecting which applications a command
		// acts on").
		if rest[0] == "" {
			return usagef("<application> cannot be empty")
		}
		o.Application = rest[0]
	}
	return nil
}
