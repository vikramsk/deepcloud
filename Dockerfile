FROM golang:1.8-alpine

RUN mkdir -p /go/src/github.com/vikramsk/deepcloud

ADD . /go/src/github.com/vikramsk/deepcloud/

RUN go install github.com/vikramsk/deepcloud/cmd/controller

ENTRYPOINT /go/bin/controller

EXPOSE 8000
