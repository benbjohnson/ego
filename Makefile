VERSION=0.1.0
GOLDFLAGS="-X main.version $(VERSION)"

default:

release: release-windows release-darwin release-linux

release-windows:
	mkdir -p bin/windows/amd64
	GOOS=windows GOARCH=amd64 go build -ldflags=$(GOLDFLAGS) -o bin/windows/amd64/ego ./cmd/ego

release-darwin:
	mkdir -p bin/darwin/amd64
	GOOS=darwin GOARCH=amd64 go build -ldflags=$(GOLDFLAGS) -o bin/darwin/amd64/ego ./cmd/ego

release-linux:
	mkdir -p bin/linux/amd64
	GOOS=linux GOARCH=amd64 go build -ldflags=$(GOLDFLAGS) -o bin/linux/amd64/ego ./cmd/ego

.PHONY: default release
