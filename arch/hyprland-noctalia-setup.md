# Hyprland + Noctalia + ly — Arch Linux Setup

> **Desktop stack:** Hyprland (WM) + Noctalia (shell) + ly (TUI login)
> Target: HUAWEI MateBook VGHH-XX (C110), Intel Arc GPU

---

## Stack Overview

```
┌──────────────────────────────────┐
│  Noctalia (bar, panel, launcher) │ ← Qt6 shell, config via GUI
├──────────────────────────────────┤
│  Hyprland (tiling compositor)    │ ← config: ~/.config/hypr/hyprland.conf
├──────────────────────────────────┤
│  ly (TUI display manager)        │ ← config: /etc/ly/config.ini
├──────────────────────────────────┤
│  Wayland                         │
└──────────────────────────────────┘
```

- **Hyprland** — tiling compositor with animations, blur, rounded corners
- **Noctalia** — beautiful desktop shell (bar, dock, control center, notifications, launcher, lock screen)
- **ly** — terminal-like login screen (no mouse needed, looks like a hacker movie)

---

## Installation

### 1. Install paru (AUR helper) — if not already

```bash
sudo pacman -S --needed base-devel git
git clone https://aur.archlinux.org/paru-bin.git /tmp/paru-bin
cd /tmp/paru-bin && makepkg -si
```

### 2. Install All Packages

```bash
# Core: Hyprland + Display Manager
sudo pacman -S hyprland ly

# Noctalia (AUR) — shell + Qt support
paru -S noctalia-shell

# Noctalia dependencies
sudo pacman -S brightnessctl imagemagick power-profiles-daemon

# Optional but recommended
sudo pacman -S cliphist wlsunset ddcutil

# Wayland screenshot tools
sudo pacman -S grim slurp wl-clipboard

# Polkit agent (needed for auth dialogs in GUI)
sudo pacman -S polkit-gnome
```

### 3. Enable ly, Disable SDDM/GDM

```bash
# If you have another DM running:
sudo systemctl disable sddm 2>/dev/null || true
sudo systemctl disable gdm 2>/dev/null || true
sudo systemctl enable ly
```

### 4. Deploy Dotfiles (Hyprland + Noctalia + all configs)

```bash
cd ~/Dotfiles
make install
# or manually: stow -d arch --target=$HOME hyprland noctalia fish nvim starship ...
```

### 5. Restore Noctalia Config

Your noctalia config (settings.json, colors.json, plugins) is pre-saved in the dotfiles.

```bash
# Symlink noctalia config via stow (already done if you ran `make install`)
# Or manually copy:
mkdir -p ~/.config/noctalia
cp -r ~/Dotfiles/arch/noctalia/.config/noctalia/* ~/.config/noctalia/
```

### 6. Reboot

```bash
sudo reboot
```

After reboot you'll see the ly login screen. Type your username, then password, and you'll be dropped into Hyprland + Noctalia.

---

## Default Keybinds

### Hyprland (Window Manager)

| Keys | Action |
|------|--------|
| `SUPER + Enter` | Terminal (Ghostty) |
| `SUPER + B` | Firefox |
| `SUPER + N` | Nautilus (files) |
| `SUPER + Q` | Close window |
| `SUPER + F` | Fullscreen |
| `SUPER + V` | Toggle floating |
| `SUPER + H/J/K/L` | Move focus (vim keys) |
| `SUPER + Shift + H/J/K/L` | Move window |
| `SUPER + 1-9` | Switch workspace |
| `SUPER + Shift + 1-9` | Move window to workspace |
| `SUPER + mouse_scroll` | Switch workspace |

### Noctalia (Shell)

| Keys | Action |
|------|--------|
| `SUPER + Space` | Launcher (app search) |
| `SUPER + S` | Control Center |
| `SUPER + ,` | Settings |

All Noctalia keybinds are configurable via the Settings GUI (`SUPER + ,`).

### Screenshots

| Keys | Action |
|------|--------|
| `SUPER + Print` | Full screen → clipboard |
| `SUPER + Shift + S` | Select area → clipboard |

---

## Configuration

### Hyprland Config

**File:** `~/.config/hypr/hyprland.conf` (managed by stow from `~/Dotfiles/arch/hyprland/`)

Key sections:
- **Monitor** — auto-detect, uncomment to override resolution/scale
- **Decoration** — Noctalia-recommended blur, rounding, shadow
- **Input** — Polish keyboard (`pl`), Caps→Escape, natural scroll, tap-to-click
- **Keybinds** — vim-style navigation, workspace switching, Noctalia IPC
- **XWayland** — `force_zero_scaling = true` for Steam/Flatpak games on scaled displays

### Noctalia Config

**Location:** `~/.config/noctalia/` (stowed from dotfiles)

Most settings are configured via the **built-in GUI**:
- `SUPER + ,` → opens Settings
- Right-click bar → Control Center → Settings gear icon
- Changes take effect immediately, saved automatically

Included plugins (pre-installed via dotfiles):
- **keybind-cheatsheet** — overlay showing all keybinds
- **privacy-indicator** — shows when mic/camera/screen-sharing is active
- **screen-recorder** — built-in screen recording
- **polkit-agent** — authentication dialogs
- **catwalk** — workspace switcher widget
- **battery-monitor-plus** — enhanced battery indicator
- **fancy-audiovisualizer** — desktop audio visualizer
- **hot-corners** — trigger actions by moving cursor to corners
- **sys-info-widget** — system resource monitor

Color scheme: **Catppuccin Mocha** (set via Noctalia Settings GUI).

## Troubleshooting

### Noctalia doesn't start after login

```bash
# Check if noctalia-shell is available
paru -Q noctalia-shell

# Run manually for debugging
qs -c noctalia-shell

# Check Hyprland logs
cat /tmp/hypr/$(ls /tmp/hypr/ | head -1)/hyprland.log | tail -50
```

### Black screen after login

```bash
# Switch to TTY (Ctrl+Alt+F3), login, check:
journalctl -u ly --no-pager -n 30
# Check if Hyprland can start manually:
Hyprland  # from TTY
```

### ly keyboard layout is wrong

ly uses the system locale. Check:

```bash
localectl status
# If needed: sudo localectl set-x11-keymap pl
```

### No blur effects

Make sure Hyprland config has `blur { enabled = true }` in the decoration block.
Also check: `hyprctl getoption decoration:blur:enabled`

### Missing app icons in Noctalia

```bash
sudo pacman -S papirus-icon-theme
# Then set in Noctalia Settings → Appearance
```

### Noctalia from AUR fails to build

```bash
# Clean build cache and retry
paru -S --cleanbuild noctalia-shell

# Check build deps
sudo pacman -S qt6-base qt6-declarative qt6-svg qt6-wayland cmake
```

---

## References

- [Noctalia Docs](https://docs.noctalia.dev/v4)
- [noctalia-shell AUR](https://aur.archlinux.org/packages/noctalia-shell)
- [Hyprland Wiki](https://wiki.hyprland.org)
- [ly GitHub](https://github.com/fairyglade/ly)
- [hyprgrass GitHub](https://github.com/horriblename/hyprgrass)
- [Noctalia Discord](https://discord.noctalia.dev)
