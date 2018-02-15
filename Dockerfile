FROM golang:1.9.2


RUN go get github.com/Appscrunch/Multy-back && \
    cd $GOPATH/src/github.com/Appscrunch/Multy-back && \
    git pull && \
    make all-with-deps && \
    make build

WORKDIR /go/src/github.com/Appscrunch/Multy-back/cmd

RUN echo "VERSION 11"

ENTRYPOINT $GOPATH/src/github.com/Appscrunch/Multy-back/cmd/multy