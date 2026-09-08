package cli

// usageText is the usage block printed for --help and for a bare invocation.
//
// appspec/02 "Argument-parser behavior" states the exact wording of usage,
// help, and parser-warning text is human-facing and deliberately not part of
// the machine-read contract; the grammar it describes is.
const usageText = `Mackup - Keep your application settings in sync.

Usage:
  mackup [options] list
  mackup [options] show <application>
  mackup [options] backup [<application>]
  mackup [options] restore [<application>]
  mackup [options] link install [<application>]
  mackup [options] link uninstall [<application>]
  mackup [options] link [<application>]
  mackup -h | --help
  mackup --version

Options:
  -h, --help                Show this help message and exit.
      --version             Show the version and exit.
  -f, --force               Answer every confirmation with "Yes".
      --force-no            Answer every confirmation with "No".
  -r, --root                Allow Mackup to be run as the superuser.
  -n, --dry-run             Show the steps that would be taken, change nothing.
  -v, --verbose             Show fuller progress, including full paths.
  -c, --config-file <path>  Use <path> as the config file instead of ~/.mackup.cfg.

Modes of operation:
  list             List every supported application key.
  show             Show one application's display name and configuration files.
  backup           Copy your configuration files into the storage folder.
  restore          Copy the configuration files from storage back into home.
  link install     Move home configuration files into storage, symlink them back.
  link             Symlink configuration files already in storage into home.
  link uninstall   Replace those symlinks with real files again.

<application> is an application key as printed by 'mackup list', not a display name.`
