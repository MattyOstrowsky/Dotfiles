---
name: linux-admin
description: Linux system administration for DevOps — package management, shell configuration, dotfiles management, systemd, performance tuning
---

## Package Management

### APT (Debian/Ubuntu)
```bash
# System update
sudo apt update && sudo apt upgrade -y

# Search and install
apt search <package>
sudo apt install -y <package>

# Remove unused deps
sudo apt autoremove --purge

# List installed, not from base
apt list --manual-installed | grep -v "base"

# Hold package at specific version
sudo apt-mark hold <package>
sudo apt-mark unhold <package>
```

### Managing Third-Party Repos
```bash
# Add repository key and source
curl -fsSL https://example.com/key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/example.gpg
echo "deb [signed-by=/usr/share/keyrings/example.gpg] https://example.com/apt stable main" | sudo tee /etc/apt/sources.list.d/example.list
sudo apt update
```

## Dotfiles Management

### GNU Stow Workflow
```bash
# Structure dotfiles as stow packages
~/dotfiles/
├── fish/
│   └── .config/
│       └── fish/
│           └── config.fish
├── nvim/
│   └── .config/
│       └── nvim/
│           └── init.lua
├── starship/
│   └── .config/
│       └── starship.toml

# Install all packages
stow -v --target=$HOME fish nvim starship

# Uninstall
stow -v -D --target=$HOME fish

# Dry run
stow -n -v --target=$HOME fish
```

### Dotfiles Hygiene
- Never store secrets (SSH keys, API tokens) in dotfiles repo
- Use `.gitignore` to exclude `.env`, `*.local.*`, `pi/` (secrets dir)
- Keep a `dependencies.md` listing required system packages
- Use `Makefile` or bootstrap script for first-time setup
- Separate per-machine config from shared config

## Shell Configuration

### Fish Shell
```fish
# Config: ~/.config/fish/config.fish
# Aliases
alias ll "ls -la"
alias gs "git status"

# Environment variables (not in config.fish — use exports.fish)
# ~/.config/fish/conf.d/exports.fish
set -gx EDITOR nvim
set -gx PATH $HOME/.local/bin $PATH

# Abbreviations (expand on space)
abbr -a gc "git commit -m"
abbr -a gst "git status"
abbr -a gp "git push"
```

### Bash/Zsh
```bash
# ~/.bashrc or ~/.zshrc
# Never add secrets or machine-specific paths here
# Use .local versions for overrides: ~/.bashrc.local, ~/.zshrc.local
export EDITOR=nvim
export PATH="$HOME/.local/bin:$PATH"

# Source local overrides if they exist
[ -f "$HOME/.bashrc.local" ] && source "$HOME/.bashrc.local"
```

## Systemd

### Service File Template
```ini
[Unit]
Description=My Service
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=appuser
Group=appuser
WorkingDirectory=/opt/app
ExecStart=/usr/local/bin/app --config /etc/app/config.yml
Restart=always
RestartSec=5
LimitNOFILE=65536
Environment=NODE_ENV=production
EnvironmentFile=/etc/app/env

[Install]
WantedBy=multi-user.target
```

### Common Operations
```bash
# Reload after modifying service file
sudo systemctl daemon-reload

# Enable/start
sudo systemctl enable --now my-service

# View logs
sudo journalctl -u my-service -f

# Restart
sudo systemctl restart my-service
```

## Performance Tuning

### Quick Checks
```bash
# Resource bottlenecks
htop                           # Real-time CPU/memory
iotop                          # I/O per process
nethogs                        # Network per process

# Disk
df -h                          # Disk usage
du -sh * --exclude=proc        # Directory sizes
iostat -x 1                    # Disk I/O stats

# Memory
free -h                        # Memory overview
vmstat 1                       # System health
slabtop                        # Kernel memory

# Network
ss -tuln                       # Listening ports
ss -tap                        # Active connections
ping -c 10 <host>              # Latency check
mtr <host>                     # Traceroute + ping
```

### sysctl Tuning
```ini
# /etc/sysctl.d/99-network-performance.conf
# Increase network buffer sizes
net.core.rmem_max = 134217728
net.core.wmem_max = 134217728
net.ipv4.tcp_rmem = 4096 87380 134217728
net.ipv4.tcp_wmem = 4096 65536 134217728

# Enable TCP BBR congestion control
net.core.default_qdisc = fq
net.ipv4.tcp_congestion_control = bbr

# Increase connection tracking
net.netfilter.nf_conntrack_max = 262144
```

## Filesystem

### Common Operations
```bash
# Find large files
find / -type f -size +100M -exec ls -lh {} \; 2>/dev/null

# Check filesystem type
df -T /var

# Mount with options
sudo mount -o noatime,nodiratime,compress=zstd /dev/sda1 /mnt/data

# Permissions audit
find /etc -perm /6000 -type f 2>/dev/null  # Find suid/sgid files
find / -writable -type f -path /proc -prune -o -print 2>/dev/null
```

## Anti-Patterns
- ❌ Piping `curl | sudo bash` without inspection
- ❌ `chmod 777` — never. Find the correct permission set
- ❌ Running everything as root — use dedicated service users
- ❌ No backup before system-wide changes
- ❌ Hardcoded paths or IPs in scripts
- ❌ Ignoring `set -e` in bash scripts
