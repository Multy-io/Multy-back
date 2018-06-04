FROM golang:1.9.4

RUN go get github.com/Appscrunch/Multy-back && \
    cd $GOPATH/src/github.com/Appscrunch && \ 
    rm -rf Multy-back && \ 
    git clone https://github.com/Appscrunch/Multy-back.git && \ 
    cd Multy-back && \ 
    git pull origin release_1.0-mempool && \ 
    git checkout release_1.0-mempool

RUN go get github.com/Appscrunch/Multy-BTC-node-service && \
    cd $GOPATH/src/github.com/Appscrunch && \
    rm -rf Multy-BTC-node-service && \
    git clone https://github.com/Appscrunch/Multy-BTC-node-service.git  && \
    cd $GOPATH/src/github.com/Appscrunch/Multy-BTC-node-service && \
    git checkout state  && \
    git pull && \

    make all-with-deps

WORKDIR /go/src/github.com/Appscrunch/Multy-BTC-node-service/cmd

RUN echo "VERSION 02"

ENTRYPOINT $GOPATH/src/github.com/Appscrunch/Multy-BTC-node-service/cmd/client