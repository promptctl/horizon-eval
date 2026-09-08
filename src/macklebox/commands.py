"""Subcommand dispatch -- step 3 of the startup pipeline (appspec/02).

Each subcommand is a leaf on one tree (appspec/01 section 1: "All five sync
commands are five leaves on one tree, not five independent programs").  The
leaves are stubs until their own tickets land; the *dispatch* is real, so the
order parse -> config load -> dispatch is observable now.
"""

from macklebox import parser
from macklebox.errors import NotYetImplemented


def _unbuilt(name, ticket):
    raise NotYetImplemented(
        "Error: '{0}' is not implemented yet ({1}).".format(name, ticket)
    )


def list_applications(invocation, config):
    _unbuilt("list", "macklebox-resolvers-rdi.4")


def show_application(invocation, config):
    _unbuilt("show", "macklebox-resolvers-rdi.4")


def backup(invocation, config):
    _unbuilt("backup", "macklebox-copy-sync-z0c.3")


def restore(invocation, config):
    _unbuilt("restore", "macklebox-copy-sync-z0c.3")


def link(invocation, config):
    _unbuilt("link", "macklebox-link-sync-8ow.3")


def link_install(invocation, config):
    _unbuilt("link install", "macklebox-link-sync-8ow.2")


def link_uninstall(invocation, config):
    _unbuilt("link uninstall", "macklebox-link-sync-8ow.4")


COMMANDS = {
    parser.LIST: list_applications,
    parser.SHOW: show_application,
    parser.BACKUP: backup,
    parser.RESTORE: restore,
    parser.LINK: link,
    parser.LINK_INSTALL: link_install,
    parser.LINK_UNINSTALL: link_uninstall,
}
