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

In progress. The command-line boundary — invocation grammar, global options,
dispatch order, exit codes, `--help` and `--version` — is implemented; the
startup resolvers and the sync operations behind it are not yet. Every
subcommand parses and dispatches, then reports that its behavior is not built
yet.

Start reading at [`appspec/00-overview.md`](appspec/00-overview.md) — the spec
reads top-down through altitudes (product contract → architecture → boundary
detail).

## Build and run

Requires Go 1.25 or newer.

```sh
make build          # builds ./bin/mackup
make check          # go vet + go test ./...
./bin/mackup --help
```

The built command is named **`mackup`**, not `macklebox`. `appspec/02` fixes the
observable surface — the command name, the grammar, literal contract tokens like
`~/.mackup.cfg` and the `Mackup <version>` line — and a conforming build must
reproduce it. `macklebox` is the project and package name only.

`--version` reports the package's own version when the build carries one and the
stable token `unknown` otherwise, per `appspec/00` "Provenance". Release builds
stamp it in:

```sh
make build VERSION=0.11.1   # ./bin/mackup --version -> "Mackup 0.11.1"
```

### Why Go

`appspec/00` states that any language reproducing the observable behavior is
conformant; the reference's Python is not a requirement. Go was chosen for a
single self-contained binary (the conformance rig can exec it under a throwaway
`HOME` with no interpreter or dependency setup), a standard library that already
covers symlinks, permission bits, effective UID, and subprocess invocation, and
`runtime/debug.ReadBuildInfo`, which resolves the version exactly the way the
spec's provenance rule describes.

## Layout

| Path | What it is |
|------|------------|
| `appspec/` | The functional specification that drives the build (source of truth) |
| `cmd/mackup/` | The `mackup` entry point |
| `internal/cli/` | The command-line boundary: argv grammar, dispatch, exit codes (`appspec/02`) |
| `internal/ui/` | Process streams; message-to-stream routing is contract (`appspec/07`) |
| `internal/version/` | Version-string resolution (`appspec/00` "Provenance") |
| `LICENSE`  | MIT |
