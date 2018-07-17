FROM golang:1.9.4

RUN go get golang.org/x/net/context
RUN go get google.golang.org/grpc && go get firebase.google.com/go && go get google.golang.org/api/option

RUN go get github.com/Multy-io/Multy-back
RUN go get github.com/satori/go.uuid

RUN cd $GOPATH/src/github.com/Multy-io && \ 
    rm -rf Multy-back && \ 
    git clone https://github.com/Multy-io/Multy-back.git && \ 
    cd Multy-back && \
    git checkout release_1.2 && \
    git pull origin release_1.2


RUN cd $GOPATH/src/github.com/Multy-io/Multy-back && \ 
    make all-with-deps 
# make all-docker 

WORKDIR /go/src/github.com/Multy-io/Multy-back/cmd

RUN echo "VERSION 02"

ENTRYPOINT $GOPATH/src/github.com/Multy-io/Multy-back/cmd/multy
