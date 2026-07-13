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
squarecloud auth login        # authenticate with your API token
squarecloud upload            # zip & upload the current directory as a new app
squarecloud commit -r         # commit the current directory and restart
squarecloud app logs          # tail the latest logs (interactive app picker)
```

Every command accepts `--json` for raw machine-readable output, and app/db
IDs can be omitted — the CLI falls back to the `squarecloud.app` config file
or an interactive picker.

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
| `app realtime [id]` | Stream realtime events (SSE) |
| `app start` / `restart` / `stop [id]` | Lifecycle signals |
| `app delete [id]` | Delete an application |
| `app domains` | Every domain configured across your apps |
| `app load-balancers` | Custom domains grouped by app |
| `app env list/set/remove/replace` | Environment variables (`set KEY=VAL...`, `--from-file .env`) |
| `app file list/read/write/move/delete` | Remote file manager |
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

Destructive commands (`delete`, `env replace`, `credentials reset`,
`purge-cache`, `snapshot restore`, ...) ask for confirmation; pass `-y` to
skip.

## Debugging

Set `SQUARECLOUD_DEBUG=1` to log every API request/response to stderr.

## Update

macOS, Linux, and WSL:

```bash
curl -fsSL https://cli.squarecloud.app/install.sh | sh
```

Windows:

```bash
squarecloud update
```
