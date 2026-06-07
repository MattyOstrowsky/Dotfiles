STOW_DIR := $(shell pwd)/arch
# Auto-detect stow packages from arch/, excluding .git and pi/ (contains secrets)
PACKAGES := $(shell ls -d arch/*/ 2>/dev/null | sed 's|arch/||g' | sed 's/\///' | grep -v '\.git' | grep -v '^pi$$' | grep -v '^scripts$$' | grep -v '^bin$$')

.PHONY: install uninstall reinstall list dry-run install-deps

install:
	@echo "=== Stow packages: $(PACKAGES) ==="
	@for pkg in $(PACKAGES); do \
		echo "  → $$pkg"; \
		stow -v -d $(STOW_DIR) --target=$(HOME) $$pkg; \
	done
	@echo "Done."

uninstall:
	@echo "=== Unstow packages: $(PACKAGES) ==="
	@for pkg in $(PACKAGES); do \
		echo "  → $$pkg"; \
		stow -v -D -d $(STOW_DIR) --target=$(HOME) $$pkg; \
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
		stow -n -v -d $(STOW_DIR) --target=$(HOME) $$pkg 2>&1 | grep -E "LINK|UNLINK|existing" || true; \
	done

install-deps:
	@echo "=== Interactive Dependency Installer ==="
	@./arch/install-deps.sh
