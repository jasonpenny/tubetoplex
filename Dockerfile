FROM golang:1.11 AS gobuild

COPY . /app

RUN cd /app && make

FROM alpine

RUN apk update \
    && apk add --no-cache ca-certificates youtube-dl cifs-utils

WORKDIR /app

# put "glibc" dependency where go binaries expect it
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY --from=gobuild /app/bin/* /app/
COPY run.sh /app/run.sh

ENTRYPOINT ["./run.sh"]
CMD ["/app/tubetoplex"]
