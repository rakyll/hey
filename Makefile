.PHONY: build clean test test-race

VERSION=0.1
BIN=target/hey

GO_ENV=CGO_ENABLED=1
GO_FLAGS=-ldflags="-X main.version=$(VERSION) -X 'main.buildTime=`date`' -extldflags -static"
GO=env $(GO_ENV) go

build: hey.go
	@$(GO) build $(GO_FLAGS) -o $(BIN) $<

docker_image: clean
	@docker build -f ./Dockerfile -t hey:$(VERSION) .

install: build
	@cp $(BIN) /usr/local/bin

test:
	@$(GO) test .

test-race:
	@$(GO) test -race .

clean:
	@$(GO) clean ./...
	@rm -f $(BIN)
