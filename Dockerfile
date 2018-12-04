FROM multyio/multy-back-builder

RUN cd $GOPATH/src/github.com/Multy-io && \
    git clone https://github.com/Multy-io/Multy-BTC-node-service.git && \
    cd $GOPATH/src/github.com/Multy-io/Multy-BTC-node-service && \
    git checkout master

RUN cd $GOPATH/src/github.com/Multy-io && \
    git clone https://github.com/Multy-io/Multy-ETH-node-service.git && \
    cd $GOPATH/src/github.com/Multy-io/Multy-ETH-node-service && \
    git checkout master

WORKDIR $GOPATH/src/github.com/Multy-io/Multy-back
COPY . $GOPATH/src/github.com/Multy-io/Multy-back
RUN go get -v ./...
RUN make build && \ 
    rm -r $GOPATH/src/github.com/Multy-io/Multy-ETH-node-service && \ 
    rm -r $GOPATH/src/github.com/Multy-io/Multy-BTC-node-service

RUN rm -rf $GOPATH/src/github.com/Multy-io

WORKDIR /go/src/github.com/Multy-io/Multy-back/cmd
RUN echo "VERSION 02"
ENTRYPOINT $GOPATH/src/github.com/Multy-io/Multy-back/cmd/multy
