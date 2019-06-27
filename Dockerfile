FROM golang:1.10.4-alpine

ENV APP_PATH /go/src/github.com/codeamp/circuit

RUN apk -U add alpine-sdk git gcc openssh docker
RUN mkdir -p $APP_PATH

WORKDIR $APP_PATH

# Install build deps
RUN go get -u github.com/cespare/reflex
RUN go get -u github.com/jteeuwen/go-bindata/...

# Static Assets
RUN mkdir -p /tmp/assets/
COPY plugins/codeamp/graphql/schema.graphql /tmp/assets/schema.graphql
COPY plugins/codeamp/graphql/static /tmp/assets/static
RUN /go/bin/go-bindata -pkg /tmp/assets -o /tmp/assets/assets.go /tmp/assets/schema.graphql

# Dex
#RUN mkdir -p /etc/dex/bootstrap/dex
#COPY bootstrap/dex /etc/dex/bootstrap/dex

# Compiled files
COPY . $APP_PATH
RUN mv /tmp/assets $APP_PATH/assets
RUN go build -o /go/bin/codeamp-circuit .
