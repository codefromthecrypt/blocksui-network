FROM curlimages/curl:latest

WORKDIR /tmp
RUN curl -L https://github.com/cortesi/modd/releases/download/v0.8/modd-0.8-linux64.tgz | tar xvz
RUN curl -L https://dist.ipfs.tech/kubo/v0.14.0/kubo_v0.14.0_linux-amd64.tar.gz | tar xvz

FROM golang:1.18.4
ARG ACCESS_TOKEN
RUN echo "machine github.com login api password ${ACCESS_TOKEN}" > ~/.netrc

COPY --from=0 /tmp/modd-0.8-linux64/modd /usr/local/bin
COPY --from=0 /tmp/kubo /tmp/kobo

RUN /tmp/kobo/install.sh
RUN ipfs init

RUN go env -w GOPRIVATE=github.com/crcls/*

RUN mkdir /tmp/cache
ENV GOMODCACHE=/tmp/cache

WORKDIR /go/src

CMD ["./start.sh", "modd.conf"]
