"""Steps 2-5 of the one startup pipeline (appspec/01 §4, appspec/02 dispatch order).

Step 1 (argv parse) is ``cli``; this module owns what happens after it: the
universal config-load gate, the application-database assembly, the environment
gate, and the fan-out to one command.

The resolvers and the command bodies are stubs at this point — the resolvers and
sync epics land them.  The command stub *raises* rather than returning ``0``, so
no command can report work it has not done (appspec/00 promise 9); the resolver
stub has nothing to resolve yet and so cannot fail.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import TextIO

from .cli import Invocation
from .errors import NotImplementedYet


@dataclass(frozen=True)
class StartupContext:
    """The decided facts every command runs against (appspec/01 §1).

    The configuration resolver's three facts (storage root, sub-directory,
    application scope) and the application database land here.
    """


def load_startup_context(invocation: Invocation) -> StartupContext:
    """Steps 2-3: load and validate the config, then assemble the app database.

    Resolving the config resolves the storage location eagerly, which is why a
    storage failure aborts every command including ``list`` and ``show``
    (appspec/02 'Command dispatch order').  That gate arrives with the resolvers;
    until then there is nothing to resolve and nothing to fail on.
    """
    return StartupContext()


def dispatch(
    invocation: Invocation,
    context: StartupContext,
    *,
    stdout: TextIO,
    stderr: TextIO,
    stdin: TextIO | None,
) -> int:
    """Steps 4-5: the environment gate, then the requested command."""
    raise NotImplementedYet(f"the {invocation.command} command")
