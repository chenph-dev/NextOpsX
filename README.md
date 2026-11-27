# NextOpsX

**English** | [简体中文](README_CN.md)

NextOpsX is a desktop operations workbench for individual operators and SREs, covering SSH, SFTP, databases, Redis, and Elasticsearch. It is built on the MyGo native UI: a single Go codebase with no WebView dependency.

## Highlights

- **One workbench instead of scattered tabs** — assets, terminals, database consoles, and AI assistance share the same window.
- **Local-first** — assets, command history, and credentials are stored on the machine by default; secrets are encrypted with AES-256-GCM.
- **Controlled execution** — AI-suggested commands follow approval policies, and model, tool, and command activity is audited.
- **Native desktop integration** — native menu, tray, notifications, themes, UI scaling, and a Chinese/English interface.

## Features

### Assets and connections

- Asset tree with folders and hosts; create, edit, move, rename, and test connections
- Protocol picker with per-protocol defaults and field schemas
- Windows assets can launch the system Remote Desktop client (only from the Windows app)
- Restores the previous layout and selected asset after restart

### Terminal and files

- Multiple SSH terminal tabs; sessions stay alive while switching assets
- SFTP drawer for browsing and transfer tasks
- Command history and favorites recorded from both the AI assistant and in-app execution, with merged, de-duplicated suggestions while typing (↑/↓ to select, Tab to insert), search, output inspection, re-run, and templates such as `{{host}}` and `{{user}}`

### Databases, caches, and search

- SQL workbench for MySQL, PostgreSQL, SQLite, ClickHouse, and SQL Server: object tree, multiple SQL tabs, result grid, paging, and CSV export
- Redis: key space, key details, TTL, command console, and INFO
- Elasticsearch: indices, Query DSL, REST console, and cluster information

### AI operations

- Multiple model providers with streaming replies
- Command and plan approval; an approved plan can continue by executing ordinary commands
- Conversation history and AI call audit
- Skills: install from ZIP, enable/disable, delete, and preview `SKILL.md`

Conversation context is sent to the configured model provider when you use the assistant.

### Desktop

- Approval policies, desktop notifications, native menu, and tray
- System/light/dark themes, UI scaling, and Chinese/English UI
- In-app updates from GitHub Releases

## Supported connections

| Protocol | Capabilities |
| --- | --- |
| SSH | Interactive terminal (multi-tab), SFTP, command execution, AI operations |
| WinRM | Command console, AI operations |
| MySQL / PostgreSQL / ClickHouse / SQL Server | SQL workbench, object tree, CSV export |
| SQLite | Local database workbench |
| Redis | Key browser, key details, TTL, command console, INFO |
| Elasticsearch | Index browser, Query DSL, REST console, cluster information |

## Requirements

- Go 1.27.3 or newer
- Linux desktop: GTK runtime; the tray also needs `libayatana-appindicator3` (without it, only the tray is unavailable)
- Packaging: Windows installers need NSIS (fetched automatically by `mygo build` on Windows); macOS signing and DMG packaging require macOS

## Getting started

Run from source:

```bash
go run .
```

For hot reload during development:

```bash
go tool mygo dev
```

Build and verify:

```bash
go build ./...
go test ./...
go tool mygo vet .
```

## Build installers

```bash
go tool mygo build                    # current platform
go tool mygo build -platform linux/amd64,linux/arm64,windows/amd64,windows/arm64,darwin/arm64,darwin/amd64
```

Artifacts are written to `build/`:

- Linux: executable, `.deb`, `.tar.gz`, and a standalone `install.sh`
- Windows: executable and NSIS installer
- macOS: `.app` and `.dmg` (signing and notarization require macOS)
- With `updates` configured: signed update archives and one `update-<platform>.json` per build target

## Data and credentials

By default, data is stored in a `nextopsx/` directory under the platform configuration directory:

| Platform | Path |
| --- | --- |
| Linux | `~/.config/nextopsx` |
| Windows | `%APPDATA%\nextopsx` |
| macOS | `~/Library/Application Support/nextopsx` |

The SQLite database is `nextopsx.db`.

- `NEXTOPSX_DATA_DIR` overrides the directory, for testing or portable deployments.
- Passwords and private keys are encrypted with AES-256-GCM. The master key is stored in the OS keyring when available, with a `master.key` file (mode `0600`) as fallback.

## Auto-update

### Installing updates

