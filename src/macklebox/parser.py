"""The argv boundary: grammar, global options, usage errors.

appspec/02-invocation.md is the contract.  The accepted forms -- and these forms
only -- are the usage block below; anything matching none of them is a usage
error.

Written by hand rather than with `argparse`: the grammar has a two-word
subcommand (`link install`) that shares its head with a one-word subcommand
taking an optional positional (`link [<application>]`), and appspec/02 and
appspec/07 pin the streams and exit codes of the error paths, which `argparse`
decides for itself.
"""

from macklebox.errors import UsageError

USAGE = """\
Usage:
  mackup [options] list
  mackup [options] show <application>
  mackup [options] backup [<application>]
  mackup [options] restore [<application>]
  mackup [options] link install [<application>]
  mackup [options] link uninstall [<application>]
  mackup [options] link [<application>]
  mackup (-h | --help)
  mackup --version
"""

HELP = """\
Keep your application settings in sync across machines, by keeping the one real
copy of each config file in a folder that is already replicated between them.

{usage}
Options:
  -h, --help               Show this help message and exit.
      --version            Show the version and exit.
  -f, --force              Answer every confirmation prompt with Yes.
      --force-no           Answer every confirmation prompt with No.
  -r, --root               Allow running as the superuser.
  -n, --dry-run            Show what would be done; change nothing.
  -v, --verbose            Show full paths and skip traces.
  -c, --config-file PATH   Use PATH as the config file instead of ~/.mackup.cfg.

<application> is an application key as shown by `mackup list` (e.g. vim, git),
not the display name.  With no <application>, a command acts on every
application in the configured scope.
""".format(usage=USAGE)


# long name -> (short name or None, takes an argument, destination)
_OPTIONS = {
    "--help": ("-h", False, "help"),
    "--version": (None, False, "version"),
    "--force": ("-f", False, "force"),
    "--force-no": (None, False, "force_no"),
    "--root": ("-r", False, "root"),
    "--dry-run": ("-n", False, "dry_run"),
    "--verbose": ("-v", False, "verbose"),
    "--config-file": ("-c", True, "config_file"),
}

_SHORT_TO_LONG = {
    short: long for long, (short, _, _) in _OPTIONS.items() if short is not None
}

# Actions the program can be asked to perform.  `usage` is the bare invocation;
# --help and --version are options, not subcommands, and are read off Options.
USAGE_ACTION = "usage"
LIST = "list"
SHOW = "show"
BACKUP = "backup"
RESTORE = "restore"
LINK = "link"
LINK_INSTALL = "link install"
LINK_UNINSTALL = "link uninstall"

# Subcommands that take an optional <application>.
_OPTIONAL_APP = (BACKUP, RESTORE, LINK_INSTALL, LINK_UNINSTALL, LINK)


class Options(object):
    """The global-options table of appspec/02, as decided data."""

    def __init__(self, **kwargs):
        self.help = kwargs.get("help", False)
        self.version = kwargs.get("version", False)
        self.force = kwargs.get("force", False)
        self.force_no = kwargs.get("force_no", False)
        self.root = kwargs.get("root", False)
        self.dry_run = kwargs.get("dry_run", False)
        self.verbose = kwargs.get("verbose", False)
        self.config_file = kwargs.get("config_file", None)

    def __repr__(self):  # pragma: no cover - debugging aid
        return "Options({0!r})".format(self.__dict__)


class Invocation(object):
    """One parsed command line: what to do, to which application, how."""

    def __init__(self, action, application=None, options=None):
        self.action = action
        self.application = application
        self.options = options if options is not None else Options()

    def __repr__(self):  # pragma: no cover - debugging aid
        return "Invocation({0!r}, {1!r}, {2!r})".format(
            self.action, self.application, self.options
        )


def parse(argv):
    """Parse the argument list (without the program name) into an Invocation.

    Raises UsageError for argv matching none of the listed usage lines.
    """
    flags, positionals = _split(list(argv))
    action, application = _interpret(positionals)
    return Invocation(action, application, Options(**flags))


def _split(tokens):
    """Separate option flags from positional arguments.

    appspec/02: "Options may appear before the subcommand.  Short and long forms
    are interchangeable and produce identical behavior."  Options after the
    subcommand are accepted too -- the grammar's `[options]` is positional
    sugar, and rejecting `mackup backup -v` would surprise every user.
    """
    flags = {}
    positionals = []
    index = 0
    while index < len(tokens):
        token = tokens[index]
        if token == "--":
            positionals.extend(tokens[index + 1 :])
            break
        if token.startswith("--"):
            index = _read_long(token, tokens, index, flags)
        elif token.startswith("-") and token != "-":
            index = _read_shorts(token, tokens, index, flags)
        else:
            positionals.append(token)
            index += 1
    return flags, positionals


def _read_long(token, tokens, index, flags):
    name, separator, inline_value = token.partition("=")
    if name not in _OPTIONS:
        raise UsageError("unrecognized option: {0}".format(name))
    _, takes_argument, dest = _OPTIONS[name]
    if not takes_argument:
        if separator:
            raise UsageError("option {0} takes no argument".format(name))
        flags[dest] = True
        return index + 1
    if separator:
        if not inline_value:
            raise UsageError("option {0} requires an argument".format(name))
        flags[dest] = inline_value
        return index + 1
    if index + 1 >= len(tokens):
        raise UsageError("option {0} requires an argument".format(name))
    flags[dest] = tokens[index + 1]
    return index + 2


def _read_shorts(token, tokens, index, flags):
    """Read a short-option cluster such as `-vn`, `-c PATH`, or `-vcPATH`."""
    cluster = token[1:]
    position = 0
    while position < len(cluster):
        short = "-" + cluster[position]
        if short not in _SHORT_TO_LONG:
            raise UsageError("unrecognized option: {0}".format(short))
        name = _SHORT_TO_LONG[short]
        _, takes_argument, dest = _OPTIONS[name]
        if not takes_argument:
            flags[dest] = True
            position += 1
            continue
        remainder = cluster[position + 1 :]
        if remainder:
            flags[dest] = remainder
            return index + 1
        if index + 1 >= len(tokens):
            raise UsageError("option {0} requires an argument".format(short))
        flags[dest] = tokens[index + 1]
        return index + 2
    return index + 1


def _interpret(positionals):
    """Match the positional arguments against the usage lines."""
    if not positionals:
        return USAGE_ACTION, None

    head = positionals[0]
    if head == LINK:
        action, rest = _interpret_link(positionals[1:])
    elif head in (LIST, SHOW, BACKUP, RESTORE):
        action, rest = head, positionals[1:]
    else:
        raise UsageError("unrecognized argument: {0}".format(head))

    return action, _take_application(action, rest)


def _interpret_link(rest):
    """`link install`, `link uninstall`, and bare `link` share a head token.

    A literal `install` / `uninstall` after `link` is the subcommand, never an
    application key -- that is what the grammar of appspec/02 says, and it is
    why no application may be keyed `install` or `uninstall`.
    """
    if rest and rest[0] == "install":
        return LINK_INSTALL, rest[1:]
    if rest and rest[0] == "uninstall":
        return LINK_UNINSTALL, rest[1:]
    return LINK, rest


def _take_application(action, rest):
    if action == SHOW:
        if not rest:
            raise UsageError("the subcommand 'show' requires an <application>")
        application = rest[0]
        rest = rest[1:]
    elif action in _OPTIONAL_APP and rest:
        application = rest[0]
        rest = rest[1:]
    else:
        application = None

    if rest:
        raise UsageError("unrecognized argument: {0}".format(rest[0]))
    return application
