# syntax=docker/dockerfile:1

FROM golang:1.19-alpine

ADD . /go/src/shoulder

WORKDIR /go/src/shoulder

COPY go.mod ./
COPY go.sum ./

RUN go mod download

RUN go build -o /ping

CMD ["/ping"]