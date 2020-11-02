###############################################################################
##  Purpose:   This Dockerfile contains the Hey executable on Alpine linux
###############################################################################
## build stage
FROM golang:1.9.2-alpine3.7 AS build-env

# Set the file maintainer (your name - the file's author)
MAINTAINER Chris Page <phriscage@gmail.com>

# app working directory
WORKDIR /app 

# Install dependencies
RUN apk --no-cache add --virtual git && \
        rm -rf /var/cache/apk/*

# Add the project
COPY requester /app/requester
COPY *.go /app/

# Install the Go dependencies
RUN go get -v -d

# Build the app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/hey .

ENTRYPOINT ["/app/hey"]


## final stage
FROM alpine:3.7
WORKDIR /app
COPY --from=build-env /app/hey /app/
ENTRYPOINT ["/app/hey"]
