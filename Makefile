BUILD_TAGS ?=

PACKAGES = $(shell go list ./... | grep -v '/vendor/')

binary = hey

release:
	GOOS=windows GOARCH=amd64 go build -o ./bin/$(binary)_windows_amd64
	GOOS=linux GOARCH=amd64 go build -o ./bin/$(binary)_linux_amd64
	GOOS=darwin GOARCH=amd64 go build -o ./bin/$(binary)_darwin_amd64

push:
	gsutil cp bin/* gs://$(binary)-release

.PHONY: test
test:
	go test -race -covermode=atomic -v -tags="$(BUILD_TAGS)" $(PACKAGES)

.PHONY: build
build:
	CGO_ENABLED=0 gox -osarch="linux/amd64" -tags="$(BUILD_TAGS)" -output hey

.PHONY: dist
dist:
	mkdir -p dist
	CGO_ENABLED=0 gox -osarch="linux/amd64" -osarch="darwin/amd64" -osarch="windows/amd64" -tags="$(BUILD_TAGS)" -output "dist/{{.Dir}}_{{.OS}}_{{.Arch}}"