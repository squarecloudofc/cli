<div align="center">
  <img alt="Square Cloud Banner" src="https://cdn.squarecloud.app/png/github-readme.png">
</div>

<h1 align="center">Square Cloud CLI</h1>

<p align="center">A command line application to manage your <a href="https://squarecloud.app" target="_blank">Square Cloud</a> applications, databases and workspaces.</p>

## Installation

macOS, Linux and WSL, without sudo:

```bash
curl -fsSL https://cli.squarecloud.app/install | sh
```

The script installs into `~/.local/bin` (or `$SQUARECLOUD_INSTALL_DIR`)
after checking the download against the release's `checksums.txt`. If that
folder is not in your `PATH`, it asks before adding it to your shell's startup
file; without a terminal to ask, it prints the line to add.

Windows (PowerShell, no administrator rights):

```powershell
irm https://raw.githubusercontent.com/squarecloudofc/cli/master/scripts/install.ps1 | iex
```

It installs into `%LocalAppData%\squarecloud\bin` and adds that folder to your
user `PATH`.

Debian, Ubuntu, Fedora and RHEL: download the `.deb` or `.rpm` for your
architecture from the [latest release](https://github.com/squarecloudofc/cli/releases/latest)
and install it:

```bash
sudo apt install ./squarecloud_*.deb   # Debian, Ubuntu
sudo dnf install ./squarecloud-*.rpm   # Fedora, RHEL
```

With [npm](https://www.npmjs.com/package/@squarecloud/cli), on any system:

```bash
npm install -g @squarecloud/cli
```

### Verifying a download

Every release signs `checksums.txt`, which lists the SHA-256 of every archive
and package, with [cosign](https://docs.sigstore.dev/) keyless signing. To
check a release by hand:

```bash
cosign verify-blob checksums.txt \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp '^https://github.com/squarecloudofc/cli/.github/workflows/release.yml@refs/tags/v' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
sha256sum --ignore-missing -c checksums.txt
```

## Getting started

```bash
squarecloud auth login        # sign in through the browser
squarecloud init              # write a squarecloud.app for the current directory
squarecloud upload            # zip & upload the current directory as a new app
squarecloud commit -r         # commit the current directory and restart
squarecloud app logs -f       # print the latest logs and follow new ones (interactive app picker)
```

Commands that print data (`list`, `info`, `status`, `logs`, `whoami`,
`upload`, ...) accept `--json` for raw machine-readable output; the others
reject it with an error. App/db IDs can be omitted: the CLI falls back to the
`squarecloud.app` config file or an interactive picker.

## Ignoring files

`upload`, `commit` and `zip` leave out what `squarecloud.ignore`, in the
project root, lists. It is the same file, with the same rules, that the
Square Cloud VS Code extension reads:

- gitignore syntax: `#` comments, `!` to re-include, `dir/` for folders only,
  `/build` for the root only, `*`, `?`, `[abc]` and `**` (`**/tmp`, `logs/**`,
  `a/**/b`). The last matching line wins, and a file inside an ignored folder
  cannot be re-included. Matching ignores case.
- Always ignored first (re-include with `!`): `node_modules`, `.git`,
  `.github`, `.vscode`, `package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`.
- Symbolic links to files are uploaded as the file; links to folders and
  broken links are skipped with a note.

```gitignore
dist/
*.log
!keep.log
/build
.env
```

Older names (`.squareignore`, `.squarecloudignore`, `square.ignore`) are
still read when `squarecloud.ignore` is missing, with a warning to rename.

## Logging in

`squarecloud auth login` shows a short code and opens
<https://squarecloud.app/account/security/authorize> in your browser. Type the
code there and approve; the CLI receives a new API key with only the
permissions its commands use. Keys issued this way expire after 90 days
(`auth whoami` shows the date); run `auth login` again to renew. Press
Ctrl+C to cancel.

| Option | Use it when |
| ------ | ----------- |
| `--no-browser` | The browser cannot be opened from this machine (SSH, containers). The URL is printed; open it anywhere. |
| `--token <key>` | You already have an API key. It is checked before being saved. |
| `--with-token` | Scripts and CI: reads the key from stdin, e.g. `echo "$KEY" \| squarecloud auth login --with-token`. |

For CI, create an API key in the dashboard and set `SQUARECLOUD_API_KEY`:
every command uses it instead of the saved key, and `auth login` is not
needed. With `CI=true` the browser login refuses to run, since its key
expires and only a person can renew it. Without a terminal and without one
of these options, `auth login` exits with a hint.

If this machine is already logged in to another account, `auth login` asks
before replacing it (`--yes` skips the question).

`auth whoami` shows the account, its plan and, for browser logins, when the
key expires and its scopes. `auth logout` removes the saved key from this
machine. If a command reports that the session expired, the key was revoked
or it lacks a permission, run `squarecloud auth login` again.

### Where the key is kept

The key goes to the system keychain: Windows Credential Manager, the macOS
Keychain, or the Secret Service on Linux (GNOME Keyring, KWallet). Where there
is none (containers, WSL without D-Bus), it stays in the CLI's config file,
readable only by your user, and `auth login` says so. A key saved in the file
by an older version moves to the keychain the next time it is used.
`auth logout` removes it from both places.

### Several accounts

`--profile <name>` (or `SQUARECLOUD_PROFILE=<name>`) picks a separate login,
with its own keychain entry; without either, the `default` profile is used.

```bash
squarecloud auth login --profile work   # log in to a second account
squarecloud app list --profile work     # use it for one command
export SQUARECLOUD_PROFILE=work         # or for every command in this shell
```

`SQUARECLOUD_API_KEY`, when set, wins over every profile.

## Commands

| Command | Description |
| ------- | ----------- |
| `auth login` / `logout` / `whoami` | Manage your Square Cloud session |
| `status` | Square Cloud platform health, open incidents and affected components |
| `init [--yes] [--force]` | Create a `squarecloud.app` for the project in the current folder |
| `upgrade [--check]` | Update the CLI to the latest release |
| `integration install` / `uninstall [--dry-run]` | Add or remove "Send to Square Cloud" in the Windows Explorer menu |
| `zip` | Zip the current folder |
| `upload` | Upload a new application (also `app upload`) |
| `commit [app id] [--restart] [--path <dir>]` | Commit the current dir or `--file` (also `app commit`) |

### `app`

| Command | Description |
| ------- | ----------- |
| `app list` | List your applications |
| `app info [id]` | Detailed application info |
| `app status [id] [--all] [--raw]` | Runtime status (one app or every app) |
| `app logs [id] [-f]` | Most recent logs; `--follow` (`-f`) keeps printing new lines until Ctrl+C |
| `app metrics [id] [--range 24h\|7d]` | CPU, RAM, network, disk I/O and storage metrics |
| `app realtime [id]` | Stream live logs (stderr lines go to stderr) and status changes until Ctrl+C |
| `app start` / `restart` / `stop [id]` | Lifecycle signals |
| `app delete [id]` | Delete an application |
| `app domains` | Every domain configured across your apps |
| `app load-balancers` | Custom domains grouped by app |
| `app env list/set/remove/replace` | Environment variables (`set KEY=VAL...`, `--from-file .env`; `replace` needs at least one variable, `remove --all` clears them all) |
| `app file list/read/write/move/delete` | Remote file manager (`write --from <file>` is binary-safe; `read -o <file>` creates new files with 0600 permissions) |
| `app deploy list/current/webhook/github link\|unlink` | Deployments & Git integrations |
| `app network dns/domain/analytics/errors/logs/performance/purge-cache` | DNS, custom domain & edge observability |
| `app snapshot list/create/restore` | Application snapshots |

### `db`

| Command | Description |
| ------- | ----------- |
| `db list` | List your databases |
| `db create --name --memory --type --version` | Create a database |
| `db info/status/metrics [id]` | Inspect a database (`metrics --range 24h\|7d`) |
| `db start/stop [id]` | Lifecycle signals |
| `db update [id] --name/--memory` | Update a database |
| `db delete [id]` | Delete a database |
| `db credentials certificate/reset` | TLS certificate & credential rotation |
| `db snapshot list/create/restore` | Database snapshots |

### `blob`

| Command | Description |
| ------- | ----------- |
| `blob list [prefix]` | List your files in Square Cloud Blob, optionally under a prefix (`avatars/`) |
| `blob upload <file> [--name <name>] [--private]` | Upload a file; prints its ID and URL |
| `blob url <id>` | Print a download link (only the link goes to stdout) |
| `blob delete <id>` | Delete a file for good (asks first; `-y` skips) |
| `blob stats` | Files stored and storage used of your plan |

### `workspace`

| Command | Description |
| ------- | ----------- |
| `workspace list/create/info/delete/leave` | Manage workspaces |
| `workspace app add/remove` | Share apps with a workspace |
| `workspace member add/update/remove/invite-code` | Manage members |

Destructive commands (`delete`, `env replace`, `env remove --all`,
`credentials reset`, `purge-cache`, `snapshot restore`, `file read -o` over an
existing file, ...) ask for confirmation; pass `-y` to skip. Answering no, or running without a terminal (CI) and without `-y`,
exits with code 1 and changes nothing. `snapshot list <id>` prints the
version IDs that `snapshot restore` takes. Starting a running app or database (or stopping a stopped one) is
reported as such, not as an error.

A snapshot restore that takes longer than about 90 seconds keeps running on
the server. `snapshot restore` then follows it, printing each step, until it
succeeds or fails (exit code 1 with the reason); Ctrl+C stops watching, not
the restore. `--no-wait` returns at that point instead of following it. A snapshot download link
is valid for 1 hour.

`app metrics --range 7d` shows hourly averages and peaks on Pro and Enterprise
plans (apps of 512 MB or more). Other plans get the last 24 hours, with a note
on stderr.

`app logs -f` and `app realtime` open the stream again when the API ends it
(after 10 minutes), so they run until Ctrl+C. They stop with an error when 3
reopens in a row bring no logs or status.

`status` shows a check mark when the platform is online, a warning sign during
maintenance or degradation, and a question mark when the check itself could
not run, which does not mean the platform is down.

`app env set` and `replace` make the application receive each value exactly as
typed. The application reads its variables through a shell, so a value with
spaces or characters such as `$`, `&`, `;` or `|` is sent in quotes: single
quotes, or double quotes when the value has a single quote. A value already in
one pair of quotes is kept as written, so `'URL="${HOST}:3000"'` still refers
to another variable.

## Blob

`blob upload` names the object after the file, without its extension (the
service keeps the extension); `--name` picks another name. A file above
90 MiB goes up in parts, six at a time, sized from the limits the service
sends, with the progress on stderr. A failed or canceled upload is aborted, so
it does not hold one of your open-upload slots. Only what is safe to repeat is
retried: reads, and parts. A private file has no public URL: `blob url <id>`
makes a temporary link.

```console
$ squarecloud blob upload backup.tar.gz
Uploading backup.tar.gz (1.20 GB)...
Sent part 77 of 77
✓ Uploaded as 123456789/backup.tar.gz
  https://blob.squarecloud.dev/123456789/backup.tar.gz
```

Keys from an `auth login` before this version lack the Blob permissions: the
first Blob command then asks you to run `squarecloud auth login` again. A key
in `SQUARECLOUD_API_KEY` needs `blob:read` and `blob:write`.

## Windows Explorer

```bash
squarecloud integration install             # add the menu entry
squarecloud integration install --dry-run   # only print the registry values
squarecloud integration uninstall           # remove every key it added
```

`integration install` adds "Send to Square Cloud" to the right-click menu of
folders and `.zip` files (on Windows 11, under "Show more options"). It writes
under `HKEY_CURRENT_USER\Software\Classes`, so it needs no administrator
rights, and it points at the `squarecloud.exe` that ran it: run it again after
moving the CLI. A folder whose `squarecloud.app` has an ID is committed to that
application; any other folder or `.zip` is uploaded as a new application. A
Windows notification shows the result.

macOS and Linux have no such command, but a shortcut takes a few lines. In
Nautilus (GNOME Files), save this as
`~/.local/share/nautilus/scripts/Send to Square Cloud` and make it executable:

```sh
#!/bin/sh
cd "$1" && squarecloud upload
```

On macOS, create a Quick Action in Automator that receives folders in Finder
and runs the shell script `cd "$1" && squarecloud upload`, with "Pass input"
set to "as arguments".

## Language

Messages, help and flag descriptions follow your system language, detected on
every run: the `LANGUAGE`, `LC_ALL`, `LC_MESSAGES` or `LANG` variable when set,
otherwise the OS display language (on Windows, the display language, not the
regional format). A language without a translation falls back to English. To pick one for a single run: `LC_ALL=pt_BR squarecloud --help`.

## Debugging

Set `SQUARECLOUD_DEBUG=1` to trace every API request to stderr: method, path,
status and duration. Headers, query strings and bodies are never logged, so the
output is safe to share.

## Init

`squarecloud init` writes the `squarecloud.app` file
(`squarecloud.config` on macOS) of the project in the current folder. It
detects `MAIN` (the `main` of `package.json`, or `index.js`, `main.py`,
`main.go`, `index.html`...), uses the folder name as `DISPLAY_NAME`, the
minimum `MEMORY` (256 MB, or 512 MB for a website or API, which also gets a
`SUBDOMAIN`) and `VERSION=recommended`, and asks for an optional `START`
command and `AUTORESTART`. It can also create a starter `squarecloud.ignore`.
`--yes` takes the detected values without asking; an existing file is kept
unless you pass `--force`.

## Update

```bash
squarecloud upgrade           # download the latest release and replace this binary
squarecloud upgrade --check   # only tell whether a new version exists
```

`upgrade` downloads the archive for your system from the GitHub release,
checks it against the release's `checksums.txt` and replaces the binary (on
Windows the running one is kept as `squarecloud.exe.old` and deleted on the
next run). A copy that npm or the system package manager installed is theirs
to update, so `upgrade` says how instead:

```bash
npm install -g @squarecloud/cli@latest   # installed with npm
sudo apt install ./squarecloud_*.deb     # from a .deb: download the new one from the release
sudo dnf install ./squarecloud-*.rpm     # from an .rpm: same
```

A binary in `/usr/bin` counts as installed from a package: the `.deb` and
`.rpm` put it there, and the install scripts never do.
