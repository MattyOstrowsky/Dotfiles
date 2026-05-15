STOW_DIR := $(shell pwd)
PACKAGES := $(shell ls -d */ | sed 's/\///' | grep -v '.git')

.PHONY: install uninstall reinstall list dry-run

install:
	@echo "Installing packages: $(PACKAGES)"
	@for pkg in $(PACKAGES); do \
		echo "  → stow $$pkg"; \
		stow -v --target=$(HOME) $$pkg; \
	done
	@echo "Done."

uninstall:
	@echo "Uninstalling packages: $(PACKAGES)"
	@for pkg in $(PACKAGES); do \
		echo "  → unstow $$pkg"; \
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
	@echo "Dry run (no changes):"
	@for pkg in $(PACKAGES); do \
		echo "  → $$pkg"; \
		stow -n -v --target=$(HOME) $$pkg 2>&1 | grep -E "LINK|UNLINK|existing" || true; \
	done
