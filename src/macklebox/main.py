"""Process entry point: the whole run, from argv to exit code.

The order here is the contract of appspec/02 'Command dispatch order':

1. parse argv -- ``--help`` / ``--version`` print and exit ``0`` right here,
   before any config read; conflicting force flags are rejected here too;
2. load the config and assemble the application database;
3. dispatch to the requested subcommand.
"""

from __future__ import annotations

import sys
from typing import Sequence, TextIO

from .cli import Command, parse_argv
from .errors import MackupError, UsageError
from .pipeline import dispatch, load_startup_context
from .usage import help_text, usage_text
from .version import version_line

#: appspec/07 names this line as contract, matched by scripts and tests.
FORCE_CONFLICT_MESSAGE = "Options --force and --force-no are mutually exclusive."

EXIT_OK = 0
EXIT_ERROR = 1


def run(
    argv: Sequence[str],
    *,
    stdout: TextIO,
    stderr: TextIO,
    stdin: TextIO | None = None,
) -> int:
    """Run one invocation against the given streams and return its exit code."""
    # 1. Parse argv.
    try:
        invocation = parse_argv(argv)
    except UsageError as error:
        # appspec/07 'Output streams': parser usage and warning text on a usage
        # error goes to stderr.
        if error.warning:
            print(error.warning, file=stderr)
        print(usage_text(), file=stderr)
        return EXIT_ERROR

    if invocation.command == Command.HELP:
        print(help_text(), file=stdout)
        return EXIT_OK
    if invocation.command == Command.VERSION:
        print(version_line(), file=stdout)
        return EXIT_OK
    if invocation.command == Command.USAGE:
        # A bare invocation is a usage *display*, not a usage error, so it goes to
        # stdout and exits 0 like --help (appspec/02 'No subcommand').
        print(usage_text(), file=stdout)
        return EXIT_OK

    if invocation.force and invocation.force_no:
        print(FORCE_CONFLICT_MESSAGE, file=stderr)
        return EXIT_ERROR

    try:
        # 2. Config load (resolving the storage location) + application database.
        context = load_startup_context(invocation)
        # 3. Dispatch, behind the per-command environment gate.
        return dispatch(
            invocation, context, stdout=stdout, stderr=stderr, stdin=stdin
        )
    except MackupError as error:
        print(f"Error: {error}", file=stderr)
        return error.exit_code


def main(argv: Sequence[str] | None = None) -> int:
    """Console-script entry point; the return value becomes the exit code."""
    return run(
        sys.argv[1:] if argv is None else argv,
        stdout=sys.stdout,
        stderr=sys.stderr,
        stdin=sys.stdin,
    )
