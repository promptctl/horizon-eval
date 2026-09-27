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

Specification complete. Implementation under way in Go: the command-line
boundary (invocation grammar, global options, dispatch order, exit codes) is in
place; the resolvers and the sync commands are not yet. Start at
[`appspec/00-overview.md`](appspec/00-overview.md) — the spec reads top-down
through altitudes (product contract → architecture → boundary detail).

The built command is named `mackup`, not `macklebox`: the observable surface is
the spec's, and `macklebox` is the project and module name only.

## Build

```sh
make build        # ./bin/mackup
make check        # gofmt + go vet + go test (includes the conformance rig)
make conformance  # the black-box rig alone, verbosely
```

Go 1.25 or newer; no third-party dependencies.

## Layout

| Path | What it is |
|------|------------|
| `appspec/` | The functional specification that drives the build (source of truth) |
| `cmd/mackup/` | The command's entry point |
| `conformance/` | The black-box rig: the real binary under a throwaway home, observed at the process boundary |
| `internal/cli/` | The command-line boundary: grammar, options, dispatch, exit codes |
| `internal/version/` | Version-string resolution |
| `LICENSE`  | MIT |
