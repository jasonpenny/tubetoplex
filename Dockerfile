FROM golang:1.22-alpine3.18 AS gobuild

COPY . /app

RUN cd /app \
      && apk add --update build-base git \
      && make

FROM alpine:3.18

RUN apk update \
      && apk add --no-cache ca-certificates cifs-utils py3-pip ffmpeg \
      && pip3 install --upgrade pip \
      && python3 -m pip install -U yt-dlp

WORKDIR /app

COPY --from=gobuild /app/bin/* /app/
COPY run.sh /app/run.sh

ENTRYPOINT ["./run.sh"]
CMD ["/app/tubetoplex"]
