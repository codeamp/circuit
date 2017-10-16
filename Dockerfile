FROM finalduty/archlinux:monthly

ENV APP_PATH /go/src/github.com/codeamp/circuit
ENV GOPATH /go
ENV PATH ${PATH}:/go/bin

RUN mkdir -p $APP_PATH
WORKDIR $APP_PATH

RUN pacman -Syu --noconfirm
RUN pacman -Sy --noconfirm libgit2 git gcc nodejs go go-tools npm base-devel
RUN go get github.com/cespare/reflex
RUN go get github.com/sirupsen/logrus

COPY . $APP_PATH
RUN go build -i -o /go/bin/codeamp-circuit .

CMD ["reflex", "-c", "reflex.conf"]
