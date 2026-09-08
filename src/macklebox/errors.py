"""Error types and the exit-code table (appspec/02-invocation.md "Exit codes")."""

EXIT_OK = 0
EXIT_FAILURE = 1


class MackupError(Exception):
    """A fatal, cleanly-handled error: diagnostic on stderr, exit 1.

    appspec/01-architecture.md section 6 calls this the *guarded* regime.  The
    shared post-condition is "no stdout, no filesystem change, non-zero exit".
    """

    exit_code = EXIT_FAILURE


class UsageError(MackupError):
    """argv did not match any listed usage line.

    appspec/02 "Argument-parser behavior": a warning line identifying the
    problem, then the usage block.  Both go to stderr (appspec/07 routes
    "argument-parser usage and warning text on a usage error" there).
    """


class NotYetImplemented(MackupError):
    """A subcommand whose behavior lands in a later ticket.

    Temporary: every subcommand is dispatched through the real pipeline already,
    so the dispatch order is observable, but the leaves are not built yet.
    """
