FROM quay.io/coreos/dex:v2.10.0
FROM golang:1.10.4-alpine

COPY --from=0 /usr/local/bin/dex /usr/local/bin/dex

ENV APP_PATH /go/src/github.com/codeamp/circuit

RUN apk -U add alpine-sdk git gcc openssh docker
RUN mkdir -p $APP_PATH

WORKDIR $APP_PATH
COPY . $APP_PATH

RUN go get -u github.com/cespare/reflex
RUN go get -u github.com/jteeuwen/go-bindata/...
RUN mkdir -p assets/
RUN /go/bin/go-bindata -pkg assets -o assets/assets.go plugins/codeamp/graphql/schema.graphql
RUN go build -i -v -o /go/bin/codeamp-circuit .
