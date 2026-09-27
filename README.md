# macklebox

An MIT-licensed tool for keeping application settings in sync across machines:
it keeps the one real copy of each config file in a folder that already syncs
between your machines (a cloud-drive folder or any replicated directory) and
wires each application to read it from there.

This repository is a clean-room reimplementation built **only** from the
behavioral specification in [`appspec/`](appspec/) — a black-box description of
an existing config-sync tool's observable behavior, written so an independent
team can rebuild a behaviorally-equivalent, command-line-compatible program
without reference to any other implementation.

## Status

Under construction. The command-line boundary is in place — the invocation
grammar, the global options, the dispatch order and the exit codes of
[`appspec/02-invocation.md`](appspec/02-invocation.md); the commands behind it
are not yet implemented and fail loudly rather than reporting success. Start at
[`appspec/00-overview.md`](appspec/00-overview.md) — the spec reads top-down
through altitudes (product contract → architecture → boundary detail).

## Layout

| Path | What it is |
|------|------------|
| `appspec/` | The functional specification that drives the build (source of truth) |
| `src/macklebox/` | The implementation. `macklebox` is the package name; the command it installs is `mackup` |
| `tests/` | The test suite |
| `LICENSE`  | MIT |

## Building and testing

Python, managed with [uv](https://docs.astral.sh/uv/):

```sh
uv sync          # create .venv and install macklebox plus its dev tools
uv run pytest    # run the suite
uv run mackup --help
```

The package version is the version string the application reports, so
`--version` prints `Mackup 0.11.1` — the reference build's string, which
`appspec/00-overview.md` pins as observable.
