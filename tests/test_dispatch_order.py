"""appspec/02 'Command dispatch order' and the exit-code table."""

from __future__ import annotations

import pytest

from macklebox.main import FORCE_CONFLICT_MESSAGE
from macklebox.usage import USAGE
from macklebox.version import DISPLAY_NAME, resolve_version


def test_help_prints_usage_to_stdout_and_exits_zero(cli):
    result = cli("--help")
    assert result.exit_code == 0
    assert USAGE in result.stdout
    assert result.stderr == ""


def test_help_short_form_matches_long_form(cli):
    assert cli("-h") == cli("--help")


def test_version_prints_the_version_line_to_stdout_and_exits_zero(cli):
    result = cli("--version")
    assert result.exit_code == 0
    assert result.stdout == f"{DISPLAY_NAME} {resolve_version()}\n"
    assert result.stderr == ""


def test_bare_invocation_shows_usage(cli):
    result = cli()
    assert result.exit_code == 0
    assert USAGE in result.stdout
    assert result.stderr == ""


def test_conflicting_force_flags_are_rejected_before_anything_else(cli):
    result = cli("--force", "--force-no", "backup")
    assert result.exit_code == 1
    assert result.stderr == f"{FORCE_CONFLICT_MESSAGE}\n"
    assert result.stdout == ""


def test_conflicting_force_flags_are_rejected_with_short_forms_too(cli):
    assert cli("-f", "--force-no", "list") == cli("--force", "--force-no", "list")


@pytest.mark.parametrize("command", [["--help"], ["--version"]])
def test_help_and_version_outrank_the_force_flag_conflict(cli, command):
    """Both short-circuit at parse time, before the conflict check (appspec/01 §4)."""
    result = cli("--force", "--force-no", *command)
    assert result.exit_code == 0
    assert result.stderr == ""


def test_unrecognized_subcommand_warns_and_shows_usage_on_stderr(cli):
    result = cli("frobnicate")
    assert result.exit_code == 1
    assert result.stdout == ""
    assert "frobnicate" in result.stderr
    assert USAGE in result.stderr


def test_show_without_an_application_is_a_usage_error(cli):
    result = cli("show")
    assert result.exit_code == 1
    assert result.stdout == ""
    assert USAGE in result.stderr


def test_unrecognized_option_is_a_usage_error(cli):
    result = cli("--bogus", "list")
    assert result.exit_code == 1
    assert result.stdout == ""
    assert "--bogus" in result.stderr


@pytest.mark.parametrize(
    "argv",
    [
        ["list"],
        ["show", "vim"],
        ["backup"],
        ["restore"],
        ["link"],
        ["link", "install"],
        ["link", "uninstall"],
    ],
)
def test_a_command_never_reports_success_it_has_not_performed(cli, argv):
    """Until the resolvers and sync epics land, every command fails loudly.

    Guards the exit-code table's line 0: a zero exit means the requested action
    completed.  A stub that returned 0 would break that before any real work
    exists, and appspec/07's shared post-condition for a failing run -- nothing on
    stdout, a diagnostic on stderr, a non-zero exit -- holds here too.
    """
    result = cli(*argv)
    assert result.exit_code != 0
    assert result.stdout == ""
    assert result.stderr.startswith("Error: ")
