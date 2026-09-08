"""`python -m macklebox` -- the same entry point as the `mackup` command."""

import sys

from macklebox.cli import main

if __name__ == "__main__":
    sys.exit(main())
