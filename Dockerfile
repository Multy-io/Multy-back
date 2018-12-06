FROM multyio/multy-back-builder

RUN cd $GOPATH/src/github.com/Multy-io && \
    git clone https://github.com/Multy-io/Multy-BTC-node-service.git && \
    cd $GOPATH/src/github.com/Multy-io/Multy-BTC-node-service && \
    git checkout release_1.4 

RUN cd $GOPATH/src/github.com/Multy-io && \
    git clone https://github.com/Multy-io/Multy-ETH-node-service.git && \
    cd $GOPATH/src/github.com/Multy-io/Multy-ETH-node-service && \
    git checkout release_1.4 

WORKDIR $GOPATH/src/github.com/Multy-io/Multy-back
COPY . $GOPATH/src/github.com/Multy-io/Multy-back
RUN go get -v ./...
RUN make build && \ 
    rm -r $GOPATH/src/github.com/Multy-io/Multy-ETH-node-service && \ 
    rm -r $GOPATH/src/github.com/Multy-io/Multy-BTC-node-service

WORKDIR /go/src/github.com/Multy-io/Multy-back/cmd
RUN echo "VERSION 02"
ENTRYPOINT $GOPATH/src/github.com/Multy-io/Multy-back/cmd/multy