STOW_DIR := $(shell pwd)
# Auto-detect stow packages, excluding .git and pi/ (contains secrets)
PACKAGES := $(shell ls -d */ | sed 's/\///' | grep -v '.git' | grep -v '^pi$$')

.PHONY: install uninstall reinstall list dry-run install-deps

install:
	@echo "=== Stow packages: $(PACKAGES) ==="
	@for pkg in $(PACKAGES); do \
		echo "  → $$pkg"; \
		stow -v --target=$(HOME) $$pkg; \
	done
	@echo "Done."

uninstall:
	@echo "=== Unstow packages: $(PACKAGES) ==="
	@for pkg in $(PACKAGES); do \
		echo "  → $$pkg"; \
		stow -v -D --target=$(HOME) $$pkg; \
	done
	@echo "Done."

reinstall: uninstall install

list:
	@echo "Available packages:"
	@for pkg in $(PACKAGES); do \
		echo "  - $$pkg"; \
	done

dry-run:
	@echo "=== Dry run ==="
	@for pkg in $(PACKAGES); do \
		echo "  → $$pkg"; \
		stow -n -v --target=$(HOME) $$pkg 2>&1 | grep -E "LINK|UNLINK|existing" || true; \
	done

install-deps:
	@echo "=== System dependencies ==="
	@echo "Run: sudo apt install -y stow git curl wget make fish fzf bat ripgrep fd-find tree htop direnv btop python3 python3-pip python3-venv"
	@echo ""
	@echo "See dependencies.md for per-package details and binary installs."
