FROM golang:1.11-alpine3.10 AS gobuild

COPY . /app

RUN cd /app \
      && apk add --update build-base git \
      && make

FROM alpine:3.10

RUN apk update \
      && apk add --no-cache ca-certificates youtube-dl cifs-utils \
      && pip3 install --upgrade youtube-dl

WORKDIR /app

COPY --from=gobuild /app/bin/* /app/
COPY run.sh /app/run.sh

ENTRYPOINT ["./run.sh"]
CMD ["/app/tubetoplex"]
