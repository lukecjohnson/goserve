VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null)

platform-all:
	rm -rf build
	make platform-macos platform-windows platform-linux

platform-macos:
	env GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o build/goserve-macos-64/bin/goserve

platform-windows:
	env GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o build/goserve-windows-64/bin/goserve.exe

platform-linux:
	env GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o build/goserve-linux-64/bin/goserve

package-all:
	rm -rf dist
	mkdir dist
	make platform-all
	make package-macos package-windows package-linux

package-macos:
	tar -czf dist/goserve-$(VERSION)-macos-64.tar.gz LICENSE -C build/goserve-macos-64 .

package-windows:
	tar -czf dist/goserve-$(VERSION)-windows-64.tar.gz LICENSE -C build/goserve-windows-64 .

package-linux:
	tar -czf dist/goserve-$(VERSION)-linux-64.tar.gz LICENSE -C build/goserve-linux-64 .
