VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null)

all: build-production package checksums

build-production:
	rm -rf build
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.currentVersion=$(VERSION) -s -w" -o build/serve-macos-64/bin/serve ./cmd/serve
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.currentVersion=$(VERSION) -s -w" -o build/serve-windows-64/bin/serve.exe ./cmd/serve
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.currentVersion=$(VERSION) -s -w" -o build/serve-linux-64/bin/serve ./cmd/serve

package:
	rm -rf dist
	mkdir dist
	tar -czf dist/serve-$(VERSION)-macos-64.tar.gz LICENSE -C build/serve-macos-64 .
	tar -czf dist/serve-$(VERSION)-windows-64.tar.gz LICENSE -C build/serve-windows-64 .
	tar -czf dist/serve-$(VERSION)-linux-64.tar.gz LICENSE -C build/serve-linux-64 .

checksums:
	cd dist && shasum -a 256 * > serve-$(VERSION)-checksums.txt