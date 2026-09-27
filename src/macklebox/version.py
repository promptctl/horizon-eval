"""Version-string resolution, per appspec/00-overview.md 'Provenance'.

The version comes from installed package metadata; when that is unavailable
(running from an uninstalled tree) it is the literal token ``unknown``.
"""

from __future__ import annotations

from importlib.metadata import PackageNotFoundError, version as _distribution_version

#: The distribution whose metadata carries the version string.
DISTRIBUTION_NAME = "macklebox"

#: The stable fallback token when no package metadata is installed.
FALLBACK_VERSION = "unknown"

#: The application's own name, as it appears in its output.
DISPLAY_NAME = "Mackup"


def resolve_version() -> str:
    try:
        return _distribution_version(DISTRIBUTION_NAME)
    except PackageNotFoundError:
        return FALLBACK_VERSION


def version_line() -> str:
    """The single line ``--version`` prints: ``Mackup <version>``."""
    return f"{DISPLAY_NAME} {resolve_version()}"
