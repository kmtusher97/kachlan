.PHONY: build build-cli build-gui dev test lint clean

build: build-cli build-gui

build-cli:
	$(MAKE) -C cli build

build-gui:
	$(MAKE) -C gui build

dev:
	$(MAKE) -C gui dev

test:
	cd media && go test ./...
	$(MAKE) -C cli test
	$(MAKE) -C gui test

lint:
	$(MAKE) -C cli lint

clean:
	$(MAKE) -C cli clean
	$(MAKE) -C gui clean
