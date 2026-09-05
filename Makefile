# Dotfiles installer entrypoint.
#
#   make install    → build the Go dashboard and run it
#   make preflight  → refresh metadata and print a read-only plan
#   make validate   → verify binaries and config symlinks
#   make yes        → non-interactive install of all applicable items
#   make clean      → remove the built binary
#
# All install/config logic lives in installer/ (Go + Bubble Tea TUI).

.PHONY: build install list preflight validate yes clean theme

build:
	@command -v go >/dev/null 2>&1 || { echo "Go not found. Install a Go toolchain first."; exit 1; }
	cd installer && CGO_ENABLED=0 go build -o dotfiles-install .

install: build
	./installer/dotfiles-install

list: build
	./installer/dotfiles-install --list

preflight: build
	./installer/dotfiles-install --preflight

validate: build
	./installer/dotfiles-install --validate

yes: build
	./installer/dotfiles-install --yes

clean:
	rm -f installer/dotfiles-install

THEME ?= catppuccin
theme: build
	./installer/dotfiles-install --theme $(THEME)