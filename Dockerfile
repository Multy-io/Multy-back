FROM golang:1.9.4

WORKDIR $GOPATH/src/github.com/Multy-io/Multy-back
COPY . $GOPATH/src/github.com/Multy-io/Multy-back

RUN go get ./... && make build

RUN echo "VERSION 03"

WORKDIR $GOPATH/src/github.com/Multy-io/Multy-back/cmd/multy-back/multy-back
ENTRYPOINT $GOPATH/src/github.com/Multy-io/Multy-back/cmd/multy-back/multy-back
