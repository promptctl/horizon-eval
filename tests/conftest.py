"""Shared fixtures for the black-box tests.

The specification is behavioral, so the tests observe the program the way
appspec/00-overview.md "Provenance" says the spec was written: run the real
command as a subprocess and assert on stdout, stderr, and the exit code.

The throwaway-home rig proper lands with macklebox-foundation-a9d.2; what is
here is the minimum the argv boundary needs.
"""

import os
import re
import shutil
import subprocess
import sys

import pytest

_SGR = re.compile(r"\033\[[0-9;]*m")

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))


def strip_ansi(text):
    """Drop SGR sequences.

    appspec/07 permits treating the exact colour bytes as cosmetic; the stream a
    message lands on is what is contract.  Assertions about *text* therefore
    compare uncoloured output, and assertions about colour look for SGR
    explicitly.
    """
    return _SGR.sub("", text)


class Result(object):
    def __init__(self, completed):
        self.exit_code = completed.returncode
        self.stdout = completed.stdout
        self.stderr = completed.stderr

    @property
    def out(self):
        return strip_ansi(self.stdout)

    @property
    def err(self):
        return strip_ansi(self.stderr)

    def __repr__(self):  # pragma: no cover - failure output
        return "Result(exit_code={0!r}, stdout={1!r}, stderr={2!r})".format(
            self.exit_code, self.stdout, self.stderr
        )


def _command():
    """The command under test.

    Prefer the installed `mackup` console script -- that is the real entry point
    and the only one with package metadata, which `--version` reads.  Fall back
    to `python -m macklebox` from the source tree so the suite still runs
    uninstalled (where `--version` correctly reports the `unknown` token).
    """
    installed = shutil.which("mackup")
    if installed:
        return [installed]
    return [sys.executable, "-m", "macklebox"]


@pytest.fixture(scope="session")
def command():
    return _command()


@pytest.fixture
def mackup(command):
    """Run the command with the given arguments and capture the boundary."""

    def run(*args, **kwargs):
        env = dict(os.environ)
        env.setdefault("PYTHONPATH", os.path.join(REPO_ROOT, "src"))
        env.update(kwargs.pop("env", {}))
        completed = subprocess.run(
            command + list(args),
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            stdin=kwargs.pop("stdin", subprocess.DEVNULL),
            universal_newlines=True,
            cwd=kwargs.pop("cwd", REPO_ROOT),
            env=env,
            **kwargs
        )
        return Result(completed)

    return run
