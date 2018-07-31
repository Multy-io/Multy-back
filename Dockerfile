FROM golang:1.9.4

RUN go get firebase.google.com/go   && \ 
    go get firebase.google.com/go/messaging  && \ 
    go get google.golang.org/api/option  && \ 
    go get github.com/satori/go.uuid && \
    go get github.com/ethereum/go-ethereum/rpc && \
    go get -u github.com/golang/protobuf/proto

RUN mkdir $GOPATH/src/github.com/Multy-io && \
    cd $GOPATH/src/github.com/Multy-io && \ 
    git clone https://github.com/Multy-io/Multy-back.git && \ 
    cd $GOPATH/src/github.com/Multy-io/Multy-back && \ 
    git checkout release_1.1.1 && \  
    git pull origin release_1.1.1

RUN cd $GOPATH/src/github.com/golang/protobuf && \
    make all

RUN apt-get update && \
    apt-get install -y protobuf-compiler

RUN cd $GOPATH/src/github.com/Multy-io && \
    git clone https://github.com/Multy-io/Multy-ETH-node-service.git && \
    cd $GOPATH/src/github.com/Multy-io/Multy-ETH-node-service/ && \
    git checkout release_1.1.1-prod && \
    git pull origin release_1.1.1-prod && \
    make build && \
    rm -r $GOPATH/src/github.com/Multy-io/Multy-back 

WORKDIR $GOPATH/src/github.com/Multy-io/Multy-ETH-node-service/cmd

RUN echo "VERSION 02"

ENTRYPOINT $GOPATH/src/github.com/Multy-io/Multy-ETH-node-service/cmd/client
