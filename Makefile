VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null)

platform-all:
	make platform-macos platform-windows platform-linux

platform-macos:
	env GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o bin/goserve-macos-64/goserve

platform-windows:
	env GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o bin/goserve-windows-64/goserve.exe

platform-linux:
	env GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o bin/goserve-linux-64/goserve