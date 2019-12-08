targets/merge-directories: sources/main.go
	CGO_ENABLED=0 go build -o targets/merge-directories ./sources

build: targets/merge-directories

TARGETS=darwin linux windows
$(TARGETS):
	GOOS=$@ GOARCH=amd64 CGO_ENABLED=0 go build -o "targets/merge-directories_$$(git describe --tags)_$@_amd64" ./sources
	zip -j targets/merge-directories_$$(git describe --tags)_$@_amd64.zip targets/merge-directories_$$(git describe --tags)_$@_amd64

targets: $(TARGETS)
