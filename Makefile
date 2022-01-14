.PHONY: build test release

build:
	@goreleaser build --skip-validate --rm-dist

test:
	go test ./...

release:
	@goreleaser release --rm-dist
