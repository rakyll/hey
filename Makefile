.PHONY: build release

build:
	@goreleaser build --skip-validate --rm-dist

release:
	@goreleaser release --rm-dist
