<div align="center">
  <img alt="Square Cloud Banner" src="https://cdn.squarecloud.app/png/github-readme.png">
</div>

<h1 align="center">Square Cloud CLI</h1>

<p align="center">A command line application to manage your <a href="https://squarecloud.app" target="_blank">Square Cloud</a> applications, databases and workspaces.</p>

## Installation

macOS, Linux, and WSL:

```bash
curl -fsSL https://cli.squarecloud.app/install | bash
```

Windows (needs [npm](https://www.npmjs.com/) installed):

```bash
npm install -g @squarecloud/cli
```

Or visit the [@squarecloud/cli npm page](https://www.npmjs.com/package/@squarecloud/cli) for more information.

## Getting started

```bash
squarecloud auth login        # sign in through the browser
squarecloud upload            # zip & upload the current directory as a new app
squarecloud commit -r         # commit the current directory and restart
squarecloud app logs          # tail the latest logs (interactive app picker)
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
machine. If a command reports that the session expired or the key was
revoked, run `squarecloud auth login` again.

## Commands

| Command | Description |
| ------- | ----------- |
| `auth login` / `logout` / `whoami` | Manage your Square Cloud session |
| `status` | Square Cloud platform health |
| `zip` | Zip the current folder |
| `upload` | Upload a new application (also `app upload`) |
| `commit [app id] [--restart] [--path <dir>]` | Commit the current dir or `--file` (also `app commit`) |

### `app`

| Command | Description |
| ------- | ----------- |
| `app list` | List your applications |
| `app info [id]` | Detailed application info |
| `app status [id] [--all] [--raw]` | Runtime status (one app or every app) |
| `app logs [id]` | Most recent logs |
| `app metrics [id]` | Last 24h of CPU/RAM/network metrics |
| `app realtime [id]` | Stream live logs (stderr lines go to stderr) and status changes |
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
| `db info/status/metrics [id]` | Inspect a database |
| `db start/stop [id]` | Lifecycle signals |
| `db update [id] --name/--memory` | Update a database |
| `db delete [id]` | Delete a database |
| `db credentials certificate/reset` | TLS certificate & credential rotation |
| `db snapshot list/create/restore` | Database snapshots |

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

`app env set` and `replace` make the application receive each value exactly as
typed. The application reads its variables through a shell, so a value with
spaces or characters such as `$`, `&`, `;` or `|` is sent in quotes: single
quotes, or double quotes when the value has a single quote. A value already in
one pair of quotes is kept as written, so `'URL="${HOST}:3000"'` still refers
to another variable.

## Language

Messages, help and flag descriptions follow your system language, detected on
every run: the `LANGUAGE`, `LC_ALL`, `LC_MESSAGES` or `LANG` variable when set,
otherwise the OS display language (on Windows, the display language, not the
regional format). A language without a translation falls back to English. To pick one for a single run: `LC_ALL=pt_BR squarecloud --help`.

## Debugging

Set `SQUARECLOUD_DEBUG=1` to trace every API request to stderr: method, path,
status and duration. Headers, query strings and bodies are never logged, so the
output is safe to share.

## Update

macOS, Linux, and WSL:

```bash
curl -fsSL https://cli.squarecloud.app/install.sh | sh
```

Windows:

```bash
npm install -g @squarecloud/cli@latest
```
