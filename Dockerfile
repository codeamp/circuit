FROM quay.io/coreos/dex:v2.10.0
FROM golang:1.14.9-alpine

COPY --from=0 /usr/local/bin/dex /usr/local/bin/dex

ENV APP_PATH /go/src/github.com/codeamp/circuit

RUN apk -U add alpine-sdk git gcc openssh docker tini
RUN mkdir -p $APP_PATH

RUN go get -u github.com/cespare/reflex
RUN go get -u github.com/jteeuwen/go-bindata/...
RUN mkdir -p assets/

WORKDIR $APP_PATH
COPY . $APP_PATH

RUN cd plugins/codeamp/graphql; /go/bin/go-bindata -pkg assets -o $APP_PATH/assets/assets.go schema.graphql
RUN go build -i -v -o /go/bin/codeamp-circuit .

# Tini is now available at /sbin/tini
ENTRYPOINT ["/sbin/tini", "--"]
