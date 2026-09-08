"""The argv boundary: appspec/02-invocation.md.

Every assertion here is made at the process boundary -- stdout, stderr, exit
code -- because that is the whole of what the specification defines.
"""

import re

import pytest

from macklebox.cli import FORCE_CONFLICT_MESSAGE

USAGE_LINES = [
    "mackup [options] list",
    "mackup [options] show <application>",
    "mackup [options] backup [<application>]",
    "mackup [options] restore [<application>]",
    "mackup [options] link install [<application>]",
    "mackup [options] link uninstall [<application>]",
    "mackup [options] link [<application>]",
]

# The subcommand forms of appspec/02 "Invocation forms".  None of them is built
# yet, so each is expected to reach dispatch and stop there -- which is exactly
# what proves the grammar accepted it.
ACCEPTED_FORMS = [
    ("list",),
    ("show", "vim"),
    ("backup",),
    ("backup", "vim"),
    ("restore",),
    ("restore", "vim"),
    ("link",),
    ("link", "vim"),
    ("link", "install"),
    ("link", "install", "vim"),
    ("link", "uninstall"),
    ("link", "uninstall", "vim"),
]


class TestHelp(object):
    """appspec/02: --help prints usage to stdout and exits 0, no config read."""

    @pytest.mark.parametrize("flag", ["-h", "--help"])
    def test_prints_usage_to_stdout_and_exits_zero(self, mackup, flag):
        result = mackup(flag)
        assert result.exit_code == 0
        assert result.stderr == ""
        for line in USAGE_LINES:
            assert line in result.out

    def test_help_wins_over_a_subcommand(self, mackup):
        """"No other action": help short-circuits before dispatch."""
        result = mackup("--help", "backup")
        assert result.exit_code == 0
        assert "Usage:" in result.out

    def test_help_wins_over_conflicting_force_flags(self, mackup):
        """Help is answered before the force-flag rejection; it touches nothing."""
        result = mackup("--force", "--force-no", "--help")
        assert result.exit_code == 0
        assert FORCE_CONFLICT_MESSAGE not in result.err

    @pytest.mark.parametrize(
        "argv",
        [
            ("--help", "show"),          # `show` wants an <application>
            ("--help", "frobnicate"),    # unrecognized subcommand
            ("--help", "backup", "a", "b"),  # extra positional
            ("--frobnicate", "--help"),  # unrecognized option, help later
        ],
        ids=lambda a: " ".join(a),
    )
    def test_help_wins_over_a_usage_error(self, mackup, argv):
        """appspec/02: "No other action" -- grammar validation is one of the
        actions help skips, not a check that runs ahead of it."""
        result = mackup(*argv)
        assert result.exit_code == 0
        assert "Usage:" in result.out
        assert result.stderr == ""

    def test_help_as_a_config_path_is_not_a_help_request(self, mackup):
        """`-c` consumes its value; that value is a path, not a flag."""
        result = mackup("-c", "--help", "list")
        assert result.exit_code == 1
        assert "is not implemented yet" in result.err


class TestVersion(object):
    """appspec/02: --version prints `Mackup <version>` to stdout, exits 0."""

    def test_prints_version_line_to_stdout_and_exits_zero(self, mackup):
        result = mackup("--version")
        assert result.exit_code == 0
        assert result.stderr == ""
        assert re.match(r"^Mackup \S+$", result.out.strip())

    def test_version_string_is_the_package_version_or_the_unknown_token(
        self, mackup
    ):
        """appspec/00 "Provenance": package metadata, else the literal token."""
        version = result_version(mackup)
        assert version == "0.11.1" or version == "unknown"

    def test_version_is_colored_even_when_piped(self, mackup):
        """appspec/07: colour is not conditioned on stdout being a TTY."""
        result = mackup("--version")
        assert "\033[" in result.stdout

    def test_version_wins_over_a_usage_error(self, mackup):
        result = mackup("--version", "frobnicate")
        assert result.exit_code == 0
        assert result.out.strip().startswith("Mackup ")
        assert result.stderr == ""


def result_version(mackup):
    return mackup("--version").out.strip().split(" ", 1)[1]


class TestBareInvocation(object):
    """appspec/02: a bare invocation is a usage display, not an error."""

    def test_prints_usage_and_exits_zero(self, mackup):
        result = mackup()
        assert result.exit_code == 0
        assert "Usage:" in result.out


