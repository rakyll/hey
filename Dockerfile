FROM golang:1.12.1 AS builder
LABEL maintainer="Ming Cheng"

# Build blade
COPY . /usr/src/hey
RUN cd /usr/src/hey && \
  go mod download && \
  GO111MODULE=on GO15VENDOREXPERIMENT=1 CGO_ENABLED=0 go build -a -ldflags="-extldflags -static" . 

# # # Stage2
FROM alpine:3.9.2
COPY --from=builder /usr/src/hey/hey /usr/bin/hey
CMD ["/usr/bin/hey"]
