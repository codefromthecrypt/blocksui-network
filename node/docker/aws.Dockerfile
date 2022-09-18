FROM curlimages/curl:latest

WORKDIR /tmp
RUN curl -L https://dist.ipfs.tech/kubo/v0.14.0/kubo_v0.14.0_linux-amd64.tar.gz | tar xvz

FROM golang:1.18.4
COPY --from=0 /tmp/kubo /tmp/kobo

RUN /tmp/kobo/install.sh
RUN ipfs --version

RUN mkdir /tmp/cache
ENV GOMODCACHE=/tmp/cache

WORKDIR /go/src
ADD config/ config/
ADD contracts/ contracts/
ADD ipfs/ ipfs/
ADD server/ server/
ADD account.go .
ADD go.mod .
ADD go.sum .
ADD main.go .
RUN go build -o /usr/bin/bui

ADD ecs.sh .

CMD ["./ecs.sh"]
