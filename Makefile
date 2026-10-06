.PHONY: build install list test clean

build:
	cd installer && CGO_ENABLED=0 go build -trimpath -o dotfiles-install .

install: build
	./installer/dotfiles-install

list: build
	./installer/dotfiles-install --list

test: build
	cd installer && go test -race ./...
	cd installer && go vet ./...
	python3 installer/tests/cli_smoke.py
	python3 installer/tests/shell_smoke.py

clean:
	rm -f installer/dotfiles-install
