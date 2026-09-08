"""Output streams and colored output.

appspec/07-output-safety-lifecycle.md, "Output streams": *the stream a message
lands on is contract, not cosmetic*.  Every message the program emits goes
through one of the helpers below, so the routing decision is made once, at the
point the message class is named, rather than at each print site.

"Colored output": all human-facing output is ANSI-coloured by level and colour
alone conveys the level -- there are no textual level labels.  Colour is emitted
unconditionally; the program does **not** condition it on whether stdout is a
TTY.

The one deliberate exception to colouring is the argument parser's usage/help
text, which is emitted verbatim (`write_raw`): it is the grammar itself, not a
levelled message.
"""

import sys

ESC = "\033["
RESET = ESC + "0m"

# SGR codes, per appspec/07 "Colored output".
NORMAL = "33"  # yellow      -- progress / info
ANOMALY = "1;33"  # bold yellow -- non-fatal anomaly ("differs between ...")
SUCCESS = "32"  # green       -- success / additions in diffs
FATAL = "91"  # bright red  -- fatal errors that exit
FAILURE = "31"  # red         -- non-fatal copy-failure lines / removed lines
VERBOSE = "35"  # magenta     -- verbose-only traces
HEADER_RULE = "34"  # blue    -- per-app verbose header rules
BOLD = "1"  # bold           -- diff file headers, app name
HUNK = "36"  # cyan          -- diff hunk headers

# Any reset sequence embedded in a message would strip colour from the rest of
# the line; colouring is reset-safe, re-applying the colour after each one.
_RESETS = (RESET, ESC + "m")


def colorize(text, code):
    """Wrap `text` in SGR `code`, re-applying it after any embedded reset."""
    prefix = ESC + code + "m"
    body = str(text)
    for reset in _RESETS:
        body = body.replace(reset, reset + prefix)
    return prefix + body + RESET


def write_raw(text, stream=None):
    """Emit `text` verbatim -- no colour, no level.  Used for usage/help text."""
    stream = sys.stdout if stream is None else stream
    stream.write(text if text.endswith("\n") else text + "\n")
    stream.flush()


def _emit(text, code, stream):
    stream.write(colorize(text, code) + "\n")
    stream.flush()


def print_message(text):
    """Normal progress / info.  stdout, yellow."""
    _emit(text, NORMAL, sys.stdout)


def print_success(text):
    """Success.  stdout, green."""
    _emit(text, SUCCESS, sys.stdout)


def print_anomaly(text):
    """Non-fatal anomaly.  stdout (NOT stderr), bold yellow.

    appspec/07: "Do not generalize warnings -> stderr."  The drift
    "differs between ..." header and the `link uninstall` "does not point to
    Mackup ... skipping" warning are both stdout messages.
    """
    _emit(text, ANOMALY, sys.stdout)


def print_verbose(text):
    """Verbose-only trace.  stdout, magenta."""
    _emit(text, VERBOSE, sys.stdout)


def print_failure(text):
    """Non-fatal per-file failure diagnostic.  stderr, red."""
    _emit(text, FAILURE, sys.stderr)


def print_fatal(text):
    """Fatal error preceding a non-zero exit.  stderr, bright red."""
    _emit(text, FATAL, sys.stderr)
