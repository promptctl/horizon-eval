"""The usage and help text.

appspec/02 'Argument-parser behavior' states the exact wording of usage, help and
parser-warning text is human-facing and deliberately not part of the contract;
the grammar it describes is.  The usage block below therefore reproduces the
invocation forms and the global-options table of appspec/02 verbatim in content,
and is free in its phrasing.
"""

from __future__ import annotations

USAGE = """\
Usage:
  mackup [options] list
  mackup [options] show <application>
  mackup [options] backup [<application>]
  mackup [options] restore [<application>]
  mackup [options] link install [<application>]
  mackup [options] link uninstall [<application>]
  mackup [options] link [<application>]
  mackup -h | --help
  mackup --version"""

_OPTIONS = """\
Options:
  -h, --help                 Print this help and exit.
  --version                  Print the version and exit.
  -f, --force                Answer every confirmation with Yes.
  --force-no                 Answer every confirmation with No.
  -r, --root                 Allow running as the superuser.
  -n, --dry-run              Print what would be done; change nothing.
  -v, --verbose              Print fuller progress detail.
  -c, --config-file <path>   Read the config from <path> instead of discovering
                             it. Must be inside your home directory."""

_DESCRIPTION = """\
Keep your application settings in sync by keeping the one real copy of each
config file in a folder that already syncs between your machines."""


def usage_text() -> str:
    """The usage block, shown for a bare invocation and for usage errors."""
    return USAGE


def help_text() -> str:
    """The full ``--help`` text: description, usage block, option table."""
    return f"{_DESCRIPTION}\n\n{USAGE}\n\n{_OPTIONS}"
