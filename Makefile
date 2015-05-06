VERSION=0.1.0
GOLDFLAGS="-X main.version $(VERSION)"

default:

bin:
	mkdir -p bin
	rm -rf bin/*

release: release-windows release-darwin release-linux

release-windows: bin
	GOOS=windows GOARCH=amd64 go build -ldflags=$(GOLDFLAGS) -o bin/ego ./cmd/ego
	cd bin && tar -cvzf ego$(VERSION).windows-amd64.tgz ego
	rm bin/ego

release-darwin: bin
	GOOS=darwin GOARCH=amd64 go build -ldflags=$(GOLDFLAGS) -o bin/ego ./cmd/ego
	cd bin && tar -cvzf ego$(VERSION).darwin-amd64.tgz ego
	rm bin/ego

release-linux: bin
	GOOS=linux GOARCH=amd64 go build -ldflags=$(GOLDFLAGS) -o bin/ego ./cmd/ego
	cd bin && tar -cvzf ego$(VERSION).linux-amd64.tgz ego
	rm bin/ego

.PHONY: bin default release
