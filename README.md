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

Under construction. The command-line boundary (invocation grammar, global
options, dispatch order, exit codes, stream routing) is in place; the
subcommands behind it are being built out. Start at
[`appspec/00-overview.md`](appspec/00-overview.md) — the spec reads top-down
through altitudes (product contract → architecture → boundary detail).

## The command is `mackup`

`macklebox` is the project and package name. The observable surface — the
command name, its output, the config filenames, the `Mackup <version>` string —
is the specification's, because command-line compatibility *is* the deliverable.

## Running it

Requires Python 3.9+ and [uv](https://docs.astral.sh/uv/).

```sh
uv run mackup --help      # run the command
uv run pytest             # run the black-box conformance suite
```

## Layout

| Path | What it is |
|------|------------|
| `appspec/` | The functional specification that drives the build (source of truth) |
| `src/macklebox/` | The implementation |
| `tests/` | Black-box tests: run the real command, assert on stdout/stderr/exit code |
| `LICENSE`  | MIT |
