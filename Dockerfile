FROM multyio/multy-back-builder

RUN cd $GOPATH/src/github.com/Multy-io && \
    git clone https://github.com/Multy-io/Multy-BTC-node-service.git && \
    cd $GOPATH/src/github.com/Multy-io/Multy-BTC-node-service && \
    git checkout master

RUN cd $GOPATH/src/github.com/Multy-io && \
    git clone https://github.com/Multy-io/Multy-ETH-node-service.git && \
    cd $GOPATH/src/github.com/Multy-io/Multy-ETH-node-service && \
    git checkout master

<<<<<<< HEAD
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
=======
WORKDIR $GOPATH/src/github.com/Multy-io/Multy-back
COPY . $GOPATH/src/github.com/Multy-io/Multy-back
RUN go get -v ./...
RUN make build && \ 
    rm -r $GOPATH/src/github.com/Multy-io/Multy-ETH-node-service && \ 
    rm -r $GOPATH/src/github.com/Multy-io/Multy-BTC-node-service

RUN rm -rf $GOPATH/src/github.com/Multy-io
>>>>>>> release_1.3

WORKDIR /go/src/github.com/Multy-io/Multy-back/cmd
RUN echo "VERSION 02"
<<<<<<< HEAD

ENTRYPOINT $GOPATH/src/github.com/Multy-io/Multy-back/cmd/multy
=======
ENTRYPOINT $GOPATH/src/github.com/Multy-io/Multy-back/cmd/multy
>>>>>>> release_1.3
