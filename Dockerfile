FROM golang:1.9.4

RUN go get github.com/Appscrunch/Multy-back && \
    rm -rf $GOPATH/src/github.com/Appscrunch/Multy-back && \
    cd $GOPATH/src/github.com/Appscrunch && \ 
    git clone https://github.com/Appscrunch/Multy-back.git 

RUN cd $GOPATH/src/github.com/Appscrunch/Multy-back && \
    git checkout eth && \
    make all-with-deps 

RUN cd $GOPATH/src/github.com/Appscrunch && \
    git clone https://github.com/Appscrunch/Multy-ETH-node-service.git 

RUN go get github.com/ethereum/go-ethereum/rpc 
RUN go get github.com/onrik/ethrpc 

RUN cd $GOPATH/src/github.com/Appscrunch/Multy-ETH-node-service && \
    make all-with-deps
    
WORKDIR /go/src/github.com/Appscrunch/Multy-ETH-node-service/cmd

RUN echo "VERSION 02"

ENTRYPOINT $GOPATH/src/github.com/Appscrunch/Multy-ETH-node-service/cmd/client
