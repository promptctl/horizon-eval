"""Shared helper: run one invocation and capture what the process boundary sees."""

from __future__ import annotations

import io
from dataclasses import dataclass

import pytest

from macklebox.main import run


@dataclass(frozen=True)
class Result:
    exit_code: int
    stdout: str
    stderr: str


@pytest.fixture
def cli():
    def invoke(*argv: str, stdin: str = "") -> Result:
        stdout, stderr = io.StringIO(), io.StringIO()
        code = run(
            list(argv), stdout=stdout, stderr=stderr, stdin=io.StringIO(stdin)
        )
        return Result(code, stdout.getvalue(), stderr.getvalue())

    return invoke
