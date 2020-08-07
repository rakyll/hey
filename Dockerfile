FROM alpine:3.11

LABEL maintainer="rakyll@rakyll.org"
LABEL maintainer="Vinícius Niche Correa viniciusnichecorrea@gmail.com"

RUN apk update && \
    rm -rf /var/cache/apk/*

RUN	addgroup hey \
    && adduser -S hey -u 1000 -G hey

USER hey

COPY --chown=hey:hey hey /usr/local/bin/

ENTRYPOINT [ "/usr/local/bin/hey" ]
