FROM golang:1.7-alpine

RUN apk update \
  && apk add git \
  && rm -rf /var/cache/apk/* \
  && go get -u github.com/rakyll/hey \
  && apk del git

ENTRYPOINT ["/go/bin/hey"]
