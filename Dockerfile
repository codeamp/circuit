FROM golang:alpine
ENV APP_PATH /go/src/github.com/codeamp/circuit
RUN apk -U add alpine-sdk git gcc

RUN mkdir -p $APP_PATH
WORKDIR $APP_PATH

COPY . $APP_PATH
RUN go get -u github.com/jteeuwen/go-bindata/...
RUN mkdir -p assets/
RUN /go/bin/go-bindata -pkg assets -o assets/assets.go plugins/codeamp/schema/schema.graphql
RUN go build -i -o /go/bin/codeamp-circuit .

COPY --from=0 /go/bin/codeamp-circuit /usr/local/bin/codeamp-circuit
RUN apk --no-cache add docker git

