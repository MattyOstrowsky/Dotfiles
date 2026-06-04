# Hyprland + Noctalia + ly — Fedora 44 Setup

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

GNOME stays installed — you can switch between GNOME and Hyprland at the login screen.

---

## Installation

### 1. Terra Repository (for Noctalia)

```bash
sudo dnf install --nogpgcheck \
  --repofrompath 'terra,https://repos.fyralabs.com/terra$releasever' \
  terra-release
```

### 2. Install All Packages

```bash
# Core: Hyprland + Noctalia
sudo dnf install hyprland noctalia-shell

# Noctalia dependencies
sudo dnf install \
  brightnessctl imagemagick \
  power-profiles-daemon

# Optional but recommended
sudo dnf install \
  cliphist wlsunset \
  ddcutil

# Wayland screenshot tools
sudo dnf install grim slurp wl-clipboard

# Polkit agent (needed for auth dialogs in GUI)
sudo dnf install polkit-gnome

# Display manager — TUI login screen
sudo dnf install ly
```

### 3. Enable ly, Disable GDM

```bash
sudo systemctl disable gdm
sudo systemctl enable ly
```

### 4. ly Nord Theme

```bash
# Copy pre-made Nord config + TTY color script
sudo cp ~/Dotfiles/ly/etc/ly/config.ini  /etc/ly/config.ini
sudo cp ~/Dotfiles/ly/etc/ly/startup.sh  /etc/ly/startup.sh
sudo chmod +x /etc/ly/startup.sh
```

### 5. Deploy Dotfiles (Hyprland config)

```bash
cd ~/Dotfiles
stow hyprland
# or: make install  (deploys all packages)
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

**File:** `~/.config/hypr/hyprland.conf` (managed by stow from `~/Dotfiles/hyprland/`)

Key sections:
- **Monitor** — auto-detect, uncomment to override resolution/scale
- **Decoration** — Noctalia-recommended blur, rounding, shadow
- **Input** — Polish keyboard (`pl`), Caps→Escape, natural scroll, tap-to-click
- **Keybinds** — vim-style navigation, workspace switching, Noctalia IPC

### Noctalia Config

**Location:** `~/.config/noctalia/` (created automatically on first run)

Most settings are configured via the **built-in GUI**:
- `SUPER + ,` → opens Settings
- Right-click bar → Control Center → Settings gear icon
- Changes take effect immediately, saved automatically

Key settings to adjust after first boot:
- Theme & color scheme
- Bar position and widgets
- Panel layout
- Notifications behavior

### ly Config — Nord Theme

ly config lives at `/etc/ly/config.ini` (system-level, cannot be stowed directly).

A **pre-made Nord-themed config** is included in the dotfiles:

```bash
# Copy Nord config + TTY color startup script
sudo cp ~/Dotfiles/ly/etc/ly/config.ini  /etc/ly/config.ini
sudo cp ~/Dotfiles/ly/etc/ly/startup.sh  /etc/ly/startup.sh
sudo chmod +x /etc/ly/startup.sh

# Apply immediately (kicks you to login screen)
sudo systemctl restart ly
```

**What the Nord config does:**

| Setting | Value | Effect |
|---------|-------|--------|
| `bg` | `#2E3440` (nord0) | Deep dark blue-gray background |
| `fg` | `#D8DEE9` (nord4) | Frosty light text |
| `border_fg` | `#5E81AC` (nord10) | Cool blue border |
| `error_fg` | `#BF616A` (nord11) | Red errors (bold) |
| `animation` | `matrix` + `#88C0D0` (nord8) | Matrix rain in Nord teal |
| `asterisk` | `•` (bullet) | Password mask |
| `clock` | `%a %d %b  %H:%M` | Mon 04 Jun  13:37 |
| `hide_version_string` | `true` | Clean top-left corner |
| `text_in_center` | `true` | Centered input labels |
| `startup.sh` | Nord TTY palette | Console colors match theme |

**To tweak:** edit `/etc/ly/config.ini` and run `ly --validate-config /etc/ly/config.ini` to check for errors, then `sudo systemctl restart ly`.

---

## Back to GNOME (if needed)

Both are installed side by side. At the ly login screen:
- Press `F2` to cycle through available desktop sessions
- Select "GNOME" instead of "Hyprland"
- Login normally

Or switch back permanently:
```bash
sudo systemctl disable ly
sudo systemctl enable gdm
sudo reboot
```

---

## Troubleshooting

### Noctalia doesn't start after login
```bash
# Check if noctalia-shell is installed
qs -c noctalia-shell  # run manually

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
sudo dnf install papirus-icon-theme
# Then set in Noctalia Settings → Appearance
```

---

## References

- [Noctalia Docs](https://docs.noctalia.dev/v4)
- [Hyprland Wiki](https://wiki.hyprland.org)
- [ly GitHub](https://github.com/fairyglade/ly)
- [Noctalia Discord](https://discord.noctalia.dev)