Builds produced by `go tool mygo build` with `updates` configured can update themselves: open **Settings → About**, check for a new version, download it, and restart. The manifest is fetched from the latest GitHub release, the download is verified against the public key embedded in the build, and the app is replaced in place.

In-app updates need a writable installation directory. Versions installed by a system package manager (for example a `.deb` under `/opt`) are updated by that package manager instead, and the app reports that self-update is unavailable.

### Publishing a release

The signing key pair is created by `mygo keygen`:

```
<config-dir>/mygo/update-keys/mygo-update.key   # secret key; back it up offline, never commit it
<config-dir>/mygo/update-keys/mygo-update.pub   # public key; already configured in mygo.json
```

`<config-dir>` is the platform configuration directory used by MyGo; on Linux it is `~/.config`.

To publish from a local machine (`gh` logged in, or `GH_TOKEN`/`GITHUB_TOKEN` set):

```bash
MYGO_UPDATER_PRIVATE_KEY="$(cat ~/.config/mygo/update-keys/mygo-update.key)" \
  go tool mygo build -platform linux/amd64,linux/arm64,windows/amd64,windows/arm64,darwin/arm64,darwin/amd64 -upload
```

Losing the secret key prevents installed users from receiving further updates; leaking it lets anyone publish updates they will accept.

CI release flow:

1. Add the secret key to the repository: `gh secret set MYGO_UPDATER_PRIVATE_KEY < ~/.config/mygo/update-keys/mygo-update.key`
2. Update `version` in `mygo.json` and the matching section in `CHANGELOG.md`, then commit.
3. Tag and push: `git tag v0.2.0 && git push origin v0.2.0`
4. `.github/workflows/release.yml` checks that the tag matches `mygo.json`, builds the Linux and Windows targets on Ubuntu and the macOS targets on macOS, then uploads the artifacts to a draft release. Review it, then **publish it and mark it as the latest release**.
5. `.github/workflows/ci.yml` runs build, tests, and vet on pushes and pull requests.

Each update manifest is tied to one build target (for example `update-linux-amd64.json`), and its URL is compiled into the binary. Published architectures must match the manifests: a `darwin/universal` build needs `update-darwin-universal.json`, or it cannot check for updates. This repository publishes `darwin/arm64` and `darwin/amd64` separately.

## Development

### Optional SSH terminal test

```bash
NEXTOPSX_SSH_TEST_HOST=<host> \
NEXTOPSX_SSH_TEST_USER=<user> \
NEXTOPSX_SSH_TEST_KEY=$HOME/.ssh/<key> \
go test -run TestSSHTerminal -v ./service/terminal
```

### Commit message hook

The repository filters automated co-author trailers and generation notices from commit messages. Enable the hook once after cloning:

```bash
git config core.hooksPath .githooks
```

The hook lives in `.githooks/commit-msg`; ordinary commits are unaffected. Use `git commit --no-verify` when the original message must be kept.

## Project layout

```text
cmd/nextopsx       Composition root: window, menu, tray, icons
ui/workspace       UI state and actions (view model); no rendering or direct I/O
ui/views           Window shell and feature panels
ui/views/ai        AI operations panel
ui/views/db        Database workbench
ui/views/redis     Redis panel
ui/views/es        Elasticsearch panel
ui/views/sftp      SFTP drawer
ui/views/hosts     Asset tree, forms, welcome page
ui/views/terminal  Terminal completion and command history panel
ui/views/winrm     WinRM console
ui/views/settings  Settings sections
ui/views/result    Shared result grid
ui/markdown        Native Markdown rendering
ui/ansi            ANSI SGR parsing and terminal colors
ui/style           UI scaling
ui/widgets         Shared status text and terminal output
ui/assets          Embedded icons
i18n               zh-CN / en-US string catalogs
service/backend    Composition root for feature services
service/hosts      Assets, folders, connection tests
service/ai         AI sessions and providers
service/db         Database queries and schema
service/sftp       SFTP browsing and transfer tasks
service/skills     Operations skills
service/snippets   Command favorites
service/settings   UI settings
service/audit      AI call audit
service/policy     Approval policies
service/history    Command history
service/rdp        External RDP launcher
service/terminal   Interactive terminal adapter
service/executor   Connector executor resolution
domain             Connectors, executors, SFTP, skills
agent              AI sessions, models, tools, prompts
store              SQLite, encryption, migrations, domain stores
resources          Application icon sources
```
