"""The user configuration file and the storage location.

STUB.  appspec/02-invocation.md "Command dispatch order and the universal
config-load gate" requires that every command except --help / --version load the
user config *before* dispatching, and that a config or storage-location failure
aborts the run whatever the subcommand was.  The dispatch order is real here;
the resolution behind it is not built yet.

Real behavior lands with the resolvers epic -- macklebox-resolvers-rdi.2 (user
config file and storage-location resolution), appspec/03 and appspec/04.
"""


class Configuration(object):
    """The three decided facts of appspec/01 section 1 -- once they are decided.

    Today it carries only what the CLI supplied, so command code written against
    it does not have to change shape when the resolvers land.
    """

    def __init__(self, config_file=None):
        self.config_file = config_file


def load(options):
    """Step 2 of the startup pipeline: load and validate the user config.

    Resolves the storage engine's folder location, which is why a failure here
    is fatal for `list` and `show` too.  Stubbed until the resolvers epic.
    """
    return Configuration(config_file=options.config_file)
