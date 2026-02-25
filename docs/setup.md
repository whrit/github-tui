# Setup and Installation Guide

`ght` is a keyboard-driven terminal UI client for GitHub. It provides issue management, comments, and GitHub Actions workflow viewing without leaving your terminal. This guide covers every supported installation method and walks through configuration so you can go from zero to a working session as quickly as possible.

---

## Table of Contents

1. [Prerequisites](#1-prerequisites)
2. [Installation](#2-installation)
   - [A. Homebrew (macOS/Linux) — Recommended](#a-homebrew-macoslinux--recommended)
   - [B. go install](#b-go-install)
   - [C. Download a Pre-built Binary](#c-download-a-pre-built-binary)
   - [D. Build from Source](#d-build-from-source)
3. [Configuration](#3-configuration)
   - [First-run Interactive Wizard](#first-run-interactive-wizard)
   - [Manual Configuration](#manual-configuration)
   - [Config File Reference](#config-file-reference)
4. [Running ght](#4-running-ght)
5. [Verifying Your Setup](#5-verifying-your-setup)
6. [Upgrading](#6-upgrading)
7. [Uninstalling](#7-uninstalling)
8. [Troubleshooting](#8-troubleshooting)

---

## 1. Prerequisites

### Go (build-from-source only)

If you intend to install via `go install` or build from source, you need **Go 1.24 or later**. The minimum is declared in `go.mod`:

```
go 1.24.2
```

Check your installed version:

```bash
go version
```

Install or upgrade Go from [https://go.dev/dl/](https://go.dev/dl/).

Homebrew users and those downloading a pre-built binary do **not** need Go installed.

### GitHub Personal Access Token

`ght` requires a GitHub **Personal Access Token (classic)** to authenticate with the GitHub API. Fine-grained PATs are also accepted, but scope validation is skipped for them — if you encounter permission errors with a fine-grained token, verify your permissions manually.

**Required scopes for a classic PAT:**

| Scope | Why it is needed |
|---|---|
| `repo` | Read issues, comments, pull requests, and Actions workflow runs/logs |
| `workflow` | Trigger and manage GitHub Actions workflows |
| `read:org` | Read organization data (projects, members) |

> **Note:** The `project` or `admin:org` scope may be substituted for `read:org` when project-level access is needed. The token validator (`github/token_validator.go`) accepts any of `project`, `read:org`, or `admin:org` to satisfy the org-read requirement.

**Creating a classic PAT:**

1. Go to [https://github.com/settings/tokens](https://github.com/settings/tokens)
2. Click **Generate new token** → **Generate new token (classic)**
3. Give it a descriptive name (e.g., `ght-terminal`)
4. Set an expiration that suits your security posture
5. Check `repo`, `workflow`, and `read:org`
6. Click **Generate token** and copy the value immediately — GitHub will not show it again

Store the token somewhere safe (a password manager, for example) before proceeding to installation.

---

## 2. Installation

Four installation paths are supported. Homebrew is the simplest for macOS and Linux users. `go install` is a single command for Go developers. Pre-built binaries require no toolchain at all. Building from source gives you the most control.

### A. Homebrew (macOS/Linux) — Recommended

`ght` is distributed through the `whrit/tap` Homebrew tap, which is updated automatically on each release by GoReleaser.

```bash
brew tap whrit/tap
brew install ght
```

Homebrew handles placing the binary on your `PATH`. No further steps are required.

### B. go install

If you have Go 1.24+ installed, this one-liner fetches, compiles, and installs the latest release into `$GOPATH/bin` (or `$HOME/go/bin` if `GOPATH` is not set):

```bash
go install github.com/whrit/github-tui/cmd/ght@latest
```

Ensure `$GOPATH/bin` is on your `PATH`:

```bash
export PATH="$HOME/go/bin:$PATH"
```

Add that line to your shell's rc file (`.bashrc`, `.zshrc`, etc.) to make it permanent.

### C. Download a Pre-built Binary

Pre-built archives are published to the GitHub Releases page for every tagged release:

[https://github.com/whrit/github-tui/releases](https://github.com/whrit/github-tui/releases)

Builds are provided for the following platform/architecture combinations:

| Archive name | Platform |
|---|---|
| `ght_darwin_arm64.tar.gz` | macOS (Apple Silicon) |
| `ght_darwin_amd64.tar.gz` | macOS (Intel) |
| `ght_linux_arm64.tar.gz` | Linux (ARM64) |
| `ght_linux_amd64.tar.gz` | Linux (x86-64) |

Each release also ships a `checksums.txt` file containing SHA-256 hashes for all archives. Verify integrity before running the binary.

**Download and install example (macOS Apple Silicon):**

```bash
# Download the archive (replace vX.Y.Z with the actual release tag)
curl -LO https://github.com/whrit/github-tui/releases/download/vX.Y.Z/ght_darwin_arm64.tar.gz

# (Optional but recommended) Verify the checksum
curl -LO https://github.com/whrit/github-tui/releases/download/vX.Y.Z/checksums.txt
shasum -a 256 --check --ignore-missing checksums.txt

# Extract
tar -xzf ght_darwin_arm64.tar.gz

# Move to a directory on your PATH
mv ght /usr/local/bin/
```

**For Linux (x86-64):**

```bash
curl -LO https://github.com/whrit/github-tui/releases/download/vX.Y.Z/ght_linux_amd64.tar.gz
tar -xzf ght_linux_amd64.tar.gz
mv ght /usr/local/bin/
```

> **Note:** If `/usr/local/bin` requires elevated permissions, use `sudo mv ght /usr/local/bin/` or place the binary in any directory that is already on your `PATH` (e.g., `$HOME/.local/bin`).

### D. Build from Source

Building from source requires Go 1.24+ and `git`.

```bash
git clone https://github.com/whrit/github-tui
cd github-tui
go build -o ght ./cmd/ght
mv ght /usr/local/bin/
```

The build produces a statically linked binary (`CGO_ENABLED=0` is set in `.goreleaser.yaml`), so there are no shared-library dependencies.

To install directly into your Go bin directory instead:

```bash
go install ./cmd/ght
```

---

## 3. Configuration

### First-run Interactive Wizard

On the very first launch, `ght` detects that no config file exists and runs an interactive setup prompt:

```
Welcome to ght! No config file found.
You need a GitHub Personal Access Token (classic) with repo, workflow, and read:org scopes.
Create one at: https://github.com/settings/tokens

Enter your GitHub token:
```

Paste your PAT and press `Enter`. `ght` writes the config file automatically with permissions set to `0600` (owner read/write only), then proceeds to start the UI. You do not need to restart.

The wizard will re-prompt if you submit an empty value.

### Manual Configuration

If you prefer to create or manage the config file yourself — or if you are deploying `ght` in a scripted/headless context — write the file before first launch.

The config file location is determined by Go's `os.UserConfigDir()`, which follows platform conventions:

| Platform | Config file path |
|---|---|
| macOS | `~/Library/Application Support/ght/config.yaml` |
| Linux / Unix | `~/.config/ght/config.yaml` (respects `$XDG_CONFIG_HOME`) |
| Windows | `%AppData%\ght\config.yaml` |

Create the directory and file:

```bash
# macOS
mkdir -p "$HOME/Library/Application Support/ght"
cat > "$HOME/Library/Application Support/ght/config.yaml" <<EOF
github:
  token: ghp_your_personal_access_token_here
EOF
chmod 600 "$HOME/Library/Application Support/ght/config.yaml"
```

```bash
# Linux
mkdir -p "${XDG_CONFIG_HOME:-$HOME/.config}/ght"
cat > "${XDG_CONFIG_HOME:-$HOME/.config}/ght/config.yaml" <<EOF
github:
  token: ghp_your_personal_access_token_here
EOF
chmod 600 "${XDG_CONFIG_HOME:-$HOME/.config}/ght/config.yaml"
```

### Config File Reference

The config file format is YAML. Only the `github.token` field is required.

```yaml
github:
  token: ghp_your_personal_access_token_here
```

| Field | Type | Required | Description |
|---|---|---|---|
| `github.token` | string | Yes | GitHub Personal Access Token. Classic PATs and fine-grained PATs are both accepted. |

> **Note:** `ght` validates the token at every startup by making a request to `https://api.github.com/user` and inspecting the `X-OAuth-Scopes` response header. If the token is invalid or missing required scopes, the application exits immediately with a descriptive error message. Fine-grained PATs do not expose scopes via this header, so scope validation is skipped for them — the app logs a warning and proceeds.

---

## 4. Running ght

```bash
# Auto-detect the repository from the current directory's git remote
ght

# Target a specific repository explicitly
ght owner/repo
```

**Auto-detection** works by running `git remote get-url origin` in the current directory and parsing the result. It supports SSH (`git@github.com:owner/repo.git`), HTTPS (`https://github.com/owner/repo`), and SCP-style remotes. If the command is not run from within a git repository and no explicit argument is supplied, `ght` exits with an error.

**Explicit target** accepts the `owner/repo` format. For example:

```bash
ght torvalds/linux
ght kubernetes/kubernetes
```

> **Note:** `ght` requires a working network connection to reach the GitHub API (`api.github.com`). Corporate proxies and VPNs that intercept TLS may cause token validation to fail. Set the standard `HTTPS_PROXY` environment variable if your environment routes traffic through a proxy.

---

## 5. Verifying Your Setup

After installation and configuration, run:

```bash
ght owner/repo
```

`ght` performs the following checks automatically before launching the UI:

1. **Config loading** — Reads `config.yaml`. Exits with a fatal error if the file exists but contains an empty token.
2. **Token validation** — Makes an authenticated `GET https://api.github.com/user` request. Exits with `Token validation failed: HTTP <status>` if the token is invalid or rejected.
3. **Scope check** — For classic PATs, verifies that `repo` and (`project` or `read:org`) scopes are present. Exits with a list of missing scopes and a link to [https://github.com/settings/tokens](https://github.com/settings/tokens) if scopes are insufficient.

A successful startup renders the TUI immediately — no output is printed to stdout on a clean run.

**Debug log:**

All log output (including startup errors in verbose form) is written to the debug log file in the same directory as the config file:

| Platform | Debug log path |
|---|---|
| macOS | `~/Library/Application Support/ght/debug.log` |
| Linux | `~/.config/ght/debug.log` |
| Windows | `%AppData%\ght\debug.log` |

If the UI does not start and the error message is not detailed enough, inspect this file:

```bash
# macOS
cat "$HOME/Library/Application Support/ght/debug.log"

# Linux
cat "${XDG_CONFIG_HOME:-$HOME/.config}/ght/debug.log"
```

---

## 6. Upgrading

### Homebrew

```bash
brew upgrade ght
```

### go install

Re-run the install command to fetch and compile the latest release:

```bash
go install github.com/whrit/github-tui/cmd/ght@latest
```

### Pre-built binary

Download the new archive from [https://github.com/whrit/github-tui/releases](https://github.com/whrit/github-tui/releases), verify the checksum, and replace the existing binary:

```bash
tar -xzf ght_darwin_arm64.tar.gz
mv ght /usr/local/bin/ght
```

### Build from source

Pull the latest changes and rebuild:

```bash
cd github-tui
git pull
go build -o ght ./cmd/ght
mv ght /usr/local/bin/
```

Your config file is not affected by upgrades — it persists across all installation methods.

---

## 7. Uninstalling

### Homebrew

```bash
brew uninstall ght
brew untap whrit/tap
```

### go install / manual binary

```bash
rm "$(which ght)"
```

### Config and log files (all methods)

Removing the binary leaves the config directory and debug log behind. Delete them if you want a clean removal:

```bash
# macOS
rm -rf "$HOME/Library/Application Support/ght"

# Linux
rm -rf "${XDG_CONFIG_HOME:-$HOME/.config}/ght"

# Windows (PowerShell)
Remove-Item -Recurse -Force "$env:APPDATA\ght"
```

> **Note:** The config directory contains `config.yaml` (which holds your GitHub token) and `debug.log`. There is no other application state — `ght` does not write a local database or cache.

---

## 8. Troubleshooting

### "invalid repo" on startup

`ght` could not determine the repository from your git remote. Either:

- You are not in a git repository — run `ght owner/repo` explicitly.
- The repository has no `origin` remote — run `git remote -v` to inspect.
- The remote URL format is unusual — file an issue with the output of `git remote get-url origin`.

### "Token validation failed: HTTP 401"

Your token is invalid, has been revoked, or was pasted with leading/trailing whitespace. Regenerate a token at [https://github.com/settings/tokens](https://github.com/settings/tokens) and update `config.yaml`.

### "missing required token scopes: repo"

Your classic PAT was created without the `repo` scope. Edit the token at [https://github.com/settings/tokens](https://github.com/settings/tokens) or create a new one with the required scopes listed in [Prerequisites](#1-prerequisites).

### Fine-grained PAT — permission errors at runtime

Fine-grained PATs do not expose scopes via the `X-OAuth-Scopes` header, so `ght` cannot validate them at startup. Instead it logs a warning and proceeds. If you see API permission errors inside the UI, verify that your fine-grained token has been granted:

- **Repository permissions:** Contents (read), Issues (read/write), Actions (read)
- **Organization permissions:** Members (read)

### The TUI renders incorrectly / garbled characters

`ght` uses `tcell`-backed rendering via `tview`. Ensure:

- Your terminal emulator supports at least 256 colors.
- `$TERM` is set to a known value (`xterm-256color`, `screen-256color`, etc.).
- Your terminal font includes the Unicode box-drawing characters used for borders.

### "cannot deserialize config file"

The YAML in your `config.yaml` is malformed. Verify the file looks exactly like:

```yaml
github:
  token: ghp_your_token_here
```

Common mistakes: tabs instead of spaces for indentation, missing `github:` top-level key, or accidental line breaks inside the token value.
