# Fedora 44 Setup — Huawei MateBook VGHH-XX (C110)

> Last updated: 2026-06-04

---

## System Info

- **Model:** HUAWEI MateBook VGHH-XX (SKU: C110)
- **CPU:** Intel Meteor Lake (Arc Graphics)
- **Kernel:** 6.19.10 / 7.0.10
- **DE:** GNOME 50 (Wayland)

---

## Installation Order

Everything is split into two layers:

| Layer | Tool | What it does |
|-------|------|-------------|
| **1. Shared core CLI** | `make install` / `installer/dotfiles-install` | Installs the shared, distro-agnostic CLI tools and applicable packages from `config/` |
| **2. Fedora/profile (this file)** | Manual commands | Docker, desktop apps, Fedora repositories, fingerprint setup, and other profile-specific steps |

**Run the shared installer first, then the manual Fedora/profile steps.**

## Layer 1 — Shared core CLI (`make install`)

The canonical Go installer handles the shared core CLI and configuration tree. It
builds and starts `installer/dotfiles-install`; applicable Stow packages come
from `config/`. Fedora desktop and profile-specific steps remain manual in
Layer 2 below.

The shared core inventory includes:

```
System:    stow, git, curl, wget, make, unzip, fish, fzf, bat, ripgrep,
           fd-find, tree, htop, direnv, btop, xclip

Runtimes:  python3, pip3, Node.js (fnm), Go (~/.local/go), Rust (rustup), Bun

Shell:     starship, zoxide, atuin, navi, fisher (fish plugin manager)

Editors:   neovim (AppImage), opencode

K8s:       kubectl, helm, kubectx, k9s, stern, kustomize, kubeconform

IaC:       terraform, terragrunt, ansible, infracost

TUI:       lazygit, lazydocker, lazysql, dive, ctop, glow

Security:  trivy, grype, checkov, tldr
```

Usage:
```bash
cd ~/Dotfiles

# Interactive — choose tools and config packages in the installer
make install

# Non-interactive — install all applicable shared items
./installer/dotfiles-install --yes

# Read-only status, plan, and validation
./installer/dotfiles-install --list
./installer/dotfiles-install --preflight
./installer/dotfiles-install --validate
```

## Layer 2 — Manual Fedora/Desktop/Profile Steps

These Fedora-specific and desktop/profile steps are intentionally manual; the
shared core installer does not replace them.


---

### 1. RPM Fusion (required for some packages)

```bash
sudo dnf install \
  https://download1.rpmfusion.org/nonfree/fedora/rpmfusion-nonfree-release-44.noarch.rpm
```

---

### 2. Docker

```bash
sudo dnf install docker-ce docker-ce-cli containerd.io
```

Adds: `docker-buildx-plugin`, `docker-compose-plugin`, `libcgroup`

Post-install:
```bash
sudo systemctl enable --now docker
sudo usermod -aG docker $USER
# Re-login required
```

---

### 3. GitHub CLI

```bash
sudo dnf install gh
```

Repo: `gh-cli` (added automatically)

---

### 4. Ghostty Terminal

```bash
sudo dnf copr enable scottames/ghostty
sudo dnf install ghostty
```

Adds dep: `gtk4-layer-shell`

---

### 5. OpenCode MCP — Playwright (browser automation)

No extra deps needed — Playwright bundles its own Chromium via npx.

---

### 6. Flatpak Apps (all system-wide)

```bash
flatpak install flathub com.google.Chrome
flatpak install flathub com.spotify.Client
flatpak install flathub com.valvesoftware.Steam
flatpak install flathub dev.zed.Zed
```

---

### 7. Fingerprint Reader — Goodix GXFP5130

#### Status: ❌ NOT SUPPORTED (no Linux driver exists)

**Sensor:** Goodix GXFP5130 (ACPI HID: `GXFP5130`)
- Detected via ACPI at `\_SB_.SPBA` (SPI bus)
- Registered as platform device: `/sys/devices/platform/GXFP5130:00`
- ACPI status: **15** (present + enabled + functional) ✅
- Physical location: SPI bus (not USB, not I2C)

**Why it doesn't work:**

| Layer | Status | Detail |
|-------|--------|--------|
| **Kernel driver** | ❌ Missing | `hid-goodix-spi.ko` only supports `GXTS7986` (touchscreen), not `GXFP5130` (fingerprint). No driver binds to this device. |
| **libfprint** | ❌ No match | `goodixmoc` driver in libfprint 1.94.10 is for **USB** Goodix sensors (e.g. 27c6:XXXX), not SPI platform devices. |
| **Upstream** | ❌ None | Zero results on kernel LKML and libfprint GitLab for "GXFP5130". No driver in development. |

**Baseline packages (already installed — just documenting):**
```bash
sudo dnf install fprintd fprintd-pam libfprint
```

**What to do:**
1. File issue at https://gitlab.freedesktop.org/libfprint/libfprint/-/issues
   - Include: `acpi:GXFP5130`, ACPI path `\_SB_.SPBA`, laptop model `HUAWEI VGHH-XX`
2. Check back periodically:
   ```bash
   dnf check-update libfprint fprintd
   ```
3. Once supported:
   ```bash
   fprintd-enroll
   sudo authselect enable-feature with-fingerprint
   ```

**Reference data for developers:**
```
ACPI modalias: acpi:GXFP5130:GXFP5130:
Physical node: /sys/devices/platform/GXFP5130:00
Waiting for supplier: 0 (no missing deps)
```

---

## Full Reinstall — Quick Reference

```bash
# 1. Enable RPM Fusion
sudo dnf install rpmfusion-nonfree-release-44

# 2. Install shared core CLI tools and configs (interactive)
cd ~/Dotfiles && make install

# 3. Docker
sudo dnf install -y docker-ce docker-ce-cli containerd.io
sudo systemctl enable --now docker
sudo usermod -aG docker $USER

# 4. GitHub CLI
sudo dnf install -y gh

# 5. Ghostty
sudo dnf copr enable scottames/ghostty -y
sudo dnf install -y ghostty

# 6. Flatpaks
flatpak install -y flathub \
  com.google.Chrome \
  com.spotify.Client \
  com.valvesoftware.Steam \
  dev.zed.Zed

# 7. Fingerprint (baseline — waiting for driver)
sudo dnf install -y fprintd fprintd-pam libfprint

# 9. Deploy dotfiles
cd ~/Dotfiles && make install

# 9. Reboot (for docker group, etc.)
sudo reboot
```

---

## Currently Enabled Repos

```
copr:copr.fedorainfracloud.org:phracek:PyCharm
copr:copr.fedorainfracloud.org:scottames:ghostty
docker-ce-stable
fedora
fedora-cisco-openh264
gh-cli
google-chrome
rpmfusion-nonfree-nvidia-driver
rpmfusion-nonfree-steam
updates
```
