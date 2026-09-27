package cli

// Help is the usage/help text. Its exact wording is human-facing and is not a
// machine-read contract (appspec/02-invocation.md, "Argument-parser behavior");
// the invocation forms and option table it lists are.
const Help = `Keep your application settings in sync.

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

Commands:
  list                  Print all supported application keys.
  show                  Print one application's display name and config files.
  backup                Copy config files from home into the storage folder.
  restore               Copy config files from the storage folder into home.
  link install          Move home config files into storage, then symlink them back.
  link uninstall        Remove those symlinks and copy the storage files back into home.
  link                  Symlink storage config files into home.

Options:
  -h, --help                 Print this help and exit.
      --version              Print the version and exit.
  -f, --force                Answer yes to every confirmation prompt.
      --force-no             Answer no to every confirmation prompt.
  -r, --root                 Allow running as the superuser.
  -n, --dry-run              Print what would be done, without doing it.
  -v, --verbose              Print fuller progress detail.
  -c, --config-file <path>   Read this config file instead of the default.

<application> is an application key as printed by 'mackup list' (e.g. vim, git),
not a display name.
`
