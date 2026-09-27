"""The observable process boundary: the real command, in a real subprocess."""

from __future__ import annotations

import shutil
import subprocess
import sys

import pytest

from macklebox.version import DISPLAY_NAME, resolve_version

EXPECTED_VERSION_LINE = f"{DISPLAY_NAME} {resolve_version()}"


def _commands():
    yield pytest.param([sys.executable, "-m", "macklebox"], id="python -m macklebox")
    console_script = shutil.which("mackup")
    yield pytest.param(
        [console_script],
        id="mackup",
        marks=pytest.mark.skipif(
            console_script is None, reason="the mackup console script is not installed"
        ),
    )


@pytest.mark.parametrize("command", list(_commands()))
def test_version_goes_to_stdout_with_exit_zero(command):
    completed = subprocess.run(
        [*command, "--version"], capture_output=True, text=True, check=False
    )
    assert completed.returncode == 0
    assert completed.stdout.strip() == EXPECTED_VERSION_LINE
    assert completed.stderr == ""


@pytest.mark.parametrize("command", list(_commands()))
def test_conflicting_force_flags_go_to_stderr_with_exit_one(command):
    completed = subprocess.run(
        [*command, "--force", "--force-no", "backup"],
        capture_output=True,
        text=True,
        check=False,
    )
    assert completed.returncode == 1
    assert completed.stdout == ""
    assert (
        completed.stderr.strip()
        == "Options --force and --force-no are mutually exclusive."
    )
