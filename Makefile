VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null)

all: build-production package checksums

build-production:
	rm -rf build
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.currentVersion=$(VERSION) -s -w" -o build/goserve-macos-64/bin/goserve
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.currentVersion=$(VERSION) -s -w" -o build/goserve-windows-64/bin/goserve.exe
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.currentVersion=$(VERSION) -s -w" -o build/goserve-linux-64/bin/goserve

package:
	rm -rf dist
	mkdir dist
	tar -czf dist/goserve-$(VERSION)-macos-64.tar.gz LICENSE -C build/goserve-macos-64 .
	tar -czf dist/goserve-$(VERSION)-windows-64.tar.gz LICENSE -C build/goserve-windows-64 .
	tar -czf dist/goserve-$(VERSION)-linux-64.tar.gz LICENSE -C build/goserve-linux-64 .

checksums:
	cd dist && shasum -a 256 * > goserve-$(VERSION)-checksums.txt