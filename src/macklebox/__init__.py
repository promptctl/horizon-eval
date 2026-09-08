"""macklebox — a clean-room reimplementation of the `mackup` command line tool.

The package name is `macklebox`; the *observable* surface (command name, output
text, config filenames) is `mackup`'s, exactly as the specification in
`appspec/` describes it.
"""

from macklebox.version import VERSION, VERSION_UNKNOWN

__all__ = ["VERSION", "VERSION_UNKNOWN"]
