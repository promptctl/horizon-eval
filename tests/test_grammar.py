"""appspec/02 'Invocation forms': every listed form parses, others are usage errors."""

from __future__ import annotations

import pytest

from macklebox.cli import Command, Invocation, parse_argv
from macklebox.errors import UsageError

ACCEPTED = [
    (["list"], Command.LIST, None),
    (["show", "vim"], Command.SHOW, "vim"),
    (["backup"], Command.BACKUP, None),
    (["backup", "vim"], Command.BACKUP, "vim"),
    (["restore"], Command.RESTORE, None),
    (["restore", "vim"], Command.RESTORE, "vim"),
    (["link", "install"], Command.LINK_INSTALL, None),
    (["link", "install", "vim"], Command.LINK_INSTALL, "vim"),
    (["link", "uninstall"], Command.LINK_UNINSTALL, None),
    (["link", "uninstall", "vim"], Command.LINK_UNINSTALL, "vim"),
    (["link"], Command.LINK, None),
    (["link", "vim"], Command.LINK, "vim"),
]


@pytest.mark.parametrize("argv, command, application", ACCEPTED)
def test_listed_forms_are_accepted(argv, command, application):
    invocation = parse_argv(argv)
    assert invocation.command == command
    assert invocation.application == application


@pytest.mark.parametrize(
    "argv",
    [
        ["frobnicate"],
        ["list", "vim"],
        ["show"],
        ["show", "vim", "git"],
        ["backup", "vim", "git"],
        ["link", "install", "vim", "git"],
        ["link", "install", "uninstall", "vim"],
        ["--bogus", "list"],
        ["-z", "list"],
        ["--config-file"],
        ["-c"],
        ["--force=yes", "list"],
    ],
)
def test_non_matching_forms_are_usage_errors(argv):
    with pytest.raises(UsageError):
        parse_argv(argv)


def test_bare_invocation_is_a_usage_display():
    assert parse_argv([]).command == Command.USAGE


def test_options_may_appear_before_or_after_the_subcommand():
    before = parse_argv(["-v", "--dry-run", "backup", "vim"])
    after = parse_argv(["backup", "vim", "-v", "--dry-run"])
    assert before == after
    assert before == Invocation(
        command=Command.BACKUP, application="vim", dry_run=True, verbose=True
    )


@pytest.mark.parametrize(
    "argv, expected",
    [
        (["-f", "backup"], {"force": True}),
        (["--force", "backup"], {"force": True}),
        (["--force-no", "backup"], {"force_no": True}),
        (["-r", "backup"], {"root": True}),
        (["--root", "backup"], {"root": True}),
        (["-n", "backup"], {"dry_run": True}),
        (["--dry-run", "backup"], {"dry_run": True}),
        (["-v", "backup"], {"verbose": True}),
        (["--verbose", "backup"], {"verbose": True}),
        (["-c", "other.cfg", "backup"], {"config_file": "other.cfg"}),
        (["--config-file", "other.cfg", "backup"], {"config_file": "other.cfg"}),
        (["--config-file=other.cfg", "backup"], {"config_file": "other.cfg"}),
        (["-cother.cfg", "backup"], {"config_file": "other.cfg"}),
        (["-vnf", "backup"], {"verbose": True, "dry_run": True, "force": True}),
    ],
)
def test_global_options_table(argv, expected):
    invocation = parse_argv(argv)
    for field, value in expected.items():
        assert getattr(invocation, field) == value, field


def test_short_and_long_forms_are_interchangeable():
    assert parse_argv(["-f", "-r", "-n", "-v", "backup"]) == parse_argv(
        ["--force", "--root", "--dry-run", "--verbose", "backup"]
    )


def test_link_install_wins_over_an_application_named_install():
    """The usage lines are ordered, so `link install` is never `link <app=install>`."""
    assert parse_argv(["link", "install"]).command == Command.LINK_INSTALL
    assert parse_argv(["link", "installer"]) == Invocation(
        command=Command.LINK, application="installer"
    )


def test_double_dash_ends_option_parsing():
    assert parse_argv(["show", "--", "-weird-app"]) == Invocation(
        command=Command.SHOW, application="-weird-app"
    )


def test_help_short_circuits_past_otherwise_invalid_argv():
    assert parse_argv(["--help", "frobnicate"]).command == Command.HELP
    assert parse_argv(["frobnicate", "-h"]).command == Command.HELP
    assert parse_argv(["--version", "--bogus"]).command == Command.VERSION
