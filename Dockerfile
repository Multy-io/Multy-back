FROM golang:1.9.4

# RUN go get golang.org/x/net/context
# RUN go get google.golang.org/grpc && go get firebase.google.com/go && go get google.golang.org/api/option

# RUN go get github.com/Multy-io/Multy-back
# RUN go get github.com/satori/go.uuid
RUN mkdir $GOPATH/src/github.com && \
    mkdir $GOPATH/src/github.com/Multy-io 
# mkdir $GOPATH/src/github.com/Multy-io/Multy-back

RUN cd $GOPATH/src/github.com/Multy-io && \
    git clone https://github.com/Multy-io/Multy-BTC-node-service.git && \
    cd $GOPATH/src/github.com/Multy-io/Multy-BTC-node-service && \
    git checkout reconnect

RUN cd $GOPATH/src/github.com/Multy-io && \
    git clone https://github.com/Multy-io/Multy-ETH-node-service.git && \
    cd $GOPATH/src/github.com/Multy-io/Multy-ETH-node-service && \
    git checkout reconnect

RUN cd $GOPATH/src/github.com/Multy-io && \ 
    git clone https://github.com/Multy-io/Multy-back.git && \ 
    cd Multy-back && \
    git checkout release_1.3


RUN cd $GOPATH/src/github.com/Multy-io/Multy-back && \ 
    go get ./... && \  
    make build && \ 
    rm -r $GOPATH/src/github.com/Multy-io/Multy-ETH-node-service && \ 
    rm -r $GOPATH/src/github.com/Multy-io//Multy-BTC-node-service


# make all-docker 

WORKDIR /go/src/github.com/Multy-io/Multy-back/cmd

RUN echo "VERSION 02"

ENTRYPOINT $GOPATH/src/github.com/Multy-io/Multy-back/cmd/multy
