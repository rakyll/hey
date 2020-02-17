binary = hey-consul

release: release-darwin release-linux
	;

release-darwin:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/$(binary)_darwin_amd64

release-linux:
	GOOS=linux GOARCH=amd64 go build -o ./bin/$(binary)_linux_amd64

push:
	gsutil cp bin/* gs://$(binary)-release
