"""The error regimes of appspec/01 §6 and the error table in appspec/07.

A ``MackupError`` is the *guarded* regime: a diagnostic on stderr and a clean
non-zero exit, never a partial effect and never anything on stdout.
"""

from __future__ import annotations


class MackupError(Exception):
    """A cleanly-handled fatal error: diagnostic to stderr, exit ``1``."""

    exit_code = 1


class UsageError(MackupError):
    """argv matched none of the usage lines in appspec/02 'Invocation forms'.

    ``warning`` is the offending argument's diagnostic line, printed above the
    usage block when the parser can name what it could not match; it is ``None``
    when the shape itself is wrong (``mackup show`` with no ``<application>``)
    and there is no single token to blame.
    """

    def __init__(self, warning: str | None = None) -> None:
        super().__init__(warning or "usage error")
        self.warning = warning


class NotImplementedYet(MackupError):
    """A pipeline stage a later ticket lands.

    It is an error, not a no-op, so that no command can exit ``0`` reporting work
    it has not performed (appspec/00 promise 9).
    """

    def __init__(self, what: str) -> None:
        super().__init__(f"{what} is not implemented yet")
