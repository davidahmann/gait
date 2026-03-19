---
title: "Install Gait"
description: "Step-by-step installation for Linux, macOS, and Windows with checksum verification and PATH setup."
---

# Install Gait

Gait's default install path is the release installer script.

## Platform Support

- `scripts/install.sh` currently supports Linux and macOS.
- Windows is supported through release assets (manual download + PATH setup), not the installer script.
- Homebrew publishing is tap-first and release-gated; see `docs/homebrew.md`.

## Recommended Path (Binary + Checksum Verification)

```bash
curl -fsSL https://raw.githubusercontent.com/Clyra-AI/gait/main/scripts/install.sh | bash
```

What the script does:

- resolves the latest release tag (or uses `--version`)
- downloads `checksums.txt` and your platform archive
- verifies SHA-256 checksum before install
- installs `gait` to `~/.local/bin` by default

## Alternate Path: Homebrew (Tap)

```bash
brew tap Clyra-AI/tap
brew install gait
```

Validate install:

```bash
brew test Clyra-AI/tap/gait
gait demo --json
```

## PATH Setup (Common First-Run Fix)

If `gait` is not found after install, add `~/.local/bin` to your shell PATH:

For `zsh`:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

For `bash`:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

For `fish`:

```fish
set -U fish_user_paths $HOME/.local/bin $fish_user_paths
```

Options:

```bash
bash scripts/install.sh --version vX.Y.Z --install-dir ~/.local/bin
```

## Verify Installation

```bash
gait version --json
gait doctor --json
gait init --json
gait check --json
gait demo
gait verify run_demo --json
```

## Three Distinct Checkpoints

Treat these as different promises, not one blended onboarding claim:

1. Fast proof: `gait version --json`, `gait doctor --json`, `gait demo`, `gait verify run_demo --json`
2. Strict inline enforcement: wire `gait gate eval` or a Gait boundary wrapper before the real tool call executes
3. Hardened `oss-prod` readiness: seed `examples/config/oss_prod_template.yaml`, run `gait check --json`, then require `gait doctor --production-readiness --json`

Managed or preloaded runtimes without an interception seam can still use the proof, capture, verify, and regress paths. They should not claim strict inline fail-closed blocking until the execution boundary is under user control.

## Dev vs Prod Mode (Important)

Use development mode for first-run validation:

```bash
gait demo
gait verify run_demo --json
```

This is the install proof path. It validates the binary, evidence, and local artifact loop, but it is not the same thing as runtime enforcement.

To move from proof to enforcement, place Gait at the real tool boundary:

```bash
gait gate eval --policy .gait.yaml --intent ./intent.json --json
```

That boundary call must happen immediately before the side effect you want to control.

Before production use, apply hardened defaults and validate readiness:

```bash
gait init --json
# From a repo checkout:
cp examples/config/oss_prod_template.yaml .gait/config.yaml

# Or, if you installed only the binary:
curl -fsSL https://raw.githubusercontent.com/Clyra-AI/gait/main/examples/config/oss_prod_template.yaml -o .gait/config.yaml

gait check --json
gait doctor --production-readiness --json
```

Use `examples/config/oss_prod_template.yaml` as the canonical hardened starting point, whether you copy it from a repo checkout or fetch that same file after a binary-only install. Then review the environment variable names, listener, and retention values for your deployment before enforcing. High-risk runtime boundaries are not production-ready until `gait doctor --production-readiness --json` reports `ok=true`.

`gait doctor --json` is truthful for binary-only installs: in a clean writable directory it returns the installed-binary lane with `status=pass|warn`, reports the checked executable via additive `binary_path`, `binary_path_source`, and `binary_version` fields, and only sets `path_binary_path` when a different PATH-resolved `gait` binary is also present. Repo-only schema/example checks stay scoped to a Gait repo checkout.

If PATH is still not updated, run directly once:

```bash
~/.local/bin/gait doctor --json
```

## One-Command Quickstart

After installation:

```bash
curl -fsSL https://raw.githubusercontent.com/Clyra-AI/gait/main/scripts/quickstart.sh | bash
```

## Windows Install (Manual Path)

1. Download the latest Windows release asset from GitHub Releases.
2. Place `gait.exe` in a local tools directory (for example `%USERPROFILE%\\bin`).
3. Add that directory to your user PATH.
4. Open a new shell and run:

```powershell
gait.exe version --json
gait.exe doctor --json
gait.exe demo
gait.exe verify run_demo
```

## Alternate Path: Build From Source

Use this only when developing Gait itself.

```bash
git clone https://github.com/Clyra-AI/gait.git
cd gait
go build -o ./gait ./cmd/gait
```

## Frequently Asked Questions

### Does Gait require Go to be installed?

No. The install script downloads a prebuilt binary. Go is only needed if building from source.

### Does Gait work on Windows?

Yes. Download the Windows binary from the GitHub release and add it to your PATH.

### How do I verify the install worked?

Run `gait version --json`, `gait doctor --json`, `gait init --json`, and `gait check --json`. Then run `gait demo` to create your first signed artifact. Before claiming high-risk production readiness, require `gait doctor --production-readiness --json` to return `ok=true`.

### Can I install via Homebrew?

Yes. See `docs/homebrew.md` for tap-based installation as an alternate path.

### How large is the Gait binary?

The compiled binary is a single static Go executable, typically under 30 MB with zero runtime dependencies.
