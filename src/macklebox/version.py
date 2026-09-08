"""Version-string resolution.

appspec/00-overview.md "Provenance": the version reported by `--version` (and
printed in `list` output) is obtained from installed package metadata.  When
that metadata is unavailable -- e.g. running straight from an uninstalled source
tree -- the version string is the literal token `unknown`.
"""

VERSION_UNKNOWN = "unknown"

# The distribution whose metadata carries our version.  `mackup` is the command
# we present as; `macklebox` is what pip/uv installed.
DISTRIBUTION_NAME = "macklebox"


def _resolve_version() -> str:
    try:
        from importlib.metadata import PackageNotFoundError, version
    except ImportError:  # pragma: no cover - importlib.metadata is 3.8+
        return VERSION_UNKNOWN

    try:
        return version(DISTRIBUTION_NAME)
    except PackageNotFoundError:
        return VERSION_UNKNOWN
    except Exception:  # pragma: no cover - defensive: broken metadata
        return VERSION_UNKNOWN


VERSION = _resolve_version()
