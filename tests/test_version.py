"""appspec/00 'Provenance': version from package metadata, else the token `unknown`."""

from __future__ import annotations

from importlib.metadata import PackageNotFoundError

from macklebox import version as version_module


def test_version_comes_from_installed_package_metadata():
    assert version_module.resolve_version() == "0.11.1"


def test_version_falls_back_to_a_stable_token_when_metadata_is_absent(monkeypatch):
    def absent(_name):
        raise PackageNotFoundError

    monkeypatch.setattr(version_module, "_distribution_version", absent)
    assert version_module.resolve_version() == "unknown"
    assert version_module.version_line() == "Mackup unknown"
