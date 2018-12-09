FROM golang:1.10 AS gobuild

RUN go get -u github.com/golang/dep/cmd/dep && \
    mkdir -p /go/src/github.com/jasonpenny/tubetoplex

COPY . /go/src/github.com/jasonpenny/tubetoplex

RUN cd /go/src/github.com/jasonpenny/tubetoplex && \
    dep ensure && \
    make


FROM alpine

RUN apk update \
    && apk add --no-cache ca-certificates youtube-dl cifs-utils

WORKDIR /app

# put "glibc" dependency where go binaries expect it
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY --from=gobuild /go/src/github.com/jasonpenny/tubetoplex/bin/* /app/
COPY run.sh /app/run.sh

ENTRYPOINT ["./run.sh"]
CMD ["/app/tubetoplex"]
