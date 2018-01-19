FROM golang:alpine
ENV APP_PATH /go/src/github.com/codeamp/circuit
RUN apk -U add alpine-sdk git gcc

RUN mkdir -p $APP_PATH
WORKDIR $APP_PATH

COPY . $APP_PATH
RUN go get -u github.com/jteeuwen/go-bindata/...
RUN mkdir -p $APP_PATH/assets/
RUN /go/bin/go-bindata -pkg assets -o $APP_PATH/assets/assets.go $APP_PATH/plugins/codeamp/schema/schema.graphql
RUN go build -i -o /go/bin/codeamp-circuit .


FROM alpine:3.4
ENV APP_PATH /go/src/github.com/codeamp/circuit
RUN apk --no-cache add docker git
COPY --from=0 /go/bin/codeamp-circuit /usr/local/bin/codeamp-circuit

