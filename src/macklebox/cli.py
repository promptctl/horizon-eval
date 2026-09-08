"""The program's outer shell: the startup pipeline of appspec/01 section 4.

    1. Parse argv.  --help / --version short-circuit to stdout, exit 0, touching
       nothing else.  Conflicting force flags are rejected here, exit 1.
    2. Load and validate the user config (resolving the storage location).
    3. Dispatch to the requested subcommand.

Steps 2 and 3 of appspec/02's numbered list also cover assembling the
application database; that resolver lands with macklebox-resolvers-rdi.3 and
hangs off the configuration.
"""

import sys

from macklebox import config as configuration
from macklebox import output, parser
from macklebox.commands import COMMANDS
from macklebox.errors import EXIT_FAILURE, EXIT_OK, MackupError, UsageError
from macklebox.version import VERSION

# appspec/07 "Error behavior summary": one of the literal tokens that IS
# contract, matched by scripts and tests.
FORCE_CONFLICT_MESSAGE = "Options --force and --force-no are mutually exclusive."


def main(argv=None):
    """Entry point for the `mackup` console command.  Returns an exit code."""
    argv = sys.argv[1:] if argv is None else list(argv)

    try:
        invocation = parser.parse(argv)
    except UsageError as error:
        return _report_usage_error(error)

    short_circuit = _short_circuit(invocation)
    if short_circuit is not None:
        return short_circuit

    try:
        return _run(invocation)
    except MackupError as error:
        output.print_fatal(str(error))
        return error.exit_code


def _short_circuit(invocation):
    """Step 1's exits: help, version, bare usage, conflicting force flags.

    Returns an exit code, or None to continue into the pipeline.  --help and
    --version are the only paths that both succeed and skip the config-load gate
    (appspec/02), so they are answered before anything else -- including the
    force-flag conflict, which is not an error the user asking for help needs.
    """
    options = invocation.options

    if options.help:
        output.write_raw(parser.HELP)
        return EXIT_OK

    if options.version:
        output.print_message("Mackup {0}".format(VERSION))
        return EXIT_OK

    if invocation.action == parser.USAGE_ACTION:
        # appspec/02: a bare invocation is a usage display, not an error --
        # usage block to the user, exit 0.
        output.write_raw(parser.USAGE)
        return EXIT_OK

    if options.force and options.force_no:
        # Rejected before config is loaded and before any action is taken.
        output.print_fatal(FORCE_CONFLICT_MESSAGE)
        return EXIT_FAILURE

    return None


def _run(invocation):
    """Steps 2 and 3: the universal config-load gate, then dispatch."""
    config = configuration.load(invocation.options)
    COMMANDS[invocation.action](invocation, config)
    return EXIT_OK


def _report_usage_error(error):
    """appspec/02 "Argument-parser behavior": a warning line naming the offending
    argument, then the usage block.  appspec/07 routes argument-parser usage and
    warning text to stderr on a usage error."""
    output.write_raw("mackup: {0}".format(error), sys.stderr)
    output.write_raw(parser.USAGE, sys.stderr)
    return EXIT_FAILURE


if __name__ == "__main__":  # pragma: no cover
    sys.exit(main())
