FROM golang:1.9.4

RUN go get github.com/Appscrunch/Multy-back && \
    rm -rf $GOPATH/src/github.com/Appscrunch/Multy-back && \
    cd $GOPATH/src/github.com/Appscrunch && \ 
    git clone https://github.com/Appscrunch/Multy-back.git && \ 
    cd Multy-back && \ 
    git checkout release_1.1 && \ 
    go get firebase.google.com/go   && \ 
    go get firebase.google.com/go/messaging  && \ 
    go get google.golang.org/api/option  && \ 
    go get github.com/satori/go.uuid

RUN cd $GOPATH/src/github.com/Appscrunch && \
    git clone https://github.com/Appscrunch/Multy-ETH-node-service.git && \
    cd $GOPATH/src/github.com/Appscrunch/Multy-ETH-node-service && \
    go get github.com/ethereum/go-ethereum/rpc
# go get ./...


RUN cd $GOPATH/src/github.com/Appscrunch/Multy-ETH-node-service && \
    make all-with-deps && \
    rm -r $GOPATH/src/github.com/Appscrunch/Multy-back 


WORKDIR $GOPATH/src/github.com/Appscrunch/Multy-ETH-node-service/cmd

RUN echo "VERSION 02"

ENTRYPOINT $GOPATH/src/github.com/Appscrunch/Multy-ETH-node-service/cmd/client
