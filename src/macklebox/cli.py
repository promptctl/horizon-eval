"""Step 1 of the startup pipeline: argv -> Invocation (appspec/02-invocation.md).

Parsing only.  Nothing here reads the filesystem, the environment or the config,
because ``--help`` and ``--version`` must short-circuit before any of that
happens (appspec/01 §4, appspec/02 'Command dispatch order').
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import Sequence

from .errors import UsageError


class Command:
    """Every command the grammar can yield, including the two non-commands."""

    #: ``-h`` / ``--help``: usage to stdout, exit 0, no config read.
    HELP = "help"
    #: ``--version``: the version line to stdout, exit 0, no config read.
    VERSION = "version"
    #: No subcommand at all: a usage *display*, not a usage error (appspec/02).
    USAGE = "usage"

    LIST = "list"
    SHOW = "show"
    BACKUP = "backup"
    RESTORE = "restore"
    LINK = "link"
    LINK_INSTALL = "link install"
    LINK_UNINSTALL = "link uninstall"


#: Application-argument arity of a subcommand form.
_NONE, _OPTIONAL, _REQUIRED = "none", "optional", "required"

# The invocation forms of appspec/02, in the order they are listed there.  Order
# is load-bearing: `link install` is listed above `link [<application>]`, so
# `mackup link install` is the install form and never `link` with an application
# named "install".
_FORMS: tuple[tuple[tuple[str, ...], str, str], ...] = (
    (("list",), Command.LIST, _NONE),
    (("show",), Command.SHOW, _REQUIRED),
    (("backup",), Command.BACKUP, _OPTIONAL),
    (("restore",), Command.RESTORE, _OPTIONAL),
    (("link", "install"), Command.LINK_INSTALL, _OPTIONAL),
    (("link", "uninstall"), Command.LINK_UNINSTALL, _OPTIONAL),
    (("link",), Command.LINK, _OPTIONAL),
)

# The global-options table of appspec/02.  Short and long forms are
# interchangeable and set the same field.
_FLAGS = {
    "-f": "force",
    "--force": "force",
    "--force-no": "force_no",
    "-r": "root",
    "--root": "root",
    "-n": "dry_run",
    "--dry-run": "dry_run",
    "-v": "verbose",
    "--verbose": "verbose",
}
_VALUE_OPTIONS = {"-c": "config_file", "--config-file": "config_file"}
_SHORT_CIRCUIT_OPTIONS = {
    "-h": Command.HELP,
    "--help": Command.HELP,
    "--version": Command.VERSION,
}


@dataclass(frozen=True)
class Invocation:
    """The decided command line: one command, at most one application, the modes."""

    command: str
    application: str | None = None
    force: bool = False
    force_no: bool = False
    root: bool = False
    dry_run: bool = False
    verbose: bool = False
    config_file: str | None = None


def parse_argv(argv: Sequence[str]) -> Invocation:
    """Parse ``argv`` (without the program name) or raise ``UsageError``.

    ``--help`` / ``--version`` return as soon as they are seen, so they hold even
    alongside argv the grammar would otherwise reject.
    """
    modes: dict[str, bool] = {
        "force": False,
        "force_no": False,
        "root": False,
        "dry_run": False,
        "verbose": False,
    }
    config_file: str | None = None
    positionals: list[str] = []

    tokens = list(argv)
    index = 0
    options_ended = False
    while index < len(tokens):
        token = tokens[index]
        index += 1

        if options_ended or not _is_option(token):
            positionals.append(token)
            continue
        if token == "--":
            options_ended = True
            continue

        if token.startswith("--"):
            name, separator, inline = token.partition("=")
            has_inline = bool(separator)
            if name in _SHORT_CIRCUIT_OPTIONS:
                return Invocation(command=_SHORT_CIRCUIT_OPTIONS[name])
            if name in _FLAGS:
                if has_inline:
                    raise UsageError(f"Option {name} takes no argument.")
                modes[_FLAGS[name]] = True
                continue
            if name in _VALUE_OPTIONS:
                if has_inline:
                    config_file = inline
                    continue
                if index >= len(tokens):
                    raise UsageError(f"Option {name} requires an argument.")
                config_file = tokens[index]
                index += 1
                continue
            raise UsageError(f"Unrecognized option: {name}")

        # A cluster of short options; only the last may take an argument.
        cluster = token[1:]
        position = 0
        while position < len(cluster):
            short = "-" + cluster[position]
            position += 1
            if short in _SHORT_CIRCUIT_OPTIONS:
                return Invocation(command=_SHORT_CIRCUIT_OPTIONS[short])
            if short in _FLAGS:
                modes[_FLAGS[short]] = True
                continue
            if short in _VALUE_OPTIONS:
                rest = cluster[position:]
                position = len(cluster)
                if rest:
                    config_file = rest
                    continue
                if index >= len(tokens):
                    raise UsageError(f"Option {short} requires an argument.")
                config_file = tokens[index]
                index += 1
                continue
            raise UsageError(f"Unrecognized option: {short}")

    command, application = _match_form(positionals)
    return Invocation(
        command=command,
        application=application,
        config_file=config_file,
        **modes,
    )


def _is_option(token: str) -> bool:
    """A lone ``-`` is a positional; anything else starting with ``-`` is an option."""
    return token.startswith("-") and token != "-"


def _match_form(positionals: list[str]) -> tuple[str, str | None]:
    if not positionals:
        return Command.USAGE, None

    for words, command, arity in _FORMS:
        if tuple(positionals[: len(words)]) != words:
            continue
        rest = positionals[len(words) :]
        if arity == _NONE:
            if rest:
                raise UsageError(_unmatched(rest[0]))
            return command, None
        if len(rest) > 1:
            raise UsageError(_unmatched(rest[1]))
        if not rest:
            if arity == _REQUIRED:
                # `show` without <application>: the shape is wrong, and no single
                # token is to blame, so the usage block stands on its own.
                raise UsageError()
            return command, None
        return command, rest[0]

    raise UsageError(_unmatched(positionals[0]))


def _unmatched(argument: str) -> str:
    return f"Warning: found unmatched argument: {argument}"
