FROM golang:1.9.4

RUN mkdir $GOPATH/src/github.com && \
    mkdir $GOPATH/src/github.com/Multy-io 

RUN cd $GOPATH/src/github.com/Multy-io && \ 
    git clone https://github.com/Multy-io/Multy-back.git && \ 
    cd Multy-back && \ 
    git checkout release_1.3

RUN go get -u github.com/golang/protobuf/proto && \
    cd $GOPATH/src/github.com/golang/protobuf && \
    make all

RUN apt-get update && \
    apt-get install -y protobuf-compiler

RUN cd $GOPATH/src/github.com/Multy-io && \
    git clone https://github.com/Multy-io/Multy-BTC-node-service.git && \
    cd $GOPATH/src/github.com/Multy-io/Multy-BTC-node-service && \
    git checkout reconnect


RUN cd $GOPATH/src/github.com/Multy-io/Multy-BTC-node-service && \
    go get ./... && \
    make build && \
    rm -r $GOPATH/src/github.com/Multy-io/Multy-back 


WORKDIR $GOPATH/src/github.com/Multy-io/Multy-BTC-node-service/cmd

RUN echo "VERSION 02"

ENTRYPOINT $GOPATH/src/github.com/Multy-io/Multy-BTC-node-service/cmd/client
