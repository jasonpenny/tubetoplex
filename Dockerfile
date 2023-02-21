FROM golang:1.19-alpine3.17 AS gobuild

COPY . /app

RUN cd /app \
      && apk add --update build-base git \
      && make

FROM alpine:3.17

RUN apk update \
      && apk add --no-cache ca-certificates youtube-dl cifs-utils py3-pip ffmpeg \
      && pip3 install --upgrade pip youtube-dl

WORKDIR /app

COPY --from=gobuild /app/bin/* /app/
COPY run.sh /app/run.sh

ENTRYPOINT ["./run.sh"]
CMD ["/app/tubetoplex"]
