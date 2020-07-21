FROM golang:1.14.6-buster@sha256:54ef9b67477c0457d87a4ae519e7a1ba67aa34e7b09d1bfc80e1538ce7c6c4d7 AS builder

WORKDIR /go/src/github.com/rakyll/hey
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go install -ldflags '-w -extldflags "-static"' \
  /go/src/github.com/rakyll/hey

FROM gcr.io/distroless/base-debian10:nonroot@sha256:78f2372169e8d9c028da3856bce864749f2bb4bbe39c69c8960a6e40498f8a88

COPY --from=builder /go/bin/hey /hey

ENV PORT 8080
EXPOSE $PORT

ENTRYPOINT ["/hey"]