class TestForceFlagExclusion(object):
    """appspec/02 "Mutually exclusive force flags"."""

    def test_single_line_on_stderr_exit_one(self, mackup):
        result = mackup("--force", "--force-no", "backup")
        assert result.exit_code == 1
        assert result.err.strip() == FORCE_CONFLICT_MESSAGE
        assert result.stdout == ""

    def test_short_form_force_is_equivalent(self, mackup):
        """Short and long forms are interchangeable."""
        result = mackup("-f", "--force-no", "list")
        assert result.exit_code == 1
        assert result.err.strip() == FORCE_CONFLICT_MESSAGE

    def test_rejected_with_no_subcommand(self, mackup):
        """appspec/02 grants the combination one exception -- --help/--version.
        A bare invocation is not it, so the usage display must not swallow it."""
        result = mackup("--force", "--force-no")
        assert result.exit_code == 1
        assert result.err.strip() == FORCE_CONFLICT_MESSAGE
        assert result.stdout == ""

    def test_fatal_message_is_colored(self, mackup):
        result = mackup("--force", "--force-no", "backup")
        assert "\033[" in result.stderr


class TestAcceptedForms(object):
    """Every listed usage line must be accepted by the parser."""

    @pytest.mark.parametrize("form", ACCEPTED_FORMS, ids=lambda f: " ".join(f))
    def test_form_is_not_a_usage_error(self, mackup, form):
        result = mackup(*form)
        assert "Usage:" not in result.err, "grammar rejected a listed form"

    @pytest.mark.parametrize("form", ACCEPTED_FORMS, ids=lambda f: " ".join(f))
    def test_form_reaches_dispatch(self, mackup, form):
        """Nothing behind the boundary is built yet, so every accepted form
        stops at its unbuilt leaf -- which is what proves it got there."""
        result = mackup(*form)
        assert result.exit_code == 1
        assert "is not implemented yet" in result.err

    @pytest.mark.parametrize(
        "option",
        ["-f", "--force", "--force-no", "-r", "--root", "-n", "--dry-run",
         "-v", "--verbose"],
    )
    def test_global_options_are_accepted_before_the_subcommand(
        self, mackup, option
    ):
        result = mackup(option, "backup")
        assert "Usage:" not in result.err

    def test_options_are_accepted_after_the_subcommand_too(self, mackup):
        result = mackup("backup", "vim", "--verbose")
        assert "Usage:" not in result.err

    @pytest.mark.parametrize(
        "form",
        [
            ("-c", "/tmp/x.cfg", "list"),
            ("--config-file", "/tmp/x.cfg", "list"),
            ("--config-file=/tmp/x.cfg", "list"),
            ("-c/tmp/x.cfg", "list"),
        ],
    )
    def test_config_file_option_takes_its_argument(self, mackup, form):
        result = mackup(*form)
        assert "Usage:" not in result.err

    def test_short_option_clusters(self, mackup):
        result = mackup("-vnf", "backup")
        assert "Usage:" not in result.err


class TestUsageErrors(object):
    """appspec/02 "Argument-parser behavior"."""

    def test_unrecognized_subcommand_warns_then_prints_usage(self, mackup):
        result = mackup("frobnicate")
        assert result.exit_code == 1
        assert "frobnicate" in result.err
        assert "Usage:" in result.err
        assert result.stdout == ""

    def test_show_requires_an_application(self, mackup):
        result = mackup("show")
        assert result.exit_code == 1
        assert "Usage:" in result.err

    def test_extra_positional_is_named(self, mackup):
        result = mackup("backup", "vim", "extra")
        assert result.exit_code == 1
        assert "extra" in result.err

    def test_list_takes_no_application(self, mackup):
        result = mackup("list", "vim")
        assert result.exit_code == 1
        assert "vim" in result.err

    def test_unknown_option_is_a_usage_error(self, mackup):
        result = mackup("--frobnicate", "list")
        assert result.exit_code == 1
        assert "--frobnicate" in result.err

    def test_config_file_option_requires_an_argument(self, mackup):
        result = mackup("list", "--config-file")
        assert result.exit_code == 1
        assert "--config-file" in result.err

    def test_config_file_option_rejects_an_empty_inline_argument(self, mackup):
        """`--config-file=` must not swallow the subcommand as its value."""
        result = mackup("--config-file=", "list")
        assert result.exit_code == 1
        assert "--config-file" in result.err

    def test_flag_option_rejects_an_inline_argument(self, mackup):
        result = mackup("--force=yes", "backup")
        assert result.exit_code == 1
        assert "--force" in result.err
