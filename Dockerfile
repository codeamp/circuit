FROM quay.io/coreos/dex:v2.8.0
FROM golang:alpine

ENV APP_PATH /go/src/github.com/codeamp/circuit

RUN apk -U add alpine-sdk git gcc
RUN go get github.com/cespare/reflex

RUN mkdir -p $APP_PATH
WORKDIR $APP_PATH

COPY . $APP_PATH
RUN go build -i -o /go/bin/codeamp-circuit .

FROM alpine:3.4
COPY --from=0 /usr/local/bin/dex /usr/local/bin/dex
COPY --from=1 /go/bin/codeamp-circuit /usr/local/bin/codeamp-circuit
