FROM golang:1.9.4

RUN go get golang.org/x/net/context
RUN go get google.golang.org/grpc && go get firebase.google.com/go && go get google.golang.org/api/option

RUN go get github.com/Multy-io/Multy-back
RUN go get github.com/satori/go.uuid

RUN go get -u github.com/golang/protobuf/proto && \
    cd $GOPATH/src/github.com/golang/protobuf && \
    make all

RUN apt-get update && \
    apt-get install -y protobuf-compiler

RUN cd $GOPATH/src/github.com/Multy-io && \ 
    rm -r Multy-back && \
    git clone https://github.com/Multy-io/Multy-back.git && \ 
    go get github.com/swaggo/gin-swagger && \
    cd $GOPATH/src/github.com/Multy-io/Multy-back && \
    git checkout dev && \
    git pull origin dev

RUN cd $GOPATH/src/github.com/Multy-io/Multy-back && \ 
    make build

WORKDIR /go/src/github.com/Multy-io/Multy-back/cmd

RUN echo "VERSION 02"

ENTRYPOINT $GOPATH/src/github.com/Multy-io/Multy-back/cmd/multy