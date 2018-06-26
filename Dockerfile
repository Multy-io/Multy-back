FROM golang:1.9.4

RUN go get golang.org/x/net/context
RUN go get google.golang.org/grpc

RUN go get github.com/Multy-io/Multy-back


RUN cd $GOPATH/src/github.com/Multy-io && \ 
    rm -rf Multy-back && \ 
    git clone https://github.com/Multy-io/Multy-back.git && \ 
    cd Multy-back
# git pull origin master
# git checkout  && \ 

RUN cd $GOPATH/src/github.com/Multy-io/Multy-back && \ 
    make all-with-deps && \ 
    make all-docker 

WORKDIR /go/src/github.com/Multy-io/Multy-back/cmd

RUN echo "VERSION 02"

ENTRYPOINT $GOPATH/src/github.com/Multy-io/Multy-back/cmd/multy
