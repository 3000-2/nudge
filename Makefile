VERSION ?= dev

.PHONY: build swift install clean test

build: swift
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o build/nudge .

swift:
	./swift/build.sh build

install: build
	sudo cp build/nudge /usr/local/bin/nudge
	sudo mkdir -p /usr/local/lib/nudge
	sudo cp -R build/Nudge.app /usr/local/lib/nudge/Nudge.app
	sudo xattr -cr /usr/local/lib/nudge/Nudge.app
	@echo "✓ nudge installed to /usr/local/bin/nudge"
	@echo "✓ Nudge.app installed to /usr/local/lib/nudge/"

clean:
	rm -rf build/

test:
	go test ./...
