# Dotfiles — DevOps Workstation

Personal dotfiles managed with [GNU Stow](https://www.gnu.org/software/stow/).
Includes an interactive dependency installer and 40+ navi cheatsheets.

## What's inside

| Package | Contents | Target |
|---------|----------|--------|
| `config/opencode/` | OpenCode AI agent config — 13 agents, 11 commands, 10 skills | `~/.config/opencode/` |
| `config/fish/` | Fish shell config — aliases, functions, plugins for DevOps | `~/.config/fish/` |
| `config/starship/` | Starship prompt — K8s, Terraform, Docker context | `~/.config/starship.toml` |
| `config/nvim/` | Neovim + NvChad config | `~/.config/nvim/` |
| `config/navi/` | Interactive cheatsheets (40+ tools) | `~/.config/navi/` |
| `config/lazygit/` | Git TUI — config with Nord theme | `~/.config/lazygit/` |
| `config/lazydocker/` | Docker TUI — config with Nord palette | `~/.config/lazydocker/` |
| `config/k9s/` | Kubernetes TUI — config, hotkeys, Nord skin | `~/.config/k9s/` |
| `config/btop/` | System monitor — config + Nord theme | `~/.config/btop/` |
| `config/bat/` | Better cat — Nord syntax theme | `~/.config/bat/` |
| `config/glow/` | Markdown renderer | `~/.config/glow/` |
| `config/atuin/` | Shell history with sync | `~/.config/atuin/` |
| `config/direnv/` | Per-directory environment | `~/.config/direnv/` |
| `config/dive/` | Docker image layer explorer | `~/.config/dive/` |

## Quick Start

### 1. Install the build prerequisites

The canonical installer is a Go program, so install Git, Make, and Go with
your distribution's package manager if they are not already available:

```bash
# Ubuntu/Debian
sudo apt install git make golang-go

# Fedora
sudo dnf install git make golang

# Arch
sudo pacman -S git make go
```

### 2. Clone the repository

```bash
git clone https://github.com/MattyOstrowsky/Dotfiles.git ~/Dotfiles
cd ~/Dotfiles
```

### 3. Install tools and shared configs

```bash
make install
```

`make install` builds `installer/dotfiles-install` and starts its interactive
installer. It detects Ubuntu/Debian, Fedora, or Arch and uses `apt`, `dnf`, or
`pacman` respectively. The installer presents a preflight plan, lets you
choose tools and applicable shared config packages, applies the selected
configs with GNU Stow, and validates the result.

The built binary can also be invoked directly from the repository root. The
following read-only and non-interactive modes are available (run `make
install` first, or build the binary from `installer/`):

```bash
./installer/dotfiles-install --list
./installer/dotfiles-install --preflight
./installer/dotfiles-install --validate
./installer/dotfiles-install --yes
./installer/dotfiles-install --yes --theme nord
```

- `--list` prints detected system, tool status, and config status without
  changing anything.
- `--preflight` refreshes package metadata and prints the installation plan;
  it does not install tools or modify config links.
- `--validate` checks installed tools and expected Stow links without changing
  anything; it exits non-zero when validation fails.
- `--yes` installs all manifest tools and applicable config packages without
  the interactive UI, then applies the default Catppuccin theme and validates.
- `--theme catppuccin|nord` applies a theme. Combined with `--yes`, it applies
  that theme as part of the installation; without `--yes`, it only updates
  the theme files.

### Install or remove one config package manually

Shared Stow packages live below `config/`, not at the repository root:

```bash
cd ~/Dotfiles
stow --dir config --target "$HOME" lazygit
stow --dir config --target "$HOME" --delete lazygit
```

Use `installer/dotfiles-install --preflight` to inspect the planned changes
before installation.

## How GNU Stow works

Stow mirrors each package's directory structure into the target home. The
shared packages are under `config/`:

```
~/Dotfiles/
├── config/
│   ├── opencode/         → ~/.config/opencode/
│   ├── fish/             → ~/.config/fish/
│   ├── starship/         → ~/.config/starship.toml
│   ├── nvim/             → ~/.config/nvim/
│   ├── navi/             → ~/.config/navi/
│   ├── lazygit/          → ~/.config/lazygit/
│   ├── lazydocker/       → ~/.config/lazydocker/
│   ├── k9s/              → ~/.config/k9s/
│   ├── btop/             → ~/.config/btop/
│   ├── bat/              → ~/.config/bat/
│   ├── glow/             → ~/.config/glow/
│   ├── atuin/            → ~/.config/atuin/
│   ├── direnv/           → ~/.config/direnv/
│   ├── dive/             → ~/.config/dive/
│   └── omp/              → ~/.omp/
└── installer/            → canonical Go installer
```

`config/` is the shared, distro-agnostic configuration tree. The
`windows-terminal/` package is host-scoped for the Windows side of a WSL
setup and is not stowed into the Linux home by the installer. Desktop-specific
material is intentionally separate: `arch/` and `fedora/` are profile and
migration areas, not part of the shared config install. The legacy
`arch/install-deps.sh` is deprecated while this migration is completed; use
`make install` and `installer/dotfiles-install` instead.

## Interactive Dependency Installer

`installer/dotfiles-install` is the canonical Go installer. Its tool manifest
is `installer/tools.yaml`; its Stow packages are scanned from `config/`.
Tools are grouped as follows:

| Category | Tools |
|----------|-------|
| **Core** | stow, git, gh, curl, wget, make, unzip, fzf, tree, htop |
| **Runtime** | python3, pip3, go, cargo, node |
| **Shell** | fish, starship, atuin, zoxide |
| **CLI** | ripgrep, fd-find, bat, btop, direnv, glow, navi, tldr, lazygit |
| **Kubernetes** | kubectl, helm, kubectx, k9s |
| **IaC** | terraform, ansible |
| **Containers** | dive, lazydocker |
| **Editors** | nvim, opencode, omp (config-only) |

Features:
- Auto-detects Ubuntu/Debian (`apt`), Fedora/RHEL-family (`dnf`), and Arch
  (`pacman`); Arch installs use official repositories only.
- Interactive TUI by default, or `--yes` for a complete non-interactive run.
- `--list`, `--preflight`, and `--validate` provide status, planning, and
  post-install checks without silently changing config links.
- Automatically skips tools that are already installed where possible and
  stows applicable packages from `config/`.
- `--theme catppuccin|nord` updates the supported tool themes.
- Downloaded and script-installed binaries are placed in `~/.local/bin/`.

The old Arch-only `arch/install-deps.sh` is retained only as a deprecated
reference for legacy/profile-specific tools outside the core manifest; it is
not the supported installer entrypoint.

## Navi Cheatsheets

40+ interactive cheatsheets for daily tools. Launch with `navi` or query directly:

```bash
navi --query "helm install"
navi --query "lazygit keybindings"
```

See `config/navi/.config/navi/cheats/` for the full list.

## OpenCode Agents

### Primary Agents (switch with Tab)
- **daily** — Personal companion, planning, delegation (po polsku)
- **architect** — Architecture planning, ADRs, red-teaming (read-only)
- **orchestrator** — Complex task breakdown, execution plans (read-only)
- **devops** — Infrastructure implementation, Docker, K8s, CI/CD
- **meta** — Agent ecosystem management

### Subagents (invoke with @name)
`@terraform` `@ansible` `@backend` `@frontend` `@data-engineer` `@security` `@cicd` `@python-dev` `@explore`

### Commands
`tf-plan` `tf-apply` `docker-build` `k8s-check` `sec-audit` `pipeline-lint` `infra-review` `cost-estimate` `self-improve` `stats` `context-check`

## Fish Shell

### Key Aliases

| Alias | Command | Category |
|-------|---------|----------|
| `k` | `kubectl` | K8s |
| `kgp` | `kubectl get pods` | K8s |
| `k9` | `k9s` | K8s |
| `ksl` | `stern` | K8s |
| `kctx` | `kubectx` | K8s |
| `kns` | `kubens` | K8s |
| `tf` | `terraform` | Terraform |
| `tfp` | `terraform plan` | Terraform |
| `d` | `docker` | Docker |
| `dc` | `docker compose` | Docker |
| `lzg` | `lazygit` | TUI |
| `lzd` | `lazydocker` | TUI |
| `lzs` | `lazysql` | TUI |
| `top` | `btop` | TUI |
| `md` | `glow` | TUI |
| `ddive` | `dive` | TUI |
| `dtop` | `ctop` | TUI |
| `gs` | `git status` | Git |
| `gcp` | `git add -A && commit && push` | Git |
